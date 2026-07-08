package config

import (
	"fmt"
	"strings"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func Validate(cfg domain.UniverseConfig) error {
	m := cfg.Metadata
	if strings.TrimSpace(m.SystemName) == "" {
		return fmt.Errorf("configuration: universe_metadata.system_name is required")
	}
	if m.SpeedOfLightKMS <= 0 {
		return fmt.Errorf("configuration: speed_of_light_kms must be positive")
	}
	if m.MaxVoidHopDistanceKM <= 0 {
		return fmt.Errorf("configuration: max_void_hop_distance_km must be positive")
	}
	if m.CoordinateScaleUnitKM <= 0 {
		return fmt.Errorf("configuration: coordinate_scale_unit_km must be positive")
	}
	if m.TowerProcessingDelayMS < 0 {
		return fmt.Errorf("configuration: tower_processing_delay_ms cannot be negative")
	}
	if m.FiberSpeedFraction <= 0 || m.FiberSpeedFraction > 1 {
		return fmt.Errorf("configuration: fiber_speed_fraction must be within (0, 1]")
	}
	if len(cfg.Nodes) < 2 {
		return fmt.Errorf("configuration: at least two planets are required")
	}

	seen := make(map[string]struct{}, len(cfg.Nodes))
	for i, p := range cfg.Nodes {
		prefix := fmt.Sprintf("configuration: nodes[%d]", i)
		if strings.TrimSpace(p.ID) == "" {
			return fmt.Errorf("%s.id is required", prefix)
		}
		if _, ok := seen[p.ID]; ok {
			return fmt.Errorf("configuration: duplicate planet id %q", p.ID)
		}
		seen[p.ID] = struct{}{}
		if p.Codex < 2 || p.Codex > 36 {
			return fmt.Errorf("%s (%s): codex must be between 2 and 36", prefix, p.ID)
		}
		if p.RadiusKM <= 0 {
			return fmt.Errorf("%s (%s): radius_km must be positive", prefix, p.ID)
		}
		if p.ActiveTowers < 4 {
			return fmt.Errorf("%s (%s): active_towers must be at least 4", prefix, p.ID)
		}
		if p.AtmosphereThicknessKM < 0 {
			return fmt.Errorf("%s (%s): atmosphere_thickness_km cannot be negative", prefix, p.ID)
		}
		if p.RefractionIndex <= 0 {
			return fmt.Errorf("%s (%s): refraction_index must be positive", prefix, p.ID)
		}
	}
	return nil
}
