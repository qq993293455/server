package gmath

import (
	"math"

	"coin-server/common/utils/generic"
)

func CeilTo[T generic.Number](v float64) T {
	return T(math.Ceil(v))
}
