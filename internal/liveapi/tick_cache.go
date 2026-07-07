package liveapi

import (
	"sync"

	"github.com/launch26/relic-ring-protocol/internal/agent"
)

// TickCache provides tick deduplication and last-known-good fallback.
type TickCache struct {
	mu          sync.RWMutex
	currentTick int64
	lastGood    *agent.NetworkState
}

// NewTickCache creates a new tick cache.
func NewTickCache() *TickCache {
	return &TickCache{}
}

// NextTick returns an incrementing tick for mock use.
func (tc *TickCache) NextTick() int64 {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.currentTick++
	return tc.currentTick
}

// Store saves a state snapshot as the last-known-good.
func (tc *TickCache) Store(state agent.NetworkState) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.lastGood = &state
	tc.currentTick = state.Tick
}

// IsDuplicate returns true if the given tick was already processed.
func (tc *TickCache) IsDuplicate(tick int64) bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tick > 0 && tick == tc.currentTick
}

// LastGood returns the last valid state, or nil if none stored.
func (tc *TickCache) LastGood() *agent.NetworkState {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	if tc.lastGood == nil {
		return nil
	}
	copy := *tc.lastGood
	return &copy
}
