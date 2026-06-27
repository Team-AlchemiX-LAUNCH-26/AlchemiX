package planet

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/launch26/relic-ring-protocol/internal/config"
	"github.com/launch26/relic-ring-protocol/internal/domain"
	codec "github.com/launch26/relic-ring-protocol/internal/encoding"
	"github.com/launch26/relic-ring-protocol/internal/geometry"
	"github.com/launch26/relic-ring-protocol/internal/latency"
	packetops "github.com/launch26/relic-ring-protocol/internal/packet"
	"github.com/launch26/relic-ring-protocol/internal/transport"
	"github.com/launch26/relic-ring-protocol/pkg/protocol"
)

type Service struct {
	Config  domain.UniverseConfig
	Planet  domain.Planet
	Planets map[string]domain.Planet
	Client  *transport.Client
	active  atomic.Bool
	Metrics Metrics
}

func NewService(cfg domain.UniverseConfig, planet domain.Planet) *Service {
	s := &Service{Config: cfg, Planet: planet, Planets: config.PlanetMap(cfg), Client: transport.NewClient(8 * time.Second)}
	s.active.Store(true)
	return s
}

func (s *Service) Active() bool { return s.active.Load() }
func (s *Service) Disable()     { s.active.Store(false) }
func (s *Service) Enable()      { s.active.Store(true) }

func (s *Service) Process(ctx context.Context, req protocol.NodePacketRequest) (domain.TransmissionResult, error) {
	started := time.Now()
	defer func() {
		s.Metrics.mu.Lock()
		s.Metrics.LastProcessing = time.Since(started)
		s.Metrics.mu.Unlock()
	}()
	if !s.Active() {
		return domain.TransmissionResult{}, fmt.Errorf("planet %s is disabled", s.Planet.ID)
	}
	p := &req.Packet
	if p.RouteIndex < 0 || p.RouteIndex >= len(req.Route.Path) {
		return domain.TransmissionResult{}, fmt.Errorf("packet route index %d is invalid", p.RouteIndex)
	}
	if req.Route.Path[p.RouteIndex] != s.Planet.ID || p.CurrentID != s.Planet.ID {
		return domain.TransmissionResult{}, fmt.Errorf("packet expected at %s but reached %s", p.CurrentID, s.Planet.ID)
	}
	disabledTowers := disabledTowerSet(req.DisabledTowers)
	currentDisabledTowers := disabledTowersForPlanet(
		disabledTowers,
		s.Planet.ID,
	)

	localPayload := p.Payload
	if p.RouteIndex > 0 {
		decoded, _, err := codec.DecodeBinary(p.BinaryStream, s.Planet.Codex)
		if err != nil {
			return domain.TransmissionResult{}, fmt.Errorf("planet %s decode payload: %w", s.Planet.ID, err)
		}
		localPayload = decoded
	}
	if err := packetops.VerifyPayload(p.Payload, localPayload); err != nil {
		return domain.TransmissionResult{}, err
	}

	var previousID, nextID string
	if p.RouteIndex > 0 {
		previousID = req.Route.Path[p.RouteIndex-1]
	}
	if p.RouteIndex+1 < len(req.Route.Path) {
		nextID = req.Route.Path[p.RouteIndex+1]
	}

	entryTower, exitTower, err := s.resolveTowers(
		previousID,
		nextID,
		disabledTowers,
	)
	if err != nil {
		return domain.TransmissionResult{}, err
	}
	planetTransit, err := latency.PlanetTransitWithFailures(
		s.Planet,
		entryTower,
		exitTower,
		s.Config.Metadata,
		currentDisabledTowers,
	)

	if err != nil {
		return domain.TransmissionResult{}, fmt.Errorf(
			"calculate internal transit for planet %s: %w",
			s.Planet.ID,
			err,
		)
	}

	if err != nil {
		return domain.TransmissionResult{}, fmt.Errorf(
			"calculate internal transit for planet %s: %w",
			s.Planet.ID,
			err,
		)
	}

	entry := domain.HopLogEntry{
		PlanetID:         s.Planet.ID,
		PreviousPlanetID: previousID,
		NextPlanetID:     nextID,
		EntryTower:       entryTower,
		ExitTower:        exitTower,

		RingPath: append(
			[]int(nil),
			planetTransit.RingPath...,
		),

		RingDirection:  planetTransit.RingDirection,
		Segments:       planetTransit.Segments,
		DistinctTowers: planetTransit.DistinctTowers,
		LocalCodex:     s.Planet.Codex,
		DecodedPayload: localPayload,
		PlanetTransit:  planetTransit,
	}

	stepLatency := planetTransit.TotalSeconds
	if nextID != "" {
		nextPlanet, ok := s.Planets[nextID]
		if !ok {
			return domain.TransmissionResult{}, fmt.Errorf("next planet %s not found", nextID)
		}
		voidDistance, err := geometry.VoidDistanceKM(s.Planet, nextPlanet, s.Config.Metadata.CoordinateScaleUnitKM)
		if err != nil {
			return domain.TransmissionResult{}, err
		}
		if voidDistance > s.Config.Metadata.MaxVoidHopDistanceKM {
			return domain.TransmissionResult{}, fmt.Errorf("planned hop %s -> %s exceeds maximum void distance", s.Planet.ID, nextID)
		}
		pair, err := geometry.ClosestActiveTowerPair(
			s.Planet,
			nextPlanet,
			s.Config.Metadata.CoordinateScaleUnitKM,
			disabledTowers,
		)
		if err != nil {
			return domain.TransmissionResult{}, err
		}
		tokens, binary, err := codec.EncodePayload(localPayload, nextPlanet.Codex)
		if err != nil {
			return domain.TransmissionResult{}, err
		}
		voidTransit := latency.VoidTransit(s.Planet, nextPlanet, voidDistance, pair.FromTower.Index, pair.ToTower.Index, s.Config.Metadata.SpeedOfLightKMS)
		entry.NextHopCodex = nextPlanet.Codex
		entry.EncodedPayload = tokens
		entry.BinaryStream = binary
		entry.VoidTransit = &voidTransit
		stepLatency += voidTransit.TotalSeconds
		p.EncodedPayload = tokens
		p.BinaryStream = binary
	}

	entry.StepLatency = stepLatency
	entry.CumulativeLatency = p.CumulativeLatencySeconds + stepLatency
	packetops.AppendHop(p, entry)
	p.Status = domain.PacketInTransit

	s.Metrics.mu.Lock()
	s.Metrics.PacketsProcessed++
	s.Metrics.mu.Unlock()
	s.emit(ctx, req.TelemetryURL, protocol.Event{Type: "packet.hop.completed", Timestamp: time.Now().UTC(), PacketID: p.ID, PlanetID: s.Planet.ID, Data: entry})

	if nextID == "" {
		p.Status = domain.PacketDelivered
		p.DecodedPayload = localPayload
		result := domain.TransmissionResult{
			Status:              domain.PacketDelivered,
			Message:             "payload delivered successfully",
			Packet:              *p,
			Route:               req.Route,
			OriginalPayload:     p.Payload,
			DecodedPayload:      localPayload,
			TotalLatencySeconds: p.CumulativeLatencySeconds,
			Latency:             req.Route.Latency,
		}
		s.emit(ctx, req.TelemetryURL, protocol.Event{Type: "packet.delivered", Timestamp: time.Now().UTC(), PacketID: p.ID, PlanetID: s.Planet.ID, Data: result})
		return result, nil
	}

	endpoint := strings.TrimRight(req.NodeEndpoints[nextID], "/")
	if endpoint == "" {
		return domain.TransmissionResult{}, fmt.Errorf("no endpoint registered for planet %s", nextID)
	}
	p.CurrentID = nextID
	p.RouteIndex++
	s.Metrics.mu.Lock()
	s.Metrics.PacketsForwarded++
	s.Metrics.mu.Unlock()
	var result domain.TransmissionResult
	if err := s.Client.PostJSON(ctx, endpoint+"/packet/receive", req, &result); err != nil {
		s.Metrics.mu.Lock()
		s.Metrics.Failures++
		s.Metrics.mu.Unlock()
		return domain.TransmissionResult{}, fmt.Errorf("forward packet from %s to %s: %w", s.Planet.ID, nextID, err)
	}
	return result, nil
}

func (s *Service) resolveTowers(
	previousID string,
	nextID string,
	disabledTowers map[string]map[int]bool,
) (entry int, exit int, err error) {
	if previousID == "" && nextID == "" {
		return 0, 0, nil
	}

	if previousID == "" {
		pair, e := geometry.ClosestActiveTowerPair(
			s.Planet,
			s.Planets[nextID],
			s.Config.Metadata.CoordinateScaleUnitKM,
			disabledTowers,
		)

		if e != nil {
			return 0, 0, e
		}

		return pair.FromTower.Index, pair.FromTower.Index, nil
	}

	incoming, e := geometry.ClosestActiveTowerPair(
		s.Planets[previousID],
		s.Planet,
		s.Config.Metadata.CoordinateScaleUnitKM,
		disabledTowers,
	)

	if e != nil {
		return 0, 0, e
	}

	entry = incoming.ToTower.Index

	if nextID == "" {
		return entry, entry, nil
	}

	outgoing, e := geometry.ClosestActiveTowerPair(
		s.Planet,
		s.Planets[nextID],
		s.Config.Metadata.CoordinateScaleUnitKM,
		disabledTowers,
	)

	if e != nil {
		return 0, 0, e
	}

	return entry, outgoing.FromTower.Index, nil
}

func (s *Service) emit(ctx context.Context, telemetryURL string, event protocol.Event) {
	if strings.TrimSpace(telemetryURL) == "" {
		return
	}
	go func() {
		telemetryCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Client.PostJSON(telemetryCtx, strings.TrimRight(telemetryURL, "/")+"/api/events/ingest", event, nil)
	}()
}
func disabledTowerSet(
	input map[string][]int,
) map[string]map[int]bool {
	output := make(
		map[string]map[int]bool,
		len(input),
	)

	for planetID, towers := range input {
		if output[planetID] == nil {
			output[planetID] = make(map[int]bool)
		}

		for _, towerIndex := range towers {
			output[planetID][towerIndex] = true
		}
	}

	return output
}

func disabledTowersForPlanet(
	input map[string]map[int]bool,
	planetID string,
) map[int]bool {
	if input == nil {
		return nil
	}

	return input[planetID]
}
