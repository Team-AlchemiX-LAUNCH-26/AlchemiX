package config

import "github.com/launch26/relic-ring-protocol/internal/domain"

const (
	DefaultSpeedOfLightKMS        = 300000.0
	DefaultMaxVoidHopDistanceKM   = 50000000.0
	DefaultTowerProcessingDelayMS = 7.0
	DefaultFiberSpeedFraction     = 0.67
)

func ApplyDocumentedDefaults(cfg *domain.UniverseConfig) {
	if cfg.Metadata.SpeedOfLightKMS == 0 {
		cfg.Metadata.SpeedOfLightKMS = DefaultSpeedOfLightKMS
	}
	if cfg.Metadata.MaxVoidHopDistanceKM == 0 {
		cfg.Metadata.MaxVoidHopDistanceKM = DefaultMaxVoidHopDistanceKM
	}
	if cfg.Metadata.TowerProcessingDelayMS == 0 {
		cfg.Metadata.TowerProcessingDelayMS = DefaultTowerProcessingDelayMS
	}
	if cfg.Metadata.FiberSpeedFraction == 0 {
		cfg.Metadata.FiberSpeedFraction = DefaultFiberSpeedFraction
	}
}
