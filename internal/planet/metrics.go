package planet

import (
	"sync"
	"time"
)

type Metrics struct {
	mu               sync.RWMutex
	PacketsProcessed int64         `json:"packets_processed"`
	PacketsForwarded int64         `json:"packets_forwarded"`
	Failures         int64         `json:"failures"`
	LastProcessing   time.Duration `json:"last_processing_ns"`
}

func (m *Metrics) snapshot() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Metrics{PacketsProcessed: m.PacketsProcessed, PacketsForwarded: m.PacketsForwarded, Failures: m.Failures, LastProcessing: m.LastProcessing}
}
