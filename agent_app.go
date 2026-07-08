package main

import (
	"context"
	"errors"
	"strings"
)

// AgentApp is a thin Wails-facing adapter.
//
// It deliberately depends on function bindings instead of importing the
// concrete internal agent package. This keeps Wails integration isolated and
// avoids circular imports. main.go converts the real agent types to these DTOs.
type AgentApp struct {
	ctx      context.Context
	bindings AgentBindings
}

type AgentBindings struct {
	Parse    func(raw string) (ParsedTransmissionRequestDTO, error)
	Execute  func(ctx context.Context, raw string) (DecisionReportDTO, error)
	State    func() AgentStateDTO
	Audit    func() []AuditRecordDTO
	Timeline func() []PacketTimelineEntryDTO
	Reset    func() error
}

type ParsedTransmissionRequestDTO struct {
	OriginID      string   `json:"origin_id"`
	DestinationID string   `json:"destination_id"`
	Payload       string   `json:"payload"`
	Confidence    float64  `json:"confidence"`
	Ambiguities   []string `json:"ambiguities"`
}

type LinkEvaluationDTO struct {
	LinkID                       string   `json:"link_id"`
	PhysicalLatencyMS            float64  `json:"physical_latency_ms,omitempty"`
	PredictedCongestionPenaltyMS float64  `json:"predicted_congestion_penalty_ms"`
	TrustScore                   float64  `json:"trust_score"`
	TargetingRiskScore           float64  `json:"targeting_risk_score"`
	UncertaintyScore             float64  `json:"uncertainty_score,omitempty"`
	CombinedCost                 float64  `json:"combined_cost"`
	Action                       string   `json:"action,omitempty"`
	Reasons                      []string `json:"reasons,omitempty"`
}

type DecisionReportDTO struct {
	OriginID              string              `json:"origin_id"`
	DestinationID         string              `json:"destination_id"`
	ParsedPayload         string              `json:"parsed_payload,omitempty"`
	BaselinePath          []string            `json:"baseline_path,omitempty"`
	ChosenPath            []string            `json:"chosen_path"`
	LinkEvaluations       []LinkEvaluationDTO `json:"link_evaluations"`
	FinalLatencyEstimateMS float64             `json:"final_latency_estimate_ms"`
	Explanation           string              `json:"explanation"`
	Status                string              `json:"status,omitempty"`
	Confidence            float64             `json:"confidence,omitempty"`
}

type AgentStateDTO struct {
	Status              string   `json:"status"`
	CurrentTick         int64    `json:"current_tick"`
	CurrentPlanet       string   `json:"current_planet,omitempty"`
	DestinationPlanet   string   `json:"destination_planet,omitempty"`
	ActiveMessageID     string   `json:"active_message_id,omitempty"`
	LastConfirmedPlanet string   `json:"last_confirmed_planet,omitempty"`
	CurrentPath         []string `json:"current_path,omitempty"`
	QueuedPackets       int      `json:"queued_packets"`
	QuarantinedLinks    []string `json:"quarantined_links,omitempty"`
	LastError           string   `json:"last_error,omitempty"`
}

type AuditRecordDTO struct {
	AuditID                     string   `json:"audit_id"`
	Tick                        int64    `json:"tick"`
	CurrentPlanet               string   `json:"current_planet"`
	LinkID                      string   `json:"link_id"`
	Action                      string   `json:"action"`
	PhysicalLatencyMS           float64  `json:"physical_latency_ms"`
	PredictedCongestionPenaltyMS float64 `json:"predicted_congestion_penalty_ms"`
	TrustScore                  float64  `json:"trust_score"`
	TargetingRiskScore          float64  `json:"targeting_risk_score"`
	UncertaintyScore            float64  `json:"uncertainty_score"`
	CombinedCost                float64  `json:"combined_cost"`
	Reasons                     []string `json:"reasons"`
	AlternativePath             []string `json:"alternative_path,omitempty"`
	Timestamp                   string   `json:"timestamp,omitempty"`
}

type PacketTimelineEntryDTO struct {
	MessageID      string `json:"message_id"`
	PacketID       string `json:"packet_id"`
	SequenceNumber int    `json:"sequence_number"`
	TotalPackets   int    `json:"total_packets"`
	State          string `json:"state"`
	PlanetID       string `json:"planet_id,omitempty"`
	LinkID         string `json:"link_id,omitempty"`
	RouteVersion   int    `json:"route_version"`
	RetryCount     int    `json:"retry_count"`
	Tick           int64  `json:"tick"`
	Detail         string `json:"detail,omitempty"`
}

func NewAgentApp(bindings AgentBindings) (*AgentApp, error) {
	switch {
	case bindings.Parse == nil:
		return nil, errors.New("agent binding Parse is required")
	case bindings.Execute == nil:
		return nil, errors.New("agent binding Execute is required")
	case bindings.State == nil:
		return nil, errors.New("agent binding State is required")
	case bindings.Audit == nil:
		return nil, errors.New("agent binding Audit is required")
	case bindings.Timeline == nil:
		return nil, errors.New("agent binding Timeline is required")
	case bindings.Reset == nil:
		return nil, errors.New("agent binding Reset is required")
	default:
		return &AgentApp{bindings: bindings}, nil
	}
}

// Startup is called by Wails when the application context becomes available.
func (a *AgentApp) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *AgentApp) ParseTransmissionRequest(raw string) (ParsedTransmissionRequestDTO, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ParsedTransmissionRequestDTO{}, errors.New("transmission request cannot be empty")
	}
	return a.bindings.Parse(raw)
}

func (a *AgentApp) EvaluateTransmission(raw string) (DecisionReportDTO, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DecisionReportDTO{}, errors.New("transmission request cannot be empty")
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.bindings.Execute(ctx, raw)
}

func (a *AgentApp) GetAgentState() AgentStateDTO {
	return a.bindings.State()
}

func (a *AgentApp) GetDecisionAudit() []AuditRecordDTO {
	records := a.bindings.Audit()
	if records == nil {
		return []AuditRecordDTO{}
	}
	return records
}

func (a *AgentApp) GetPacketTimeline() []PacketTimelineEntryDTO {
	entries := a.bindings.Timeline()
	if entries == nil {
		return []PacketTimelineEntryDTO{}
	}
	return entries
}

func (a *AgentApp) ResetAgentHistory() error {
	return a.bindings.Reset()
}
