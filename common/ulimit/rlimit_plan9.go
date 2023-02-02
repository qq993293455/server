//go:build plan9
// +build plan9

package ulimit

func SetRLimit() error {
	return nil
}

func SetOpenCoreDump() error {
	return nil
}
