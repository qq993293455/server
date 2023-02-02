package service

import (
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/eventloop"
	"coin-server/common/gopool"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	"coin-server/common/natsclient"
	"coin-server/common/network/stdtcp"
	"coin-server/common/proto/broadcast"
	"coin-server/common/proto/models"
	readpb "coin-server/common/proto/role_state_read"
	"coin-server/common/protocol"
	"coin-server/common/safego"
	"coin-server/common/startcheck"
	system_info "coin-server/common/system-info"
	"coin-server/common/timer"
	"coin-server/common/values"
	env2 "coin-server/common/values/env"

	"github.com/gogo/protobuf/types"
	"github.com/nats-io/nats.go"

	"github.com/rs/xid"

	"github.com/gogo/protobuf/proto"

	"go.uber.org/zap"
)

type roleChannel struct {
	closed     int32
	queue      chan *ctx.Context
	lastTime   time.Time
	roleId     values.RoleId
	gateId     values.GatewayId
	online     bool
	eventCount int64
}

type Service struct {
	system_info.ServiceStatusChangeBaseEvent
	nc              *natsclient.ClusterClient
	log             *logger.Logger
	serverId        values.ServerId
	serverType      models.ServerType
	el              *eventloop.EventLoop
	roleDispatchMap map[values.RoleId]*roleChannel
	handler         *handler.Handler
	roleCount       int64
	maxCount        int64
	uniqueId        string
	statusServer    *system_info.ServiceMgr
	onServiceNew    func(s *broadcast.Stats_ServerStats)
	onServiceLose   func(s *broadcast.Stats_ServerStats)
}

func (this_ *Service) SetOnServerNew(onServiceNew func(s *broadcast.Stats_ServerStats)) {
	this_.onServiceNew = onServiceNew
}

func (this_ *Service) SetOnServerLose(onServiceLose func(s *broadcast.Stats_ServerStats)) {
	this_.onServiceLose = onServiceLose
}

func NewService(
	urls []string,
	log *logger.Logger,
	serverId values.ServerId,
	serverType models.ServerType,
	isDebug bool,
	openLogMid bool,
	wares ...handler.MiddleWare,
) *Service {
	if serverType == models.ServerType_DungeonMatchServer || serverId > 0 {
		startcheck.StartCheck(serverType, serverId)
	}
	nc := natsclient.NewClusterClient(serverType, serverId, urls, log)
	el := eventloop.NewEventLoop(log)
	midS := make([]handler.MiddleWare, 0, 4)
	if openLogMid {
		midS = append(midS, handler.Logger, handler.LogServer, handler.LogServer2)
	}
	midS = append(midS, handler.Recover, handler.OpenGMHandler, handler.Metrics, handler.Tracing, handler.UnLocker, handler.DoWriteDB)
	midS = append(midS, wares...)

	s := &Service{
		nc:              nc,
		log:             log,
		serverId:        serverId,
		serverType:      serverType,
		el:              el,
		roleDispatchMap: map[values.RoleId]*roleChannel{},
		handler:         handler.NewHandler(el, isDebug, env2.GetIsOpenMiddleError(), midS...),
		maxCount:        int64(runtime.NumCPU() * 5000),
		uniqueId:        xid.New().String(),
	}
	s.statusServer = system_info.NewServiceMgr(s.uniqueId, log, s)
	serverIdHeader.ServerId = serverId
	serverIdHeader.ServerType = serverType
	return s
}

var statsMessageName = proto.MessageName(&broadcast.Stats_ServerStats{})

func (this_ *Service) startStats() {
	this_.nc.Subscribe(statsMessageName, func(msg *nats.Msg) {
		stats := &broadcast.Stats_ServerStats{}
		err := protocol.DecodeInternal(msg.Data, nil, stats)
		if err != nil {
			this_.log.Error("unmarshal stats error", zap.Error(err))
			return
		}
		this_.el.PostFuncQueue(func() {
			this_.statusServer.AddOrSet(stats)
		})
	})
	this_.el.TickQueue(time.Second*5, func() bool {
		this_.statusServer.CheckLose()
		return !this_.el.Stopped()
	})
	safego.GOWithLogger(this_.log, func() {
		this_.sendStats()
		t := time.NewTimer(time.Second * 1)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				this_.sendStats()
				t.Reset(time.Second * 1)
			}
		}
	})
}

func (this_ *Service) sendStats() {
	stats := &broadcast.Stats_ServerStats{
		UniqueId:   this_.uniqueId,
		ServerId:   this_.serverId,
		ServerType: this_.serverType,
		MaxCount:   this_.maxCount,
		CurrCount:  this_.GetRoleCount(),
		Timestamp:  timer.Now().UnixNano(),
	}
	si := system_info.StatsInfo()
	stats.Stats = si
	err := this_.nc.Publish(0, nil, stats)
	if err != nil {
		this_.log.Error("syncStats err", zap.Error(err), zap.Int64("server_id", this_.serverId), zap.String("server_type", this_.serverType.String()))
	}
}

func (this_ *Service) OnServiceNew(s *broadcast.Stats_ServerStats) {
	this_.ServiceStatusChangeBaseEvent.OnServiceNew(s)
	if this_.onServiceNew != nil {
		this_.onServiceNew(s)
	}
}

func (this_ *Service) OnServiceLose(s *broadcast.Stats_ServerStats) {
	this_.ServiceStatusChangeBaseEvent.OnServiceLose(s)
	if this_.onServiceLose != nil {
		this_.onServiceLose(s)
	}
}

func (this_ *Service) RegisterFunc(desc string, function interface{}, midWares ...handler.MiddleWare) {
	this_.handler.RegisterFunc(desc, function, midWares...)
}

func (this_ *Service) RegisterEvent(desc string, function interface{}, midWares ...handler.MiddleWare) {
	this_.handler.RegisterEvent(desc, function, midWares...)
}

func (this_ *Service) Group(mid handler.MiddleWare, midWares ...handler.MiddleWare) *handler.Handler {
	return this_.handler.Group(mid, midWares...)
}

func (this_ *Service) Start(f func(interface{}), syncStats bool) {
	es := this_.handler.GetHandlers()
	for _, v := range es {
		this_.log.Info("registered : " + v.String())
	}
	els := eventlocal.GetAllEventLocal()
	for _, v := range els {
		this_.log.Info("subscribe event_local : " + v)
	}
	subjS := this_.handler.GetSubjArray()
	// if len(subjS) == 0 {
	// 	panic("no subj found")
	// }
	for _, v := range subjS {
		if v == protocol.TopicBroadcast {
			continue
		}
		subj := v + ".>"
		if this_.serverId != 0 {
			subj = v + "." + strconv.Itoa(int(this_.serverId)) + ".>"
		}
		this_.log.Info("nats sub", zap.String("subj", subj))
		this_.Subscribe(subj)
	}
	if len(subjS) > 0 {
		bcs := strings.Join([]string{protocol.TopicBroadcast, this_.serverType.String(), ">"}, ".")
		this_.log.Info("nats sub", zap.String("subj", bcs))
		this_.SubscribeBroadcast(bcs)
	}

	this_.el.Start(func(event interface{}) {
		switch t := event.(type) {
		case *ctx.Context:
			this_.dispatchCtx(t)
		case []*readpb.RoleStateROnly_PushToClient:
			this_.dispatchPushMessages(t)
		case []*ctx.EventRemote:
			this_.dispatchEventRemote(t)
		default:
			if f != nil {
				f(event)
			}
		}
	})
	this_.startStats()
}

var serverIdHeader = &models.ServerHeader{}

func (this_ *Service) dispatchPushMessages(pms []*readpb.RoleStateROnly_PushToClient) {
	gopool.Submit(func() {
		if err := this_.nc.Publish(0, serverIdHeader, &readpb.RoleStateROnly_PushManyToClient{
			Pcs: pms,
		}); err != nil {
			this_.log.Error("push msg to role state server error", zap.Any("data", pms))
		}
	})
}

func (this_ *Service) dispatchEventRemote(list []*ctx.EventRemote) {
	gopool.Submit(func() {
		for _, evt := range list {
			if err := this_.nc.Publish(evt.ServerId, ctx.NewHeaderWithOutCtx(evt.RoleId, evt.UserId, this_.serverId, this_.serverType, evt.StartTime, evt.TraceId, evt.RuleVersion), evt.Data); err != nil {
				this_.log.Error("push remote event to other role error", zap.Any("evt", evt))
			}
		}
	})
}

// PushToAllOnlineClient 不走事务，调用则即时发送。
func (this_ *Service) PushToAllOnlineClient(msgs []proto.Message) {
	if len(msgs) > 0 {
		gopool.Submit(func() {
			msgList := make([]*types.Any, 0, len(msgs))
			for _, msg := range msgs {
				msgList = append(msgList, protocol.MsgToAny(msg))
			}
			if err := this_.nc.Publish(0, &models.ServerHeader{RoleId: "all"}, &readpb.RoleStateROnly_PushToAllClient{
				Messages: msgList,
			}); err != nil {
				this_.log.Error("push msg to role state server error", zap.Any("data", msgs))
			}
		})
	}
}

func (this_ *Service) Subscribe(subj string) {
	this_.nc.SubscribeHandler(subj, func(ctx *ctx.Context) {
		this_.el.PostEventQueue(ctx)
	})
}

func (this_ *Service) SubscribeBroadcast(subj string) {
	this_.nc.SubscribeBroadcast(subj, func(ctx *ctx.Context) {
		this_.el.PostEventQueue(ctx)
	})
}

func (this_ *Service) HandleTCPData(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) *errmsg.ErrMsg {
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
	uc.TraceLogger.ResetInitFiledS(header.TraceId, header.RoleId)

	if rpcIndex > 0 {
		uc.RespondFunc = func(respMsgName string, respData []byte) *errmsg.ErrMsg {
			return session.RPCResponseTCP(rpcIndex, respMsgName, respData)
		}
	}
	this_.el.PostEventQueue(uc)
	return nil
}

func (this_ *Service) NewHeader(roleId values.RoleId, c *ctx.Context) *models.ServerHeader {
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

func (this_ *Service) syncStats() {
	safego.GO(func(i interface{}) {
		this_.log.Error("syncStats panic", zap.Any("panic info", i))
		this_.syncStats()
	}, func() {
		this_.syncStatsOne(timer.Now())
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()
		for t := range ticker.C {
			this_.syncStatsOne(t)
		}
	})
}

func (this_ *Service) syncStatsOne(t time.Time) {
	ss := &broadcast.Stats_ServerStats{
		ServerId:   this_.serverId,
		ServerType: this_.serverType,
		MaxCount:   this_.maxCount,
		CurrCount:  this_.GetRoleCount(),
		Timestamp:  t.Unix(),
	}

	err := this_.nc.Publish(0, nil, ss)
	if err != nil {
		this_.log.Error("syncStats err", zap.Error(err), zap.Int64("server_id", this_.serverId), zap.String("server_type", this_.serverType.String()))
	}
}

func (this_ *Service) CloseRoleRoutine(roleId string) {
	this_.el.PostEventQueue(func() {
		c, ok := this_.roleDispatchMap[roleId]
		if ok && atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
			c.online = false
			delete(this_.roleDispatchMap, roleId)
			close(c.queue)
			this_.AddRoleCount(-1)
		}
	})
}

func (this_ *Service) GetRoleCount() int64 {
	return atomic.LoadInt64(&this_.roleCount)
}

func (this_ *Service) AddRoleCount(count int64) int64 {
	return atomic.AddInt64(&this_.roleCount, count)
}

func (this_ *Service) dispatchCtx(uc *ctx.Context) {
	roleId := uc.RoleId
	if roleId == "" {
		if uc.TraceLogger == nil {
			uc.TraceLogger = logger.GetTraceLoggerWith(uc.TraceId, "")
		}

		gopool.Submit(func() {
			if uc.F == nil {
				this_.handler.Handle(uc)
			} else {
				uc.StartTime = timer.Now().UnixNano()
				if uc.TraceLogger == nil {
					uc.TraceLogger = logger.GetTraceLoggerWith(uc.TraceId, "")
				}
				this_.handler.Handle(uc)
			}

		})
	} else {
		c, ok := this_.roleDispatchMap[roleId]
		if !ok {
			c = &roleChannel{
				queue:    make(chan *ctx.Context, 10),
				lastTime: timer.Now(),
				roleId:   roleId,
				gateId:   uc.GateId,
			}
			if uc.F == nil && uc.ServerHeader != nil && uc.GateId != 0 {
				c.online = true
			}
			safego.GOWithLogger(this_.log, func() {
				this_.log.Info("start role", zap.String("role_id", c.roleId))
				for v := range c.queue {
					if v.F == nil {
						this_.handler.Handle(v)
					} else {
						v.StartTime = timer.Now().UnixNano()
						v.TraceLogger.ResetInitFiledS(v.TraceId, v.RoleId)
						this_.handler.Handle(v)
					}
				}
				this_.log.Info("end role", zap.String("role_id", roleId))
			})

			this_.el.TickQueue(time.Minute, func() bool {
				if timer.Now().Sub(c.lastTime) >= time.Minute*2 {
					if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
						delete(this_.roleDispatchMap, roleId)
						close(c.queue)
						this_.AddRoleCount(-1)
					}
					return false
				}
				return true
			})
			this_.AddRoleCount(1)
			this_.roleDispatchMap[roleId] = c
		}
		if uc.F == nil && uc.ServerHeader != nil && uc.GateId != 0 {
			if c.gateId != uc.GateId {
				c.gateId = uc.GateId
			}
			if uc.GateId != 0 {
				c.online = true
			}
		}
		select {
		case c.queue <- uc:
		default:
			gopool.Submit(func() {
				uc.Error("user queue full", zap.String("req", proto.MessageName(uc.Req)))
				handler.RespErr(uc, &errmsg.ErrMsg{
					ErrCode:         models.ErrorType_ErrorNormal,
					ErrMsg:          errmsg.InternalErrMsg,
					ErrInternalInfo: "queue full,maybe block",
				})
			})
		}

		c.lastTime = timer.Now()
	}
}

// AfterFuncCtx 会继承ctx里面的ServerHeader
func (this_ *Service) AfterFuncCtx(c *ctx.Context, duration time.Duration, f func(ctx *ctx.Context)) {
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

func (this_ *Service) AfterFunc(duration time.Duration, f func(ctx *ctx.Context)) {
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
func (this_ *Service) UntilFuncCtx(c *ctx.Context, t time.Time, f func(ctx *ctx.Context)) {
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

func (this_ *Service) UntilFunc(t time.Time, f func(ctx *ctx.Context)) {
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

func (this_ *Service) TickFuncCtx(c *ctx.Context, d time.Duration, f func(ctx *ctx.Context) bool) {
	this_.AfterFuncCtx(c, d, func(ctx *ctx.Context) {
		if f(ctx) {
			this_.TickFuncCtx(ctx, d, f)
		}
	})
}

func (this_ *Service) TickFunc(d time.Duration, f func(ctx *ctx.Context) bool) {
	this_.AfterFunc(d, func(ctx *ctx.Context) {
		if f(ctx) {
			this_.TickFunc(d, f)
		}
	})
}

func (this_ *Service) GetNatsClient() *natsclient.ClusterClient {
	return this_.nc
}

func (this_ *Service) Close() {
	this_.nc.Close()
	this_.el.Stop()
	this_.nc.Shutdown()
	if this_.serverId > 0 {
		startcheck.StopCheck(this_.serverType, this_.serverId)
	}
}
func (this_ *Service) GetEventLoop() *eventloop.EventLoop {
	return this_.el
}

func (this_ *Service) AddBussMid(h handler.MiddleWare) {
	this_.handler.AddHandler(h)
}
