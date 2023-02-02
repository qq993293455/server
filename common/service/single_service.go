package service

import (
	"strconv"
	"strings"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/eventloop"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	"coin-server/common/natsclient"
	"coin-server/common/network/stdtcp"
	"coin-server/common/proto/models"
	"coin-server/common/protocol"
	"coin-server/common/startcheck"
	"coin-server/common/timer"
	"coin-server/common/values"
	env2 "coin-server/common/values/env"

	"github.com/rs/xid"
	"go.uber.org/zap"
)

type SingleService struct {
	nc         *natsclient.ClusterClient
	log        *logger.Logger
	serverId   values.ServerId
	serverType models.ServerType
	el         *eventloop.EventLoop
	handler    *handler.Handler
}

func NewSingleService(
	urls []string,
	log *logger.Logger,
	serverId values.ServerId,
	serverType models.ServerType,
	isDebug bool,
	openLogMid bool,
	wares ...handler.MiddleWare,
) *SingleService {
	if serverType == models.ServerType_DungeonMatchServer || serverId > 0 {
		startcheck.StartCheck(serverType, serverId)
	}
	nc := natsclient.NewClusterClient(serverType, serverId, urls, log)
	el := eventloop.NewEventLoop(log)
	midS := make([]handler.MiddleWare, 0, 4)
	if openLogMid {
		midS = append(midS, handler.Logger, handler.LogServer, handler.LogServer2)
	}
	midS = append(midS, handler.Recover, handler.Tracing, handler.UnLocker, handler.DoWriteDB)
	midS = append(midS, wares...)
	s := &SingleService{
		nc:         nc,
		log:        log,
		serverId:   serverId,
		serverType: serverType,
		el:         el,
		handler:    handler.NewHandler(el, isDebug, env2.GetIsOpenMiddleError(), midS...),
	}
	return s
}

func (this_ *SingleService) RegisterFunc(desc string, function interface{}, midWares ...handler.MiddleWare) {
	this_.handler.RegisterFunc(desc, function, midWares...)
}

func (this_ *SingleService) RegisterEvent(desc string, function interface{}, midWares ...handler.MiddleWare) {
	this_.handler.RegisterEvent(desc, function, midWares...)
}

func (this_ *SingleService) Group(mid handler.MiddleWare, midWares ...handler.MiddleWare) *handler.Handler {
	return this_.handler.Group(mid, midWares...)
}

func (this_ *SingleService) Start(f func(interface{})) {
	es := this_.handler.GetHandlers()
	for _, v := range es {
		this_.log.Info("registered : " + v.String())
	}
	els := eventlocal.GetAllEventLocal()
	for _, v := range els {
		this_.log.Info("subscribe event_local : " + v)
	}
	subjS := this_.handler.GetSubjArray()
	for _, v := range subjS {
		subj := v + ".>"
		if this_.serverId != 0 {
			subj = v + "." + strconv.Itoa(int(this_.serverId)) + ".>"
		}
		this_.log.Info("nats sub", zap.String("subj", subj))
		this_.Subscribe(subj)
	}
	bcs := strings.Join([]string{protocol.TopicBroadcast, this_.serverType.String()}, ".")
	this_.log.Info("nats sub", zap.String("subj", bcs))
	this_.Subscribe(bcs)

	this_.el.Start(func(event interface{}) {
		switch t := event.(type) {
		case *ctx.Context:
			this_.dispatchCtx(t)
		default:
			if f != nil {
				f(event)
			}
		}
	})
}

func (this_ *SingleService) Subscribe(subj string) {
	this_.nc.SubscribeHandler(subj, func(ctx *ctx.Context) {
		this_.el.PostEventQueue(ctx)
	})
}

var PongMsgName = (&models.PONG{}).XXX_MessageName()

const TCPSessionKey = "tcp_session"

func (this_ *SingleService) HandleTCPData(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) *errmsg.ErrMsg {
	if msgName == PongMsgName {
		return nil
	}

	msg := msgcreate.NewMessage(msgName)
	header := &models.ServerHeader{}
	err := protocol.DecodeInternal(frame, header, msg)
	if err != nil {
		return err
	}
	if header.StartTime == 0 {
		header.StartTime = timer.Now().UnixNano()
	}
	if header.TraceId == "" {
		header.TraceId = xid.NewWithTime(timer.Now()).String()
	}
	uc := ctx.GetPoolContext()
	uc.ServerHeader = header
	uc.Req = msg
	uc.SetValue(TCPSessionKey, session)
	if uc.TraceLogger == nil {
		uc.TraceLogger = this_.log.WithTrace(header.TraceId, header.RoleId)
	} else {
		uc.TraceLogger.ResetInitFiledS(header.TraceId, header.RoleId)
	}

	if rpcIndex > 0 {
		uc.RespondFunc = func(respMsgName string, respData []byte) *errmsg.ErrMsg {
			return session.RPCResponseTCP(rpcIndex, respMsgName, respData)
		}
	}
	this_.el.PostEventQueue(uc)
	return nil
}

func (this_ *SingleService) NewHeader(roleId values.RoleId, c *ctx.Context) *models.ServerHeader {
	newC := &models.ServerHeader{}
	newC.RoleId = roleId
	newC.ServerId = this_.serverId
	newC.ServerType = this_.serverType
	if c != nil {
		newC.StartTime = c.StartTime
		newC.TraceId = c.TraceId
		newC.RuleVersion = c.RuleVersion
	}
	return newC
}

func (this_ *SingleService) dispatchCtx(uc *ctx.Context) {
	if uc.TraceLogger == nil {
		uc.TraceLogger = this_.log.WithTrace(uc.TraceId, "")
	}
	if uc.F != nil {
		uc.StartTime = timer.Now().UnixNano()
	}
	this_.handler.Handle(uc)
}

// AfterFuncCtx 会继承ctx里面的ServerHeader
func (this_ *SingleService) AfterFuncCtx(c *ctx.Context, duration time.Duration, f func(ctx *ctx.Context)) {
	h := *c.ServerHeader
	h.ServerId = this_.serverId
	h.ServerType = this_.serverType
	ac := ctx.GetPoolContext()
	ac.F = f
	ac.ServerHeader = &h
	ac.TraceLogger.ResetInitFiledS(c.TraceId, c.RoleId)
	timer.AfterFunc(duration, func() {
		this_.el.PostEventQueue(ac)
	})
}

func (this_ *SingleService) AfterFunc(duration time.Duration, f func(ctx *ctx.Context)) {
	timer.AfterFunc(duration, func() {
		h := &models.ServerHeader{}
		h.ServerId = this_.serverId
		h.ServerType = this_.serverType
		h.TraceId = xid.NewWithTime(time.Now()).String()
		ac := ctx.GetPoolContext()
		ac.F = f
		ac.ServerHeader = h
		ac.TraceLogger.ResetInitFiledS(h.TraceId, "")
		this_.el.PostEventQueue(ac)
	})
}

// UntilFuncCtx 会继承ctx里面的ServerHeader
func (this_ *SingleService) UntilFuncCtx(c *ctx.Context, t time.Time, f func(ctx *ctx.Context)) {
	h := *c.ServerHeader
	h.ServerId = this_.serverId
	h.ServerType = this_.serverType
	ac := ctx.GetPoolContext()
	ac.F = f
	ac.ServerHeader = &h
	ac.TraceLogger.ResetInitFiledS(c.TraceId, c.RoleId)
	timer.UntilFunc(t, func() {
		this_.el.PostEventQueue(ac)
	})
}

func (this_ *SingleService) UntilFunc(t time.Time, f func(ctx *ctx.Context)) {
	timer.UntilFunc(t, func() {
		h := &models.ServerHeader{}
		h.ServerId = this_.serverId
		h.ServerType = this_.serverType
		h.TraceId = xid.NewWithTime(time.Now()).String()
		ac := ctx.GetPoolContext()
		ac.F = f
		ac.ServerHeader = h
		ac.TraceLogger.ResetInitFiledS(h.TraceId, "")
		this_.el.PostEventQueue(ac)
	})
}

func (this_ *SingleService) TickFuncCtx(c *ctx.Context, d time.Duration, f func(ctx *ctx.Context) bool) {
	this_.AfterFuncCtx(c, d, func(ctx *ctx.Context) {
		if f(ctx) {
			this_.TickFuncCtx(ctx, d, f)
		}
	})
}

func (this_ *SingleService) TickFunc(d time.Duration, f func(ctx *ctx.Context) bool) {
	this_.AfterFunc(d, func(ctx *ctx.Context) {
		if f(ctx) {
			this_.TickFunc(d, f)
		}
	})
}

func (this_ *SingleService) GetNatsClient() *natsclient.ClusterClient {
	return this_.nc
}

func (this_ *SingleService) Close() {
	this_.nc.Close()
	this_.el.Stop()
	this_.nc.Shutdown()
}

func (this_ *SingleService) GetEventLoop() *eventloop.EventLoop {
	return this_.el
}
