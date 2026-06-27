package geometry

import (
	"fmt"
	"math"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func CenterDistanceKM(a, b domain.Planet, scale float64) float64 {
	return math.Hypot((b.X-a.X)*scale, (b.Y-a.Y)*scale)
}

func VoidDistanceKM(a, b domain.Planet, scale float64) (float64, error) {
	center := CenterDistanceKM(a, b, scale)
	void := center - (a.RadiusKM + a.AtmosphereThicknessKM) - (b.RadiusKM + b.AtmosphereThicknessKM)
	if void < 0 && void > -1e-7 {
		return 0, nil
	}
	if void < 0 {
		return 0, fmt.Errorf("planets %s and %s overlap after radius and atmosphere subtraction", a.ID, b.ID)
	}
	return void, nil
}
