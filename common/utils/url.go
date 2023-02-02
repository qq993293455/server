package utils

import (
	"path"
	"strings"
)

func JoinURL(base string, paths ...string) string {
	p := path.Join(paths...)
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(p, "/")
}
