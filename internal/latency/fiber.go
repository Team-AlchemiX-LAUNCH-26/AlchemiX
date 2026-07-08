package latency

import "math"

func FiberSeconds(radiusKM float64, towerCount, segments int, fiberFraction, speedOfLightKMS float64) float64 {
	if segments == 0 {
		return 0
	}
	arcDistance := 2 * math.Pi * radiusKM * float64(segments) / float64(towerCount)
	return arcDistance / (fiberFraction * speedOfLightKMS)
}
