package audit

import (
	"fmt"
	"strings"
)

// GenerateReport produces a human-readable decision timeline.
func GenerateReport(records []Record) string {
	if len(records) == 0 {
		return "No decisions recorded."
	}

	var sb strings.Builder
	sb.WriteString("=== DECISION AUDIT REPORT ===\n\n")

	for i, r := range records {
		sb.WriteString(fmt.Sprintf("Decision %d [Tick %d]\n", i+1, r.Tick))
		sb.WriteString(fmt.Sprintf("  Planet:     %s\n", r.CurrentPlanet))
		sb.WriteString(fmt.Sprintf("  Link:       %s\n", r.LinkID))
		sb.WriteString(fmt.Sprintf("  Action:     %s\n", r.Action))
		sb.WriteString(fmt.Sprintf("  Next hop:   %s\n", r.ChosenNextPlanet))
		sb.WriteString(fmt.Sprintf("  True Cost:  %.1f ms\n", r.CombinedCost))
		sb.WriteString(fmt.Sprintf("  Trust:      %.2f\n", r.TrustScore))
		sb.WriteString(fmt.Sprintf("  Risk:       %.2f\n", r.TargetingRiskScore))
		sb.WriteString(fmt.Sprintf("  Reasons:    %s\n", strings.Join(r.Reasons, "; ")))
		if len(r.AlternativePath) > 0 {
			sb.WriteString(fmt.Sprintf("  Alt path:   %s\n", strings.Join(r.AlternativePath, " → ")))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Total decisions: %d\n", len(records)))
	return sb.String()
}
