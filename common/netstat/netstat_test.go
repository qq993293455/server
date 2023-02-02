package netstat

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUsePorts(t *testing.T) {
	r := require.New(t)
	now := time.Now()
	m, err := UsedPorts(40000, 31000, "127.0.0.1", "10.23.50.230")
	fmt.Println(time.Now().Sub(now))
	r.NoError(err)
	fmt.Println(m)
}
