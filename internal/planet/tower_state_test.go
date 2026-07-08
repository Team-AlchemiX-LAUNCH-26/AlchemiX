package planet

import (
	"testing"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func TestDisabledTowerSet(t *testing.T) {
	input := map[string][]int{
		"Aegis": {2, 5},
		"Dawn":  {1},
	}

	got := disabledTowerSet(input)

	if !got["Aegis"][2] {
		t.Fatal("expected Aegis tower 2 to be disabled")
	}

	if !got["Aegis"][5] {
		t.Fatal("expected Aegis tower 5 to be disabled")
	}

	if !got["Dawn"][1] {
		t.Fatal("expected Dawn tower 1 to be disabled")
	}

	if got["Dawn"][2] {
		t.Fatal("did not expect Dawn tower 2 to be disabled")
	}
}

func TestDisabledTowersForPlanet(t *testing.T) {
	input := map[string]map[int]bool{
		"Aegis": {
			2: true,
		},
	}

	got := disabledTowersForPlanet(input, "Aegis")

	if !got[2] {
		t.Fatal("expected Aegis tower 2 to be disabled")
	}

	missing := disabledTowersForPlanet(input, "Dawn")

	if missing != nil {
		t.Fatalf("expected nil for Dawn, got %#v", missing)
	}
}
func TestResolveTowersAvoidsDisabledTower(t *testing.T) {
	cfg := domain.UniverseConfig{
		Metadata: domain.UniverseMetadata{
			CoordinateScaleUnitKM: 1,
		},
		Nodes: []domain.Planet{
			{
				ID:           "Left",
				X:            0,
				Y:            0,
				RadiusKM:     1,
				ActiveTowers: 4,
			},
			{
				ID:           "Right",
				X:            10,
				Y:            0,
				RadiusKM:     1,
				ActiveTowers: 4,
			},
		},
	}

	service := NewService(
		cfg,
		cfg.Nodes[0],
	)

	disabled := map[string]map[int]bool{
		"Left": {
			1: true,
		},
	}

	entry, exit, err := service.resolveTowers(
		"",
		"Right",
		disabled,
	)

	if err != nil {
		t.Fatal(err)
	}

	if entry == 1 || exit == 1 {
		t.Fatal("disabled tower 1 must not be selected")
	}

	if entry != exit {
		t.Fatalf(
			"source planet should use same entry and exit tower, got %d and %d",
			entry,
			exit,
		)
	}
}
