package protocol

import (
	"time"
)

type Event struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	PacketID  string    `json:"packet_id,omitempty"`
	PlanetID  string    `json:"planet_id,omitempty"`
	Message   string    `json:"message,omitempty"`
	Data      any       `json:"data,omitempty"`
}

func NewEvent(eventType string, data any) Event {
	return Event{Type: eventType, Timestamp: time.Now().UTC(), Data: data}
}
