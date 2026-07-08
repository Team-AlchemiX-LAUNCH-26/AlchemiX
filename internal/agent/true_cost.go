package agent

// TrueCostCalculator computes the True Cost for each link.
type TrueCostCalculator struct {
	config AgentConfig
}

// NewTrueCostCalculator creates a calculator with the given config.
func NewTrueCostCalculator(config AgentConfig) *TrueCostCalculator {
	return &TrueCostCalculator{config: config}
}

// Calculate computes True Cost for a link given all model outputs.
func (tc *TrueCostCalculator) Calculate(
	physicalLatencyMS float64,
	congestionPenaltyMS float64,
	trustScore float64,
	targetingRiskScore float64,
	uncertaintyScore float64,
	isSwitching bool,
) LinkEvaluation {
	// Trust penalty: higher cost when trust is low.
	trustPenalty := tc.config.TrustWeightMS * (1.0 - trustScore)

	// Targeting risk penalty.
	riskPenalty := tc.config.TargetingWeightMS * targetingRiskScore

	// Uncertainty penalty.
	uncertaintyPenalty := tc.config.UncertaintyWeightMS * uncertaintyScore

	// Switching penalty (cost of changing from current route).
	switchingPenalty := 0.0
	if isSwitching {
		switchingPenalty = tc.config.SwitchingWeightMS
	}

	trueCost := physicalLatencyMS +
		congestionPenaltyMS +
		trustPenalty +
		riskPenalty +
		uncertaintyPenalty +
		switchingPenalty

	return LinkEvaluation{
		PhysicalLatencyMS:            physicalLatencyMS,
		PredictedCongestionPenaltyMS: congestionPenaltyMS,
		TrustScore:                   trustScore,
		TargetingRiskScore:           targetingRiskScore,
		UncertaintyScore:             uncertaintyScore,
		CombinedCost:                 trueCost,
	}
}

// EvaluateLink performs the full evaluation pipeline for a single link.
func (tc *TrueCostCalculator) EvaluateLink(
	linkID string,
	obs LinkObservation,
	congestion CongestionPrediction,
	trust ScoreResult,
	targeting ScoreResult,
	uncertainty float64,
	isSwitching bool,
) LinkEvaluation {
	eval := tc.Calculate(
		obs.PhysicalLatencyMS,
		congestion.PenaltyMS,
		trust.Score,
		targeting.Score,
		uncertainty,
		isSwitching,
	)
	eval.LinkID = linkID
	return eval
}
