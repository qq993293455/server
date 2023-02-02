package system_info

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystemInfo(t *testing.T) {
	r := require.New(t)
	mp := MemoryUsedPercent()
	r.True(mp > 0 && mp < 1)

	cp := CPUUsedPercent()
	r.True(cp > 0 && cp < 1)

	dp := DiskUsedInfo()
	fmt.Println(dp)
	r.True(len(dp) > 0)

	nii := NetworkIOInfo()
	fmt.Println(nii)
	r.True(nii != nil)

	dii := DiskIOInfo()
	fmt.Println(dii)
	r.True(len(dii) > 0)

	fmt.Println(NumGoroutine())
	fmt.Println(SelfInfo())
	fmt.Println(StatsInfo())
}
