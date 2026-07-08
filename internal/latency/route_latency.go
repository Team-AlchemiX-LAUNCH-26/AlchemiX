package latency

import "github.com/launch26/relic-ring-protocol/internal/domain"

func Aggregate(planets []domain.PlanetTransitBreakdown, voids []domain.VoidTransitBreakdown) domain.LatencyBreakdown {
	var out domain.LatencyBreakdown
	for _, p := range planets {
		out.FiberSeconds += p.FiberSeconds
		out.TowerSeconds += p.TowerSeconds
	}
	for _, v := range voids {
		out.AtmosphereSeconds += v.SourceAtmosphereSeconds + v.DestinationAtmosphereSeconds
		out.VoidSeconds += v.VacuumSeconds
	}
	out.TotalSeconds = out.FiberSeconds + out.TowerSeconds + out.AtmosphereSeconds + out.VoidSeconds
	return out
}
