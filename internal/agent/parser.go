package agent

import (
	"fmt"
	"regexp"
	"strings"
)

// DeterministicParser parses controlled natural-language transmission requests.
type DeterministicParser struct {
	validPlanets map[string]bool
	canonical    map[string]string
}

// NewParser creates a parser with the known planet names.
func NewParser(planetNames []string) *DeterministicParser {
	valid := make(map[string]bool, len(planetNames))
	canonical := make(map[string]string, len(planetNames))
	for _, name := range planetNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		lower := strings.ToLower(name)
		valid[lower] = true
		canonical[lower] = name
	}
	return &DeterministicParser{
		validPlanets: valid,
		canonical:    canonical,
	}
}

// Parse extracts origin, destination, and payload from a controlled request.
func (p *DeterministicParser) Parse(raw string) (ParsedTransmissionRequest, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ParsedTransmissionRequest{}, fmt.Errorf("empty request")
	}

	payload, masked, err := extractQuotedPayload(raw)
	if err != nil {
		return ParsedTransmissionRequest{}, err
	}
	if payload == "" {
		return ParsedTransmissionRequest{}, fmt.Errorf("missing payload")
	}

	tokens := tokenizeParserInput(masked)
	route, err := p.extractRoute(tokens)
	if err != nil {
		return ParsedTransmissionRequest{}, err
	}
	if route.origin == route.destination {
		return ParsedTransmissionRequest{}, fmt.Errorf("origin and destination must be different: %s", route.origin)
	}

	return ParsedTransmissionRequest{
		OriginID:      p.canonicalize(route.origin),
		DestinationID: p.canonicalize(route.destination),
		Payload:       payload,
	}, nil
}

type parserToken struct {
	text  string
	lower string
}

type parsedRoute struct {
	origin      string
	destination string
}

var parserTokenPattern = regexp.MustCompile(`[A-Za-z0-9]+|->`)

func tokenizeParserInput(text string) []parserToken {
	matches := parserTokenPattern.FindAllString(text, -1)
	tokens := make([]parserToken, 0, len(matches))
	for _, match := range matches {
		tokens = append(tokens, parserToken{
			text:  match,
			lower: strings.ToLower(match),
		})
	}
	return tokens
}

func (p *DeterministicParser) extractRoute(tokens []parserToken) (parsedRoute, error) {
	if route, ok, err := p.extractArrowRoute(tokens); ok || err != nil {
		return route, err
	}
	return p.extractFromToRoute(tokens)
}

func (p *DeterministicParser) extractArrowRoute(tokens []parserToken) (parsedRoute, bool, error) {
	arrowIndexes := tokenIndexes(tokens, "->")
	if len(arrowIndexes) == 0 {
		return parsedRoute{}, false, nil
	}
	if len(arrowIndexes) > 1 {
		return parsedRoute{}, true, fmt.Errorf("ambiguous route: multiple arrow operators")
	}
	if len(tokenIndexes(tokens, "from")) > 0 || len(tokenIndexes(tokens, "to")) > 0 {
		return parsedRoute{}, true, fmt.Errorf("ambiguous route: do not mix arrow syntax with from/to syntax")
	}

	arrow := arrowIndexes[0]
	origin, err := p.singlePlanet(tokens[:arrow], "origin")
	if err != nil {
		return parsedRoute{}, true, err
	}
	destination, err := p.singlePlanet(tokens[arrow+1:], "destination")
	if err != nil {
		return parsedRoute{}, true, err
	}
	return parsedRoute{origin: origin, destination: destination}, true, nil
}

func (p *DeterministicParser) extractFromToRoute(tokens []parserToken) (parsedRoute, error) {
	fromIndexes := tokenIndexes(tokens, "from")
	toIndexes := tokenIndexes(tokens, "to")

	if len(fromIndexes) == 0 {
		return parsedRoute{}, fmt.Errorf("could not identify origin planet")
	}
	if len(toIndexes) == 0 {
		return parsedRoute{}, fmt.Errorf("could not identify destination planet")
	}
	if len(fromIndexes) > 1 {
		return parsedRoute{}, fmt.Errorf("ambiguous origin: multiple from clauses")
	}
	if len(toIndexes) > 1 {
		return parsedRoute{}, fmt.Errorf("ambiguous destination: multiple to clauses")
	}

	fromIndex := fromIndexes[0]
	toIndex := toIndexes[0]
	if fromIndex == toIndex {
		return parsedRoute{}, fmt.Errorf("ambiguous route")
	}

	var originSegment []parserToken
	var destinationSegment []parserToken
	if fromIndex < toIndex {
		originSegment = tokens[fromIndex+1 : toIndex]
		destinationSegment = tokens[toIndex+1:]
	} else {
		destinationSegment = tokens[toIndex+1 : fromIndex]
		originSegment = tokens[fromIndex+1:]
	}

	origin, err := p.singlePlanet(originSegment, "origin")
	if err != nil {
		return parsedRoute{}, err
	}
	destination, err := p.singlePlanet(destinationSegment, "destination")
	if err != nil {
		return parsedRoute{}, err
	}

	return parsedRoute{origin: origin, destination: destination}, nil
}

func (p *DeterministicParser) singlePlanet(tokens []parserToken, role string) (string, error) {
	var found []string
	for _, token := range tokens {
		if p.validPlanets[token.lower] {
			found = append(found, token.lower)
		}
	}
	if len(found) == 0 {
		return "", fmt.Errorf("missing or unknown %s planet", role)
	}
	if len(found) > 1 {
		return "", fmt.Errorf("ambiguous %s planet", role)
	}
	return found[0], nil
}

func tokenIndexes(tokens []parserToken, lower string) []int {
	indexes := make([]int, 0)
	for index, token := range tokens {
		if token.lower == lower {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func extractQuotedPayload(raw string) (payload string, masked string, err error) {
	start := strings.Index(raw, `"`)
	if start == -1 {
		return "", raw, nil
	}

	end := strings.Index(raw[start+1:], `"`)
	if end == -1 {
		return "", "", fmt.Errorf("malformed payload: missing closing quote")
	}
	end += start + 1

	if strings.Contains(raw[end+1:], `"`) {
		return "", "", fmt.Errorf("ambiguous payload: multiple quoted payloads")
	}

	payload = strings.TrimSpace(raw[start+1 : end])
	if payload == "" {
		return "", "", fmt.Errorf("payload cannot be empty")
	}

	maskedBytes := []byte(raw)
	for i := start; i <= end; i++ {
		maskedBytes[i] = ' '
	}
	return payload, string(maskedBytes), nil
}

// canonicalize returns the configured planet capitalization when known.
func (p *DeterministicParser) canonicalize(lower string) string {
	if canonical, ok := p.canonical[lower]; ok {
		return canonical
	}
	if lower == "" {
		return ""
	}
	return strings.ToUpper(lower[:1]) + lower[1:]
}
