package maps

import "coin-server/common/utils/generic"

// InIf 根据f函数检查kv是否在map中
func InIf[K comparable, V any](m map[K]V, f func(k K, v V) bool) bool {
	for k, v := range m {
		if f(k, v) {
			return true
		}
	}
	return false
}

func Merge[K comparable, V generic.Number](ret map[K]V, other map[K]V) {
	for k, v := range other {
		ret[k] += v
	}
}

func Copy[K comparable, V any](source map[K]V) map[K]V {
	ret := make(map[K]V, len(source))
	for k, v := range source {
		ret[k] = v
	}
	return ret
}
