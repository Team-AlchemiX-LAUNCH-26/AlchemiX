package routing

import (
	"fmt"

	"github.com/launch26/relic-ring-protocol/internal/latency"
)

func transitionCost(
	g *Graph,
	previous string,
	current string,
	next string,
) (float64, error) {
	currentPlanet, ok := g.Planets[current]
	if !ok {
		return 0, fmt.Errorf(
			"planet %s is unavailable",
			current,
		)
	}

	outgoing, err := g.Link(current, next)
	if err != nil {
		return 0, err
	}

	exitTower, receiveTower, ok := outgoing.TowersFor(
		current,
		next,
	)

	if !ok {
		return 0, fmt.Errorf(
			"link tower direction missing for %s -> %s",
			current,
			next,
		)
	}

	entryTower := exitTower

	if previous != "" {
		incoming, err := g.Link(previous, current)
		if err != nil {
			return 0, err
		}

		_, entryTower, ok = incoming.TowersFor(
			previous,
			current,
		)

		if !ok {
			return 0, fmt.Errorf(
				"link tower direction missing for %s -> %s",
				previous,
				current,
			)
		}
	}

	planetTransit, err := latency.PlanetTransitWithFailures(
		currentPlanet,
		entryTower,
		exitTower,
		g.Metadata,
		g.DisabledTowersForPlanet(current),
	)

	if err != nil {
		return 0, fmt.Errorf(
			"planet %s internal ring unavailable: %w",
			current,
			err,
		)
	}

	nextPlanet := g.Planets[next]

	voidTransit := latency.VoidTransit(
		currentPlanet,
		nextPlanet,
		outgoing.VoidDistanceKM,
		exitTower,
		receiveTower,
		g.Metadata.SpeedOfLightKMS,
	)

	return planetTransit.TotalSeconds +
		voidTransit.TotalSeconds, nil
}

func destinationCost(
	g *Graph,
	previous string,
	destination string,
) (float64, error) {
	incoming, err := g.Link(previous, destination)
	if err != nil {
		return 0, err
	}

	_, entryTower, ok := incoming.TowersFor(
		previous,
		destination,
	)

	if !ok {
		return 0, fmt.Errorf(
			"link tower direction missing for %s -> %s",
			previous,
			destination,
		)
	}

	transit, err := latency.PlanetTransitWithFailures(
		g.Planets[destination],
		entryTower,
		entryTower,
		g.Metadata,
		g.DisabledTowersForPlanet(destination),
	)

	if err != nil {
		return 0, fmt.Errorf(
			"destination planet %s internal transit unavailable: %w",
			destination,
			err,
		)
	}

	return transit.TotalSeconds, nil
}
