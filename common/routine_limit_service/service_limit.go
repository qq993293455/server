package routine_limit_service

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
	"coin-server/common/natsclient"
	"coin-server/common/proto/broadcast"
	"coin-server/common/proto/models"
	readpb "coin-server/common/proto/role_state_read"
	"coin-server/common/protocol"
	"coin-server/common/safego"
	"coin-server/common/startcheck"
	system_info "coin-server/common/system-info"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values"
	env2 "coin-server/common/values/env"

	"github.com/nats-io/nats.go"

	"github.com/rs/xid"

	"github.com/gogo/protobuf/proto"

	"go.uber.org/zap"
)

const (
	dispatchMask = 0b11111111111 // 2047
	dispatchCnt  = dispatchMask + 1
)

type roleStatus struct {
	lastTime time.Time
	roleId   values.RoleId
	gateId   values.GatewayId
	online   bool
	hashIdx  int64
}

type RoutineLimitService struct {
	system_info.ServiceStatusChangeBaseEvent
	nc            *natsclient.ClusterClient
	log           *logger.Logger
	serverId      values.ServerId
	serverType    models.ServerType
	el            *eventloop.EventLoop
	hashDispatch  [dispatchCnt]chan *ctx.Context
	roleMap       map[values.RoleId]*roleStatus
	handler       *handler.Handler
	roleCount     int64
	maxCount      int64
	uniqueId      string
	statusServer  *system_info.ServiceMgr
	onServiceNew  func(s *broadcast.Stats_ServerStats)
	onServiceLose func(s *broadcast.Stats_ServerStats)
}

func (this_ *RoutineLimitService) SetOnServerNew(onServiceNew func(s *broadcast.Stats_ServerStats)) {
	this_.onServiceNew = onServiceNew
}

func (this_ *RoutineLimitService) SetOnServerLose(onServiceLose func(s *broadcast.Stats_ServerStats)) {
	this_.onServiceLose = onServiceLose
}

func NewRoutineLimitService(
	urls []string,
	log *logger.Logger,
	serverId values.ServerId,
	serverType models.ServerType,
	isDebug bool,
	openLogMid bool,
	wares ...handler.MiddleWare,
) *RoutineLimitService {
	if serverType == models.ServerType_DungeonMatchServer || serverId > 0 {
		startcheck.StartCheck(serverType, serverId)
	}
	nc := natsclient.NewClusterClient(serverType, serverId, urls, log)
	el := eventloop.NewEventLoop(log)
	midS := make([]handler.MiddleWare, 0, 4)
	if openLogMid {
		midS = append(midS, handler.Logger)
	}
	midS = append(midS, handler.Recover, handler.OpenGMHandler, handler.Tracing, handler.UnLocker, handler.DoWriteDB)
	midS = append(midS, wares...)
	s := &RoutineLimitService{
		nc:           nc,
		log:          log,
		serverId:     serverId,
		serverType:   serverType,
		el:           el,
		roleMap:      map[values.RoleId]*roleStatus{},
		hashDispatch: [dispatchCnt]chan *ctx.Context{},
		handler:      handler.NewHandler(el, isDebug, env2.GetIsOpenMiddleError(), midS...),
		maxCount:     int64(runtime.NumCPU() * 5000),
		uniqueId:     xid.New().String(),
	}
	s.statusServer = system_info.NewServiceMgr(s.uniqueId, log, s)
	serverIdHeader.ServerId = serverId
	serverIdHeader.ServerType = serverType
	for idx := range s.hashDispatch {
		queue := make(chan *ctx.Context, 100)
		s.hashDispatch[idx] = queue
		safego.GOWithLogger(s.log, func() {
			for c := range queue {
				if c.F == nil {
					s.handler.Handle(c)
				} else {
					c.StartTime = timer.Now().UnixNano()
					if c.TraceLogger == nil {
						c.TraceLogger = logger.GetTraceLogger()
					}
					c.TraceLogger.ResetInitFiledS(c.TraceId, c.RoleId)
					s.handler.Handle(c)
				}
			}
		})
	}
	return s
}

var statsMessageName = proto.MessageName(&broadcast.Stats_ServerStats{})

func (this_ *RoutineLimitService) startStats() {
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

func (this_ *RoutineLimitService) sendStats() {
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

func (this_ *RoutineLimitService) OnServiceNew(s *broadcast.Stats_ServerStats) {
	this_.ServiceStatusChangeBaseEvent.OnServiceNew(s)
	if this_.onServiceNew != nil {
		this_.onServiceNew(s)
	}
}

func (this_ *RoutineLimitService) OnServiceLose(s *broadcast.Stats_ServerStats) {
	this_.ServiceStatusChangeBaseEvent.OnServiceLose(s)
	if this_.onServiceLose != nil {
		this_.onServiceLose(s)
	}
}

func (this_ *RoutineLimitService) RegisterFunc(desc string, function interface{}, midWares ...handler.MiddleWare) {
	this_.handler.RegisterFunc(desc, function, midWares...)
}

func (this_ *RoutineLimitService) RegisterEvent(desc string, function interface{}, midWares ...handler.MiddleWare) {
	this_.handler.RegisterEvent(desc, function, midWares...)
}

func (this_ *RoutineLimitService) Group(mid handler.MiddleWare, midWares ...handler.MiddleWare) *handler.Handler {
	return this_.handler.Group(mid, midWares...)
}

func (this_ *RoutineLimitService) Start(f func(interface{}), syncStats bool) {
	es := this_.handler.GetHandlers()
	for _, v := range es {
		this_.log.Info("registered : " + v.String())
	}
	els := eventlocal.GetAllEventLocal()
	for _, v := range els {
		this_.log.Info("subscribe event_local : " + v)
	}
	subjS := this_.handler.GetSubjArray()
	if len(subjS) == 0 {
		panic("no subj found")
	}
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
		case []*readpb.RoleStateROnly_PushToClient:
			this_.dispatchPushMessages(t)
		default:
			if f != nil {
				f(event)
			}
		}
	})
	this_.startStats()
}

var serverIdHeader = &models.ServerHeader{}

func (this_ *RoutineLimitService) dispatchPushMessages(pms []*readpb.RoleStateROnly_PushToClient) {
	gopool.Submit(func() {
		if err := this_.nc.Publish(0, serverIdHeader, &readpb.RoleStateROnly_PushManyToClient{
			Pcs: pms,
		}); err != nil {
			this_.log.Error("push msg to role state server error", zap.Any("data", pms))
		}
	})
}

func (this_ *RoutineLimitService) Subscribe(subj string) {
	this_.nc.SubscribeHandler(subj, func(ctx *ctx.Context) {
		this_.el.PostEventQueue(ctx)
	})
}

func (this_ *RoutineLimitService) syncStats() {
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

func (this_ *RoutineLimitService) syncStatsOne(t time.Time) {
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

func (this_ *RoutineLimitService) GetRoleCount() int64 {
	return atomic.LoadInt64(&this_.roleCount)
}

func (this_ *RoutineLimitService) AddRoleCount(count int64) int64 {
	return atomic.AddInt64(&this_.roleCount, count)
}

func (this_ *RoutineLimitService) dispatchCtx(uc *ctx.Context) {
	roleId := uc.RoleId
	if roleId == "" {
		uc.TraceLogger = this_.log.WithTrace(uc.TraceId, "")
		gopool.Submit(func() {
			if uc.F == nil {
				this_.handler.Handle(uc)
			} else {
				uc.StartTime = timer.Now().UnixNano()
				if uc.TraceLogger == nil {
					uc.TraceLogger = this_.log.WithTrace(uc.TraceId, "")
				}
				this_.handler.Handle(uc)
			}

		})
	} else {
		status, ok := this_.roleMap[roleId]
		if !ok {
			status = &roleStatus{
				lastTime: timer.Now(),
				roleId:   roleId,
				gateId:   uc.GateId,
				hashIdx:  hash(roleId),
			}
			if uc.F == nil && uc.ServerHeader != nil && uc.GateId != 0 {
				status.online = true
			}

			this_.el.TickQueue(time.Minute, func() bool {
				if timer.Now().Sub(status.lastTime) >= time.Minute*30 {
					delete(this_.roleMap, roleId)
					this_.AddRoleCount(-1)
					return false
				}
				return true
			})
			this_.AddRoleCount(1)
			this_.roleMap[roleId] = status
		}
		if uc.F == nil && uc.ServerHeader != nil && uc.GateId != 0 {
			if status.gateId != uc.GateId {
				status.gateId = uc.GateId
			}
			if uc.GateId != 0 {
				status.online = true
			}
		}
		select {
		case this_.hashDispatch[status.hashIdx] <- uc:
		default:
			gopool.Submit(func() {
				uc.Error("hash queue full", zap.String("req", proto.MessageName(uc.Req)))
				handler.RespErr(uc, &errmsg.ErrMsg{
					ErrCode:         models.ErrorType_ErrorNormal,
					ErrMsg:          errmsg.InternalErrMsg,
					ErrInternalInfo: "queue full,maybe block",
				})
			})
		}

		status.lastTime = timer.Now()
	}
}

// AfterFuncCtx 会继承ctx里面的ServerHeader
func (this_ *RoutineLimitService) AfterFuncCtx(c *ctx.Context, duration time.Duration, f func(ctx *ctx.Context)) {
	h := *c.ServerHeader
	h.ServerId = this_.serverId
	h.ServerType = this_.serverType
	ac := ctx.GetPoolContext()
	ac.F = f
	ac.ServerHeader = &h
	ac.TraceLogger.ResetInitFiledS(c.TraceId, c.RoleId)
	ac.StartTime = c.StartTime
	timer.AfterFunc(duration, func() {
		this_.el.PostEventQueue(ac)
	})
}

func (this_ *RoutineLimitService) AfterFunc(duration time.Duration, f func(ctx *ctx.Context)) {
	timer.AfterFunc(duration, func() {
		h := &models.ServerHeader{}
		h.ServerId = this_.serverId
		h.ServerType = this_.serverType
		h.TraceId = xid.NewWithTime(time.Now()).String()
		ac := ctx.GetPoolContext()
		ac.F = f
		ac.ServerHeader = h
		ac.TraceLogger.ResetInitFiledS(h.TraceId, "")
		ac.StartTime = timer.Now().UnixNano()
		this_.el.PostEventQueue(ac)
	})
}

// UntilFuncCtx 会继承ctx里面的ServerHeader
func (this_ *RoutineLimitService) UntilFuncCtx(c *ctx.Context, t time.Time, f func(ctx *ctx.Context)) {
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

func (this_ *RoutineLimitService) UntilFunc(t time.Time, f func(ctx *ctx.Context)) {
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

func (this_ *RoutineLimitService) TickFuncCtx(c *ctx.Context, d time.Duration, f func(ctx *ctx.Context) bool) {
	this_.AfterFuncCtx(c, d, func(ctx *ctx.Context) {
		if f(ctx) {
			this_.TickFuncCtx(ctx, d, f)
		}
	})
}

func (this_ *RoutineLimitService) TickFunc(d time.Duration, f func(ctx *ctx.Context) bool) {
	this_.AfterFunc(d, func(ctx *ctx.Context) {
		if f(ctx) {
			this_.TickFunc(d, f)
		}
	})
}

func (this_ *RoutineLimitService) GetNatsClient() *natsclient.ClusterClient {
	return this_.nc
}

func (this_ *RoutineLimitService) Close() {
	this_.nc.Close()
	this_.el.Stop()
	this_.nc.Shutdown()
	if this_.serverId > 0 {
		startcheck.StopCheck(this_.serverType, this_.serverId)
	}
}
func (this_ *RoutineLimitService) GetEventLoop() *eventloop.EventLoop {
	return this_.el
}

func hash(roleId values.RoleId) int64 {
	return int64(utils.Base34DecodeString(roleId)) & dispatchMask
}
