package intelligence

import (
	"math"

	"github.com/launch26/relic-ring-protocol/internal/agent"
)

// TargetingScorer implements agent.TargetingModel using HistGradientBoosting JSON trees.
type TargetingScorer struct {
	model   *ModelFile
	history *FeatureHistory
}

// NewTargetingScorer loads the targeting model from a JSON file.
func NewTargetingScorer(path string) (*TargetingScorer, error) {
	m, err := LoadModelFile(path)
	if err != nil {
		return nil, err
	}
	return &TargetingScorer{
		model:   m,
		history: NewFeatureHistory(),
	}, nil
}

// Score estimates the probability a link will be jammed.
func (t *TargetingScorer) Score(obs agent.LinkObservation, usage agent.UsageHistory) agent.ScoreResult {
	features := t.buildFeatures(obs)

	jamProb := 0.0
	if len(t.model.HGBRTrees) > 0 {
		rawScore := PredictHGBEnsemble(t.model.HGBRTrees, features, t.model.BasePrediction, t.model.LearningRate)
		
		// Platt Calibration (sigmoid)
		a := t.model.PlattA
		b := t.model.PlattB
		
		if a != 0 || b != 0 {
			jamProb = 1.0 / (1.0 + math.Exp(-(a*rawScore + b)))
		} else {
			jamProb = rawScore // Fallback if no calibration available
		}
	}

	confidence := 0.85
	reasons := []string{}
	if jamProb > 0.50 {
		reasons = append(reasons, "High structural targeting risk based on historical traffic patterns")
	}

	t.history.AddObservation(obs)

	return agent.ScoreResult{
		Score:      jamProb,
		Confidence: confidence,
		Reasons:    reasons,
	}
}

func (t *TargetingScorer) buildFeatures(obs agent.LinkObservation) map[string]float64 {
	linkEncoded := float64(t.model.LinkIDMap[obs.LinkID])

	share := obs.TrafficShare
	missingShare := 0.0
	// For parity, if share is missing, python code used median. But live it's always populated by state provider.

	extractShare := func(o agent.LinkObservation) float64 { return o.TrafficShare }
	extractJam := func(o agent.LinkObservation) float64 {
		if o.Status == "jammed" {
			return 1.0
		}
		return 0.0
	}

	prevShare := t.history.LastValue(obs.LinkID, extractShare, share)
	rollingShare := t.history.LaggedMean(obs.LinkID, 5, extractShare)
	if rollingShare == 0 && prevShare == share {
		rollingShare = share
	}
	shareChange := share - prevShare

	histJamRate := t.history.LaggedMean(obs.LinkID, 100, extractJam)
	rollingJamCount := t.history.LaggedSum(obs.LinkID, 10, extractJam)

	// consecutive_high_usage: how many of the last 10 ticks was traffic_share above median (approx).
	// We use 1.0 / num_links as the threshold.
	nLinks := len(t.model.LinkIDMap)
	if nLinks == 0 {
		nLinks = 1
	}
	medianShare := 1.0 / float64(nLinks)
	
	extractHighUsage := func(o agent.LinkObservation) float64 {
		if o.TrafficShare > medianShare {
			return 1.0
		}
		return 0.0
	}
	consecutiveHighUsage := t.history.LaggedSum(obs.LinkID, 10, extractHighUsage)

	return map[string]float64{
		"traffic_share":           share,
		"traffic_share_missing":   missingShare,
		"traffic_share_rank":      1.0, // rank is complex to compute live without full state; use constant fallback
		"prev_traffic_share":      prevShare,
		"traffic_share_change":    shareChange,
		"rolling_traffic_share":   rollingShare,
		"historical_jam_rate":     histJamRate,
		"rolling_jam_count":       rollingJamCount,
		"consecutive_high_usage":  consecutiveHighUsage,
		"link_id_encoded":         linkEncoded,
	}
}
