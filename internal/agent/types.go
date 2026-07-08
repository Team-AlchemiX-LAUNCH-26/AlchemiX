// Package agent implements the Chimera-Resilient Analytical Co-Pilot Agent.
// This file defines shared types and interfaces used across all agent sub-packages.
package agent

import "context"

// ParsedTransmissionRequest is the structured output from natural-language parsing.
type ParsedTransmissionRequest struct {
	OriginID      string `json:"origin_id"`
	DestinationID string `json:"destination_id"`
	Payload       string `json:"payload"`
}

// RequestParser converts natural-language input into a structured request.
type RequestParser interface {
	Parse(raw string) (ParsedTransmissionRequest, error)
}

// CandidateRoute represents a single candidate path with physical latency.
type CandidateRoute struct {
	Path             []string `json:"path"`
	PhysicalLatencyMS float64  `json:"physical_latency_ms"`
}

// BaselineRouter wraps the Phase 1 physics-based router.
type BaselineRouter interface {
	FindRoute(originID, destinationID string) (CandidateRoute, error)
}

// CandidateGenerator produces K alternative routes.
type CandidateGenerator interface {
	TopKPaths(originID, destinationID string, k int, excludedLinks map[string]bool) ([]CandidateRoute, error)
}

// LinkObservation represents the live state of a single interplanetary link.
type LinkObservation struct {
	Tick                  int64    `json:"tick"`
	LinkID                string   `json:"link_id"`
	PlanetA               string   `json:"planet_a"`
	PlanetB               string   `json:"planet_b"`
	CapacityUnits         float64  `json:"capacity_units"`
	CurrentLoad           float64  `json:"current_load"`
	LoadRatio             float64  `json:"load_ratio"`
	SelfReportedLatencyMS *float64 `json:"self_reported_latency_ms"`
	TrafficShare          float64  `json:"traffic_share"`
	Status                string   `json:"status"`
	PhysicalLatencyMS     float64  `json:"physical_latency_ms"`
}

// CongestionPrediction is produced by the congestion model for a link.
type CongestionPrediction struct {
	PenaltyMS             float64 `json:"predicted_congestion_penalty_ms"`
	SaturationProbability float64 `json:"saturation_probability"`
	Confidence            float64 `json:"confidence"`
}

// CongestionModel predicts congestion for a link.
type CongestionModel interface {
	Predict(observation LinkObservation) CongestionPrediction
}

// ScoreResult is produced by trust and targeting models.
type ScoreResult struct {
	Score      float64  `json:"score"`
	Confidence float64  `json:"confidence"`
	Reasons    []string `json:"reasons"`
}

// TrustModel estimates telemetry trust for a link.
type TrustModel interface {
	Score(observation LinkObservation, congestion CongestionPrediction) ScoreResult
}

// UsageHistory tracks recent link selections for targeting-risk analysis.
type UsageHistory struct {
	RecentSelections map[string]int   `json:"recent_selections"`
	ConsecutiveUses  map[string]int   `json:"consecutive_uses"`
	LastSelectedTick map[string]int64 `json:"last_selected_tick"`
}

// TargetingModel estimates the probability a link will be jammed.
type TargetingModel interface {
	Score(observation LinkObservation, usage UsageHistory) ScoreResult
}

// StateProvider delivers the current live network state.
type StateProvider interface {
	GetLatestState(ctx context.Context) (NetworkState, error)
}

// NetworkState represents the validated live state of all links.
type NetworkState struct {
	Tick         int64                      `json:"tick"`
	Links        map[string]LinkObservation `json:"links"`
	ValidLinks   map[string]bool            `json:"valid_links"`
	Uncertainties map[string]float64        `json:"uncertainties"`
}

// LinkEvaluation is the per-link scoring output.
type LinkEvaluation struct {
	LinkID                       string  `json:"link_id"`
	PhysicalLatencyMS            float64 `json:"physical_latency_ms"`
	PredictedCongestionPenaltyMS float64 `json:"predicted_congestion_penalty_ms"`
	TrustScore                   float64 `json:"trust_score"`
	TargetingRiskScore           float64 `json:"targeting_risk_score"`
	UncertaintyScore             float64 `json:"uncertainty_score"`
	CombinedCost                 float64 `json:"combined_cost"`
}

// DecisionAction represents the agent's action for a hop.
type DecisionAction string

const (
	ActionContinue DecisionAction = "CONTINUE"
	ActionReroute  DecisionAction = "REROUTE"
	ActionQueue    DecisionAction = "QUEUE"
)

// HopDecision records the agent's decision for a single hop.
type HopDecision struct {
	Tick            int64            `json:"tick"`
	CurrentPlanet   string           `json:"current_planet"`
	NextPlanet      string           `json:"next_planet"`
	LinkID          string           `json:"link_id"`
	Action          DecisionAction   `json:"action"`
	Evaluation      LinkEvaluation   `json:"evaluation"`
	Reasons         []string         `json:"reasons"`
	AlternativePath []string         `json:"alternative_path,omitempty"`
}

// DecisionReport is the final standardized output of the agent.
type DecisionReport struct {
	OriginID             string           `json:"origin_id"`
	DestinationID        string           `json:"destination_id"`
	ChosenPath           []string         `json:"chosen_path"`
	LinkEvaluations      []LinkEvaluation `json:"link_evaluations"`
	FinalLatencyEstimate float64          `json:"final_latency_estimate_ms"`
	Explanation          string           `json:"explanation"`
	HopDecisions         []HopDecision    `json:"hop_decisions"`
}

// AgentConfig holds tunable weights and thresholds.
type AgentConfig struct {
	TrustWeightMS       float64 `json:"trust_weight_ms"`
	TargetingWeightMS   float64 `json:"targeting_weight_ms"`
	UncertaintyWeightMS float64 `json:"uncertainty_weight_ms"`
	SwitchingWeightMS   float64 `json:"switching_weight_ms"`
	SaturationLoadRatio float64 `json:"saturation_load_ratio"`
	TopKRoutes          int     `json:"top_k_routes"`
	TrustHardThreshold  float64 `json:"trust_hard_threshold"`
	RiskHardThreshold   float64 `json:"risk_hard_threshold"`
	AckTimeoutMS        int     `json:"ack_timeout_ms"`
	MaxRetries          int     `json:"max_retries"`
	PacketTTL           int     `json:"packet_ttl"`
	PollIntervalMS      int     `json:"poll_interval_ms"`
}

// DefaultConfig returns a safe default configuration.
func DefaultConfig() AgentConfig {
	return AgentConfig{
		TrustWeightMS:       40.0,
		TargetingWeightMS:   25.0,
		UncertaintyWeightMS: 20.0,
		SwitchingWeightMS:   10.0,
		SaturationLoadRatio: 0.90,
		TopKRoutes:          5,
		TrustHardThreshold:  0.40,
		RiskHardThreshold:   0.80,
		AckTimeoutMS:        5000,
		MaxRetries:          3,
		PacketTTL:           10,
		PollIntervalMS:      2000,
	}
}
