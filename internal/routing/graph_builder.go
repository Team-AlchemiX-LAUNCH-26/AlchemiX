package routing

import (
	"fmt"
	"sort"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

func BuildGraph(
	cfg domain.UniverseConfig,
	unavailableNodes map[string]bool,
	disabledLinks map[string]bool,
	disabledTowers map[string]map[int]bool,
) (*Graph, error) {
	graph := &Graph{
		Metadata:       cfg.Metadata,
		Planets:        make(map[string]domain.Planet),
		Adjacency:      make(map[string][]domain.Link),
		Links:          make([]domain.Link, 0),
		DisabledTowers: cloneDisabledTowers(disabledTowers),
	}

	// Add all available planets as graph nodes.
	for _, planet := range cfg.Nodes {
		if unavailableNodes[planet.ID] {
			continue
		}

		graph.Planets[planet.ID] = planet
		graph.Adjacency[planet.ID] = []domain.Link{}
	}

	// Compare each unique pair of planets.
	for i := 0; i < len(cfg.Nodes); i++ {
		firstPlanet := cfg.Nodes[i]

		if unavailableNodes[firstPlanet.ID] {
			continue
		}

		for j := i + 1; j < len(cfg.Nodes); j++ {
			secondPlanet := cfg.Nodes[j]

			if unavailableNodes[secondPlanet.ID] {
				continue
			}

			linkID := LinkID(
				firstPlanet.ID,
				secondPlanet.ID,
			)

			if disabledLinks[linkID] {
				continue
			}

			link, valid, err := BuildLink(
				firstPlanet,
				secondPlanet,
				cfg.Metadata,
				disabledTowers,
			)

			if err != nil {
				return nil, fmt.Errorf(
					"build link %s-%s: %w",
					firstPlanet.ID,
					secondPlanet.ID,
					err,
				)
			}

			if !valid {
				continue
			}

			graph.Links = append(
				graph.Links,
				link,
			)

			graph.Adjacency[firstPlanet.ID] = append(
				graph.Adjacency[firstPlanet.ID],
				link,
			)

			graph.Adjacency[secondPlanet.ID] = append(
				graph.Adjacency[secondPlanet.ID],
				link,
			)
		}
	}

	// Keep neighbour ordering deterministic.
	for planetID := range graph.Adjacency {
		sort.Slice(
			graph.Adjacency[planetID],
			func(i int, j int) bool {
				firstNeighbour, _ :=
					graph.Adjacency[planetID][i].
						Other(planetID)

				secondNeighbour, _ :=
					graph.Adjacency[planetID][j].
						Other(planetID)

				return firstNeighbour < secondNeighbour
			},
		)
	}

	return graph, nil
}

func cloneDisabledTowers(
	input map[string]map[int]bool,
) map[string]map[int]bool {
	output := make(
		map[string]map[int]bool,
		len(input),
	)

	for planetID, towers := range input {
		output[planetID] = make(
			map[int]bool,
			len(towers),
		)

		for towerIndex, disabled := range towers {
			output[planetID][towerIndex] = disabled
		}
	}

	return output
}
