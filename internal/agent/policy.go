package agent

import "sort"

// PolicyEngine implements the continue/reroute/queue decision logic.
type PolicyEngine struct {
	config AgentConfig
}

// NewPolicyEngine creates a policy engine with the given thresholds.
func NewPolicyEngine(config AgentConfig) *PolicyEngine {
	return &PolicyEngine{config: config}
}

// ScoredHop represents a next-hop candidate with its evaluation.
type ScoredHop struct {
	NextPlanet string
	LinkID     string
	Evaluation LinkEvaluation
	Reasons    []string
}

// SelectNextHop picks the safest next hop from scored candidates.
// Returns the decision action and the chosen hop.
func (p *PolicyEngine) SelectNextHop(candidates []ScoredHop) (DecisionAction, *ScoredHop) {
	if len(candidates) == 0 {
		return ActionQueue, nil
	}

	// Filter out hops that violate hard thresholds.
	var viable []ScoredHop
	for _, c := range candidates {
		if c.Evaluation.TrustScore < p.config.TrustHardThreshold {
			continue
		}
		if c.Evaluation.TargetingRiskScore > p.config.RiskHardThreshold {
			continue
		}
		viable = append(viable, c)
	}

	if len(viable) == 0 {
		// All candidates violate thresholds — reroute or queue.
		// Return the least-bad option with REROUTE action.
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Evaluation.CombinedCost < candidates[j].Evaluation.CombinedCost
		})
		best := candidates[0]
		best.Reasons = append(best.Reasons, "all candidates below safety thresholds")
		return ActionReroute, &best
	}

	// Sort by combined cost ascending.
	sort.Slice(viable, func(i, j int) bool {
		return viable[i].Evaluation.CombinedCost < viable[j].Evaluation.CombinedCost
	})

	best := viable[0]

	// Check if a significantly safer alternative exists.
	if len(viable) > 1 {
		secondBest := viable[1]
		costDiff := best.Evaluation.CombinedCost - secondBest.Evaluation.CombinedCost
		if costDiff > 0 && costDiff/best.Evaluation.CombinedCost > 0.3 {
			best.Reasons = append(best.Reasons, "significantly safer alternative considered")
		}
	}

	return ActionContinue, &best
}

// ShouldQueue determines if the agent should queue rather than transmit.
func (p *PolicyEngine) ShouldQueue(state NetworkState, currentPlanet, destination string) bool {
	// Queue if no valid links exist from the current planet.
	hasValidLink := false
	for _, obs := range state.Links {
		if !state.ValidLinks[obs.LinkID] {
			continue
		}
		if obs.PlanetA == currentPlanet || obs.PlanetB == currentPlanet {
			hasValidLink = true
			break
		}
	}
	return !hasValidLink
}
