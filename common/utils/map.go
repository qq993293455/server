package utils

import "coin-server/common/utils/generic"

func MergeMapNumber[K comparable, V generic.Number](m1 map[K]V, m2 map[K]V) map[K]V {
	for k, v := range m2 {
		m1[k] += v
	}
	return m1
}
