package handler

import (
	"fmt"
	"strings"
	"unsafe"

	"coin-server/common/msgcreate"
	"coin-server/common/statistical2"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/metrics"
	"coin-server/common/proto/gatewaytcp"
	"coin-server/common/proto/models"
	"coin-server/common/statistical"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values/env"

	"github.com/gogo/protobuf/proto"
	"github.com/petermattis/goid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var kafkaTracer = otel.Tracer("statistic/kafka")

func Recover(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		defer func() {
			if e := recover(); e != nil {
				err = errmsg.NewErrorPanic(e)
			}
		}()
		return next(ctx)
	}
}

func OpenGMHandler(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		if env.GetString(env.OPEN_GM_HANDLER) == "1" {
			return next(ctx)
		}
		msgName := msgcreate.MessageName(ctx.Req)
		n := strings.LastIndexByte(msgName, '.')
		if n != -1 {
			name := msgName[n+1:]
			if strings.HasPrefix(name, "Cheat") {
				return errmsg.NewNormalErr("gm_handler_closed", "gm handler closed")
			}
		}

		return next(ctx)
	}
}

var noLog = proto.MessageName(&gatewaytcp.GatewayStdTcp_PushManyToClient{})

func Logger(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) *errmsg.ErrMsg {
		gid := goid.Get()
		now := timer.Now()
		var reqMsgName string
		if ctx.Req != nil {
			reqMsgName = proto.MessageName(ctx.Req)
			ctx.Info("handle start", zap.Int64("goid", gid), zap.String("uri", reqMsgName), zap.Any("req_data", ctx.Req), zap.Any("header", ctx.ServerHeader))
		} else if ctx.F != nil {
			reqMsgName = "timer_func_call"
			ctx.Info("handle timer start", zap.Int64("goid", gid), zap.String("uri", reqMsgName), zap.Any("header", ctx.ServerHeader))
		}

		err := next(ctx)

		fs := zapFiledPool.Get().([]zap.Field)
		fs = fs[:0]
		defer zapFiledPool.Put(fs)
		fs = append(fs, zap.Int64("goid", gid))
		fs = append(fs, zap.String("req", reqMsgName))
		if ctx.Req != nil {
			fs = append(fs, zap.Any("req_data", ctx.Req))
		}
		if ctx.Resp != nil {
			fs = append(fs, zap.String("resp", proto.MessageName(ctx.Resp)))
			fs = append(fs, zap.Any("resp_data", ctx.Resp))
		}
		if len(ctx.OtherRequest) > 0 {
			fs = append(fs, zap.Strings("other_request", ctx.OtherRequest))
		}
		if len(ctx.OtherMsg) > 0 {
			for _, v := range ctx.OtherMsg {
				fs = append(fs, zap.String("other_msg", proto.MessageName(v)))
				fs = append(fs, zap.Any("other_msg_data", v))
			}
		}
		fs = append(fs, zap.Duration("cost", timer.Now().Sub(now)))
		if err != nil {
			if len(err.StackStace) > 0 {
				ss := err.StackStace
				err.StackStace = nil
				fs = append(fs, zap.Error(err))
				ctx.AddStack(fmt.Sprintf("%+v", (*utils.Stack)(unsafe.Pointer(&ss)).StackTrace())).Warn("handle error", fs...)
			} else {
				fs = append(fs, zap.Error(err))
				ctx.Warn("handle error", fs...)
			}
		} else {
			ctx.Info("handle success", fs...)
		}
		return err
	}
}

func LogServer(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) *errmsg.ErrMsg {
		err := next(ctx)
		if err != nil {
			return err
		}
		if statistical.Disabled("all") {
			return nil
		}

		if env.OpenTracing() {
			var span trace.Span
			ctx.Context, span = kafkaTracer.Start(ctx.Context, "StatisticKafkaWrite ")
			defer span.End()
		}

		ls := ctx.NewLogServer()
		if len(ls.GetCache()) == 0 {
			return nil
		}

		realErr := statistical.Flush(ls)
		if realErr != nil {
			return errmsg.NewInternalErr(realErr.Error())
		}
		return nil
	}
}

func LogServer2(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) *errmsg.ErrMsg {
		err := next(ctx)
		if err != nil {
			return err
		}
		if statistical2.Disabled("all") {
			return nil
		}

		if env.OpenTracing() {
			var span trace.Span
			ctx.Context, span = kafkaTracer.Start(ctx.Context, "Statistic2KafkaWrite ")
			defer span.End()
		}

		ls2 := ctx.NewLogServer2()
		if len(ls2.GetCache()) == 0 {
			return nil
		}

		realErr := statistical2.Flush(ls2)
		if realErr != nil {
			return errmsg.NewInternalErr(realErr.Error())
		}
		return nil
	}
}

func DoWriteDB(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		err = next(ctx)
		if err == nil {
			orm := ctx.GetOrmForMiddleWare()
			if orm != nil {
				err = orm.Do()
			}
		}
		return
	}
}

func UnLocker(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		defer func() {
			e := ctx.DRUnlock()
			if e != nil && err == nil {
				err = e
			}
		}()
		err = next(ctx)
		return err
	}
}

func Tracing(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		if !env.OpenTracing() {
			return next(ctx)
		}
		var span trace.Span
		defer func() {
			if span != nil {
				span.End()
			}
		}()
		tp := otel.GetTracerProvider()
		if tp != nil && ctx.Req != nil {
			name := models.ServerType_name[int32(ctx.ServerType)]
			tr := tp.Tracer(name)
			ctx.Context, span = tr.Start(ctx.Context, proto.MessageName(ctx.Req))
			span.SetAttributes(attribute.Key("roleId").String(ctx.RoleId))
			span.SetAttributes(attribute.Key("traceId").String(ctx.TraceId))
		}
		return next(ctx)
	}
}

func Metrics(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		if !metrics.IsOpen() {
			return next(ctx)
		}

		now := timer.Now()
		var reqMsgName string
		if ctx.Req != nil {
			reqMsgName = proto.MessageName(ctx.Req)
			//ctx.Debug("handle start", zap.String("uri", reqMsgName), zap.Any("req_data", ctx.Req), zap.Any("header", ctx.ServerHeader))
		} else if ctx.F != nil {
			reqMsgName = "timer_func_call"
			//ctx.Debug("handle timer start", zap.String("uri", reqMsgName), zap.Any("header", ctx.ServerHeader))
		}
		metrics.RequestsTotal.WithLabelValues(reqMsgName).Inc()

		err = next(ctx)
		if err != nil {
			metrics.RequestsErrorTotal.WithLabelValues(err.ErrCode.String(), err.ErrMsg).Inc()
		}

		metrics.RequestsLatencyHistogram.WithLabelValues(reqMsgName).Observe(timer.Now().Sub(now).Seconds())

		return err
	}
}
