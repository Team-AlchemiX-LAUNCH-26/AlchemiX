package geometry

import (
	"github.com/launch26/relic-ring-protocol/internal/domain"
	"math"
	"testing"
)

func TestTowerZeroAtTop(t *testing.T) {
	p := domain.Planet{ID: "A", RadiusKM: 10, ActiveTowers: 8}
	towers := GenerateTowers(p, 1)
	if math.Abs(towers[0].XKM) > 1e-9 || math.Abs(towers[0].YKM-10) > 1e-9 {
		t.Fatalf("tower 0 not at top: %+v", towers[0])
	}
	if towers[1].AngleDegrees != 45 {
		t.Fatalf("expected 45 degree spacing")
	}
}
func TestRingSegments(t *testing.T) {
	s, m := ShortestRingSegments(7, 1, 8)
	if s != 2 || m != 3 {
		t.Fatalf("expected 2 segments/3 towers, got %d/%d", s, m)
	}
	s, m = ShortestRingSegments(3, 3, 8)
	if s != 0 || m != 1 {
		t.Fatalf("same tower dedup failed")
	}
}
