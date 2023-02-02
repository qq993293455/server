package percent

import "math"

const BASE = float64(10000)

// Addition 加成计算
func Addition(origin, percent int64) int64 {
	return int64(math.Ceil(float64(origin) * float64(percent) / BASE))
}

// AdditionFloat 加成计算
func AdditionFloat(origin, percent int64) float64 {
	return float64(origin) * float64(percent) / BASE
}
