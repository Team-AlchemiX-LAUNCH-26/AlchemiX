package latency

import (
	"github.com/launch26/relic-ring-protocol/internal/domain"
	"github.com/launch26/relic-ring-protocol/internal/geometry"
)

// PlanetTransit preserves the original function used by the
// current routing and packet code.
//
// It assumes that no towers are disabled.
//
// This compatibility function will be migrated to the
// failure-aware version in the routing integration phase.
func PlanetTransit(
	p domain.Planet,
	entryTower int,
	exitTower int,
	metadata domain.UniverseMetadata,
) domain.PlanetTransitBreakdown {
	transit, _ := PlanetTransitWithFailures(
		p,
		entryTower,
		exitTower,
		metadata,
		nil,
	)

	return transit
}

// PlanetTransitWithFailures calculates the official Internal
// Crust Transit Time while avoiding disabled towers.
//
// disabledTowers example:
//
//	map[int]bool{
//	    2: true,
//	    6: true,
//	}
func PlanetTransitWithFailures(
	p domain.Planet,
	entryTower int,
	exitTower int,
	metadata domain.UniverseMetadata,
	disabledTowers map[int]bool,
) (domain.PlanetTransitBreakdown, error) {
	ringPath, err := geometry.ShortestOperationalRingPath(
		entryTower,
		exitTower,
		p.ActiveTowers,
		disabledTowers,
	)

	if err != nil {
		return domain.PlanetTransitBreakdown{}, err
	}

	fiberSeconds := FiberSeconds(
		p.RadiusKM,
		p.ActiveTowers,
		ringPath.Segments,
		metadata.FiberSpeedFraction,
		metadata.SpeedOfLightKMS,
	)

	towerSeconds := TowerDelaySeconds(
		ringPath.DistinctTowers,
		metadata.TowerProcessingDelayMS,
	)

	return domain.PlanetTransitBreakdown{
		PlanetID:       p.ID,
		EntryTower:     entryTower,
		ExitTower:      exitTower,
		RingPath:       append([]int(nil), ringPath.Towers...),
		RingDirection:  ringPath.Direction,
		Segments:       ringPath.Segments,
		DistinctTowers: ringPath.DistinctTowers,
		FiberSeconds:   fiberSeconds,
		TowerSeconds:   towerSeconds,
		TotalSeconds:   fiberSeconds + towerSeconds,
	}, nil
}
