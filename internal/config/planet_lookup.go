package config

import (
	"fmt"
	"strings"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func FindPlanetByID(cfg domain.UniverseConfig, id string) (domain.Planet, error) {
	for _, p := range cfg.Nodes {
		if strings.EqualFold(p.ID, id) {
			return p, nil
		}
	}
	return domain.Planet{}, fmt.Errorf("planet %q not found", id)
}

func PlanetMap(cfg domain.UniverseConfig) map[string]domain.Planet {
	out := make(map[string]domain.Planet, len(cfg.Nodes))
	for _, p := range cfg.Nodes {
		out[p.ID] = p
	}
	return out
}
