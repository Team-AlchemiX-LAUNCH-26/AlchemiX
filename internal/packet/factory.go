package packet

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func New(origin, destination, payload string, route []string) (domain.Packet, error) {
	if strings.TrimSpace(origin) == "" || strings.TrimSpace(destination) == "" {
		return domain.Packet{}, fmt.Errorf("origin and destination are required")
	}
	if payload == "" {
		return domain.Packet{}, fmt.Errorf("payload cannot be empty")
	}
	if len(route) == 0 || route[0] != origin || route[len(route)-1] != destination {
		return domain.Packet{}, fmt.Errorf("route does not match origin and destination")
	}
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return domain.Packet{}, fmt.Errorf("create packet id: %w", err)
	}
	return domain.Packet{
		ID:            "pkt-" + hex.EncodeToString(idBytes),
		OriginID:      origin,
		DestinationID: destination,
		CurrentID:     origin,
		Payload:       payload,
		HopLog:        []domain.HopLogEntry{},
		Route:         append([]string(nil), route...),
		Status:        domain.PacketCreated,
	}, nil
}
