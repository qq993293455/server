package gopool

import (
	"fmt"

	"coin-server/common/logger"

	"github.com/panjf2000/ants/v2"
)

var pool *ants.Pool

func init() {
	var err error
	pool, err = ants.NewPool(50000, ants.WithLogger(log{}), ants.WithPanicHandler(func(interface{}) {
		l := log{}
		l.Printf("ants panic:%v", err)
	}))
	if err != nil {
		panic(err)
	}
}

type log struct{}

func (log) Printf(format string, args ...interface{}) {
	if logger.DefaultLogger != nil {
		logger.DefaultLogger.Error(fmt.Sprintf(format, args...))
	} else {
		fmt.Printf(format, args...)
	}
}

func Submit(f func()) {
	_ = pool.Submit(f)
}
