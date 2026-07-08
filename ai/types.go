package ai

// LinkState holds the live network telemetry for a single link,
// as fetched from the competition GET /state API.
// All fields mirror the column names from the training datasets.
type LinkState struct {
	// Identification
	LinkID string `json:"link_id"` // e.g. "Aegis-Boreas"

	// From link_traffic_history (Congestion model inputs)
	LoadUnits float64 `json:"load_units"`
	LoadRatio float64 `json:"load_ratio"` // 0.0–1.0
	Status    string  `json:"status"`     // "ok" | "saturated"

	// From link_telemetry (Trust model inputs)
	SelfReportedLatencyMS float64 `json:"self_reported_latency_ms"`
	MeasuredLatencyMS     float64 `json:"measured_latency_ms"`

	// From link_incident_history (Targeting model inputs)
	TrafficShare float64 `json:"traffic_share"` // 0.0–1.0
}

// LinkScore is the AI Agent's assessment of a single link.
// All penalty fields are in milliseconds so they are directly
// comparable to the Phase 1 physics latency.
type LinkScore struct {
	LinkID string `json:"link_id"`

	// Phase 1 physics latency for this void transit (ms).
	// Derived from domain.VoidTransitBreakdown.TotalSeconds × 1000.
	PhysicsLatencyMS float64 `json:"physics_latency_ms"`

	// Model 1 – Congestion
	CongestionPenaltyMS float64 `json:"congestion_penalty_ms"`

	// Model 2 – Trust
	TrustScore      float64 `json:"trust_score"`       // 0–1; 1 = fully honest
	TrustPenaltyMS  float64 `json:"trust_penalty_ms"`  // (1−trust)×TrustWeight

	// Model 3 – Targeting
	JamProbability      float64 `json:"jam_probability"`       // 0–1
	TargetingPenaltyMS  float64 `json:"targeting_penalty_ms"`  // P(jam)×TargetingWeight

	// Sum
	TrueCostMS float64 `json:"true_cost_ms"`

	// Human-readable explanation for this link's score.
	Explanation string `json:"explanation"`
}

// RouteAssessment is the complete AI Agent output for a route.
type RouteAssessment struct {
	// The planet path chosen by Phase 1 router (unchanged).
	Path []string `json:"path"`

	// Per-link scores in path order.
	LinkScores []LinkScore `json:"link_scores"`

	// Sum of physics latency across all links (mirrors Phase 1 total, ms).
	TotalPhysicsMS float64 `json:"total_physics_ms"`

	// Sum of all AI penalty adjustments (ms).
	TotalAIAdjustmentMS float64 `json:"total_ai_adjustment_ms"`

	// Combined True Cost (ms).
	TotalTrueCostMS float64 `json:"total_true_cost_ms"`

	// Qualitative safety label.
	SafetyLabel string `json:"safety_label"` // "safe" | "caution" | "danger"

	// Free-text summary the UI can display.
	Summary string `json:"summary"`

	// Whether the AI agent was available. False = fallback to physics only.
	AgentActive bool `json:"agent_active"`
}

// InferenceRequest is what the Go agent sends to the Python inference server.
type InferenceRequest struct {
	LinkID                string  `json:"link_id"`
	LoadUnits             float64 `json:"load_units"`
	LoadRatio             float64 `json:"load_ratio"`
	Status                string  `json:"status"`
	SelfReportedLatencyMS float64 `json:"self_reported_latency_ms"`
	MeasuredLatencyMS     float64 `json:"measured_latency_ms"`
	TrafficShare          float64 `json:"traffic_share"`
}

// InferenceResponse is what the Python inference server returns.
type InferenceResponse struct {
	LinkID              string  `json:"link_id"`
	CongestionPenaltyMS float64 `json:"congestion_penalty_ms"`
	TrustScore          float64 `json:"trust_score"`
	TrustPenaltyMS      float64 `json:"trust_penalty_ms"`
	JamProbability      float64 `json:"jam_probability"`
	TargetingPenaltyMS  float64 `json:"targeting_penalty_ms"`
}

// LiveLinkState is the raw response body from the competition GET /state API.
// The actual competition schema may differ; fields are mapped by best guess
// from the training dataset column names.
type LiveLinkState struct {
	LinkID                string  `json:"link_id"`
	Tick                  int     `json:"tick"`
	LoadUnits             float64 `json:"load_units"`
	LoadRatio             float64 `json:"load_ratio"`
	Status                string  `json:"status"`
	ObservedLatencyMS     float64 `json:"observed_latency_ms"`
	SelfReportedLatencyMS float64 `json:"self_reported_latency_ms"`
	MeasuredLatencyMS     float64 `json:"measured_latency_ms"`
	TrafficShare          float64 `json:"traffic_share"`
	JammedFlag            bool    `json:"jammed_flag"`
}

// LiveStateResponse wraps the top-level competition API response.
type LiveStateResponse struct {
	Tick  int             `json:"tick"`
	Links []LiveLinkState `json:"links"`
}
