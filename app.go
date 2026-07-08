package main

import (
	"context"
	"os"
	"time"

	"github.com/launch26/relic-ring-protocol/ai"
	"github.com/launch26/relic-ring-protocol/internal/domain"
	"github.com/launch26/relic-ring-protocol/internal/orchestrator"
	"github.com/launch26/relic-ring-protocol/pkg/protocol"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	client *orchestrator.RemoteClient
	cancel context.CancelFunc
}

func NewApp() *App {
	base := os.Getenv("ORCHESTRATOR_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	client := orchestrator.NewRemoteClient(base)
	return &App{client: client}
}
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	streamCtx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	go func() {
		for {
			err := a.client.StreamEvents(streamCtx, func(event protocol.Event) { runtime.EventsEmit(ctx, "network:event", event) })
			if err != nil {
				runtime.EventsEmit(ctx, "network:event", protocol.Event{Type: "event.stream.error", Timestamp: time.Now().UTC(), Message: err.Error()})
			}
			select {
			case <-streamCtx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}()
}
func (a *App) shutdown(context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
}
func (a *App) GetUniverse() (protocol.UniverseResponse, error) { return a.client.Universe(a.ctx) }
func (a *App) StartTransmission(req protocol.TransmissionRequest) (domain.TransmissionResult, error) {
	return a.client.Transmit(a.ctx, req)
}
func (a *App) DisableNode(id string) error {
	return a.client.Action(a.ctx, "/api/nodes/"+id+"/disable", map[string]any{})
}
func (a *App) EnableNode(id string) error {
	return a.client.Action(a.ctx, "/api/nodes/"+id+"/enable", map[string]any{})
}
func (a *App) DisableLink(aID, bID string) error {
	return a.client.Action(a.ctx, "/api/links/disable", protocol.LinkRequest{A: aID, B: bID})
}
func (a *App) EnableLink(aID, bID string) error {
	return a.client.Action(a.ctx, "/api/links/enable", protocol.LinkRequest{A: aID, B: bID})
}
func (a *App) DisableTower(
	planetID string,
	towerIndex int,
) error {
	return a.client.Action(
		a.ctx,
		"/api/towers/disable",
		protocol.TowerRequest{
			PlanetID:   planetID,
			TowerIndex: towerIndex,
		},
	)
}

func (a *App) EnableTower(
	planetID string,
	towerIndex int,
) error {
	return a.client.Action(
		a.ctx,
		"/api/towers/enable",
		protocol.TowerRequest{
			PlanetID:   planetID,
			TowerIndex: towerIndex,
		},
	)
}
func (a *App) ResetNetwork() error {
	return a.client.Action(a.ctx, "/api/network/reset", map[string]any{})
}
