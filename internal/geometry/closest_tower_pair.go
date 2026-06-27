package geometry

import (
	"fmt"
	"math"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

// ClosestTowerPair preserves the original behaviour.
//
// Existing code that does not provide tower-failure information
// can continue using this function.
func ClosestTowerPair(
	from domain.Planet,
	to domain.Planet,
	scale float64,
) (domain.TowerPair, error) {
	return ClosestActiveTowerPair(
		from,
		to,
		scale,
		nil,
	)
}

// ClosestActiveTowerPair finds the closest operational tower pair
// between two planets.

func ClosestActiveTowerPair(
	from domain.Planet,
	to domain.Planet,
	scale float64,
	disabledTowers map[string]map[int]bool,
) (domain.TowerPair, error) {
	fromTowers := GenerateTowers(from, scale)
	toTowers := GenerateTowers(to, scale)

	if len(fromTowers) == 0 {
		return domain.TowerPair{}, fmt.Errorf(
			"planet %s has no configured towers",
			from.ID,
		)
	}

	if len(toTowers) == 0 {
		return domain.TowerPair{}, fmt.Errorf(
			"planet %s has no configured towers",
			to.ID,
		)
	}

	activeFromCount := countOperationalTowers(
		from.ID,
		fromTowers,
		disabledTowers,
	)

	if activeFromCount == 0 {
		return domain.TowerPair{}, fmt.Errorf(
			"planet %s has no operational towers",
			from.ID,
		)
	}

	activeToCount := countOperationalTowers(
		to.ID,
		toTowers,
		disabledTowers,
	)

	if activeToCount == 0 {
		return domain.TowerPair{}, fmt.Errorf(
			"planet %s has no operational towers",
			to.ID,
		)
	}

	best := domain.TowerPair{
		Distance: math.Inf(1),
	}

	found := false

	for _, fromTower := range fromTowers {
		if IsTowerDisabled(
			disabledTowers,
			from.ID,
			fromTower.Index,
		) {
			continue
		}

		for _, toTower := range toTowers {
			if IsTowerDisabled(
				disabledTowers,
				to.ID,
				toTower.Index,
			) {
				continue
			}

			distance := math.Hypot(
				toTower.XKM-fromTower.XKM,
				toTower.YKM-fromTower.YKM,
			)

			if isBetterTowerPair(
				distance,
				fromTower,
				toTower,
				best,
			) {
				best = domain.TowerPair{
					FromTower: fromTower,
					ToTower:   toTower,
					Distance:  distance,
				}

				found = true
			}
		}
	}

	if !found {
		return domain.TowerPair{}, fmt.Errorf(
			"no operational tower pair exists between %s and %s",
			from.ID,
			to.ID,
		)
	}

	return best, nil
}

func IsTowerDisabled(
	disabledTowers map[string]map[int]bool,
	planetID string,
	towerIndex int,
) bool {
	if disabledTowers == nil {
		return false
	}

	planetTowers, exists := disabledTowers[planetID]
	if !exists {
		return false
	}

	return planetTowers[towerIndex]
}

func countOperationalTowers(
	planetID string,
	towers []domain.Tower,
	disabledTowers map[string]map[int]bool,
) int {
	count := 0

	for _, tower := range towers {
		if !IsTowerDisabled(
			disabledTowers,
			planetID,
			tower.Index,
		) {
			count++
		}
	}

	return count
}

func isBetterTowerPair(
	distance float64,
	fromTower domain.Tower,
	toTower domain.Tower,
	currentBest domain.TowerPair,
) bool {
	const epsilon = 1e-9

	if distance < currentBest.Distance-epsilon {
		return true
	}

	if math.Abs(distance-currentBest.Distance) > epsilon {
		return false
	}

	// Deterministic tie-breaking:
	// 1. Lower sending tower index
	// 2. Lower receiving tower index
	if fromTower.Index < currentBest.FromTower.Index {
		return true
	}

	return fromTower.Index == currentBest.FromTower.Index &&
		toTower.Index < currentBest.ToTower.Index
}
