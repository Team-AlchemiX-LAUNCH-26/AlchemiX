package intelligence

import (
	"math"
	"sync"

	"github.com/launch26/relic-ring-protocol/internal/agent"
)

// FeatureHistory maintains the last N observations per link to compute leakage-safe rolling features.
type FeatureHistory struct {
	mu       sync.RWMutex
	history  map[string][]agent.LinkObservation
	lastTick map[string]int64
}

// NewFeatureHistory creates a new rolling history cache.
func NewFeatureHistory() *FeatureHistory {
	return &FeatureHistory{
		history:  make(map[string][]agent.LinkObservation),
		lastTick: make(map[string]int64),
	}
}

// AddObservation adds an observation to the history. It deduplicates by tick.
func (h *FeatureHistory) AddObservation(obs agent.LinkObservation) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.lastTick[obs.LinkID] >= obs.Tick {
		return // Already processed or out of order.
	}
	h.lastTick[obs.LinkID] = obs.Tick

	list := h.history[obs.LinkID]
	list = append(list, obs)
	if len(list) > 10 {
		list = list[len(list)-10:] // Keep max 10 for rolling_jam_count
	}
	h.history[obs.LinkID] = list
}

// LaggedMean returns the mean of a metric over the last N ticks (excluding current).
func (h *FeatureHistory) LaggedMean(linkID string, window int, extractor func(agent.LinkObservation) float64) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	list := h.history[linkID]
	if len(list) == 0 {
		return 0
	}
	if window > len(list) {
		window = len(list)
	}
	sum := 0.0
	for i := len(list) - window; i < len(list); i++ {
		sum += extractor(list[i])
	}
	return sum / float64(window)
}

// LaggedStd returns the std dev of a metric over the last N ticks (excluding current).
func (h *FeatureHistory) LaggedStd(linkID string, window int, extractor func(agent.LinkObservation) float64) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	list := h.history[linkID]
	if len(list) < 2 {
		return 0
	}
	if window > len(list) {
		window = len(list)
	}
	if window < 2 {
		return 0
	}

	sum := 0.0
	for i := len(list) - window; i < len(list); i++ {
		sum += extractor(list[i])
	}
	mean := sum / float64(window)

	variance := 0.0
	for i := len(list) - window; i < len(list); i++ {
		val := extractor(list[i])
		variance += (val - mean) * (val - mean)
	}
	// Pandas uses n-1 for std dev (sample std)
	return math.Sqrt(variance / float64(window-1))
}

// LaggedSum returns the sum of a metric over the last N ticks (excluding current).
func (h *FeatureHistory) LaggedSum(linkID string, window int, extractor func(agent.LinkObservation) float64) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	list := h.history[linkID]
	if len(list) == 0 {
		return 0
	}
	if window > len(list) {
		window = len(list)
	}
	sum := 0.0
	for i := len(list) - window; i < len(list); i++ {
		sum += extractor(list[i])
	}
	return sum
}

// LastValue returns the immediately preceding tick's value (excluding current).
func (h *FeatureHistory) LastValue(linkID string, extractor func(agent.LinkObservation) float64, fallback float64) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	list := h.history[linkID]
	if len(list) == 0 {
		return fallback
	}
	return extractor(list[len(list)-1])
}

// GetHistoryCopy returns a copy of the slice of past observations.
func (h *FeatureHistory) GetHistoryCopy(linkID string) []agent.LinkObservation {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	list := h.history[linkID]
	copyList := make([]agent.LinkObservation, len(list))
	copy(copyList, list)
	return copyList
}
