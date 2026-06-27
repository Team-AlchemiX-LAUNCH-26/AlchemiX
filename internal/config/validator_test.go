package config

import (
	"github.com/launch26/relic-ring-protocol/internal/domain"
	"testing"
)

func validConfig() domain.UniverseConfig {
	return domain.UniverseConfig{Metadata: domain.UniverseMetadata{SystemName: "x", SpeedOfLightKMS: 1, MaxVoidHopDistanceKM: 1, CoordinateScaleUnitKM: 1, FiberSpeedFraction: .5}, Nodes: []domain.Planet{{ID: "A", Codex: 2, RadiusKM: 1, ActiveTowers: 4, RefractionIndex: 1}, {ID: "B", Codex: 2, RadiusKM: 1, ActiveTowers: 4, RefractionIndex: 1}}}
}
func TestDuplicatePlanetRejected(t *testing.T) {
	cfg := validConfig()
	cfg.Nodes[1].ID = "A"
	if Validate(cfg) == nil {
		t.Fatal("expected duplicate ID error")
	}
}
func TestTooFewTowersRejected(t *testing.T) {
	cfg := validConfig()
	cfg.Nodes[0].ActiveTowers = 3
	if Validate(cfg) == nil {
		t.Fatal("expected tower count error")
	}
}
