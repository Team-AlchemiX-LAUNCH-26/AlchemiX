package latency

func TowerDelaySeconds(distinctTowers int, delayMS float64) float64 {
	return float64(distinctTowers) * delayMS / 1000.0
}
