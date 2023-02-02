//go:build !windows && !plan9
// +build !windows,!plan9

package ulimit

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"syscall"
)

var rLimit sync.Once

func SetRLimit() error {
	var err error

	rLimit.Do(func() {
		err = setRLimit()
	})

	return err
}

func setRLimit() error {
	fmt.Println(runtime.GOOS)
	return set(uint64(runtime.NumCPU() * 50000))
}

func set(limit uint64) error {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return fmt.Errorf("error getting rlimit: %w", err)
	}

	oldMax := rLimit.Max
	if rLimit.Cur < limit {
		if rLimit.Max < limit {
			rLimit.Max = limit
		}
		rLimit.Cur = limit
	}

	// If we're on darwin, work around the fact that Getrlimit reports the wrong
	// value. See https://github.com/golang/go/issues/30401
	if runtime.GOOS == "darwin" && rLimit.Cur > 10240 {
		// The max file limit is 10240, even though the max returned by
		// Getrlimit is 1<<63-1. This is OPEN_MAX in sys/syslimits.h.
		rLimit.Max = 10240
		rLimit.Cur = 10240
	}

	// Try updating the limit. If it fails, try using the previous maximum
	// instead of our new maximum. Not all users have permissions to increase
	// the maximum.
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		rLimit.Max = oldMax
		rLimit.Cur = oldMax
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			return fmt.Errorf("error setting ulimit: %w", err)
		}
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return fmt.Errorf("error getting rlimit: %w", err)
	}
	fmt.Printf("rlimit info: cur %d, max %d\n", rLimit.Cur, rLimit.Max)
	return nil
}

func SetOpenCoreDump() error {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_CORE, &rLimit)
	if err != nil {
		return fmt.Errorf("error getting rlimit RLIMIT_CORE: %w", err)
	}

	ulimited := math.MaxInt
	rLimit.Cur = uint64(ulimited)
	rLimit.Max = uint64(ulimited)

	if err := syscall.Setrlimit(syscall.RLIMIT_CORE, &rLimit); err != nil {
		return fmt.Errorf("error setting rlimit RLIMIT_CORE: %w", err)
	}

	err = syscall.Getrlimit(syscall.RLIMIT_CORE, &rLimit)
	if err != nil {
		return fmt.Errorf("error getting rlimit: %w", err)
	}
	fmt.Printf("rlimit coredump file size info: cur %d, max %d\n", rLimit.Cur, rLimit.Max)
	return nil
}
