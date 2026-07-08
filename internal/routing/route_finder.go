package routing

import (
	"fmt"

	"github.com/launch26/relic-ring-protocol/internal/domain"
	"github.com/launch26/relic-ring-protocol/internal/latency"
)

func FindLowestLatencyRoute(
	g *Graph,
	source string,
	destination string,
) (domain.Route, error) {
	path, _, err := FindPath(g, source, destination)
	if err != nil {
		return domain.Route{}, err
	}

	return BuildRouteDetails(g, path)
}

func BuildRouteDetails(
	g *Graph,
	path []string,
) (domain.Route, error) {
	if len(path) == 0 {
		return domain.Route{}, fmt.Errorf(
			"route path cannot be empty",
		)
	}

	if len(path) == 1 {
		p := g.Planets[path[0]]

		transit, err := latency.PlanetTransitWithFailures(
			p,
			0,
			0,
			g.Metadata,
			g.DisabledTowersForPlanet(p.ID),
		)

		if err != nil {
			return domain.Route{}, err
		}

		breakdown := latency.Aggregate(
			[]domain.PlanetTransitBreakdown{transit},
			nil,
		)

		return domain.Route{
			Path:           path,
			PlanetTransits: []domain.PlanetTransitBreakdown{transit},
			Latency:        breakdown,
		}, nil
	}

	planetTransits := make(
		[]domain.PlanetTransitBreakdown,
		0,
		len(path),
	)

	voidTransits := make(
		[]domain.VoidTransitBreakdown,
		0,
		len(path)-1,
	)

	for i, id := range path {
		planet := g.Planets[id]

		var entry int
		var exit int

		if i == 0 {
			link, err := g.Link(id, path[i+1])
			if err != nil {
				return domain.Route{}, err
			}

			exit, _, _ = link.TowersFor(id, path[i+1])
			entry = exit
		} else if i == len(path)-1 {
			link, err := g.Link(path[i-1], id)
			if err != nil {
				return domain.Route{}, err
			}

			_, entry, _ = link.TowersFor(path[i-1], id)
			exit = entry
		} else {
			incoming, err := g.Link(path[i-1], id)
			if err != nil {
				return domain.Route{}, err
			}

			_, entry, _ = incoming.TowersFor(path[i-1], id)

			outgoing, err := g.Link(id, path[i+1])
			if err != nil {
				return domain.Route{}, err
			}

			exit, _, _ = outgoing.TowersFor(id, path[i+1])
		}

		planetTransit, err := latency.PlanetTransitWithFailures(
			planet,
			entry,
			exit,
			g.Metadata,
			g.DisabledTowersForPlanet(id),
		)

		if err != nil {
			return domain.Route{}, fmt.Errorf(
				"build route details for planet %s: %w",
				id,
				err,
			)
		}

		planetTransits = append(
			planetTransits,
			planetTransit,
		)

		if i < len(path)-1 {
			next := path[i+1]

			link, err := g.Link(id, next)
			if err != nil {
				return domain.Route{}, err
			}

			send, receive, _ := link.TowersFor(id, next)

			voidTransits = append(
				voidTransits,
				latency.VoidTransit(
					planet,
					g.Planets[next],
					link.VoidDistanceKM,
					send,
					receive,
					g.Metadata.SpeedOfLightKMS,
				),
			)
		}
	}

	breakdown := latency.Aggregate(
		planetTransits,
		voidTransits,
	)

	return domain.Route{
		Path:           path,
		PlanetTransits: planetTransits,
		VoidTransits:   voidTransits,
		Latency:        breakdown,
	}, nil
}
