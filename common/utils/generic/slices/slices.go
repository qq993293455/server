package slices

func Remove[T any](slice []T, idx int) []T {
	return append(slice[:idx], slice[idx+1:]...)
}

func Delete[T any](slice []T, f func(T) bool) []T {
	for i, v := range slice {
		if f(v) {
			return Remove[T](slice, i)
		}
	}
	return slice
}

func In[T comparable](slice []T, v T) bool {
	for _, vv := range slice {
		if vv == v {
			return true
		}
	}
	return false
}

func InIf[T any](slice []T, f func(v T) bool) bool {
	for _, vv := range slice {
		if f(vv) {
			return true
		}
	}
	return false
}
