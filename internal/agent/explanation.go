package agent

import (
	"fmt"
	"strings"
)

// ExplainDecision generates a human-readable explanation of a routing decision.
func ExplainDecision(report DecisionReport) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Route from %s to %s: %s\n",
		report.OriginID, report.DestinationID,
		strings.Join(report.ChosenPath, " → ")))

	sb.WriteString(fmt.Sprintf("Estimated latency: %.1f ms\n\n", report.FinalLatencyEstimate))

	for _, eval := range report.LinkEvaluations {
		sb.WriteString(fmt.Sprintf("Link %s:\n", eval.LinkID))
		sb.WriteString(fmt.Sprintf("  Physical:   %.1f ms\n", eval.PhysicalLatencyMS))
		sb.WriteString(fmt.Sprintf("  Congestion: +%.1f ms\n", eval.PredictedCongestionPenaltyMS))
		sb.WriteString(fmt.Sprintf("  Trust:      %.2f\n", eval.TrustScore))
		sb.WriteString(fmt.Sprintf("  Risk:       %.2f\n", eval.TargetingRiskScore))
		sb.WriteString(fmt.Sprintf("  True Cost:  %.1f ms\n\n", eval.CombinedCost))
	}

	if len(report.HopDecisions) > 0 {
		sb.WriteString("Hop decisions:\n")
		for _, hop := range report.HopDecisions {
			sb.WriteString(fmt.Sprintf("  %s → %s [%s]: %s\n",
				hop.CurrentPlanet, hop.NextPlanet,
				hop.Action, strings.Join(hop.Reasons, "; ")))
		}
	}

	return sb.String()
}

// BuildPublicReport creates the standardized competition-facing JSON output.
func BuildPublicReport(
	originID, destinationID string,
	chosenPath []string,
	evaluations []LinkEvaluation,
	totalLatency float64,
	hopDecisions []HopDecision,
) DecisionReport {
	explanation := summarizeDecision(chosenPath, evaluations)

	return DecisionReport{
		OriginID:             originID,
		DestinationID:        destinationID,
		ChosenPath:           chosenPath,
		LinkEvaluations:      evaluations,
		FinalLatencyEstimate: totalLatency,
		Explanation:          explanation,
		HopDecisions:         hopDecisions,
	}
}

func summarizeDecision(path []string, evals []LinkEvaluation) string {
	if len(path) < 2 {
		return "Direct delivery, no routing needed."
	}

	// Find the best and worst link.
	var bestLink, worstLink string
	bestCost := 1e18
	worstCost := 0.0
	for _, e := range evals {
		if e.CombinedCost < bestCost {
			bestCost = e.CombinedCost
			bestLink = e.LinkID
		}
		if e.CombinedCost > worstCost {
			worstCost = e.CombinedCost
			worstLink = e.LinkID
		}
	}

	parts := []string{}
	parts = append(parts, fmt.Sprintf("Selected %s route", strings.Join(path[1:len(path)-1], "-")))

	if bestLink != "" {
		parts = append(parts, fmt.Sprintf("strongest link: %s (cost %.1f)", bestLink, bestCost))
	}
	if worstLink != "" && worstLink != bestLink {
		parts = append(parts, fmt.Sprintf("weakest link: %s (cost %.1f)", worstLink, worstCost))
	}

	return strings.Join(parts, "; ")
}

// CanonicalLinkID ensures consistent link naming with alphabetical ordering.
func CanonicalLinkID(a, b string) string {
	if a < b {
		return a + "-" + b
	}
	return b + "-" + a
}
