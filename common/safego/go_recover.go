package safego

import (
	"coin-server/common/logger"

	"go.uber.org/zap"
)

func Recover(panicHandler func(e interface{})) {
	if e := recover(); e != nil {
		if panicHandler != nil {
			panicHandler(e)
		}
	}
}

func RecoverWithLogger(log *logger.Logger) {
	if e := recover(); e != nil {
		log.Error("panic", zap.Any("panic info", e))
	}
}

func RecoverWithTraceLogger(log *logger.TraceLogger) {
	if e := recover(); e != nil {
		log.Error("panic", zap.Any("panic info", e))
	}
}

func GOWithLogger(log *logger.Logger, f func()) {
	go func() {
		defer RecoverWithLogger(log)
		f()
	}()
}

func GO(errF func(interface{}), f func()) {
	go func() {
		defer Recover(errF)
		f()
	}()
}
