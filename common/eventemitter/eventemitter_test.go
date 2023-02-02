package eventemitter

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"coin-server/common/logger"

	"go.uber.org/zap/zapcore"
)

func TestEventEmitter(t *testing.T) {
	log := logger.MustNew(zapcore.DebugLevel, &logger.Options{
		Console:      "",
		FilePath:     nil,
		RemoteAddr:   nil,
		InitFields:   nil,
		Development:  true,
		RootCallPath: "",
	})
	el := NewEventEmitter(log)
	count := int64(0)

	el.On("test", func(data interface{}) {
		fmt.Println("test", data)
		atomic.AddInt64(&count, 1)
	})

	el.Emit("test", "hello")
	time.Sleep(10 * time.Millisecond)
	if count != 1 {
		t.Error("count is not 1")
	}

	el.EmitFunc(func() {
		fmt.Println("test")
		atomic.AddInt64(&count, 1)
	})
	time.Sleep(10 * time.Millisecond)
	if count != 2 {
		t.Error("count is not 2")
	}
}
