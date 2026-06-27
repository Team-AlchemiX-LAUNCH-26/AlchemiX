package geometry

import (
	"reflect"
	"testing"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func TestClosestActiveTowerPairUsesNormalPairWhenAllActive(
	t *testing.T,
) {
	left := domain.Planet{
		ID:           "Left",
		X:            0,
		Y:            0,
		RadiusKM:     1,
		ActiveTowers: 4,
	}

	right := domain.Planet{
		ID:           "Right",
		X:            10,
		Y:            0,
		RadiusKM:     1,
		ActiveTowers: 4,
	}

	pair, err := ClosestActiveTowerPair(
		left,
		right,
		1,
		nil,
	)

	if err != nil {
		t.Fatal(err)
	}

	// For a planet on the left, Tower 1 points right.
	// For a planet on the right, Tower 3 points left.
	if pair.FromTower.Index != 1 {
		t.Fatalf(
			"expected Left Tower 1, got Tower %d",
			pair.FromTower.Index,
		)
	}

	if pair.ToTower.Index != 3 {
		t.Fatalf(
			"expected Right Tower 3, got Tower %d",
			pair.ToTower.Index,
		)
	}
}

func TestClosestActiveTowerPairIgnoresDisabledTower(
	t *testing.T,
) {
	left := domain.Planet{
		ID:           "Left",
		X:            0,
		Y:            0,
		RadiusKM:     1,
		ActiveTowers: 4,
	}

	right := domain.Planet{
		ID:           "Right",
		X:            10,
		Y:            0,
		RadiusKM:     1,
		ActiveTowers: 4,
	}

	disabled := map[string]map[int]bool{
		"Left": {
			1: true,
		},
	}

	pair, err := ClosestActiveTowerPair(
		left,
		right,
		1,
		disabled,
	)

	if err != nil {
		t.Fatal(err)
	}

	if pair.FromTower.Index == 1 {
		t.Fatal(
			"disabled Left Tower 1 must not be selected",
		)
	}

	// Towers 0 and 2 are equally close after Tower 1 fails.
	// Deterministic tie-breaking selects the lower index.
	if pair.FromTower.Index != 0 {
		t.Fatalf(
			"expected fallback Left Tower 0, got Tower %d",
			pair.FromTower.Index,
		)
	}

	if pair.ToTower.Index != 3 {
		t.Fatalf(
			"expected Right Tower 3, got Tower %d",
			pair.ToTower.Index,
		)
	}
}

func TestClosestActiveTowerPairFailsWhenAllTowersDisabled(
	t *testing.T,
) {
	left := domain.Planet{
		ID:           "Left",
		RadiusKM:     1,
		ActiveTowers: 4,
	}

	right := domain.Planet{
		ID:           "Right",
		X:            10,
		RadiusKM:     1,
		ActiveTowers: 4,
	}

	disabled := map[string]map[int]bool{
		"Left": {
			0: true,
			1: true,
			2: true,
			3: true,
		},
	}

	_, err := ClosestActiveTowerPair(
		left,
		right,
		1,
		disabled,
	)

	if err == nil {
		t.Fatal(
			"expected an error when all Left towers are disabled",
		)
	}
}

func TestOperationalRingPathWithoutFailures(
	t *testing.T,
) {
	got, err := ShortestOperationalRingPath(
		1,
		4,
		8,
		nil,
	)

	if err != nil {
		t.Fatal(err)
	}

	wantTowers := []int{1, 2, 3, 4}

	if !reflect.DeepEqual(got.Towers, wantTowers) {
		t.Fatalf(
			"expected path %v, got %v",
			wantTowers,
			got.Towers,
		)
	}

	if got.Segments != 3 {
		t.Fatalf(
			"expected 3 segments, got %d",
			got.Segments,
		)
	}

	if got.DistinctTowers != 4 {
		t.Fatalf(
			"expected 4 distinct towers, got %d",
			got.DistinctTowers,
		)
	}

	if got.Direction != domain.RingDirectionClockwise {
		t.Fatalf(
			"expected clockwise, got %s",
			got.Direction,
		)
	}
}

func TestOperationalRingPathAvoidsBrokenTower(
	t *testing.T,
) {
	disabled := map[int]bool{
		2: true,
	}

	got, err := ShortestOperationalRingPath(
		1,
		4,
		8,
		disabled,
	)

	if err != nil {
		t.Fatal(err)
	}

	wantTowers := []int{1, 0, 7, 6, 5, 4}

	if !reflect.DeepEqual(got.Towers, wantTowers) {
		t.Fatalf(
			"expected path %v, got %v",
			wantTowers,
			got.Towers,
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

	if got.Direction !=
		domain.RingDirectionCounterClockwise {
		t.Fatalf(
			"expected counter-clockwise, got %s",
			got.Direction,
		)
	}
}

func TestOperationalRingPathFailsWhenBothDirectionsBlocked(
	t *testing.T,
) {
	disabled := map[int]bool{
		0: true,
		2: true,
	}

	_, err := ShortestOperationalRingPath(
		1,
		4,
		8,
		disabled,
	)

	if err == nil {
		t.Fatal(
			"expected no operational ring path",
		)
	}
}

func TestOperationalRingPathRejectsDisabledEndpoint(
	t *testing.T,
) {
	disabled := map[int]bool{
		1: true,
	}

	_, err := ShortestOperationalRingPath(
		1,
		4,
		8,
		disabled,
	)

	if err == nil {
		t.Fatal(
			"expected disabled entry tower to be rejected",
		)
	}
}

func TestOperationalRingPathSameTower(
	t *testing.T,
) {
	got, err := ShortestOperationalRingPath(
		3,
		3,
		8,
		nil,
	)

	if err != nil {
		t.Fatal(err)
	}

	wantTowers := []int{3}

	if !reflect.DeepEqual(got.Towers, wantTowers) {
		t.Fatalf(
			"expected %v, got %v",
			wantTowers,
			got.Towers,
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

	if got.Direction != domain.RingDirectionStationary {
		t.Fatalf(
			"expected stationary, got %s",
			got.Direction,
		)
	}
}

func TestEqualLengthRingPathsChooseClockwise(
	t *testing.T,
) {
	got, err := ShortestOperationalRingPath(
		0,
		4,
		8,
		nil,
	)

	if err != nil {
		t.Fatal(err)
	}

	want := []int{0, 1, 2, 3, 4}

	if !reflect.DeepEqual(got.Towers, want) {
		t.Fatalf(
			"expected deterministic clockwise path %v, got %v",
			want,
			got.Towers,
		)
	}
}
