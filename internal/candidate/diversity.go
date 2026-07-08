package candidate

import "github.com/launch26/relic-ring-protocol/internal/agent"

// DiversityScore calculates how edge-diverse a backup route is compared to the primary.
// Returns a value in [0, 1] where 1 means fully diverse (no shared edges).
func DiversityScore(primary, backup agent.CandidateRoute) float64 {
	if len(primary.Path) < 2 || len(backup.Path) < 2 {
		return 1.0
	}

	primaryEdges := make(map[string]bool)
	for i := 0; i < len(primary.Path)-1; i++ {
		primaryEdges[canonicalLinkID(primary.Path[i], primary.Path[i+1])] = true
	}

	backupEdges := 0
	sharedEdges := 0
	for i := 0; i < len(backup.Path)-1; i++ {
		backupEdges++
		lid := canonicalLinkID(backup.Path[i], backup.Path[i+1])
		if primaryEdges[lid] {
			sharedEdges++
		}
	}

	if backupEdges == 0 {
		return 1.0
	}
	return 1.0 - float64(sharedEdges)/float64(backupEdges)
}

// RankByDiversity reorders candidates so that edge-diverse routes rank higher.
// The primary (index 0) is kept in place.
func RankByDiversity(candidates []agent.CandidateRoute) []agent.CandidateRoute {
	if len(candidates) <= 1 {
		return candidates
	}

	primary := candidates[0]
	rest := make([]agent.CandidateRoute, len(candidates)-1)
	copy(rest, candidates[1:])

	// Sort rest by diversity descending, then by cost ascending.
	for i := 0; i < len(rest); i++ {
		for j := i + 1; j < len(rest); j++ {
			di := DiversityScore(primary, rest[i])
			dj := DiversityScore(primary, rest[j])
			if dj > di || (dj == di && rest[j].PhysicalLatencyMS < rest[i].PhysicalLatencyMS) {
				rest[i], rest[j] = rest[j], rest[i]
			}
		}
	}

	result := make([]agent.CandidateRoute, 0, len(candidates))
	result = append(result, primary)
	result = append(result, rest...)
	return result
}
