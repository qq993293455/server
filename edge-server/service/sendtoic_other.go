//go:build !linux

package service

func send2IC(tag string, params map[string]string) error {
	return nil
}
