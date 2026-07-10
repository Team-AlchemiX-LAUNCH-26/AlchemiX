package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// Agent is the Chimera-Resilient Analytical Co-Pilot.
type Agent struct {
	Config     AgentConfig
	Parser     *DeterministicParser
	Router     *RouterAdapter
	Candidates CandidateGenerator
	Congestion CongestionModel
	Trust      TrustModel
	Targeting  TargetingModel
	LiveState  StateProvider
	Loop       *DecisionLoop
	OnHop      func(context.Context, HopDecision)
}

// LoadConfig reads agent configuration from a JSON file.
func LoadConfig(path string) (AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig(), nil // Fall back to defaults.
	}
	var cfg AgentConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("parse agent config: %w", err)
	}
	return cfg, nil
}

// Execute processes a natural-language request and produces a decision report.
func (a *Agent) Execute(ctx context.Context, naturalLanguageRequest string) (DecisionReport, error) {
	// Step 1: Parse the request.
	parsed, err := a.Parser.Parse(naturalLanguageRequest)
	if err != nil {
		return DecisionReport{}, fmt.Errorf("parse request: %w", err)
	}

	// Step 2: Generate baseline route.
	baseline, err := a.Router.FindRoute(parsed.OriginID, parsed.DestinationID)
	if err != nil {
		return DecisionReport{}, fmt.Errorf("baseline route: %w", err)
	}

	// Step 3: Generate candidate routes.
	candidates, err := a.Candidates.TopKPaths(
		parsed.OriginID, parsed.DestinationID,
		a.Config.TopKRoutes, nil,
	)
	if err != nil {
		// Fall back to baseline only.
		candidates = []CandidateRoute{baseline}
	}

	// Step 4-10: Run the sequential decision loop.
	chosenPath, evaluations, hopDecisions, err := a.Loop.RunFullRoute(
		ctx, parsed.OriginID, parsed.DestinationID, candidates,
	)
	if err != nil {
		return DecisionReport{}, fmt.Errorf("decision loop: %w", err)
	}

	// Deduplicate evaluations by link ID.
	evalMap := make(map[string]LinkEvaluation)
	for _, e := range evaluations {
		if existing, ok := evalMap[e.LinkID]; !ok || e.CombinedCost < existing.CombinedCost {
			evalMap[e.LinkID] = e
		}
	}
	var dedupedEvals []LinkEvaluation
	for _, e := range evalMap {
		dedupedEvals = append(dedupedEvals, e)
	}

	totalLatency := latencyEstimateFromChosenHops(hopDecisions)

	// Build the report.
	report := BuildPublicReport(
		parsed.OriginID, parsed.DestinationID, parsed.Payload,
		chosenPath, dedupedEvals, totalLatency, hopDecisions,
	)

	return report, nil
}

func latencyEstimateFromChosenHops(hops []HopDecision) float64 {
	total := 0.0
	for _, hop := range hops {
		if hop.LinkID == "" || hop.NextPlanet == "" {
			continue
		}
		total += hop.Evaluation.PhysicalLatencyMS +
			hop.Evaluation.PredictedCongestionPenaltyMS
	}
	return total
}

// GetState returns a summary of the agent's current state.
func (a *Agent) GetState() map[string]interface{} {
	return map[string]interface{}{
		"config":  a.Config,
		"planets": a.Parser.validPlanets,
	}
}
