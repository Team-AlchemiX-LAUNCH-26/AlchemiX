package intelligence

import (
	"github.com/launch26/relic-ring-protocol/internal/agent"
)

// UncertaintyDetector validates live observations and flags anomalies.
type UncertaintyDetector struct{}

// NewUncertaintyDetector creates a new detector.
func NewUncertaintyDetector() *UncertaintyDetector {
	return &UncertaintyDetector{}
}

// Evaluate returns an uncertainty score in [0, 1] for a link observation.
func (u *UncertaintyDetector) Evaluate(obs agent.LinkObservation) float64 {
	score := 0.0
	checks := 0.0

	// Missing self-reported latency.
	if obs.SelfReportedLatencyMS == nil {
		score += 0.3
	}
	checks++

	// Unknown status.
	if obs.Status != "ok" && obs.Status != "saturated" {
		score += 0.25
	}
	checks++

	// Impossible latency: self-reported below physical minimum.
	if obs.SelfReportedLatencyMS != nil && obs.PhysicalLatencyMS > 0 {
		if *obs.SelfReportedLatencyMS < obs.PhysicalLatencyMS*0.5 {
			score += 0.2
		}
	}
	checks++

	// Invalid load ratio range.
	if obs.LoadRatio < 0 || obs.LoadRatio > 1.5 {
		score += 0.2
	}
	checks++

	// Zero capacity (contradictory).
	if obs.CapacityUnits <= 0 {
		score += 0.15
	}
	checks++

	// Negative current load.
	if obs.CurrentLoad < 0 {
		score += 0.15
	}
	checks++

	// Stale tick (tick == 0 is suspicious in live).
	if obs.Tick <= 0 {
		score += 0.1
	}
	checks++

	// Contradictory: load ratio doesn't match load/capacity.
	if obs.CapacityUnits > 0 {
		expectedRatio := obs.CurrentLoad / obs.CapacityUnits
		diff := obs.LoadRatio - expectedRatio
		if diff > 0.1 || diff < -0.1 {
			score += 0.1
		}
	}
	checks++

	// Normalize to [0, 1].
	if score > 1 {
		score = 1
	}
	return score
}
