package routing

import (
	"sort"

	"github.com/launch26/relic-ring-protocol/internal/domain"
	"github.com/launch26/relic-ring-protocol/internal/geometry"
)

func LinkID(a, b string) string {
	pair := []string{a, b}
	sort.Strings(pair)

	return pair[0] + "::" + pair[1]
}

func BuildLink(
	a domain.Planet,
	b domain.Planet,
	metadata domain.UniverseMetadata,
	disabledTowers map[string]map[int]bool,
) (domain.Link, bool, error) {
	voidDistance, err := geometry.VoidDistanceKM(
		a,
		b,
		metadata.CoordinateScaleUnitKM,
	)

	if err != nil {
		return domain.Link{}, false, err
	}

	if voidDistance > metadata.MaxVoidHopDistanceKM {
		return domain.Link{}, false, nil
	}

	pair, err := geometry.ClosestActiveTowerPair(
		a,
		b,
		metadata.CoordinateScaleUnitKM,
		disabledTowers,
	)

	// If towers are unavailable, this specific link cannot be used.
	// This should not crash the whole graph. The router can still
	// try other links and routes.
	if err != nil {
		return domain.Link{}, false, nil
	}

	return domain.Link{
		A:              a.ID,
		B:              b.ID,
		VoidDistanceKM: voidDistance,
		ATower:         pair.FromTower.Index,
		BTower:         pair.ToTower.Index,
	}, true, nil
}
