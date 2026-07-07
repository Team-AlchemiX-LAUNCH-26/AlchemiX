// Package liveapi provides a client for the Chimera live state API.
package liveapi

import "github.com/launch26/relic-ring-protocol/internal/agent"

// StateResponse represents the JSON response from GET /state.
type StateResponse struct {
	Tick  int64      `json:"tick"`
	Links []LinkData `json:"links"`
}

// LinkData represents a single link in the live API response.
type LinkData struct {
	LinkID                string   `json:"link_id"`
	PlanetA               string   `json:"planet_a"`
	PlanetB               string   `json:"planet_b"`
	CapacityUnits         float64  `json:"capacity_units"`
	CurrentLoad           float64  `json:"current_load"`
	LoadRatio             float64  `json:"load_ratio"`
	SelfReportedLatencyMS *float64 `json:"self_reported_latency_ms"`
	TrafficShare          float64  `json:"traffic_share"`
	Status                string   `json:"status"`
}

// ToObservation converts API link data to an agent.LinkObservation.
func (ld LinkData) ToObservation(tick int64, physicalLatencyMS float64) agent.LinkObservation {
	return agent.LinkObservation{
		Tick:                  tick,
		LinkID:                ld.LinkID,
		PlanetA:               ld.PlanetA,
		PlanetB:               ld.PlanetB,
		CapacityUnits:         ld.CapacityUnits,
		CurrentLoad:           ld.CurrentLoad,
		LoadRatio:             ld.LoadRatio,
		SelfReportedLatencyMS: ld.SelfReportedLatencyMS,
		TrafficShare:          ld.TrafficShare,
		Status:                ld.Status,
		PhysicalLatencyMS:     physicalLatencyMS,
	}
}
