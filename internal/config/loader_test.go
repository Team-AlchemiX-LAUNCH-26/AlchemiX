package config

import (
	"path/filepath"
	"testing"
)

func TestLoadProvidedUniverse(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "universe-config.json")
	cfg, err := LoadUniverseConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Metadata.SystemName != "Zeta-26" {
		t.Fatalf("expected Zeta-26, got %s", cfg.Metadata.SystemName)
	}
	if len(cfg.Nodes) != 6 {
		t.Fatalf("expected 6 nodes, got %d", len(cfg.Nodes))
	}
}
