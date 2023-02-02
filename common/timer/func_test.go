package timer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBeginOfWeek(t *testing.T) {
	r := require.New(t)
	start := time.Date(2022, 11, 28, 0, 0, 0, 0, time.UTC)
	tt := time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)
	t1 := BeginOfWeek(tt)
	r.Equal(start, t1)

	tt = time.Date(2022, 11, 30, 13, 26, 0, 0, time.UTC)
	t1 = BeginOfWeek(tt)
	r.Equal(start, t1)

	tt = time.Date(2022, 12, 4, 23, 59, 59, 999, time.UTC)
	t1 = BeginOfWeek(tt)
	r.Equal(start, t1)

	tt = time.Date(2022, 12, 4, 23, 59, 60, 0, time.UTC)
	t1 = BeginOfWeek(tt)
	r.NotEqual(start, t1)
}
