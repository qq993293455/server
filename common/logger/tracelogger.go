package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type TraceLogger struct {
	zapcore.Core
	fields  []zap.Field
	log     *zap.Logger
	stack   string
	discard bool
}

func (this_ *TraceLogger) AddStack(stack string) *TraceLogger {
	this_.stack = stack
	return this_
}

//func (this_ *TraceLogger) Clone() *TraceLogger {
//	return &TraceLogger{
//		Core:    this_.Core,
//		fields:  this_.fields,
//		log:     this_.log,
//		discard: this_.discard,
//	}
//}

var zapTraceLoggerPool = sync.Pool{New: func() interface{} {
	return DefaultLogger.WithTraceEmpty()
}}

func GetTraceLoggerWith(traceId, roleId string) *TraceLogger {
	l := zapTraceLoggerPool.Get().(*TraceLogger)
	l.ResetInitFiledS(traceId, roleId)
	return l
}

func GetTraceLogger() *TraceLogger {
	l := zapTraceLoggerPool.Get().(*TraceLogger)
	return l
}

func (this_ *TraceLogger) ReleasePool() {
	this_.fields = this_.fields[:0]
	zapTraceLoggerPool.Put(this_)
}

func (this_ *TraceLogger) ResetInitFiledS(traceId, roleId string) {
	this_.fields = this_.fields[:0]
	this_.fields = append(this_.fields, zap.String("trace_id", traceId), zap.String("role_id", roleId))
}

func (this_ *TraceLogger) poolFunc(f func(msg string, fields ...zap.Field), msg string, fields ...zap.Field) {
	if this_.discard {
		return
	}
	this_.fields = append(this_.fields, fields...)
	f(msg, this_.fields...)
	this_.fields = this_.fields[:2]
}

func (this_ *TraceLogger) Debug(msg string, fields ...zap.Field) {
	if this_.discard {
		return
	}
	this_.poolFunc(func(msg string, fields ...zap.Field) {
		this_.log.Debug(msg, fields...)
	}, msg, fields...)
}

func (this_ *TraceLogger) Info(msg string, fields ...zap.Field) {
	if this_.discard {
		return
	}
	this_.poolFunc(func(msg string, fields ...zap.Field) {
		this_.log.Info(msg, fields...)
	}, msg, fields...)
}

func (this_ *TraceLogger) Warn(msg string, fields ...zap.Field) {
	if this_.discard {
		return
	}
	this_.poolFunc(func(msg string, fields ...zap.Field) {
		this_.log.Warn(msg, fields...)
	}, msg, fields...)
}

func (this_ *TraceLogger) Error(msg string, fields ...zap.Field) {
	if this_.discard {
		return
	}
	this_.poolFunc(func(msg string, fields ...zap.Field) {
		this_.log.Error(msg, fields...)
	}, msg, fields...)
}

func (this_ *TraceLogger) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	// Let the wrapped Core decide whether to log this message or not. This
	// also gives the downstream a chance to register itself directly with the
	// CheckedEntry.
	if downstream := this_.Core.Check(ent, ce); downstream != nil {
		if this_.stack != "" {
			downstream.Stack = this_.stack
			this_.stack = ""
		}
		return downstream
	}
	return ce
}
