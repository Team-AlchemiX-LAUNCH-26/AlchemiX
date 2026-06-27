package routing

import (
	"testing"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func testMetadata() domain.UniverseMetadata {
	return domain.UniverseMetadata{
		SpeedOfLightKMS:        300000,
		MaxVoidHopDistanceKM:   1000000,
		CoordinateScaleUnitKM:  1,
		TowerProcessingDelayMS: 7,
		FiberSpeedFraction:     0.67,
	}
}

func TestBuildLinkUsesAlternativeTowerWhenClosestIsDisabled(
	t *testing.T,
) {
	left := domain.Planet{
		ID:                    "Left",
		X:                     0,
		Y:                     0,
		RadiusKM:              1,
		ActiveTowers:          4,
		AtmosphereThicknessKM: 0,
		RefractionIndex:       1,
	}

	right := domain.Planet{
		ID:                    "Right",
		X:                     10,
		Y:                     0,
		RadiusKM:              1,
		ActiveTowers:          4,
		AtmosphereThicknessKM: 0,
		RefractionIndex:       1,
	}

	disabled := map[string]map[int]bool{
		"Left": {
			1: true,
		},
	}

	link, valid, err := BuildLink(
		left,
		right,
		testMetadata(),
		disabled,
	)

	if err != nil {
		t.Fatal(err)
	}

	if !valid {
		t.Fatal("expected link to remain valid")
	}

	if link.ATower == 1 {
		t.Fatal("disabled tower must not be selected")
	}

	if link.ATower != 0 {
		t.Fatalf(
			"expected fallback tower 0, got %d",
			link.ATower,
		)
	}

	if link.BTower != 3 {
		t.Fatalf(
			"expected right-side receiving tower 3, got %d",
			link.BTower,
		)
	}
}

func TestBuildLinkBecomesInvalidWhenAllSourceTowersDisabled(
	t *testing.T,
) {
	left := domain.Planet{
		ID:                    "Left",
		X:                     0,
		Y:                     0,
		RadiusKM:              1,
		ActiveTowers:          4,
		AtmosphereThicknessKM: 0,
		RefractionIndex:       1,
	}

	right := domain.Planet{
		ID:                    "Right",
		X:                     10,
		Y:                     0,
		RadiusKM:              1,
		ActiveTowers:          4,
		AtmosphereThicknessKM: 0,
		RefractionIndex:       1,
	}

	disabled := map[string]map[int]bool{
		"Left": {
			0: true,
			1: true,
			2: true,
			3: true,
		},
	}

	_, valid, err := BuildLink(
		left,
		right,
		testMetadata(),
		disabled,
	)

	if err != nil {
		t.Fatal(err)
	}

	if valid {
		t.Fatal(
			"expected link to be invalid when all source towers are disabled",
		)
	}
}

func TestDijkstraSkipsRouteWhenRelayRingIsBlocked(
	t *testing.T,
) {
	metadata := testMetadata()

	planets := map[string]domain.Planet{
		"A": {
			ID:                    "A",
			RadiusKM:              1,
			ActiveTowers:          8,
			RefractionIndex:       1,
			AtmosphereThicknessKM: 0,
		},
		"B": {
			ID:                    "B",
			RadiusKM:              1,
			ActiveTowers:          8,
			RefractionIndex:       1,
			AtmosphereThicknessKM: 0,
		},
		"C": {
			ID:                    "C",
			RadiusKM:              1,
			ActiveTowers:          8,
			RefractionIndex:       1,
			AtmosphereThicknessKM: 0,
		},
		"D": {
			ID:                    "D",
			RadiusKM:              1,
			ActiveTowers:          8,
			RefractionIndex:       1,
			AtmosphereThicknessKM: 0,
		},
	}

	// Candidate path A -> B -> C is blocked inside B.
	//
	// A -> B enters B at Tower 1.
	// B -> C exits B at Tower 4.
	// Broken towers 0 and 2 block both directions from 1 to 4.
	//
	// Alternative path A -> D -> C should be selected.
	ab := domain.Link{
		A:              "A",
		B:              "B",
		ATower:         0,
		BTower:         1,
		VoidDistanceKM: 1,
	}

	bc := domain.Link{
		A:              "B",
		B:              "C",
		ATower:         4,
		BTower:         0,
		VoidDistanceKM: 1,
	}

	ad := domain.Link{
		A:              "A",
		B:              "D",
		ATower:         0,
		BTower:         0,
		VoidDistanceKM: 2,
	}

	dc := domain.Link{
		A:              "D",
		B:              "C",
		ATower:         0,
		BTower:         0,
		VoidDistanceKM: 2,
	}

	g := &Graph{
		Metadata: metadata,
		Planets:  planets,
		Adjacency: map[string][]domain.Link{
			"A": {ab, ad},
			"B": {ab, bc},
			"C": {bc, dc},
			"D": {ad, dc},
		},
		Links: []domain.Link{
			ab,
			bc,
			ad,
			dc,
		},
		DisabledTowers: map[string]map[int]bool{
			"B": {
				0: true,
				2: true,
			},
		},
	}

	route, err := FindLowestLatencyRoute(g, "A", "C")
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"A", "D", "C"}

	if len(route.Path) != len(want) {
		t.Fatalf(
			"expected path %v, got %v",
			want,
			route.Path,
		)
	}

	for i := range want {
		if route.Path[i] != want[i] {
			t.Fatalf(
				"expected path %v, got %v",
				want,
				route.Path,
			)
		}
	}
}
