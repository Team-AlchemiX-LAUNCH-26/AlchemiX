package domain

type Tower struct {
	Index        int     `json:"index"`
	AngleDegrees float64 `json:"angle_degrees"`
	XKM          float64 `json:"x_km"`
	YKM          float64 `json:"y_km"`
}

type TowerPair struct {
	FromTower Tower   `json:"from_tower"`
	ToTower   Tower   `json:"to_tower"`
	Distance  float64 `json:"distance_km"`
}

type RingDirection string

const (
	RingDirectionStationary       RingDirection = "stationary"
	RingDirectionClockwise        RingDirection = "clockwise"
	RingDirectionCounterClockwise RingDirection = "counter_clockwise"
)

// RingPath describes how a packet travels through a planet's
// underground fibre ring.
type RingPath struct {
	Towers         []int         `json:"towers"`
	Segments       int           `json:"segments"`
	DistinctTowers int           `json:"distinct_towers"`
	Direction      RingDirection `json:"direction"`
}
