package liveapi

import (
	"context"
	"math/rand"

	"github.com/launch26/relic-ring-protocol/internal/agent"
)

// Client fetches live network state from the Chimera API.
//
// # API Key Configuration
//
// Set the following environment variables in your .env file:
//
//	CHIMERA_API_KEY=your-secret-key
//	CHIMERA_BASE_URL=https://chimera.launch26.space
//
// The client sends the key as:
//
//	GET /state
//	X-Team-Key: <CHIMERA_API_KEY>
//
// Currently using mock responses for development.
type Client struct {
	baseURL    string
	apiKey     string
	tickCache  *TickCache
	baselines  map[string]float64
	capacities map[string]float64
}

// NewClient creates a new live API client.
// In production, set baseURL from CHIMERA_BASE_URL and apiKey from CHIMERA_API_KEY.
func NewClient(baseURL, apiKey string, baselines, capacities map[string]float64) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		tickCache:  NewTickCache(),
		baselines:  baselines,
		capacities: capacities,
	}
}

// GetLatestState fetches the current network state.
// Currently returns mock data; replace with real HTTP call when API is available.
func (c *Client) GetLatestState(ctx context.Context) (agent.NetworkState, error) {
	// TODO: Replace with real HTTP call:
	//
	//   req, _ := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/state", nil)
	//   req.Header.Set("X-Team-Key", c.apiKey)
	//   resp, err := httpClient.Do(req)
	//   ...
	//
	// For now, generate mock state.
	return c.mockState(), nil
}

// mockState generates a realistic mock network state for development.
func (c *Client) mockState() agent.NetworkState {
	tick := c.tickCache.NextTick()

	links := map[string]agent.LinkObservation{}
	validLinks := map[string]bool{}
	uncertainties := map[string]float64{}

	linkDefs := []struct {
		id, a, b string
	}{
		{"Aegis-Boreas", "Aegis", "Boreas"},
		{"Aegis-Dawn", "Aegis", "Dawn"},
		{"Aegis-Elysium", "Aegis", "Elysium"},
		{"Boreas-Dawn", "Boreas", "Dawn"},
		{"Boreas-Elysium", "Boreas", "Elysium"},
		{"Boreas-Fenix", "Boreas", "Fenix"},
		{"Caelum-Dawn", "Caelum", "Dawn"},
		{"Caelum-Elysium", "Caelum", "Elysium"},
		{"Caelum-Fenix", "Caelum", "Fenix"},
		{"Dawn-Elysium", "Dawn", "Elysium"},
		{"Dawn-Fenix", "Dawn", "Fenix"},
		{"Elysium-Fenix", "Elysium", "Fenix"},
	}

	for _, ld := range linkDefs {
		loadRatio := rand.Float64() * 0.85
		capacity := c.capacities[ld.id]
		if capacity == 0 {
			capacity = 100
		}
		currentLoad := loadRatio * capacity
		status := "ok"
		if loadRatio >= 0.90 {
			status = "saturated"
		}

		baseline := c.baselines[ld.id]
		selfReported := baseline * (1 + rand.Float64()*0.3)
		trafficShare := 1.0 / float64(len(linkDefs))

		obs := agent.LinkObservation{
			Tick:                  tick,
			LinkID:                ld.id,
			PlanetA:               ld.a,
			PlanetB:               ld.b,
			CapacityUnits:         capacity,
			CurrentLoad:           currentLoad,
			LoadRatio:             loadRatio,
			SelfReportedLatencyMS: &selfReported,
			TrafficShare:          trafficShare,
			Status:                status,
			PhysicalLatencyMS:     baseline,
		}

		links[ld.id] = obs
		validLinks[ld.id] = status == "ok"
		uncertainties[ld.id] = 0.05
	}

	state := agent.NetworkState{
		Tick:          tick,
		Links:         links,
		ValidLinks:    validLinks,
		Uncertainties: uncertainties,
	}

	c.tickCache.Store(state)
	return state
}
