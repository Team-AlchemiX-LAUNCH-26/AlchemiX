package ai

import (
	"fmt"
	"strings"
)

// Penalty weights — tunable constants used in the True Cost formula.
// All weights are in milliseconds so they are directly comparable to
// the Phase 1 physics latency measured by domain.LatencyBreakdown.
const (
	// TrustWeight: maximum extra latency added when trust_score = 0 (fully dishonest).
	// A link that wildly under-reports its latency could hide 50 s of real delay.
	TrustWeight float64 = 50_000

	// TargetingWeight: maximum extra latency added when jam_probability = 1.0.
	// A certain Chimera jam effectively makes the link unusable → 200 s penalty.
	TargetingWeight float64 = 200_000

	// SafetyThresholds for the qualitative safety label.
	safeThresholdMS   float64 = 100_000  // < 100 s AI adjustment → "safe"
	cautionThresholdMS float64 = 300_000 // 100–300 s → "caution";  > 300 s → "danger"
)

// ScoreLink computes the True Cost for a single link.
//
// Parameters
//   - physicsLatencyMS: Phase 1 void-transit total latency for this link (ms).
//     Derived from domain.VoidTransitBreakdown.TotalSeconds × 1000.
//   - inf: the InferenceResponse returned by the Python model server.
//
// The formula is evaluated sequentially — one link at a time — never
// across the whole route at once.
func ScoreLink(physicsLatencyMS float64, inf InferenceResponse) LinkScore {
	// Trust penalty: dishonest links get penalised proportionally.
	trustPenalty := (1.0 - clamp01(inf.TrustScore)) * TrustWeight

	// Targeting penalty: high jam probability makes the link expensive.
	targetingPenalty := clamp01(inf.JamProbability) * TargetingWeight

	trueCost := physicsLatencyMS + inf.CongestionPenaltyMS + trustPenalty + targetingPenalty

	return LinkScore{
		LinkID:             inf.LinkID,
		PhysicsLatencyMS:   physicsLatencyMS,
		CongestionPenaltyMS: inf.CongestionPenaltyMS,
		TrustScore:         inf.TrustScore,
		TrustPenaltyMS:     trustPenalty,
		JamProbability:     inf.JamProbability,
		TargetingPenaltyMS: targetingPenalty,
		TrueCostMS:         trueCost,
		Explanation:        buildLinkExplanation(inf, physicsLatencyMS, trustPenalty, targetingPenalty),
	}
}

// ScoreRoute aggregates per-link scores into a RouteAssessment.
// It also computes the qualitative safety label and human-readable summary.
func ScoreRoute(path []string, scores []LinkScore) RouteAssessment {
	var totalPhysics, totalAI, totalTrue float64
	for _, s := range scores {
		totalPhysics += s.PhysicsLatencyMS
		aiAdjust := s.CongestionPenaltyMS + s.TrustPenaltyMS + s.TargetingPenaltyMS
		totalAI += aiAdjust
		totalTrue += s.TrueCostMS
	}

	label := safetyLabel(totalAI)
	summary := buildRouteSummary(path, scores, totalPhysics, totalAI, label)

	return RouteAssessment{
		Path:                path,
		LinkScores:          scores,
		TotalPhysicsMS:      totalPhysics,
		TotalAIAdjustmentMS: totalAI,
		TotalTrueCostMS:     totalTrue,
		SafetyLabel:         label,
		Summary:             summary,
		AgentActive:         true,
	}
}

// FallbackAssessment returns a RouteAssessment that only uses physics latency.
// Used when the inference server is unavailable.
func FallbackAssessment(path []string, physicsLatencyMS float64) RouteAssessment {
	return RouteAssessment{
		Path:                path,
		LinkScores:          nil,
		TotalPhysicsMS:      physicsLatencyMS,
		TotalAIAdjustmentMS: 0,
		TotalTrueCostMS:     physicsLatencyMS,
		SafetyLabel:         "unknown",
		Summary:             "AI Agent unavailable — route based on physics latency only.",
		AgentActive:         false,
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func safetyLabel(totalAIAdjustmentMS float64) string {
	switch {
	case totalAIAdjustmentMS < safeThresholdMS:
		return "safe"
	case totalAIAdjustmentMS < cautionThresholdMS:
		return "caution"
	default:
		return "danger"
	}
}

func buildLinkExplanation(
	inf InferenceResponse,
	physicsMS, trustPenalty, targetingPenalty float64,
) string {
	parts := make([]string, 0, 4)

	if inf.CongestionPenaltyMS > 5_000 {
		parts = append(parts, fmt.Sprintf(
			"congestion adds %.1f s", inf.CongestionPenaltyMS/1000,
		))
	}
	if inf.TrustScore < 0.80 {
		parts = append(parts, fmt.Sprintf(
			"low trust (%.0f%%) adds %.1f s", inf.TrustScore*100, trustPenalty/1000,
		))
	}
	if inf.JamProbability > 0.20 {
		parts = append(parts, fmt.Sprintf(
			"%.0f%% jam risk adds %.1f s", inf.JamProbability*100, targetingPenalty/1000,
		))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("Link is healthy (physics %.1f s).", physicsMS/1000)
	}
	return fmt.Sprintf("Warnings: %s.", strings.Join(parts, "; "))
}

func buildRouteSummary(
	path []string,
	scores []LinkScore,
	totalPhysicsMS, totalAIMS float64,
	label string,
) string {
	// Find worst link by AI adjustment.
	var worstLink string
	var worstAI float64
	for _, s := range scores {
		adj := s.CongestionPenaltyMS + s.TrustPenaltyMS + s.TargetingPenaltyMS
		if adj > worstAI {
			worstAI = adj
			worstLink = s.LinkID
		}
	}

	route := strings.Join(path, " → ")
	base := fmt.Sprintf(
		"Route [%s]: physics %.1f s, AI adjustment +%.1f s → true cost %.1f s [%s].",
		route,
		totalPhysicsMS/1000,
		totalAIMS/1000,
		(totalPhysicsMS+totalAIMS)/1000,
		strings.ToUpper(label),
	)
	if worstLink != "" && worstAI > 5_000 {
		base += fmt.Sprintf(" Highest risk link: %s (+%.1f s).", worstLink, worstAI/1000)
	}
	return base
}
