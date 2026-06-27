package domain

const (
	PacketCreated   = "created"
	PacketInTransit = "in_transit"
	PacketDelivered = "delivered"
	PacketFailed    = "failed"
)

type Packet struct {
	ID                       string        `json:"id"`
	OriginID                 string        `json:"origin_id"`
	DestinationID            string        `json:"destination_id"`
	CurrentID                string        `json:"current_id"`
	Payload                  string        `json:"payload"`
	EncodedPayload           []string      `json:"encoded_payload,omitempty"`
	BinaryStream             string        `json:"binary_stream,omitempty"`
	HopLog                   []HopLogEntry `json:"hop_log"`
	Route                    []string      `json:"route"`
	RouteIndex               int           `json:"route_index"`
	Status                   string        `json:"status"`
	DecodedPayload           string        `json:"decoded_payload,omitempty"`
	CumulativeLatencySeconds float64       `json:"cumulative_latency_seconds"`
}

type TransmissionResult struct {
	Status              string           `json:"status"`
	Message             string           `json:"message,omitempty"`
	Packet              Packet           `json:"packet"`
	Route               Route            `json:"route"`
	OriginalPayload     string           `json:"original_payload"`
	DecodedPayload      string           `json:"decoded_payload"`
	TotalLatencySeconds float64          `json:"total_latency_seconds"`
	Latency             LatencyBreakdown `json:"latency"`
}
