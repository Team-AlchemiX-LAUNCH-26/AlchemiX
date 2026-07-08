package orchestrator

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/launch26/relic-ring-protocol/internal/transport"
	"github.com/launch26/relic-ring-protocol/pkg/protocol"
)

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		transport.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /api/universe", s.handleUniverse)
	mux.HandleFunc("POST /api/transmissions", s.handleTransmission)
	mux.HandleFunc("POST /api/nodes/{id}/disable", s.handleDisableNode)
	mux.HandleFunc("POST /api/nodes/{id}/enable", s.handleEnableNode)
	mux.HandleFunc("POST /api/links/disable", s.handleDisableLink)
	mux.HandleFunc("POST /api/links/enable", s.handleEnableLink)
	mux.HandleFunc("POST /api/towers/disable", s.handleDisableTower)
	mux.HandleFunc("POST /api/towers/enable", s.handleEnableTower)
	mux.HandleFunc("POST /api/network/reset", s.handleReset)
	mux.HandleFunc("POST /api/events/ingest", s.handleEventIngest)
	mux.Handle("GET /api/events", s.Events)
	return cors(mux)
}
func (s *Service) handleDisableTower(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req protocol.TowerRequest

	if err := transport.DecodeJSON(r, &req); err != nil {
		transport.WriteJSON(
			w,
			http.StatusBadRequest,
			protocol.APIError{
				Status:  "failed",
				Message: err.Error(),
			},
		)

		return
	}

	if err := s.DisableTower(
		req.PlanetID,
		req.TowerIndex,
	); err != nil {
		transport.WriteJSON(
			w,
			http.StatusBadRequest,
			protocol.APIError{
				Status:  "failed",
				Message: err.Error(),
			},
		)

		return
	}

	transport.WriteJSON(
		w,
		http.StatusOK,
		protocol.ActionResponse{
			Status: "ok",
			Message: fmt.Sprintf(
				"%s tower %d disabled",
				req.PlanetID,
				req.TowerIndex,
			),
		},
	)
}

func (s *Service) handleEnableTower(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req protocol.TowerRequest

	if err := transport.DecodeJSON(r, &req); err != nil {
		transport.WriteJSON(
			w,
			http.StatusBadRequest,
			protocol.APIError{
				Status:  "failed",
				Message: err.Error(),
			},
		)

		return
	}

	if err := s.EnableTower(
		req.PlanetID,
		req.TowerIndex,
	); err != nil {
		transport.WriteJSON(
			w,
			http.StatusBadRequest,
			protocol.APIError{
				Status:  "failed",
				Message: err.Error(),
			},
		)

		return
	}

	transport.WriteJSON(
		w,
		http.StatusOK,
		protocol.ActionResponse{
			Status: "ok",
			Message: fmt.Sprintf(
				"%s tower %d restored",
				req.PlanetID,
				req.TowerIndex,
			),
		},
	)
}
func (s *Service) handleUniverse(w http.ResponseWriter, _ *http.Request) {
	snapshot, err := s.Snapshot()
	if err != nil {
		transport.WriteJSON(w, 500, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	transport.WriteJSON(w, 200, protocol.UniverseResponse{Status: "ok", Snapshot: snapshot})
}

func (s *Service) handleTransmission(w http.ResponseWriter, r *http.Request) {
	var req protocol.TransmissionRequest
	if err := transport.DecodeJSON(r, &req); err != nil {
		transport.WriteJSON(w, 400, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	result, err := s.StartTransmission(r.Context(), req)
	if err != nil {
		transport.WriteJSON(w, 422, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	transport.WriteJSON(w, 200, result)
}

func (s *Service) handleDisableNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.DisableNode(r.Context(), id); err != nil {
		transport.WriteJSON(w, 400, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	transport.WriteJSON(w, 200, protocol.ActionResponse{Status: "ok", Message: id + " disabled"})
}
func (s *Service) handleEnableNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.EnableNode(r.Context(), id); err != nil {
		transport.WriteJSON(w, 400, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	transport.WriteJSON(w, 200, protocol.ActionResponse{Status: "ok", Message: id + " enabled"})
}
func (s *Service) handleDisableLink(w http.ResponseWriter, r *http.Request) {
	var req protocol.LinkRequest
	if err := transport.DecodeJSON(r, &req); err != nil {
		transport.WriteJSON(w, 400, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	if err := s.DisableLink(req.A, req.B); err != nil {
		transport.WriteJSON(w, 400, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	transport.WriteJSON(w, 200, protocol.ActionResponse{Status: "ok", Message: "link disabled"})
}
func (s *Service) handleEnableLink(w http.ResponseWriter, r *http.Request) {
	var req protocol.LinkRequest
	if err := transport.DecodeJSON(r, &req); err != nil {
		transport.WriteJSON(w, 400, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	if err := s.EnableLink(req.A, req.B); err != nil {
		transport.WriteJSON(w, 400, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	transport.WriteJSON(w, 200, protocol.ActionResponse{Status: "ok", Message: "link enabled"})
}
func (s *Service) handleReset(w http.ResponseWriter, r *http.Request) {
	s.Reset(r.Context())
	transport.WriteJSON(w, 200, protocol.ActionResponse{Status: "ok", Message: "network restored"})
}
func (s *Service) handleEventIngest(w http.ResponseWriter, r *http.Request) {
	var event protocol.Event
	if err := transport.DecodeJSON(r, &event); err != nil {
		transport.WriteJSON(w, 400, protocol.APIError{Status: "failed", Message: err.Error()})
		return
	}
	s.Events.Publish(event)
	transport.WriteJSON(w, 202, protocol.ActionResponse{Status: "ok", Message: "event accepted"})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if strings.EqualFold(r.Method, http.MethodOptions) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
