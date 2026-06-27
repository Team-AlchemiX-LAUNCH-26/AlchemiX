package protocol

import "github.com/launch26/relic-ring-protocol/internal/domain"

type TransmissionRequest struct {
	OriginID      string `json:"origin_id"`
	DestinationID string `json:"destination_id"`
	Payload       string `json:"payload"`
}

type LinkRequest struct {
	A string `json:"a"`
	B string `json:"b"`
}

type TowerRequest struct {
	PlanetID   string `json:"planet_id"`
	TowerIndex int    `json:"tower_index"`
}

type NodePacketRequest struct {
	Packet        domain.Packet     `json:"packet"`
	Route         domain.Route      `json:"route"`
	NodeEndpoints map[string]string `json:"node_endpoints"`
	TelemetryURL  string            `json:"telemetry_url,omitempty"`

	DisabledTowers map[string][]int `json:"disabled_towers,omitempty"`
}
