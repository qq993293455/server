package logger

import (
	"net/url"
	"strings"
	"time"

	"coin-server/common/values"
	"coin-server/common/values/env"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var DefaultLogger *Logger

func SetDefaultLogger(log *Logger) {
	DefaultLogger = log
}

//
//func init() {
//	DefaultLogger = MustNew(zap.DebugLevel, &Options{
//		Console:     "",
//		FilePath:    nil,
//		RemoteAddr:  nil,
//		InitFields:  nil,
//		Development: true,
//	})
//}

// systemClock implements default Clock that uses system time.
type loggerClock struct{}

func (loggerClock) Now() time.Time {
	if env.GetInteger(env.LOG_UTC_TIME) > 0 {
		return time.Now().UTC()
	}
	return time.Now()
}

func (loggerClock) NewTicker(duration time.Duration) *time.Ticker {
	return time.NewTicker(duration)
}

type Logger struct {
	log *zap.Logger
	//as  *AsyncSink
	discard bool
}

func (this_ *Logger) Debug(msg string, fields ...zap.Field) {
	if this_.discard {
		return
	}
	this_.log.Debug(msg, fields...)
}

func (this_ *Logger) Info(msg string, fields ...zap.Field) {
	if this_.discard {
		return
	}
	this_.log.Info(msg, fields...)
}

func (this_ *Logger) Warn(msg string, fields ...zap.Field) {
	if this_.discard {
		return
	}
	this_.log.Warn(msg, fields...)
}

func (this_ *Logger) Error(msg string, fields ...zap.Field) {
	if this_.discard {
		return
	}
	this_.log.Error(msg, fields...)
}

func (this_ *Logger) With(options ...zap.Option) *Logger {
	if this_.discard {
		return this_
	}
	newLog := this_.log.WithOptions(options...)
	return &Logger{log: newLog}
}

func (this_ *Logger) WithFields(fields ...zap.Field) *Logger {
	newLog := this_.log.With(fields...)
	return &Logger{log: newLog}
}

func (this_ *Logger) WithTrace(traceId string, roleId values.RoleId) *TraceLogger {
	if this_.discard {
		return &TraceLogger{discard: true}
	}

	tl := &TraceLogger{
		fields: make([]zap.Field, 0, 32),
	}
	tl.log = this_.log.WithOptions(zap.AddCallerSkip(2), zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return tl
	}))
	tl.Core = this_.log.Core()

	tl.fields = append(tl.fields, zap.String("trace_id", traceId),
		zap.String("roleId", roleId))
	return tl
}

func (this_ *Logger) WithTraceEmpty() *TraceLogger {
	if this_.discard {
		return &TraceLogger{discard: true}
	}

	tl := &TraceLogger{
		fields: make([]zap.Field, 0, 32),
	}
	tl.log = this_.log.WithOptions(zap.AddCallerSkip(2), zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return tl
	}))
	tl.Core = this_.log.Core()

	return tl
}

func (this_ *Logger) WithGnet(options ...zap.Option) *zap.SugaredLogger {
	newLog := this_.log.WithOptions(options...)
	return newLog.Sugar()
}

func (this_ *Logger) Sync() error {
	if this_.discard {
		return nil
	}
	return this_.log.Sync()
}

type Options struct {
	// 如果没有其他IO，会强制使用控制台输出
	Console    string
	FilePath   []string
	RemoteAddr []string
	// 默认字段
	InitFields map[string]interface{}

	// 如果是production ,则以json格式输出，否则以console格式输出
	Development bool

	RootCallPath string

	Discard bool
}

var _pool = buffer.NewPool()

func rootPath(rootStr string, ec zapcore.EntryCaller) string {
	if !ec.Defined {
		return "undefined"
	}
	idx := strings.Index(ec.File, rootStr)
	if idx == -1 {
		return ec.FullPath()
	}
	buf := _pool.Get()
	// Keep everything after the penultimate separator.
	buf.AppendString(ec.File[idx+len(rootStr)+1:])
	buf.AppendByte(':')
	buf.AppendInt(int64(ec.Line))
	caller := buf.String()
	buf.Free()
	return caller
}

func MustNew(level zapcore.Level, opt *Options) *Logger {
	envLevel := env.GetLogLevel()
	if envLevel > level {
		level = envLevel
	}
	if opt.Discard {
		return &Logger{discard: opt.Discard}
	}
	encoding := "json"
	encodeCall := zapcore.FullCallerEncoder
	if opt.RootCallPath != "" {
		encodeCall = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(rootPath(opt.RootCallPath, caller))
		}
	}
	ec := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   encodeCall,
	}

	if opt.Development {
		encoding = "console"
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	canColor := true
	//outPath := []string{AsyncSinkName + "://"}
	//as := NewAsyncSink(opt.RemoteAddr, opt.FilePath, opt.Console)
	//err := zap.RegisterSink(AsyncSinkName, func(url *url.URL) (zap.Sink, error) {
	//	return as, nil
	//})
	//if err != nil {
	//	panic(err)
	//}
	outPath := append(opt.FilePath, opt.RemoteAddr...)
	if opt.Console != "" {
		outPath = append(outPath, opt.Console)
	}
	if len(opt.RemoteAddr) > 0 {
		canColor = false
	}
	if len(opt.FilePath) > 0 {
		canColor = false
	}

	if !canColor {
		ec.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         encoding,
		EncoderConfig:    ec,
		OutputPaths:      outPath,
		ErrorOutputPaths: []string{"stderr"},
	}
	cfg.InitialFields = opt.InitFields

	l, err := cfg.Build(zap.AddStacktrace(zap.ErrorLevel), zap.AddCallerSkip(1), zap.WithClock(&loggerClock{}))
	if err != nil {
		panic(err) // 这里不允许出错
	}
	return &Logger{log: l}
}

func MustNewAsync(level zapcore.Level, opt *Options) *Logger {

	envLevel := env.GetLogLevel()
	if envLevel > level {
		level = envLevel
	}

	if env.GetInteger(env.LOG_STDOUT) <= 0 {
		opt.Console = ""
	}

	if opt.Discard {
		return &Logger{discard: opt.Discard}
	}
	encoding := "json"
	encodeCall := zapcore.FullCallerEncoder
	if opt.RootCallPath != "" {
		encodeCall = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(rootPath(opt.RootCallPath, caller))
		}
	}
	ec := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   encodeCall,
	}

	if opt.Development {
		encoding = "console"
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	canColor := true
	outPath := []string{AsyncSinkName + "://"}
	as := NewAsyncSink(opt.RemoteAddr, opt.FilePath, opt.Console)
	err := zap.RegisterSink(AsyncSinkName, func(url *url.URL) (zap.Sink, error) {
		return as, nil
	})
	if err != nil {
		panic(err)
	}

	if len(opt.RemoteAddr) > 0 {
		canColor = false
	}
	if len(opt.FilePath) > 0 {
		canColor = false
	}

	if !canColor {
		ec.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         encoding,
		EncoderConfig:    ec,
		OutputPaths:      outPath,
		ErrorOutputPaths: []string{"stderr"},
	}
	cfg.InitialFields = opt.InitFields

	l, err := cfg.Build(zap.AddStacktrace(zap.ErrorLevel), zap.AddCallerSkip(1), zap.WithClock(&loggerClock{}))
	if err != nil {
		panic(err) // 这里不允许出错
	}
	return &Logger{log: l}
}
