package geometry

import (
	"fmt"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

// ShortestRingSegments preserves the original behaviour.
//
// Existing code can continue calling this function until the
// latency engine is updated in Phase 3.
func ShortestRingSegments(
	entryTower int,
	exitTower int,
	towerCount int,
) (segments int, distinctTowers int) {
	path, err := ShortestOperationalRingPath(
		entryTower,
		exitTower,
		towerCount,
		nil,
	)

	if err != nil {
		return 0, 0
	}

	return path.Segments, path.DistinctTowers
}

// ShortestOperationalRingPath finds the shortest valid path
// between two towers while avoiding disabled towers.

func ShortestOperationalRingPath(
	entryTower int,
	exitTower int,
	towerCount int,
	disabled map[int]bool,
) (domain.RingPath, error) {
	if towerCount <= 0 {
		return domain.RingPath{}, fmt.Errorf(
			"tower count must be greater than zero",
		)
	}

	if entryTower < 0 || entryTower >= towerCount {
		return domain.RingPath{}, fmt.Errorf(
			"entry tower %d is outside valid range 0-%d",
			entryTower,
			towerCount-1,
		)
	}

	if exitTower < 0 || exitTower >= towerCount {
		return domain.RingPath{}, fmt.Errorf(
			"exit tower %d is outside valid range 0-%d",
			exitTower,
			towerCount-1,
		)
	}

	if disabled[entryTower] {
		return domain.RingPath{}, fmt.Errorf(
			"entry tower %d is disabled",
			entryTower,
		)
	}

	if disabled[exitTower] {
		return domain.RingPath{}, fmt.Errorf(
			"exit tower %d is disabled",
			exitTower,
		)
	}

	// No fibre movement is needed when the packet enters
	// and leaves through the same operational tower.
	if entryTower == exitTower {
		return domain.RingPath{
			Towers:         []int{entryTower},
			Segments:       0,
			DistinctTowers: 1,
			Direction:      domain.RingDirectionStationary,
		}, nil
	}

	clockwise, clockwiseValid := buildRingPath(
		entryTower,
		exitTower,
		towerCount,
		1,
		domain.RingDirectionClockwise,
		disabled,
	)

	counterClockwise, counterClockwiseValid := buildRingPath(
		entryTower,
		exitTower,
		towerCount,
		-1,
		domain.RingDirectionCounterClockwise,
		disabled,
	)

	switch {
	case !clockwiseValid && !counterClockwiseValid:
		return domain.RingPath{}, fmt.Errorf(
			"no operational ring path exists between towers %d and %d",
			entryTower,
			exitTower,
		)

	case clockwiseValid && !counterClockwiseValid:
		return clockwise, nil

	case !clockwiseValid && counterClockwiseValid:
		return counterClockwise, nil

	case clockwise.Segments < counterClockwise.Segments:
		return clockwise, nil

	case counterClockwise.Segments < clockwise.Segments:
		return counterClockwise, nil

	default:
		// Deterministic tie-breaking:
		// choose clockwise when both paths have equal length.
		return clockwise, nil
	}
}

func buildRingPath(
	entryTower int,
	exitTower int,
	towerCount int,
	step int,
	direction domain.RingDirection,
	disabled map[int]bool,
) (domain.RingPath, bool) {
	towers := []int{entryTower}
	current := entryTower

	// At most towerCount steps are needed to travel around
	// one complete ring.
	for travelled := 0; travelled < towerCount; travelled++ {
		current = normalizeTowerIndex(
			current+step,
			towerCount,
		)

		// A broken tower blocks this direction.
		if disabled[current] {
			return domain.RingPath{}, false
		}

		towers = append(towers, current)

		if current == exitTower {
			return domain.RingPath{
				Towers:         towers,
				Segments:       len(towers) - 1,
				DistinctTowers: len(towers),
				Direction:      direction,
			}, true
		}
	}

	return domain.RingPath{}, false
}

func normalizeTowerIndex(
	index int,
	towerCount int,
) int {
	index %= towerCount

	if index < 0 {
		index += towerCount
	}

	return index
}
