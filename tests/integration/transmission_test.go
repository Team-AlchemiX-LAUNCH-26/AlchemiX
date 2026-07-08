package integration

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/launch26/relic-ring-protocol/internal/config"
	"github.com/launch26/relic-ring-protocol/internal/domain"
	packetops "github.com/launch26/relic-ring-protocol/internal/packet"
	"github.com/launch26/relic-ring-protocol/internal/planet"
	"github.com/launch26/relic-ring-protocol/internal/routing"
	"github.com/launch26/relic-ring-protocol/internal/transport"
	"github.com/launch26/relic-ring-protocol/pkg/protocol"
)

func TestDistributedThreeNodeTransmission(t *testing.T) {
	cfg, err := config.LoadUniverseConfig(filepath.Join("..", "..", "configs", "universe-config.json"))
	if err != nil {
		t.Fatal(err)
	}

	graph, err := routing.BuildGraph(cfg, map[string]bool{}, map[string]bool{}, map[string]map[int]bool{})
	if err != nil {
		t.Fatal(err)
	}
	route, err := routing.FindLowestLatencyRoute(graph, "Aegis", "Caelum")
	if err != nil {
		t.Fatal(err)
	}

	servers := map[string]*httptest.Server{}
	endpoints := map[string]string{}
	for _, id := range route.Path {
		p, err := config.FindPlanetByID(cfg, id)
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewServer(planet.NewService(cfg, p).Handler())
		servers[id] = server
		endpoints[id] = server.URL
		defer server.Close()
	}

	packet, err := packetops.New("Aegis", "Caelum", "Hello world", route.Path)
	if err != nil {
		t.Fatal(err)
	}
	req := protocol.NodePacketRequest{
		Packet:        packet,
		Route:         route,
		NodeEndpoints: endpoints,
	}

	client := transport.NewClient(5 * time.Second)
	var result domain.TransmissionResult
	if err := client.PostJSON(context.Background(), endpoints["Aegis"]+"/packet/receive", req, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.PacketDelivered {
		t.Fatalf("expected delivered, got %s", result.Status)
	}
	if result.DecodedPayload != "Hello world" {
		t.Fatalf("payload changed: %q", result.DecodedPayload)
	}
	if len(result.Packet.HopLog) != len(route.Path) {
		t.Fatalf("expected %d hop-log entries, got %d", len(route.Path), len(result.Packet.HopLog))
	}
}
