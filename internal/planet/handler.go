package planet

import (
	"net/http"

	"github.com/launch26/relic-ring-protocol/internal/transport"
	"github.com/launch26/relic-ring-protocol/pkg/protocol"
)

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /planet", s.handlePlanet)
	mux.HandleFunc("GET /metrics", s.handleMetrics)
	mux.HandleFunc("POST /packet/receive", s.handlePacket)
	mux.HandleFunc("POST /admin/disable", s.handleDisable)
	mux.HandleFunc("POST /admin/enable", s.handleEnable)
	return requestLogging(mux)
}

func (s *Service) handleHealth(w http.ResponseWriter, _ *http.Request) {
	status := "ready"
	code := http.StatusOK
	if !s.Active() {
		status = "disabled"
		code = http.StatusServiceUnavailable
	}
	transport.WriteJSON(w, code, map[string]any{"status": status, "planet_id": s.Planet.ID, "active": s.Active()})
}

func (s *Service) handlePlanet(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, s.Planet)
}

func (s *Service) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	transport.WriteJSON(w, http.StatusOK, s.Metrics.snapshot())
}

func (s *Service) handlePacket(w http.ResponseWriter, r *http.Request) {
	var req protocol.NodePacketRequest
	if err := transport.DecodeJSON(r, &req); err != nil {
		transport.WriteJSON(w, http.StatusBadRequest, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	result, err := s.Process(r.Context(), req)
	if err != nil {
		transport.WriteJSON(w, http.StatusBadGateway, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	transport.WriteJSON(w, http.StatusOK, result)
}

func (s *Service) handleDisable(w http.ResponseWriter, _ *http.Request) {
	s.Disable()
	transport.WriteJSON(w, http.StatusOK, protocol.ActionResponse{Status: "ok", Message: s.Planet.ID + " disabled"})
}

func (s *Service) handleEnable(w http.ResponseWriter, _ *http.Request) {
	s.Enable()
	transport.WriteJSON(w, http.StatusOK, protocol.ActionResponse{Status: "ok", Message: s.Planet.ID + " enabled"})
}

func requestLogging(next http.Handler) http.Handler { return next }
