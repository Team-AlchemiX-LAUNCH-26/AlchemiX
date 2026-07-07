package liveapi

import (
	"github.com/launch26/relic-ring-protocol/internal/agent"
)

// Validator checks the integrity of live state data.
type Validator struct {
	saturationThreshold float64
}

// NewValidator creates a validator with the given saturation threshold.
func NewValidator(saturationThreshold float64) *Validator {
	return &Validator{saturationThreshold: saturationThreshold}
}

// Validate processes raw state, removes unusable links, and returns cleaned state.
func (v *Validator) Validate(state agent.NetworkState) agent.NetworkState {
	validated := agent.NetworkState{
		Tick:          state.Tick,
		Links:         make(map[string]agent.LinkObservation, len(state.Links)),
		ValidLinks:    make(map[string]bool, len(state.Links)),
		Uncertainties: make(map[string]float64, len(state.Uncertainties)),
	}

	for id, obs := range state.Links {
		validated.Links[id] = obs

		valid := true

		// Hard rule: saturated links are unavailable.
		if obs.LoadRatio >= v.saturationThreshold {
			valid = false
		}
		if obs.Status == "saturated" {
			valid = false
		}

		// Links with unknown status are suspect.
		if obs.Status != "ok" && obs.Status != "saturated" {
			valid = false
		}

		validated.ValidLinks[id] = valid
		validated.Uncertainties[id] = state.Uncertainties[id]
	}

	return validated
}
