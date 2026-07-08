package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

// Agent is the Phase 2 AI Copilot. It enriches a Phase 1 route with
// True Cost analysis by evaluating each link sequentially using three
// ML models served by the Python inference server.
//
// The Agent NEVER modifies the Phase 1 route. It only adds a
// RouteAssessment alongside the existing domain.Route.
type Agent struct {
	live      *LiveAPIClient
	inference *InferenceClient
}

// New creates a fully configured Agent.
// Call once at application startup (e.g. in NewApp or main.go).
func New() *Agent {
	return &Agent{
		live:      NewLiveAPIClient(),
		inference: NewInferenceClient(),
	}
}

// Ping checks whether the Python inference server is alive.
// Returns nil if ready, error otherwise.
func (a *Agent) Ping(ctx context.Context) error {
	return a.inference.Ping(ctx)
}

// Assess analyses a Phase 1 route and returns a RouteAssessment.
//
// Flow:
//  1. Extract the list of links from path (consecutive planet pairs).
//  2. Fetch live state for each link (with safe fallback on failure).
//  3. For each link — sequentially — call the inference server.
//  4. Apply the True Cost formula via cost.go.
//  5. Return the RouteAssessment (never modifies the Phase 1 route).
func (a *Agent) Assess(ctx context.Context, route domain.Route) RouteAssessment {
	path := route.Path
	if len(path) < 2 {
		// Single-planet route — no links to evaluate.
		physicsMS := route.Latency.TotalSeconds * 1000
		return FallbackAssessment(path, physicsMS)
	}

	// Build the ordered list of link IDs from the path.
	linkIDs := pathToLinkIDs(path)

	// Fetch live network state (with fallback).
	liveStates := a.live.FetchStateWithFallback(ctx, linkIDs)

	// Extract physics latency per void transit from the Phase 1 route.
	physicsPerLink := extractPhysicsLatency(route)

	// Score each link sequentially.
	scores := make([]LinkScore, 0, len(linkIDs))
	for i, linkID := range linkIDs {
		state, ok := liveStates[linkID]
		if !ok {
			state = safeDefaultState(linkID)
		}

		req := InferenceRequest{
			LinkID:                linkID,
			LoadUnits:             state.LoadUnits,
			LoadRatio:             state.LoadRatio,
			Status:                state.Status,
			SelfReportedLatencyMS: state.SelfReportedLatencyMS,
			MeasuredLatencyMS:     state.MeasuredLatencyMS,
			TrafficShare:          state.TrafficShare,
		}

		inf, err := a.inference.Predict(ctx, req)
		if err != nil {
			// Inference server unavailable for this link — use safe zeros
			// so the congestion/trust/targeting penalties default to 0.
			fmt.Fprintf(os.Stderr, "[ai/agent] WARNING: %v\n", err)
			inf = InferenceResponse{
				LinkID:              linkID,
				CongestionPenaltyMS: 0,
				TrustScore:          1.0,
				TrustPenaltyMS:      0,
				JamProbability:      0,
				TargetingPenaltyMS:  0,
			}
		}

		physicsMS := 0.0
		if i < len(physicsPerLink) {
			physicsMS = physicsPerLink[i]
		}

		scores = append(scores, ScoreLink(physicsMS, inf))
	}

	return ScoreRoute(path, scores)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// pathToLinkIDs converts ["Aegis","Boreas","Dawn"] →
// ["Aegis-Boreas","Boreas-Dawn"] using the canonical sorted link ID format.
func pathToLinkIDs(path []string) []string {
	ids := make([]string, 0, len(path)-1)
	for i := 0; i < len(path)-1; i++ {
		ids = append(ids, canonicalLinkID(path[i], path[i+1]))
	}
	return ids
}

// canonicalLinkID produces "A-B" with A < B alphabetically,
// matching the link_id format in the training datasets.
func canonicalLinkID(a, b string) string {
	if strings.Compare(a, b) <= 0 {
		return a + "-" + b
	}
	return b + "-" + a
}

// extractPhysicsLatency returns the void-transit latency in milliseconds
// for each hop, in path order. The Phase 1 domain.Route.VoidTransits slice
// has one entry per link (len(path)-1).
func extractPhysicsLatency(route domain.Route) []float64 {
	out := make([]float64, len(route.VoidTransits))
	for i, vt := range route.VoidTransits {
		out[i] = vt.TotalSeconds * 1000
	}
	return out
}
