package intelligence

import (
	"github.com/launch26/relic-ring-protocol/internal/agent"
)

// CongestionPredictor implements agent.CongestionModel using HistGradientBoosting JSON trees.
type CongestionPredictor struct {
	model   *ModelFile
	history *FeatureHistory
}

// NewCongestionPredictor loads the congestion model from a JSON file.
func NewCongestionPredictor(path string) (*CongestionPredictor, error) {
	m, err := LoadModelFile(path)
	if err != nil {
		return nil, err
	}
	return &CongestionPredictor{
		model:   m,
		history: NewFeatureHistory(),
	}, nil
}

// Predict returns congestion prediction for a link observation.
func (c *CongestionPredictor) Predict(obs agent.LinkObservation) agent.CongestionPrediction {
	features := c.buildFeatures(obs)

	// In v2, saturation classifier was removed in favor of a hard rule.
	// We handle saturation probabilistically based on load_ratio.
	satProb := 0.0
	if obs.LoadRatio >= 0.90 || obs.Status == "saturated" {
		satProb = 1.0
	} else if obs.LoadRatio > 0.80 {
		satProb = (obs.LoadRatio - 0.80) / 0.10 // Linear scale from 80% to 90%
	}

	// Congestion penalty (HGBR Ensemble)
	penalty := 0.0
	if len(c.model.HGBRTrees) > 0 {
		penalty = PredictHGBEnsemble(c.model.HGBRTrees, features, c.model.BasePrediction, c.model.LearningRate)
		if penalty < 0 {
			penalty = 0
		}
	}

	// Confidence based on how far from saturation.
	confidence := 1.0 - satProb*0.5
	if satProb == 1.0 {
		confidence = 0.95 // High confidence it is saturated.
	}

	// Important: Add observation to history AFTER building features (shift(1) parity).
	c.history.AddObservation(obs)

	return agent.CongestionPrediction{
		PenaltyMS:             penalty,
		SaturationProbability: satProb,
		Confidence:            confidence,
	}
}

func (c *CongestionPredictor) buildFeatures(obs agent.LinkObservation) map[string]float64 {
	linkEncoded := float64(c.model.LinkIDMap[obs.LinkID])
	capacity := c.model.Capacities[obs.LinkID]
	if capacity == 0 {
		capacity = 100 // Fallback.
	}

	// Extractors
	extractLoad := func(o agent.LinkObservation) float64 { return o.LoadRatio }

	prevLoad := c.history.LastValue(obs.LinkID, extractLoad, obs.LoadRatio)
	rollingMean := c.history.LaggedMean(obs.LinkID, 5, extractLoad)
	if rollingMean == 0 && prevLoad == obs.LoadRatio {
		rollingMean = obs.LoadRatio // Fallback for first tick
	}
	rollingStd := c.history.LaggedStd(obs.LinkID, 5, extractLoad)

	// Rate of increase = diff of lagged rolling mean.
	// Since we don't store rolling means historically in our cache, we approximate it 
	// by comparing rolling mean(window=5) to rolling mean(window=6 offset by 1).
	// To be perfectly accurate with Python diff(), we can just use (rollingMean - oldRollingMean).
	// For simplicity, we approximate: (currentLoad - prevLoad) / 1.0
	// But to match python exact parity: rate_of_load_increase = rolling_mean_load_ratio.diff()
	// To avoid adding another cache, we'll store the last rolling mean in features map... wait!
	// It's just easier to compute mean of [t-6, t-2] and subtract from [t-5, t-1].
	oldRollingMean := 0.0
	// We can compute oldRollingMean over ticks [t-6 to t-2] by shifting the extractor
	hList := c.history.GetHistoryCopy(obs.LinkID)
	if len(hList) >= 2 {
		sum := 0.0
		count := 0
		for i := len(hList) - 6; i < len(hList)-1; i++ {
			if i >= 0 {
				sum += hList[i].LoadRatio
				count++
			}
		}
		if count > 0 {
			oldRollingMean = sum / float64(count)
		}
	} else {
		oldRollingMean = rollingMean
	}
	rateOfIncrease := rollingMean - oldRollingMean

	distToSat := 0.90 - obs.LoadRatio
	if distToSat < 0 {
		distToSat = 0
	}

	missingUnits := 0.0
	if obs.CurrentLoad == 0 && obs.LoadRatio > 0 {
		// Just a heuristic for "was missing" since Go doesn't use NaNs easily
		// This should match how it's fed. In live, we usually have load_units.
		missingUnits = 0.0
	}

	return map[string]float64{
		"load_ratio":              obs.LoadRatio,
		"load_units":              obs.CurrentLoad,
		"capacity_units":          capacity,
		"link_id_encoded":         linkEncoded,
		"previous_load_ratio":     prevLoad,
		"load_ratio_change":       obs.LoadRatio - prevLoad,
		"rolling_mean_load_ratio": rollingMean,
		"rolling_std_load_ratio":  rollingStd,
		"rate_of_load_increase":   rateOfIncrease,
		"distance_to_saturation":  distToSat,
		"load_units_missing":      missingUnits,
	}
}
