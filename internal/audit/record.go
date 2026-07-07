// Package audit provides decision auditing for the agent.
package audit

import "time"

// Record captures a single routing decision for auditing.
type Record struct {
	Timestamp                time.Time `json:"timestamp"`
	Tick                     int64     `json:"tick"`
	CurrentPlanet            string    `json:"current_planet"`
	LinkID                   string    `json:"link_id"`
	Action                   string    `json:"action"`
	PhysicalLatencyMS        float64   `json:"physical_latency_ms"`
	PredictedCongestionMS    float64   `json:"predicted_congestion_penalty_ms"`
	TrustScore               float64   `json:"trust_score"`
	TargetingRiskScore       float64   `json:"targeting_risk_score"`
	UncertaintyScore         float64   `json:"uncertainty_score"`
	CombinedCost             float64   `json:"combined_cost"`
	Reasons                  []string  `json:"reasons"`
	AlternativePath          []string  `json:"alternative_path,omitempty"`
	ChosenNextPlanet         string    `json:"chosen_next_planet,omitempty"`
	CandidatesEvaluated      int       `json:"candidates_evaluated"`
	Confidence               float64   `json:"confidence"`
}

// NewRecord creates a record from the agent's hop decision data.
func NewRecord(
	tick int64,
	currentPlanet, linkID, action string,
	physicalMS, congestionMS, trust, risk, uncertainty, cost float64,
	reasons []string,
	nextPlanet string,
	altPath []string,
	candidateCount int,
	confidence float64,
) Record {
	return Record{
		Timestamp:             time.Now().UTC(),
		Tick:                  tick,
		CurrentPlanet:         currentPlanet,
		LinkID:                linkID,
		Action:                action,
		PhysicalLatencyMS:     physicalMS,
		PredictedCongestionMS: congestionMS,
		TrustScore:            trust,
		TargetingRiskScore:    risk,
		UncertaintyScore:      uncertainty,
		CombinedCost:          cost,
		Reasons:               reasons,
		ChosenNextPlanet:      nextPlanet,
		AlternativePath:       altPath,
		CandidatesEvaluated:   candidateCount,
		Confidence:            confidence,
	}
}
