package routing

import (
	"github.com/launch26/relic-ring-protocol/internal/config"
	"path/filepath"
	"testing"
)

func TestAegisToCaelumRoute(t *testing.T) {
	cfg, err := config.LoadUniverseConfig(filepath.Join("..", "..", "configs", "universe-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := BuildGraph(
		cfg,
		map[string]bool{},
		map[string]bool{},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	route, err := FindLowestLatencyRoute(g, "Aegis", "Caelum")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"Aegis", "Dawn", "Caelum"}
	if len(route.Path) != len(expected) {
		t.Fatalf("unexpected path %v", route.Path)
	}
	for i := range expected {
		if route.Path[i] != expected[i] {
			t.Fatalf("expected %v got %v", expected, route.Path)
		}
	}
	if route.Latency.TotalSeconds <= 0 {
		t.Fatal("latency must be positive")
	}
}
func TestDawnFailureReroutes(t *testing.T) {
	cfg, _ := config.LoadUniverseConfig(filepath.Join("..", "..", "configs", "universe-config.json"))
	g, err := BuildGraph(cfg, map[string]bool{"Dawn": true}, map[string]bool{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	route, err := FindLowestLatencyRoute(g, "Aegis", "Caelum")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range route.Path {
		if id == "Dawn" {
			t.Fatalf("failed node appears in route %v", route.Path)
		}
	}
}
