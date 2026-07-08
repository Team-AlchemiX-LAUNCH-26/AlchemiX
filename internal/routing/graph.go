package routing

import (
	"fmt"
	"sort"

	"github.com/launch26/relic-ring-protocol/internal/domain"
)

type Graph struct {
	Metadata  domain.UniverseMetadata
	Planets   map[string]domain.Planet
	Adjacency map[string][]domain.Link
	Links     []domain.Link

	// Planet ID -> tower index -> disabled status.
	DisabledTowers map[string]map[int]bool
}

func (g *Graph) Neighbors(id string) []domain.Link {
	return g.Adjacency[id]
}

func (g *Graph) Link(
	from string,
	to string,
) (domain.Link, error) {
	for _, link := range g.Adjacency[from] {
		other, ok := link.Other(from)

		if ok && other == to {
			return link, nil
		}
	}

	return domain.Link{}, fmt.Errorf(
		"no active link between %s and %s",
		from,
		to,
	)
}

func (g *Graph) SortedPlanetIDs() []string {
	ids := make([]string, 0, len(g.Planets))

	for id := range g.Planets {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	return ids
}

func (g *Graph) DisabledTowersForPlanet(
	planetID string,
) map[int]bool {
	if g.DisabledTowers == nil {
		return nil
	}

	return g.DisabledTowers[planetID]
}
