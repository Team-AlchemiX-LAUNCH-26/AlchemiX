package orchestrator

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/launch26/relic-ring-protocol/internal/config"
	"github.com/launch26/relic-ring-protocol/internal/domain"
	packetops "github.com/launch26/relic-ring-protocol/internal/packet"
	"github.com/launch26/relic-ring-protocol/internal/resilience"
	"github.com/launch26/relic-ring-protocol/internal/routing"
	"github.com/launch26/relic-ring-protocol/internal/transport"
	"github.com/launch26/relic-ring-protocol/pkg/protocol"
)

type Service struct {
	Config        domain.UniverseConfig
	Endpoints     map[string]string
	State         *resilience.State
	Client        *transport.Client
	Events        *transport.EventHub
	PublicBaseURL string
}

func New(configPath, publicBaseURL string, failureThreshold int) (*Service, error) {
	cfg, err := config.LoadUniverseConfig(configPath)
	if err != nil {
		return nil, err
	}
	endpoints, err := ResolveNodeEndpoints(cfg)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(cfg.Nodes))
	for i, p := range cfg.Nodes {
		ids[i] = p.ID
	}
	return &Service{
		Config:        cfg,
		Endpoints:     endpoints,
		State:         resilience.NewState(ids, failureThreshold),
		Client:        transport.NewClient(4 * time.Second),
		Events:        transport.NewEventHub(),
		PublicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}, nil
}

func (s *Service) RefreshHealth(ctx context.Context) {
	for id, endpoint := range s.Endpoints {
		id, endpoint := id, endpoint
		go func() {
			checkCtx, cancel := context.WithTimeout(ctx, 1200*time.Millisecond)
			defer cancel()
			var response struct {
				Active bool `json:"active"`
			}
			err := s.Client.GetJSON(checkCtx, strings.TrimRight(endpoint, "/")+"/health", &response)
			healthy := err == nil && response.Active
			if s.State.RecordHealth(id, healthy) {
				typeName := "node.failed"
				if healthy {
					typeName = "node.restored"
				}
				s.Events.Publish(protocol.Event{Type: typeName, Timestamp: time.Now().UTC(), PlanetID: id})
			}
		}()
	}
}

func (s *Service) StartHealthMonitor(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.RefreshHealth(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.RefreshHealth(ctx)
		}
	}
}

func (s *Service) graph() (*routing.Graph, error) {
	healthy, manual, links := s.State.Snapshot()
	unavailable := make(map[string]bool)
	for _, p := range s.Config.Nodes {
		if !healthy[p.ID] || manual[p.ID] {
			unavailable[p.ID] = true
		}
	}
	return routing.BuildGraph(
		s.Config,
		unavailable,
		links,
		s.State.DisabledTowerSetSnapshot(),
	)
}

func (s *Service) Snapshot() (domain.NetworkSnapshot, error) {
	g, err := s.graph()
	if err != nil {
		return domain.NetworkSnapshot{}, err
	}
	healthy, manual, links := s.State.Snapshot()
	statuses := make([]domain.PlanetStatus, 0, len(s.Config.Nodes))
	for _, p := range s.Config.Nodes {
		statuses = append(statuses, domain.PlanetStatus{ID: p.ID, Endpoint: s.Endpoints[p.ID], Healthy: healthy[p.ID], ManuallyDisabled: manual[p.ID], Available: healthy[p.ID] && !manual[p.ID]})
	}
	disabledLinks := make([]string, 0, len(links))
	for id, disabled := range links {
		if disabled {
			disabledLinks = append(disabledLinks, id)
		}
	}
	sort.Strings(disabledLinks)
	return domain.NetworkSnapshot{
		Metadata:       s.Config.Metadata,
		Planets:        s.Config.Nodes,
		PlanetStatus:   statuses,
		Links:          g.Links,
		DisabledLinks:  disabledLinks,
		DisabledTowers: s.State.DisabledTowersSnapshot(),
	}, nil
}

func (s *Service) StartTransmission(ctx context.Context, req protocol.TransmissionRequest) (domain.TransmissionResult, error) {
	if strings.TrimSpace(req.Payload) == "" {
		return domain.TransmissionResult{}, fmt.Errorf("payload cannot be empty")
	}
	// Perform a synchronous health pass so route decisions use current service status.
	s.refreshHealthSynchronously(ctx)
	g, err := s.graph()
	if err != nil {
		return domain.TransmissionResult{}, err
	}
	route, err := routing.FindLowestLatencyRoute(g, req.OriginID, req.DestinationID)
	if err != nil {
		return domain.TransmissionResult{}, err
	}
	p, err := packetops.New(req.OriginID, req.DestinationID, req.Payload, route.Path)
	if err != nil {
		return domain.TransmissionResult{}, err
	}
	s.Events.Publish(protocol.Event{Type: "route.calculated", Timestamp: time.Now().UTC(), PacketID: p.ID, Data: route})
	s.Events.Publish(protocol.Event{Type: "packet.created", Timestamp: time.Now().UTC(), PacketID: p.ID, Data: p})
	endpoint := strings.TrimRight(s.Endpoints[req.OriginID], "/")
	if endpoint == "" {
		return domain.TransmissionResult{}, fmt.Errorf("no endpoint registered for source %s", req.OriginID)
	}
	nodeReq := protocol.NodePacketRequest{
		Packet:         p,
		Route:          route,
		NodeEndpoints:  s.Endpoints,
		TelemetryURL:   s.PublicBaseURL,
		DisabledTowers: s.State.DisabledTowersSnapshot(),
	}
	var result domain.TransmissionResult
	if err := s.Client.PostJSON(ctx, endpoint+"/packet/receive", nodeReq, &result); err != nil {
		s.Events.Publish(protocol.Event{Type: "packet.failed", Timestamp: time.Now().UTC(), PacketID: p.ID, Message: err.Error()})
		return domain.TransmissionResult{}, err
	}
	// Use the route engine's aggregate as authoritative and validate node total within tolerance.
	if diff := result.TotalLatencySeconds - route.Latency.TotalSeconds; diff > 1e-6 || diff < -1e-6 {
		return domain.TransmissionResult{}, fmt.Errorf("node latency total %.9f does not match route engine %.9f", result.TotalLatencySeconds, route.Latency.TotalSeconds)
	}
	result.Route = route
	result.Latency = route.Latency
	return result, nil
}

func (s *Service) refreshHealthSynchronously(ctx context.Context) {
	for id, endpoint := range s.Endpoints {
		checkCtx, cancel := context.WithTimeout(ctx, 1200*time.Millisecond)
		var response struct {
			Active bool `json:"active"`
		}
		err := s.Client.GetJSON(checkCtx, strings.TrimRight(endpoint, "/")+"/health", &response)
		cancel()
		s.State.SetHealth(id, err == nil && response.Active)
	}
}

func (s *Service) DisableNode(ctx context.Context, id string) error {
	if _, err := config.FindPlanetByID(s.Config, id); err != nil {
		return err
	}
	s.State.DisableNode(id)
	endpoint := strings.TrimRight(s.Endpoints[id], "/")
	_ = s.Client.PostJSON(ctx, endpoint+"/admin/disable", map[string]any{}, nil)
	s.Events.Publish(protocol.Event{Type: "node.failed", Timestamp: time.Now().UTC(), PlanetID: id, Message: "manually disabled"})
	return nil
}

func (s *Service) EnableNode(ctx context.Context, id string) error {
	if _, err := config.FindPlanetByID(s.Config, id); err != nil {
		return err
	}
	endpoint := strings.TrimRight(s.Endpoints[id], "/")
	if err := s.Client.PostJSON(ctx, endpoint+"/admin/enable", map[string]any{}, nil); err != nil {
		return err
	}
	s.State.EnableNode(id)
	s.State.SetHealth(id, true)
	s.Events.Publish(protocol.Event{Type: "node.restored", Timestamp: time.Now().UTC(), PlanetID: id})
	return nil
}

func (s *Service) DisableLink(a, b string) error {
	if a == b {
		return fmt.Errorf("link endpoints must be different")
	}
	if _, err := config.FindPlanetByID(s.Config, a); err != nil {
		return err
	}
	if _, err := config.FindPlanetByID(s.Config, b); err != nil {
		return err
	}
	id := routing.LinkID(a, b)
	s.State.DisableLink(id)
	s.Events.Publish(protocol.Event{Type: "link.failed", Timestamp: time.Now().UTC(), Data: map[string]string{"id": id, "a": a, "b": b}})
	return nil
}

func (s *Service) EnableLink(a, b string) error {
	id := routing.LinkID(a, b)
	s.State.EnableLink(id)
	s.Events.Publish(protocol.Event{Type: "link.restored", Timestamp: time.Now().UTC(), Data: map[string]string{"id": id, "a": a, "b": b}})
	return nil
}
func (s *Service) DisableTower(
	planetID string,
	towerIndex int,
) error {
	planet, err := config.FindPlanetByID(
		s.Config,
		planetID,
	)

	if err != nil {
		return err
	}

	if towerIndex < 0 ||
		towerIndex >= planet.ActiveTowers {
		return fmt.Errorf(
			"tower index %d is invalid for planet %s; valid range is 0-%d",
			towerIndex,
			planet.ID,
			planet.ActiveTowers-1,
		)
	}

	// Use the canonical ID from the configuration.
	s.State.DisableTower(
		planet.ID,
		towerIndex,
	)

	if s.Events != nil {
		s.Events.Publish(protocol.Event{
			Type:      "tower.failed",
			Timestamp: time.Now().UTC(),
			PlanetID:  planet.ID,
			Message: fmt.Sprintf(
				"tower %d manually disabled",
				towerIndex,
			),
			Data: map[string]any{
				"planet_id":   planet.ID,
				"tower_index": towerIndex,
				"tower_count": planet.ActiveTowers,
				"disabled":    true,
			},
		})
	}

	return nil
}

func (s *Service) EnableTower(
	planetID string,
	towerIndex int,
) error {
	planet, err := config.FindPlanetByID(
		s.Config,
		planetID,
	)

	if err != nil {
		return err
	}

	if towerIndex < 0 ||
		towerIndex >= planet.ActiveTowers {
		return fmt.Errorf(
			"tower index %d is invalid for planet %s; valid range is 0-%d",
			towerIndex,
			planet.ID,
			planet.ActiveTowers-1,
		)
	}

	s.State.EnableTower(
		planet.ID,
		towerIndex,
	)

	if s.Events != nil {
		s.Events.Publish(protocol.Event{
			Type:      "tower.restored",
			Timestamp: time.Now().UTC(),
			PlanetID:  planet.ID,
			Message: fmt.Sprintf(
				"tower %d restored",
				towerIndex,
			),
			Data: map[string]any{
				"planet_id":   planet.ID,
				"tower_index": towerIndex,
				"tower_count": planet.ActiveTowers,
				"disabled":    false,
			},
		})
	}

	return nil
}
func (s *Service) Reset(ctx context.Context) {
	s.State.Reset()
	for id, endpoint := range s.Endpoints {
		_ = s.Client.PostJSON(ctx, strings.TrimRight(endpoint, "/")+"/admin/enable", map[string]any{}, nil)
		s.State.SetHealth(id, true)
	}
	s.Events.Publish(protocol.Event{Type: "network.reset", Timestamp: time.Now().UTC()})
}
