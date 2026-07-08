package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func LoadUniverseConfig(path string) (domain.UniverseConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.UniverseConfig{}, fmt.Errorf("read universe config %q: %w", path, err)
	}

	var cfg domain.UniverseConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return domain.UniverseConfig{}, fmt.Errorf("decode universe config %q: %w", path, err)
	}

	ApplyDocumentedDefaults(&cfg)
	if err := Validate(cfg); err != nil {
		return domain.UniverseConfig{}, err
	}
	return cfg, nil
}
