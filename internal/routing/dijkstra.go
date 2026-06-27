package routing

import (
	"container/heap"
	"fmt"
	"math"
	"strings"
)

const routeEpsilon = 1e-9

func stateLess(a, b *routeState) bool {
	if math.Abs(a.cost-b.cost) > routeEpsilon {
		return a.cost < b.cost
	}
	if len(a.path) != len(b.path) {
		return len(a.path) < len(b.path)
	}
	return strings.Join(a.path, "\x00") < strings.Join(b.path, "\x00")
}

func FindPath(g *Graph, source, destination string) ([]string, float64, error) {
	if _, ok := g.Planets[source]; !ok {
		return nil, 0, fmt.Errorf("source planet %q is unavailable or unknown", source)
	}
	if _, ok := g.Planets[destination]; !ok {
		return nil, 0, fmt.Errorf("destination planet %q is unavailable or unknown", destination)
	}
	if source == destination {
		transit := 0.0
		return []string{source}, transit, nil
	}

	pq := &priorityQueue{}
	heap.Init(pq)
	pushState(pq, &routeState{current: source, path: []string{source}})
	best := map[string]float64{"\x00" + source: 0}

	for pq.Len() > 0 {
		state := popState(pq)
		key := state.previous + "\x00" + state.current
		if known, ok := best[key]; ok && state.cost > known+routeEpsilon {
			continue
		}
		if state.current == destination {
			return state.path, state.cost, nil
		}

		for _, edge := range g.Neighbors(state.current) {
			next, _ := edge.Other(state.current)
			if contains(state.path, next) {
				continue
			}
			step, err := transitionCost(
				g,
				state.previous,
				state.current,
				next,
			)

			if err != nil {
				// This candidate transition may be impossible because
				// broken towers block the internal ring path.
				//
				// Do not fail the whole route search. Just skip this
				// candidate and allow Dijkstra to try another path.
				continue
			}
			newCost := state.cost + step
			if next == destination {
				final, err := destinationCost(
					g,
					state.current,
					destination,
				)

				if err != nil {
					// Destination entry may be impossible due to tower failure.
					// Skip this candidate route.
					continue
				}
				newCost += final
			}
			newPath := append(append([]string(nil), state.path...), next)
			newKey := state.current + "\x00" + next
			old, exists := best[newKey]
			candidate := &routeState{previous: state.current, current: next, cost: newCost, path: newPath}
			if !exists || newCost < old-routeEpsilon {
				best[newKey] = newCost
				pushState(pq, candidate)
			} else if math.Abs(newCost-old) <= routeEpsilon {
				pushState(pq, candidate)
			}
		}
	}
	return nil, 0, fmt.Errorf("no valid route exists between %s and %s", source, destination)
}

func contains(path []string, id string) bool {
	for _, item := range path {
		if item == id {
			return true
		}
	}
	return false
}
