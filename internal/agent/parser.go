package agent

import (
	"fmt"
	"regexp"
	"strings"
)

// DeterministicParser parses natural-language transmission requests.
type DeterministicParser struct {
	validPlanets map[string]bool
}

// NewParser creates a parser with the known planet names.
func NewParser(planetNames []string) *DeterministicParser {
	valid := make(map[string]bool, len(planetNames))
	for _, name := range planetNames {
		valid[strings.ToLower(name)] = true
	}
	return &DeterministicParser{validPlanets: valid}
}

// Parse extracts origin, destination, and payload from natural-language input.
func (p *DeterministicParser) Parse(raw string) (ParsedTransmissionRequest, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ParsedTransmissionRequest{}, fmt.Errorf("empty request")
	}

	// Extract quoted payload.
	payload := ""
	quoteRe := regexp.MustCompile(`"([^"]+)"`)
	if matches := quoteRe.FindStringSubmatch(raw); len(matches) > 1 {
		payload = matches[1]
	}

	normalized := strings.ToLower(raw)

	// Find origin (after "from").
	origin := p.findPlanetAfter(normalized, "from")

	// Find destination (after "to").
	destination := p.findPlanetAfter(normalized, "to")

	// Fallback: find any two planets mentioned in order.
	if origin == "" || destination == "" {
		planets := p.findAllPlanets(normalized)
		if len(planets) >= 2 {
			if origin == "" {
				origin = planets[0]
			}
			if destination == "" {
				for _, pl := range planets {
					if pl != origin {
						destination = pl
						break
					}
				}
			}
		}
	}

	if origin == "" {
		return ParsedTransmissionRequest{}, fmt.Errorf("could not identify origin planet")
	}
	if destination == "" {
		return ParsedTransmissionRequest{}, fmt.Errorf("could not identify destination planet")
	}
	if origin == destination {
		return ParsedTransmissionRequest{}, fmt.Errorf("origin and destination must be different: %s", origin)
	}

	// Default payload if none quoted.
	if payload == "" {
		payload = "transmission"
	}

	return ParsedTransmissionRequest{
		OriginID:      p.canonicalize(origin),
		DestinationID: p.canonicalize(destination),
		Payload:       payload,
	}, nil
}

// findPlanetAfter looks for a planet name after the given keyword.
func (p *DeterministicParser) findPlanetAfter(text, keyword string) string {
	idx := strings.Index(text, keyword)
	if idx == -1 {
		return ""
	}
	after := text[idx+len(keyword):]
	words := strings.Fields(after)
	for _, word := range words {
		cleaned := strings.Trim(word, ".,;:!?\"'")
		if p.validPlanets[cleaned] {
			return cleaned
		}
	}
	return ""
}

// findAllPlanets returns all planet names found in the text, in order.
func (p *DeterministicParser) findAllPlanets(text string) []string {
	words := strings.Fields(text)
	var found []string
	for _, word := range words {
		cleaned := strings.Trim(word, ".,;:!?\"'")
		if p.validPlanets[cleaned] {
			found = append(found, cleaned)
		}
	}
	return found
}

// canonicalize returns the properly capitalized planet name.
func (p *DeterministicParser) canonicalize(lower string) string {
	if lower == "" {
		return ""
	}
	return strings.ToUpper(lower[:1]) + lower[1:]
}
