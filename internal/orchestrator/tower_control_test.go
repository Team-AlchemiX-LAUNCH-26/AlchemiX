package orchestrator

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/launch26/relic-ring-protocol/internal/domain"
	"github.com/launch26/relic-ring-protocol/internal/resilience"
	"github.com/launch26/relic-ring-protocol/internal/transport"
)

func newTowerControlTestService() *Service {
	cfg := domain.UniverseConfig{
		Nodes: []domain.Planet{
			{
				ID:           "Aegis",
				ActiveTowers: 8,
			},
			{
				ID:           "Dawn",
				ActiveTowers: 6,
			},
		},
	}

	return &Service{
		Config: cfg,
		State: resilience.NewState(
			[]string{"Aegis", "Dawn"},
			1,
		),
		Events: transport.NewEventHub(),
	}
}

func TestDisableAndEnableTower(t *testing.T) {
	service := newTowerControlTestService()

	err := service.DisableTower(
		"Aegis",
		2,
	)

	if err != nil {
		t.Fatal(err)
	}

	if !service.State.IsTowerDisabled(
		"Aegis",
		2,
	) {
		t.Fatal(
			"expected Aegis tower 2 to be disabled",
		)
	}

	err = service.EnableTower(
		"Aegis",
		2,
	)

	if err != nil {
		t.Fatal(err)
	}

	if service.State.IsTowerDisabled(
		"Aegis",
		2,
	) {
		t.Fatal(
			"expected Aegis tower 2 to be restored",
		)
	}
}

func TestDisableTowerRejectsInvalidIndex(t *testing.T) {
	service := newTowerControlTestService()

	err := service.DisableTower(
		"Aegis",
		8,
	)

	if err == nil {
		t.Fatal(
			"expected tower index 8 to be rejected",
		)
	}
}

func TestDisableTowerRejectsUnknownPlanet(t *testing.T) {
	service := newTowerControlTestService()

	err := service.DisableTower(
		"Unknown",
		0,
	)

	if err == nil {
		t.Fatal(
			"expected unknown planet to be rejected",
		)
	}
}

func TestDisableTowerHTTPHandler(t *testing.T) {
	service := newTowerControlTestService()

	body := []byte(`{
		"planet_id": "Aegis",
		"tower_index": 3
	}`)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/towers/disable",
		bytes.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	service.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	if !service.State.IsTowerDisabled(
		"Aegis",
		3,
	) {
		t.Fatal(
			"expected HTTP endpoint to disable Aegis tower 3",
		)
	}
}

func TestEnableTowerHTTPHandler(t *testing.T) {
	service := newTowerControlTestService()

	service.State.DisableTower(
		"Aegis",
		3,
	)

	body := []byte(`{
		"planet_id": "Aegis",
		"tower_index": 3
	}`)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/towers/enable",
		bytes.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	response := httptest.NewRecorder()

	service.Handler().ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	if service.State.IsTowerDisabled(
		"Aegis",
		3,
	) {
		t.Fatal(
			"expected HTTP endpoint to restore Aegis tower 3",
		)
	}
}
