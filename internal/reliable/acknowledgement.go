package reliable

import "sync"

// AckTracker manages acknowledgement state for in-flight packets.
type AckTracker struct {
	mu       sync.Mutex
	pending  map[string]bool   // packetID -> awaiting ACK
	received map[string]bool   // packetID -> ACK received
	duplicates map[string]int  // packetID -> duplicate count
}

// NewAckTracker creates a new ACK tracker.
func NewAckTracker() *AckTracker {
	return &AckTracker{
		pending:    make(map[string]bool),
		received:   make(map[string]bool),
		duplicates: make(map[string]int),
	}
}

// MarkPending registers a packet as awaiting ACK.
func (at *AckTracker) MarkPending(packetID string) {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.pending[packetID] = true
}

// ReceiveAck processes an ACK for a packet.
func (at *AckTracker) ReceiveAck(packetID string) bool {
	at.mu.Lock()
	defer at.mu.Unlock()

	if at.received[packetID] {
		at.duplicates[packetID]++
		return false // Duplicate ACK.
	}
	at.received[packetID] = true
	delete(at.pending, packetID)
	return true
}

// IsPending returns true if the packet is still awaiting ACK.
func (at *AckTracker) IsPending(packetID string) bool {
	at.mu.Lock()
	defer at.mu.Unlock()
	return at.pending[packetID]
}

// PendingCount returns the number of packets awaiting ACK.
func (at *AckTracker) PendingCount() int {
	at.mu.Lock()
	defer at.mu.Unlock()
	return len(at.pending)
}
