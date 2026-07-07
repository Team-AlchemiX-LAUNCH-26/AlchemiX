// Package candidate implements Yen's K-shortest paths for route generation.
package candidate

import (
	"fmt"
	"math"
	"sort"

	"github.com/launch26/relic-ring-protocol/internal/agent"
)

// Generator produces top-K candidate routes using Yen's algorithm.
type Generator struct {
	planets   []string
	neighbors map[string][]string
	weights   map[string]float64 // linkID -> physical latency ms
}

// NewGenerator creates a candidate route generator from the universe topology.
func NewGenerator(planets []string, links []LinkDef) *Generator {
	neighbors := make(map[string][]string)
	weights := make(map[string]float64)

	for _, p := range planets {
		neighbors[p] = []string{}
	}

	for _, l := range links {
		neighbors[l.A] = append(neighbors[l.A], l.B)
		neighbors[l.B] = append(neighbors[l.B], l.A)
		weights[canonicalLinkID(l.A, l.B)] = l.LatencyMS
	}

	return &Generator{
		planets:   planets,
		neighbors: neighbors,
		weights:   weights,
	}
}

// LinkDef defines a link in the topology.
type LinkDef struct {
	A, B      string
	LatencyMS float64
}

// TopKPaths returns up to k shortest paths using Yen's algorithm.
func (g *Generator) TopKPaths(originID, destinationID string, k int, excludedLinks map[string]bool) ([]agent.CandidateRoute, error) {
	if _, ok := g.neighbors[originID]; !ok {
		return nil, fmt.Errorf("unknown planet: %s", originID)
	}
	if _, ok := g.neighbors[destinationID]; !ok {
		return nil, fmt.Errorf("unknown planet: %s", destinationID)
	}

	// Find shortest path.
	shortest, cost, err := g.dijkstra(originID, destinationID, excludedLinks, nil)
	if err != nil {
		return nil, err
	}

	results := []agent.CandidateRoute{{Path: shortest, PhysicalLatencyMS: cost}}
	candidates := []agent.CandidateRoute{}

	for i := 1; i < k; i++ {
		prevPath := results[len(results)-1].Path

		for j := 0; j < len(prevPath)-1; j++ {
			spurNode := prevPath[j]
			rootPath := prevPath[:j+1]

			// Exclude edges from root to spur used by existing results.
			edgeExclusions := make(map[string]bool)
			for lid, excl := range excludedLinks {
				if excl {
					edgeExclusions[lid] = true
				}
			}
			for _, r := range results {
				if len(r.Path) > j && pathEqual(r.Path[:j+1], rootPath) {
					edgeExclusions[canonicalLinkID(r.Path[j], r.Path[j+1])] = true
				}
			}

			// Exclude nodes in root path (except spur node).
			nodeExclusions := make(map[string]bool)
			for _, n := range rootPath[:len(rootPath)-1] {
				nodeExclusions[n] = true
			}

			spurPath, spurCost, err := g.dijkstra(spurNode, destinationID, edgeExclusions, nodeExclusions)
			if err != nil {
				continue
			}

			// Combine root + spur.
			totalPath := make([]string, len(rootPath)-1)
			copy(totalPath, rootPath[:len(rootPath)-1])
			totalPath = append(totalPath, spurPath...)

			rootCost := g.pathCost(rootPath)
			totalCost := rootCost + spurCost

			candidate := agent.CandidateRoute{Path: totalPath, PhysicalLatencyMS: totalCost}

			// Avoid duplicates.
			if !containsRoute(candidates, candidate) && !containsRoute(results, candidate) {
				candidates = append(candidates, candidate)
			}
		}

		if len(candidates) == 0 {
			break
		}

		// Pick the lowest-cost candidate.
		sort.Slice(candidates, func(a, b int) bool {
			return candidates[a].PhysicalLatencyMS < candidates[b].PhysicalLatencyMS
		})

		results = append(results, candidates[0])
		candidates = candidates[1:]
	}

	return results, nil
}

func (g *Generator) dijkstra(src, dst string, excludedEdges, excludedNodes map[string]bool) ([]string, float64, error) {
	dist := map[string]float64{src: 0}
	prev := map[string]string{}
	visited := map[string]bool{}

	for {
		// Find unvisited node with smallest distance.
		current := ""
		best := math.Inf(1)
		for n, d := range dist {
			if !visited[n] && d < best {
				best = d
				current = n
			}
		}
		if current == "" {
			break
		}
		if current == dst {
			break
		}
		visited[current] = true

		for _, neighbor := range g.neighbors[current] {
			if excludedNodes != nil && excludedNodes[neighbor] {
				continue
			}
			lid := canonicalLinkID(current, neighbor)
			if excludedEdges != nil && excludedEdges[lid] {
				continue
			}
			w := g.weights[lid]
			if w == 0 {
				w = 1
			}
			alt := dist[current] + w
			if d, ok := dist[neighbor]; !ok || alt < d {
				dist[neighbor] = alt
				prev[neighbor] = current
			}
		}
	}

	if _, ok := dist[dst]; !ok {
		return nil, 0, fmt.Errorf("no path from %s to %s", src, dst)
	}

	// Reconstruct path.
	path := []string{}
	for n := dst; n != ""; n = prev[n] {
		path = append([]string{n}, path...)
	}
	return path, dist[dst], nil
}

func (g *Generator) pathCost(path []string) float64 {
	cost := 0.0
	for i := 0; i < len(path)-1; i++ {
		lid := canonicalLinkID(path[i], path[i+1])
		cost += g.weights[lid]
	}
	return cost
}

func canonicalLinkID(a, b string) string {
	if a < b {
		return a + "-" + b
	}
	return b + "-" + a
}

func pathEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsRoute(routes []agent.CandidateRoute, r agent.CandidateRoute) bool {
	for _, existing := range routes {
		if pathEqual(existing.Path, r.Path) {
			return true
		}
	}
	return false
}
