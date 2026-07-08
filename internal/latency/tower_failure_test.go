package latency

import (
	"math"
	"reflect"
	"testing"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func TestPlanetTransitWithFailuresUsesAlternativeRingPath(
	t *testing.T,
) {
	planet := domain.Planet{
		ID:           "Aegis",
		RadiusKM:     100,
		ActiveTowers: 8,
	}

	metadata := domain.UniverseMetadata{
		SpeedOfLightKMS:        300000,
		FiberSpeedFraction:     0.67,
		TowerProcessingDelayMS: 7,
	}

	disabled := map[int]bool{
		2: true,
	}

	got, err := PlanetTransitWithFailures(
		planet,
		1,
		4,
		metadata,
		disabled,
	)

	if err != nil {
		t.Fatal(err)
	}

	wantPath := []int{1, 0, 7, 6, 5, 4}

	if !reflect.DeepEqual(got.RingPath, wantPath) {
		t.Fatalf(
			"expected ring path %v, got %v",
			wantPath,
			got.RingPath,
		)
	}

	if got.RingDirection !=
		domain.RingDirectionCounterClockwise {
		t.Fatalf(
			"expected counter-clockwise, got %s",
			got.RingDirection,
		)
	}

	if got.Segments != 5 {
		t.Fatalf(
			"expected 5 segments, got %d",
			got.Segments,
		)
	}

	if got.DistinctTowers != 6 {
		t.Fatalf(
			"expected 6 distinct towers, got %d",
			got.DistinctTowers,
		)
	}

	expectedFiber := FiberSeconds(
		planet.RadiusKM,
		planet.ActiveTowers,
		5,
		metadata.FiberSpeedFraction,
		metadata.SpeedOfLightKMS,
	)

	expectedTower :=
		6 * metadata.TowerProcessingDelayMS / 1000

	if math.Abs(got.FiberSeconds-expectedFiber) > 1e-12 {
		t.Fatalf(
			"expected fiber %.12f, got %.12f",
			expectedFiber,
			got.FiberSeconds,
		)
	}

	if math.Abs(got.TowerSeconds-expectedTower) > 1e-12 {
		t.Fatalf(
			"expected tower %.12f, got %.12f",
			expectedTower,
			got.TowerSeconds,
		)
	}

	expectedTotal := expectedFiber + expectedTower

	if math.Abs(got.TotalSeconds-expectedTotal) > 1e-12 {
		t.Fatalf(
			"expected total %.12f, got %.12f",
			expectedTotal,
			got.TotalSeconds,
		)
	}
}

func TestPlanetTransitWithFailuresReturnsErrorWhenRingBlocked(
	t *testing.T,
) {
	planet := domain.Planet{
		ID:           "Aegis",
		RadiusKM:     100,
		ActiveTowers: 8,
	}

	metadata := domain.UniverseMetadata{
		SpeedOfLightKMS:        300000,
		FiberSpeedFraction:     0.67,
		TowerProcessingDelayMS: 7,
	}

	disabled := map[int]bool{
		0: true,
		2: true,
	}

	_, err := PlanetTransitWithFailures(
		planet,
		1,
		4,
		metadata,
		disabled,
	)

	if err == nil {
		t.Fatal(
			"expected an error because both ring directions are blocked",
		)
	}
}

func TestPlanetTransitWithFailuresSameTower(
	t *testing.T,
) {
	planet := domain.Planet{
		ID:           "Aegis",
		RadiusKM:     100,
		ActiveTowers: 8,
	}

	metadata := domain.UniverseMetadata{
		SpeedOfLightKMS:        300000,
		FiberSpeedFraction:     0.67,
		TowerProcessingDelayMS: 7,
	}

	got, err := PlanetTransitWithFailures(
		planet,
		3,
		3,
		metadata,
		nil,
	)

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got.RingPath, []int{3}) {
		t.Fatalf(
			"expected stationary path [3], got %v",
			got.RingPath,
		)
	}

	if got.RingDirection !=
		domain.RingDirectionStationary {
		t.Fatalf(
			"expected stationary direction, got %s",
			got.RingDirection,
		)
	}

	if got.Segments != 0 {
		t.Fatalf(
			"expected zero segments, got %d",
			got.Segments,
		)
	}

	if got.DistinctTowers != 1 {
		t.Fatalf(
			"expected one distinct tower, got %d",
			got.DistinctTowers,
		)
	}

	if math.Abs(got.TotalSeconds-0.007) > 1e-12 {
		t.Fatalf(
			"expected 0.007 seconds, got %.12f",
			got.TotalSeconds,
		)
	}
}

func TestPlanetTransitCompatibilityFunctionStillWorks(
	t *testing.T,
) {
	planet := domain.Planet{
		ID:           "Aegis",
		RadiusKM:     100,
		ActiveTowers: 8,
	}

	metadata := domain.UniverseMetadata{
		SpeedOfLightKMS:        300000,
		FiberSpeedFraction:     0.67,
		TowerProcessingDelayMS: 7,
	}

	got := PlanetTransit(
		planet,
		3,
		3,
		metadata,
	)

	if !reflect.DeepEqual(got.RingPath, []int{3}) {
		t.Fatalf(
			"expected ring path [3], got %v",
			got.RingPath,
		)
	}

	if got.RingDirection !=
		domain.RingDirectionStationary {
		t.Fatalf(
			"expected stationary direction, got %s",
			got.RingDirection,
		)
	}
}
