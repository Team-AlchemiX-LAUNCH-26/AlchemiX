package agent

import (
	"fmt"
	"sort"
)

// RouterAdapter wraps the Phase 1 router to implement BaselineRouter.
// Uses the candidate.Generator's dijkstra to find shortest paths
// without depending directly on the existing routing package,
// keeping the agent fully decoupled.
type RouterAdapter struct {
	planets   []string
	neighbors map[string][]string
	weights   map[string]float64
}

// NewRouterAdapter creates an adapter from the universe topology.
func NewRouterAdapter(planets []string, links []RouterLink) *RouterAdapter {
	neighbors := make(map[string][]string)
	weights := make(map[string]float64)
	for _, p := range planets {
		neighbors[p] = []string{}
	}
	for _, l := range links {
		neighbors[l.A] = append(neighbors[l.A], l.B)
		neighbors[l.B] = append(neighbors[l.B], l.A)
		weights[CanonicalLinkID(l.A, l.B)] = l.LatencyMS
	}
	// Sort neighbors for determinism.
	for p := range neighbors {
		sort.Strings(neighbors[p])
	}
	return &RouterAdapter{planets: planets, neighbors: neighbors, weights: weights}
}

// RouterLink defines a link for the router adapter.
type RouterLink struct {
	A, B      string
	LatencyMS float64
}

// FindRoute returns the shortest path between origin and destination.
func (ra *RouterAdapter) FindRoute(originID, destinationID string) (CandidateRoute, error) {
	// Simple Dijkstra, reusing the same approach as candidate.Generator.
	path, cost, err := dijkstraSimple(ra.neighbors, ra.weights, originID, destinationID)
	if err != nil {
		return CandidateRoute{}, err
	}
	return CandidateRoute{Path: path, PhysicalLatencyMS: cost}, nil
}

func dijkstraSimple(neighbors map[string][]string, weights map[string]float64, src, dst string) ([]string, float64, error) {
	if _, ok := neighbors[src]; !ok {
		return nil, 0, fmt.Errorf("unknown planet: %s", src)
	}
	if _, ok := neighbors[dst]; !ok {
		return nil, 0, fmt.Errorf("unknown planet: %s", dst)
	}

	type entry struct {
		node string
		cost float64
	}

	dist := map[string]float64{src: 0}
	prev := map[string]string{}
	visited := map[string]bool{}

	for {
		var current string
		best := 1e18
		for n, d := range dist {
			if !visited[n] && d < best {
				best = d
				current = n
			}
		}
		if current == "" || current == dst {
			break
		}
		visited[current] = true

		for _, nb := range neighbors[current] {
			lid := CanonicalLinkID(current, nb)
			w := weights[lid]
			if w == 0 {
				w = 1
			}
			alt := dist[current] + w
			if d, ok := dist[nb]; !ok || alt < d {
				dist[nb] = alt
				prev[nb] = current
			}
		}
	}

	if _, ok := dist[dst]; !ok {
		return nil, 0, fmt.Errorf("no route from %s to %s", src, dst)
	}

	path := []string{}
	for n := dst; n != ""; n = prev[n] {
		path = append([]string{n}, path...)
	}
	return path, dist[dst], nil
}
