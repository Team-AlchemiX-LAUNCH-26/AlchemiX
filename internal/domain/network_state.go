package domain

type PlanetStatus struct {
	ID               string `json:"id"`
	Endpoint         string `json:"endpoint"`
	Healthy          bool   `json:"healthy"`
	ManuallyDisabled bool   `json:"manually_disabled"`
	Available        bool   `json:"available"`
}

type NetworkSnapshot struct {
	Metadata       UniverseMetadata `json:"metadata"`
	Planets        []Planet         `json:"planets"`
	PlanetStatus   []PlanetStatus   `json:"planet_status"`
	Links          []Link           `json:"links"`
	DisabledLinks  []string         `json:"disabled_links"`
	DisabledTowers map[string][]int `json:"disabled_towers"`
}
