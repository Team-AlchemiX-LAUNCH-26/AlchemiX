// Package reliable implements reliable packet transmission with ACK checkpointing.
package reliable

import "time"

// PacketStatus represents the lifecycle state of a packet.
type PacketStatus string

const (
	StatusCreated           PacketStatus = "CREATED"
	StatusChunked           PacketStatus = "CHUNKED"
	StatusQueued            PacketStatus = "QUEUED"
	StatusInFlight          PacketStatus = "IN_FLIGHT"
	StatusAcknowledged      PacketStatus = "ACKNOWLEDGED"
	StatusDeliveryUncertain PacketStatus = "DELIVERY_UNCERTAIN"
	StatusRerouting         PacketStatus = "REROUTING"
	StatusNetworkPartition  PacketStatus = "NETWORK_PARTITION"
	StatusDelivered         PacketStatus = "DELIVERED"
	StatusFailedSafe        PacketStatus = "FAILED_SAFE"
	StatusExpired           PacketStatus = "EXPIRED"
)

// ReliablePacket extends the basic packet with reliability fields.
type ReliablePacket struct {
	MessageID           string       `json:"message_id"`
	PacketID            string       `json:"packet_id"`
	SequenceNumber      int          `json:"sequence_number"`
	TotalPackets        int          `json:"total_packets"`
	Payload             []byte       `json:"payload"`
	Checksum            string       `json:"checksum"`
	CurrentPlanet       string       `json:"current_planet"`
	LastConfirmedPlanet string       `json:"last_confirmed_planet"`
	RouteVersion        int          `json:"route_version"`
	RetryCount          int          `json:"retry_count"`
	TTL                 int          `json:"ttl"`
	Priority            string       `json:"priority"`
	Status              PacketStatus `json:"status"`
	CreatedAt           time.Time    `json:"created_at"`
}

// Sender handles reliable packet transmission with ACK confirmation.
type Sender struct {
	ackTimeout  time.Duration
	maxRetries  int
	checkpoint  *Checkpoint
}

// NewSender creates a sender with the given timeout and retry settings.
func NewSender(ackTimeoutMS, maxRetries int) *Sender {
	return &Sender{
		ackTimeout: time.Duration(ackTimeoutMS) * time.Millisecond,
		maxRetries: maxRetries,
		checkpoint: NewCheckpoint(),
	}
}

// Send transmits a packet and waits for acknowledgement.
// Returns true if ACK received, false if timeout/failure.
func (s *Sender) Send(packet *ReliablePacket, nextPlanet string) bool {
	packet.Status = StatusInFlight

	// In a real implementation, this would:
	// 1. Serialize the packet
	// 2. Send via HTTP to the next planet's endpoint
	// 3. Wait for ACK with timeout
	// 4. Handle retries

	// For now, simulate successful transmission.
	packet.Status = StatusAcknowledged
	packet.LastConfirmedPlanet = nextPlanet
	packet.CurrentPlanet = nextPlanet
	s.checkpoint.Advance(packet.MessageID, nextPlanet)

	return true
}

// SendWithRetry attempts to send with retries on failure.
func (s *Sender) SendWithRetry(packet *ReliablePacket, nextPlanet string) bool {
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		packet.RetryCount = attempt
		if s.Send(packet, nextPlanet) {
			return true
		}
	}
	packet.Status = StatusDeliveryUncertain
	return false
}

// GetCheckpoint returns the checkpoint manager.
func (s *Sender) GetCheckpoint() *Checkpoint {
	return s.checkpoint
}
