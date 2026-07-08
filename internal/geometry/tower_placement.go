package geometry

import (
	"math"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func GenerateTowers(p domain.Planet, scale float64) []domain.Tower {
	centerX := ScaleCoordinate(p.X, scale)
	centerY := ScaleCoordinate(p.Y, scale)
	towers := make([]domain.Tower, p.ActiveTowers)
	for i := 0; i < p.ActiveTowers; i++ {
		angle := 2 * math.Pi * float64(i) / float64(p.ActiveTowers)
		towers[i] = domain.Tower{
			Index:        i,
			AngleDegrees: 360 * float64(i) / float64(p.ActiveTowers),
			XKM:          centerX + p.RadiusKM*math.Sin(angle),
			YKM:          centerY + p.RadiusKM*math.Cos(angle),
		}
	}
	return towers
}
