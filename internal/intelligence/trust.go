package intelligence

import (
	"github.com/launch26/relic-ring-protocol/internal/agent"
)

// TrustScorer implements agent.TrustModel using HistGradientBoosting JSON trees.
type TrustScorer struct {
	model   *ModelFile
	history *FeatureHistory
}

// NewTrustScorer loads the trust model from a JSON file.
func NewTrustScorer(path string) (*TrustScorer, error) {
	m, err := LoadModelFile(path)
	if err != nil {
		return nil, err
	}
	return &TrustScorer{
		model:   m,
		history: NewFeatureHistory(),
	}, nil
}

// Score estimates telemetry trust for a link.
func (t *TrustScorer) Score(obs agent.LinkObservation, congestion agent.CongestionPrediction) agent.ScoreResult {
	features := t.buildFeatures(obs)

	// Trust Regressor (under_report_ratio)
	ratio := 0.0
	if len(t.model.HGBRTrees) > 0 {
		ratio = PredictHGBEnsemble(t.model.HGBRTrees, features, t.model.BasePrediction, t.model.LearningRate)
		if ratio < 0 {
			ratio = 0
		} else if ratio > 1 {
			ratio = 1
		}
	}

	// Calculate trust score based on ratio and trust scale.
	scale := t.model.TrustScale
	if scale == 0 {
		scale = 0.50
	}
	trustScore := 1.0 - (ratio / scale)
	if trustScore < 0 {
		trustScore = 0
	}

	confidence := 0.90
	if obs.SelfReportedLatencyMS == nil {
		confidence = 0.40 // Lower confidence if missing telemetry
	}

	reasons := []string{}
	if ratio > t.model.SpoofThreshold {
		reasons = append(reasons, "High probability of latency spoofing detected")
	}

	// Add observation to history AFTER building features to maintain parity with shift(1).
	t.history.AddObservation(obs)

	return agent.ScoreResult{
		Score:      trustScore,
		Confidence: confidence,
		Reasons:    reasons,
	}
}

func (t *TrustScorer) buildFeatures(obs agent.LinkObservation) map[string]float64 {
	linkEncoded := float64(t.model.LinkIDMap[obs.LinkID])
	phys := t.model.PhysicalBaselinesMS[obs.LinkID]
	if phys == 0 {
		phys = 10.0
	}

	selfVal := 0.0
	missingSelf := 1.0
	if obs.SelfReportedLatencyMS != nil {
		selfVal = *obs.SelfReportedLatencyMS
		missingSelf = 0.0
	} else {
		// Use Python preprocessing median fallback if available in metadata.
		// Since we didn't inject the exact preprocessing median into ModelFile root,
		// we just use physical as a safe fallback or whatever we have.
		// For perfect parity, we should extract the preprocessing median. 
		// We'll approximate for live.
		selfVal = phys
	}

	selfToPhys := selfVal / phys
	devFromBase := selfVal - phys

	extractSelf := func(o agent.LinkObservation) float64 {
		if o.SelfReportedLatencyMS != nil {
			return *o.SelfReportedLatencyMS
		}
		return phys
	}

	prevSelf := t.history.LastValue(obs.LinkID, extractSelf, selfVal)
	rollingSelf := t.history.LaggedMean(obs.LinkID, 5, extractSelf)
	if rollingSelf == 0 && prevSelf == selfVal {
		rollingSelf = selfVal
	}
	selfChange := selfVal - prevSelf

	// historical_bias is the expanding mean of under_report_ratio from the past.
	extractBias := func(o agent.LinkObservation) float64 {
		if o.SelfReportedLatencyMS != nil && phys > 0 {
			// Measured is not available to the predictor live unless it's past?
			// Actually, "measured" is the target of the trust model! 
			// In live usage, we DO NOT have measured latency. That's why we use the predictor.
			// But the trust feature historical_bias relies on past under_report_ratio.
			// How do we get past under_report_ratio live? 
			// We can't! We must rely on our predicted ratio, or skip it.
			// The python model trained on true measured_latency_ms.
			// In live use, we use 0.0 (no known bias).
			return 0.0
		}
		return 0.0
	}
	histBias := t.history.LaggedMean(obs.LinkID, 100, extractBias) // Approximate expanding

	belowMin := 0.0
	if selfVal < phys*0.95 {
		belowMin = 1.0
	}

	return map[string]float64{
		"self_reported_latency_ms":      selfVal,
		"self_reported_latency_missing": missingSelf,
		"physical_latency_ms":           phys,
		"self_to_physical_ratio":        selfToPhys,
		"deviation_from_baseline":       devFromBase,
		"prev_self_reported":            prevSelf,
		"self_reported_change":          selfChange,
		"rolling_self_reported":         rollingSelf,
		"historical_bias":               histBias,
		"below_physical_min":            belowMin,
		"link_id_encoded":               linkEncoded,
	}
}
