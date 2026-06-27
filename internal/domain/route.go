package domain

type Link struct {
	A              string  `json:"a"`
	B              string  `json:"b"`
	VoidDistanceKM float64 `json:"void_distance_km"`
	ATower         int     `json:"a_tower"`
	BTower         int     `json:"b_tower"`
}

func (l Link) Other(id string) (string, bool) {
	switch id {
	case l.A:
		return l.B, true
	case l.B:
		return l.A, true
	default:
		return "", false
	}
}

func (l Link) TowersFor(from, to string) (sending, receiving int, ok bool) {
	if from == l.A && to == l.B {
		return l.ATower, l.BTower, true
	}
	if from == l.B && to == l.A {
		return l.BTower, l.ATower, true
	}
	return 0, 0, false
}

type PlanetTransitBreakdown struct {
	PlanetID   string `json:"planet_id"`
	EntryTower int    `json:"entry_tower"`
	ExitTower  int    `json:"exit_tower"`

	// The exact sequence of towers used inside the planet.
	RingPath []int `json:"ring_path"`

	// stationary, clockwise, or counter_clockwise.
	RingDirection RingDirection `json:"ring_direction"`

	Segments       int     `json:"segments"`
	DistinctTowers int     `json:"distinct_towers"`
	FiberSeconds   float64 `json:"fiber_seconds"`
	TowerSeconds   float64 `json:"tower_seconds"`
	TotalSeconds   float64 `json:"total_seconds"`
}

type VoidTransitBreakdown struct {
	FromID                       string  `json:"from_id"`
	ToID                         string  `json:"to_id"`
	SendingTower                 int     `json:"sending_tower"`
	ReceivingTower               int     `json:"receiving_tower"`
	VoidDistanceKM               float64 `json:"void_distance_km"`
	SourceAtmosphereSeconds      float64 `json:"source_atmosphere_seconds"`
	VacuumSeconds                float64 `json:"vacuum_seconds"`
	DestinationAtmosphereSeconds float64 `json:"destination_atmosphere_seconds"`
	TotalSeconds                 float64 `json:"total_seconds"`
}

type LatencyBreakdown struct {
	FiberSeconds      float64 `json:"fiber_seconds"`
	TowerSeconds      float64 `json:"tower_seconds"`
	AtmosphereSeconds float64 `json:"atmosphere_seconds"`
	VoidSeconds       float64 `json:"void_seconds"`
	TotalSeconds      float64 `json:"total_seconds"`
}

type Route struct {
	Path           []string                 `json:"path"`
	PlanetTransits []PlanetTransitBreakdown `json:"planet_transits"`
	VoidTransits   []VoidTransitBreakdown   `json:"void_transits"`
	Latency        LatencyBreakdown         `json:"latency"`
}
