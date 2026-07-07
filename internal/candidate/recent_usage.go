package candidate

import "sync"

// RecentUsageTracker tracks which links have been recently selected.
type RecentUsageTracker struct {
	mu               sync.Mutex
	selections       map[string]int   // linkID -> count in recent window
	consecutiveUses  map[string]int   // linkID -> consecutive use count
	lastSelectedTick map[string]int64 // linkID -> tick of last selection
	lastLinks        []string         // ordered list of recently used links
	windowSize       int
}

// NewRecentUsageTracker creates a tracker with the given window size.
func NewRecentUsageTracker(windowSize int) *RecentUsageTracker {
	if windowSize < 1 {
		windowSize = 10
	}
	return &RecentUsageTracker{
		selections:       make(map[string]int),
		consecutiveUses:  make(map[string]int),
		lastSelectedTick: make(map[string]int64),
		lastLinks:        make([]string, 0),
		windowSize:       windowSize,
	}
}

// Record records a link selection at the given tick.
func (r *RecentUsageTracker) Record(linkID string, tick int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.selections[linkID]++
	r.lastSelectedTick[linkID] = tick

	// Track consecutive uses.
	if len(r.lastLinks) > 0 && r.lastLinks[len(r.lastLinks)-1] == linkID {
		r.consecutiveUses[linkID]++
	} else {
		r.consecutiveUses[linkID] = 1
	}

	r.lastLinks = append(r.lastLinks, linkID)

	// Trim window.
	if len(r.lastLinks) > r.windowSize {
		removed := r.lastLinks[0]
		r.lastLinks = r.lastLinks[1:]
		r.selections[removed]--
		if r.selections[removed] <= 0 {
			delete(r.selections, removed)
		}
	}
}

// Selections returns the current selection counts.
func (r *RecentUsageTracker) Selections() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]int, len(r.selections))
	for k, v := range r.selections {
		out[k] = v
	}
	return out
}

// ConsecutiveUses returns the current consecutive use counts.
func (r *RecentUsageTracker) ConsecutiveUses() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]int, len(r.consecutiveUses))
	for k, v := range r.consecutiveUses {
		out[k] = v
	}
	return out
}

// LastSelectedTick returns the last tick each link was selected.
func (r *RecentUsageTracker) LastSelectedTick() map[string]int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]int64, len(r.lastSelectedTick))
	for k, v := range r.lastSelectedTick {
		out[k] = v
	}
	return out
}

// Reset clears all tracking state.
func (r *RecentUsageTracker) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.selections = make(map[string]int)
	r.consecutiveUses = make(map[string]int)
	r.lastSelectedTick = make(map[string]int64)
	r.lastLinks = r.lastLinks[:0]
}
