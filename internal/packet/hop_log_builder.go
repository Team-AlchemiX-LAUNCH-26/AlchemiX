package packet

import "github.com/launch26/relic-ring-protocol/internal/domain"

func AppendHop(p *domain.Packet, entry domain.HopLogEntry) {
	entry.Sequence = len(p.HopLog) + 1
	p.HopLog = append(p.HopLog, entry)
	p.CumulativeLatencySeconds = entry.CumulativeLatency
}
