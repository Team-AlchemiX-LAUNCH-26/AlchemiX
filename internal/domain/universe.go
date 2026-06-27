package domain

type UniverseMetadata struct {
	SystemName             string  `json:"system_name"`
	SpeedOfLightKMS        float64 `json:"speed_of_light_kms"`
	MaxVoidHopDistanceKM   float64 `json:"max_void_hop_distance_km"`
	CoordinateScaleUnitKM  float64 `json:"coordinate_scale_unit_km"`
	TowerProcessingDelayMS float64 `json:"tower_processing_delay_ms"`
	FiberSpeedFraction     float64 `json:"fiber_speed_fraction"`
}

type UniverseConfig struct {
	Metadata UniverseMetadata `json:"universe_metadata"`
	Nodes    []Planet         `json:"nodes"`
}
