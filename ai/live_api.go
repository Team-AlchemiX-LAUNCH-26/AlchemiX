package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// LiveAPIClient fetches real-time link telemetry from the competition
// GET /state endpoint during the live evaluation phase.
//
// IMPORTANT: This data is ONLY used as inference input, never for training.
type LiveAPIClient struct {
	baseURL string
	client  *http.Client
}

// NewLiveAPIClient creates a client pointing at the competition state API.
// The URL is read from the CHIMERA_API_URL environment variable.
// Falls back to a sensible local default for development.
func NewLiveAPIClient() *LiveAPIClient {
	base := strings.TrimRight(os.Getenv("CHIMERA_API_URL"), "/")
	if base == "" {
		base = "http://localhost:9090"
	}
	return &LiveAPIClient{
		baseURL: base,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// FetchState calls GET /state and returns a map of link_id → LinkState
// so callers can look up any link by its canonical ID in O(1).
func (c *LiveAPIClient) FetchState(ctx context.Context) (map[string]LinkState, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/state", nil)
	if err != nil {
		return nil, fmt.Errorf("live_api: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("live_api: GET /state: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("live_api: GET /state returned HTTP %d", resp.StatusCode)
	}

	var state LiveStateResponse
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("live_api: decode response: %w", err)
	}

	out := make(map[string]LinkState, len(state.Links))
	for _, lls := range state.Links {
		out[lls.LinkID] = LinkState{
			LinkID:                lls.LinkID,
			LoadUnits:             lls.LoadUnits,
			LoadRatio:             lls.LoadRatio,
			Status:                lls.Status,
			SelfReportedLatencyMS: lls.SelfReportedLatencyMS,
			MeasuredLatencyMS:     lls.MeasuredLatencyMS,
			TrafficShare:          lls.TrafficShare,
		}
	}
	return out, nil
}

// FetchStateWithFallback calls FetchState. On any error it logs a warning
// and returns a map of safe default values for all known links so the agent
// can still produce a degraded-but-functional assessment.
func (c *LiveAPIClient) FetchStateWithFallback(
	ctx context.Context,
	knownLinks []string,
) map[string]LinkState {
	states, err := c.FetchState(ctx)
	if err != nil {
		// Log without crashing — the agent continues in degraded mode.
		fmt.Fprintf(os.Stderr, "[ai/live_api] WARNING: %v — using safe defaults\n", err)
		states = make(map[string]LinkState, len(knownLinks))
		for _, id := range knownLinks {
			states[id] = safeDefaultState(id)
		}
	}
	// Fill any links missing from the API response with safe defaults.
	for _, id := range knownLinks {
		if _, ok := states[id]; !ok {
			states[id] = safeDefaultState(id)
		}
	}
	return states
}

// safeDefaultState returns conservative link values that produce
// near-zero penalty adjustments when the live API is unavailable.
func safeDefaultState(linkID string) LinkState {
	return LinkState{
		LinkID:                linkID,
		LoadUnits:             50.0,
		LoadRatio:             0.30,
		Status:                "ok",
		SelfReportedLatencyMS: 100_000,
		MeasuredLatencyMS:     100_000,
		TrafficShare:          0.08,
	}
}
