package domain

type HopLogEntry struct {
	Sequence         int    `json:"sequence"`
	PlanetID         string `json:"planet_id"`
	PreviousPlanetID string `json:"previous_planet_id,omitempty"`
	NextPlanetID     string `json:"next_planet_id,omitempty"`

	EntryTower int `json:"entry_tower"`
	ExitTower  int `json:"exit_tower"`

	// Exact internal tower route followed by the packet.
	RingPath []int `json:"ring_path"`

	RingDirection RingDirection `json:"ring_direction"`

	Segments       int `json:"segments"`
	DistinctTowers int `json:"distinct_towers"`

	LocalCodex   int `json:"local_codex"`
	NextHopCodex int `json:"next_hop_codex,omitempty"`

	DecodedPayload string   `json:"decoded_payload"`
	EncodedPayload []string `json:"encoded_payload,omitempty"`
	BinaryStream   string   `json:"binary_stream,omitempty"`

	PlanetTransit PlanetTransitBreakdown `json:"planet_transit"`
	VoidTransit   *VoidTransitBreakdown  `json:"void_transit,omitempty"`

	StepLatency       float64 `json:"step_latency_seconds"`
	CumulativeLatency float64 `json:"cumulative_latency_seconds"`
}
