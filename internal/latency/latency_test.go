package latency

import (
	"github.com/launch26/relic-ring-protocol/internal/domain"
	"math"
	"testing"
)

func TestSameTowerOneDelay(t *testing.T) {
	p := domain.Planet{ID: "A", RadiusKM: 100, ActiveTowers: 8}
	m := domain.UniverseMetadata{SpeedOfLightKMS: 300000, FiberSpeedFraction: .67, TowerProcessingDelayMS: 7}
	got := PlanetTransit(p, 2, 2, m)
	if got.Segments != 0 || got.DistinctTowers != 1 {
		t.Fatalf("dedup failed %+v", got)
	}
	if math.Abs(got.TotalSeconds-.007) > 1e-12 {
		t.Fatalf("expected .007, got %.12f", got.TotalSeconds)
	}
}
