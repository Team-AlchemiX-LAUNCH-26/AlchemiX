package latency

import "github.com/launch26/relic-ring-protocol/internal/domain"

func VacuumSeconds(voidDistanceKM, speedOfLightKMS float64) float64 {
	return voidDistanceKM / speedOfLightKMS
}

func VoidTransit(a, b domain.Planet, voidDistanceKM float64, sendingTower, receivingTower int, speedOfLightKMS float64) domain.VoidTransitBreakdown {
	sourceAtmosphere := AtmosphereSeconds(a, speedOfLightKMS)
	destinationAtmosphere := AtmosphereSeconds(b, speedOfLightKMS)
	vacuum := VacuumSeconds(voidDistanceKM, speedOfLightKMS)
	return domain.VoidTransitBreakdown{
		FromID:                       a.ID,
		ToID:                         b.ID,
		SendingTower:                 sendingTower,
		ReceivingTower:               receivingTower,
		VoidDistanceKM:               voidDistanceKM,
		SourceAtmosphereSeconds:      sourceAtmosphere,
		VacuumSeconds:                vacuum,
		DestinationAtmosphereSeconds: destinationAtmosphere,
		TotalSeconds:                 sourceAtmosphere + vacuum + destinationAtmosphere,
	}
}
