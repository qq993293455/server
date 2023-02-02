package eventloop

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"coin-server/common/errmsg"
	"coin-server/common/gopool"
	"coin-server/common/logger"

	"github.com/stretchr/testify/require"

	"go.uber.org/zap/zapcore"
)

func TestNewEventLoop(t *testing.T) {
	log := logger.MustNew(zapcore.DebugLevel, &logger.Options{
		Console:      "",
		FilePath:     nil,
		RemoteAddr:   nil,
		InitFields:   nil,
		Development:  true,
		RootCallPath: "",
	})
	el := NewEventLoop(log)
	count := int64(0)
	go func() {
		for {
			el.PostFuncQueue(func() {
				atomic.AddInt64(&count, 1)
			})
		}
	}()

	el.Start(func(event interface{}) {

	})
	for {
		fmt.Println(atomic.LoadInt64(&count))
		time.Sleep(time.Second)
	}
}

type testS1 struct {
	int1 int64
	str  string
}

type testS2 struct {
	int1 int64
	str  string
}

func TestChanEventLoop_Start(t *testing.T) {
	go http.ListenAndServe(":6063", nil)
	r := require.New(t)
	log := logger.MustNew(zapcore.DebugLevel, &logger.Options{
		Console:      "stdout",
		FilePath:     nil,
		RemoteAddr:   nil,
		InitFields:   nil,
		Development:  true,
		RootCallPath: "",
	})
	el := NewChanEventLoop(log)
	RegisterFuncChanEventLoop[int, string](el, func(i int) (string, *errmsg.ErrMsg) {
		return strconv.Itoa(i), nil
	})

	RegisterFuncChanEventLoop[string, string](el, func(i string) (string, *errmsg.ErrMsg) {
		return i, nil
	})

	RegisterFuncChanEventLoop[testS1, testS2](el, func(i testS1) (testS2, *errmsg.ErrMsg) {
		return testS2{}, nil
	})

	RegisterFuncChanEventLoop[*testS1, *testS2](el, func(i *testS1) (*testS2, *errmsg.ErrMsg) {
		return &testS2{}, nil
	})

	go func() {
		el.Start(func(event interface{}) {

		})
	}()

	out, err := RequestChanEventLoop[int, string](el, 1)
	r.Equal((*errmsg.ErrMsg)(nil), err)
	r.Equal("1", out)

	out, err = RequestChanEventLoop[string, string](el, "1")
	r.Equal((*errmsg.ErrMsg)(nil), err)
	r.Equal("1", out)

	out1, err := RequestChanEventLoop[testS1, testS2](el, testS1{})
	r.Equal((*errmsg.ErrMsg)(nil), err)
	r.Equal(testS2{}, out1)

	out2, err := RequestChanEventLoop[*testS1, *testS2](el, &testS1{})
	r.Equal((*errmsg.ErrMsg)(nil), err)
	r.Equal(testS2{}, *out2)

	now := time.Now()
	count := int64(0)
	for i := 0; i < 10000000; i++ {
		gopool.Submit(func() {
			for x := 0; x < 100; x++ {
				CallChanEventLoop[*testS1, *testS2](el, &testS1{}, func(s1 *testS1) (*testS2, *errmsg.ErrMsg) {
					return &testS2{}, nil
				})

				nc := atomic.AddInt64(&count, 1)
				if nc%1000000 == 0 {
					fmt.Println(time.Now().Sub(now), nc)
				}
			}

		})

	}

	fmt.Println(time.Now().Sub(now))
}
