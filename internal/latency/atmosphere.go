package latency

import "github.com/launch26/relic-ring-protocol/internal/domain"

func AtmosphereSeconds(p domain.Planet, speedOfLightKMS float64) float64 {
	return p.AtmosphereThicknessKM * p.RefractionIndex / speedOfLightKMS
}
