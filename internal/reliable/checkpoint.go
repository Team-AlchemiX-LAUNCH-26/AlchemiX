package reliable

import "sync"

// Checkpoint tracks the last confirmed relay position per message.
type Checkpoint struct {
	mu        sync.RWMutex
	positions map[string]string // messageID -> last confirmed planet
}

// NewCheckpoint creates a new checkpoint manager.
func NewCheckpoint() *Checkpoint {
	return &Checkpoint{positions: make(map[string]string)}
}

// Advance moves the checkpoint forward for a message.
func (cp *Checkpoint) Advance(messageID, planet string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.positions[messageID] = planet
}

// LastConfirmed returns the last confirmed planet for a message.
func (cp *Checkpoint) LastConfirmed(messageID string) string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.positions[messageID]
}

// Clear removes the checkpoint for a message (after delivery).
func (cp *Checkpoint) Clear(messageID string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	delete(cp.positions, messageID)
}

// All returns all current checkpoints.
func (cp *Checkpoint) All() map[string]string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	out := make(map[string]string, len(cp.positions))
	for k, v := range cp.positions {
		out[k] = v
	}
	return out
}
