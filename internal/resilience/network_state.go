package resilience

import (
	"sort"
	"sync"
)

type State struct {
	mu sync.RWMutex

	healthy        map[string]bool
	manualDisabled map[string]bool
	disabledLinks  map[string]bool

	// Planet ID -> tower index -> disabled status.
	disabledTowers map[string]map[int]bool

	failures  map[string]int
	threshold int
}

func NewState(planetIDs []string, threshold int) *State {
	if threshold < 1 {
		threshold = 1
	}

	state := &State{
		healthy:        make(map[string]bool, len(planetIDs)),
		manualDisabled: make(map[string]bool),
		disabledLinks:  make(map[string]bool),
		disabledTowers: make(map[string]map[int]bool),
		failures:       make(map[string]int),
		threshold:      threshold,
	}

	for _, id := range planetIDs {
		state.healthy[id] = true
	}

	return state
}

func (s *State) RecordHealth(id string, healthy bool) (changed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	old := s.healthy[id]

	if healthy {
		s.failures[id] = 0
		s.healthy[id] = true
	} else {
		s.failures[id]++

		if s.failures[id] >= s.threshold {
			s.healthy[id] = false
		}
	}

	return old != s.healthy[id]
}

func (s *State) SetHealth(id string, healthy bool) (changed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	old := s.healthy[id]
	s.healthy[id] = healthy

	if healthy {
		s.failures[id] = 0
	} else {
		s.failures[id] = s.threshold
	}

	return old != healthy
}

func (s *State) DisableNode(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.manualDisabled[id] = true
}

func (s *State) EnableNode(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.manualDisabled, id)
}

func (s *State) DisableLink(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.disabledLinks[id] = true
}

func (s *State) EnableLink(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.disabledLinks, id)
}

func (s *State) DisableTower(planetID string, towerIndex int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.disabledTowers[planetID] == nil {
		s.disabledTowers[planetID] = make(map[int]bool)
	}

	s.disabledTowers[planetID][towerIndex] = true
}

func (s *State) EnableTower(planetID string, towerIndex int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	towers := s.disabledTowers[planetID]
	if towers == nil {
		return
	}

	delete(towers, towerIndex)

	// Remove the empty planet entry to keep the snapshot clean.
	if len(towers) == 0 {
		delete(s.disabledTowers, planetID)
	}
}

func (s *State) IsTowerDisabled(
	planetID string,
	towerIndex int,
) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.disabledTowers[planetID][towerIndex]
}

func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.manualDisabled = make(map[string]bool)
	s.disabledLinks = make(map[string]bool)
	s.disabledTowers = make(map[string]map[int]bool)

	for id := range s.healthy {
		s.healthy[id] = true
		s.failures[id] = 0
	}
}

func (s *State) Snapshot() (
	healthy map[string]bool,
	manual map[string]bool,
	links map[string]bool,
) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	healthy = clone(s.healthy)
	manual = clone(s.manualDisabled)
	links = clone(s.disabledLinks)

	return
}

// DisabledTowerSetSnapshot returns a structure designed for Go calculations.

func (s *State) DisabledTowerSetSnapshot() map[string]map[int]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(
		map[string]map[int]bool,
		len(s.disabledTowers),
	)

	for planetID, towers := range s.disabledTowers {
		copiedTowers := make(map[int]bool, len(towers))

		for towerIndex, disabled := range towers {
			if disabled {
				copiedTowers[towerIndex] = true
			}
		}

		result[planetID] = copiedTowers
	}

	return result
}

// DisabledTowersSnapshot returns a JSON-friendly structure for the API and React frontend.

func (s *State) DisabledTowersSnapshot() map[string][]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(
		map[string][]int,
		len(s.disabledTowers),
	)

	for planetID, towers := range s.disabledTowers {
		indices := make([]int, 0, len(towers))

		for towerIndex, disabled := range towers {
			if disabled {
				indices = append(indices, towerIndex)
			}
		}

		sort.Ints(indices)

		if len(indices) > 0 {
			result[planetID] = indices
		}
	}

	return result
}

func clone(input map[string]bool) map[string]bool {
	output := make(map[string]bool, len(input))

	for key, value := range input {
		output[key] = value
	}

	return output
}
