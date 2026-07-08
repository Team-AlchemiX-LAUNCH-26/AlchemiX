package protocol

import "github.com/launch26/relic-ring-protocol/internal/domain"

type APIError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type UniverseResponse struct {
	Status   string                 `json:"status"`
	Snapshot domain.NetworkSnapshot `json:"snapshot"`
}

type ActionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
