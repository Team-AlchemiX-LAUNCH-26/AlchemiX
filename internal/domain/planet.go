package domain

type Planet struct {
	ID                    string  `json:"id"`
	Codex                 int     `json:"codex"`
	X                     float64 `json:"x"`
	Y                     float64 `json:"y"`
	RadiusKM              float64 `json:"radius_km"`
	ActiveTowers          int     `json:"active_towers"`
	AtmosphereThicknessKM float64 `json:"atmosphere_thickness_km"`
	RefractionIndex       float64 `json:"refraction_index"`
}
