package agent

import (
	"context"
	"fmt"
)

// DecisionLoop implements the hop-by-hop sequential decision process (§9).
type DecisionLoop struct {
	config      AgentConfig
	congestion  CongestionModel
	trust       TrustModel
	targeting   TargetingModel
	stateProvider StateProvider
	costCalc    *TrueCostCalculator
	policy      *PolicyEngine
	usage       UsageHistory
}

// NewDecisionLoop creates a decision loop with all required components.
func NewDecisionLoop(
	config AgentConfig,
	congestion CongestionModel,
	trust TrustModel,
	targeting TargetingModel,
	stateProvider StateProvider,
) *DecisionLoop {
	return &DecisionLoop{
		config:        config,
		congestion:    congestion,
		trust:         trust,
		targeting:     targeting,
		stateProvider: stateProvider,
		costCalc:      NewTrueCostCalculator(config),
		policy:        NewPolicyEngine(config),
		usage: UsageHistory{
			RecentSelections: make(map[string]int),
			ConsecutiveUses:  make(map[string]int),
			LastSelectedTick: make(map[string]int64),
		},
	}
}

func (dl *DecisionLoop) EvaluateHop(
	ctx context.Context,
	currentPlanet string,
	destination string,
	candidates []CandidateRoute,
	visited map[string]bool,
) (DecisionAction, *ScoredHop, []LinkEvaluation, error) {
	state, err := dl.stateProvider.GetLatestState(ctx)
	if err != nil {
		return ActionQueue, nil, nil, fmt.Errorf("fetch live state: %w", err)
	}

	// Find all possible next hops from current planet.
	nextHops := make(map[string]bool)
	for _, route := range candidates {
		for i, planet := range route.Path {
			if planet == currentPlanet && i < len(route.Path)-1 {
				next := route.Path[i+1]
				if !visited[next] {
					nextHops[next] = true
				}
			}
		}
	}

	var scoredHops []ScoredHop
	var allEvals []LinkEvaluation

	for nextPlanet := range nextHops {
		linkID := CanonicalLinkID(currentPlanet, nextPlanet)

		obs, ok := state.Links[linkID]
		if !ok {
			continue
		}

		// Skip invalid links.
		if !state.ValidLinks[linkID] {
			continue
		}

		// Run all models.
		congPred := dl.congestion.Predict(obs)
		trustResult := dl.trust.Score(obs, congPred)
		targetResult := dl.targeting.Score(obs, dl.usage)
		uncertainty := state.Uncertainties[linkID]

		eval := dl.costCalc.EvaluateLink(
			linkID, obs, congPred, trustResult, targetResult, uncertainty, false,
		)

		allEvals = append(allEvals, eval)

		var reasons []string
		reasons = append(reasons, trustResult.Reasons...)
		reasons = append(reasons, targetResult.Reasons...)

		scoredHops = append(scoredHops, ScoredHop{
			NextPlanet: nextPlanet,
			LinkID:     linkID,
			Evaluation: eval,
			Reasons:    reasons,
		})
	}

	action, chosen := dl.policy.SelectNextHop(scoredHops)

	// Record usage if continuing.
	if chosen != nil && action == ActionContinue {
		dl.usage.RecentSelections[chosen.LinkID]++
		dl.usage.LastSelectedTick[chosen.LinkID] = state.Tick
	}

	return action, chosen, allEvals, nil
}

// RunFullRoute executes the complete hop-by-hop decision process.
func (dl *DecisionLoop) RunFullRoute(
	ctx context.Context,
	origin, destination string,
	candidates []CandidateRoute,
) ([]string, []LinkEvaluation, []HopDecision, error) {
	currentPlanet := origin
	var chosenPath []string
	var allEvaluations []LinkEvaluation
	var hopDecisions []HopDecision

	chosenPath = append(chosenPath, currentPlanet)

	visited := make(map[string]bool)
	visited[currentPlanet] = true

	maxHops := 20 // Safety limit.
	for i := 0; i < maxHops && currentPlanet != destination; i++ {
		action, chosen, evals, err := dl.EvaluateHop(ctx, currentPlanet, destination, candidates, visited)
		if err != nil {
			return chosenPath, allEvaluations, hopDecisions, err
		}

		allEvaluations = append(allEvaluations, evals...)

		if chosen == nil {
			hopDecisions = append(hopDecisions, HopDecision{
				CurrentPlanet: currentPlanet,
				Action:        ActionQueue,
				Reasons:       []string{"no viable next hop found"},
			})
			break
		}

		hopDecisions = append(hopDecisions, HopDecision{
			CurrentPlanet: currentPlanet,
			NextPlanet:    chosen.NextPlanet,
			LinkID:        chosen.LinkID,
			Action:        action,
			Evaluation:    chosen.Evaluation,
			Reasons:       chosen.Reasons,
		})

		if action == ActionQueue {
			break
		}

		chosenPath = append(chosenPath, chosen.NextPlanet)
		visited[chosen.NextPlanet] = true
		currentPlanet = chosen.NextPlanet
	}

	return chosenPath, allEvaluations, hopDecisions, nil
}
