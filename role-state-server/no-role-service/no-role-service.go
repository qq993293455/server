package no_role_service

import (
	"math/rand"
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

type NoRoleService struct {
	system_info.ServiceStatusChangeBaseEvent
	nc             *natsclient.ClusterClient
	log            *logger.Logger
	serverId       values.ServerId
	serverType     models.ServerType
	el             *eventloop.EventLoop
	hashDispatch   [dispatchCnt]chan *ctx.Context
	handler        *handler.Handler
	roleCount      int64
	maxCount       int64
	uniqueId       string
	statusServer   *system_info.ServiceMgr
	onServiceNew   func(s *broadcast.Stats_ServerStats)
	onServiceLose  func(s *broadcast.Stats_ServerStats)
	leaderNotifyCh chan bool
}

func (this_ *NoRoleService) SetOnServerNew(onServiceNew func(s *broadcast.Stats_ServerStats)) {
	this_.onServiceNew = onServiceNew
}

func (this_ *NoRoleService) SetOnServerLose(onServiceLose func(s *broadcast.Stats_ServerStats)) {
	this_.onServiceLose = onServiceLose
}

func NewNoRoleService(
	urls []string,
	log *logger.Logger,
	serverId values.ServerId,
	serverType models.ServerType,
	isDebug bool,
	openLogMid bool,
	wares ...handler.MiddleWare,
) *NoRoleService {
	/*if serverType == models.ServerType_DungeonMatchServer || serverId > 0 {
		startcheck.StartCheck(serverType, serverId)
	}*/
	nc := natsclient.NewClusterClient(serverType, serverId, urls, log)
	el := eventloop.NewEventLoop(log)
	midS := make([]handler.MiddleWare, 0, 4)
	if openLogMid {
		midS = append(midS, handler.Logger)
	}
	midS = append(midS, handler.Recover, handler.Tracing, handler.UnLocker, handler.DoWriteDB)
	midS = append(midS, wares...)
	s := &NoRoleService{
		nc:             nc,
		log:            log,
		serverId:       serverId,
		serverType:     serverType,
		el:             el,
		hashDispatch:   [dispatchCnt]chan *ctx.Context{},
		handler:        handler.NewHandler(el, isDebug, env2.GetIsOpenMiddleError(), midS...),
		maxCount:       int64(runtime.NumCPU() * 5000),
		uniqueId:       xid.New().String(),
		leaderNotifyCh: make(chan bool, 1),
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
						c.TraceLogger = logger.GetTraceLoggerWith(c.TraceId, c.RoleId)
					}
					s.handler.Handle(c)
				}
			}
		})
	}
	return s
}

var statsMessageName = proto.MessageName(&broadcast.Stats_ServerStats{})

func (this_ *NoRoleService) startStats() {
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

func (this_ *NoRoleService) GetLeaderNotifyCh() chan bool {
	return this_.leaderNotifyCh
}

func (this_ *NoRoleService) sendStats() {
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

func (this_ *NoRoleService) OnServiceNew(s *broadcast.Stats_ServerStats) {
	this_.ServiceStatusChangeBaseEvent.OnServiceNew(s)
	if this_.onServiceNew != nil {
		this_.onServiceNew(s)
	}
}

func (this_ *NoRoleService) OnServiceLose(s *broadcast.Stats_ServerStats) {
	this_.ServiceStatusChangeBaseEvent.OnServiceLose(s)
	if this_.onServiceLose != nil {
		this_.onServiceLose(s)
	}
}

func (this_ *NoRoleService) RegisterFunc(desc string, function interface{}, midWares ...handler.MiddleWare) {
	this_.handler.RegisterFunc(desc, function, midWares...)
}

func (this_ *NoRoleService) RegisterEvent(desc string, function interface{}, midWares ...handler.MiddleWare) {
	this_.handler.RegisterEvent(desc, function, midWares...)
}

func (this_ *NoRoleService) Group(mid handler.MiddleWare, midWares ...handler.MiddleWare) *handler.Handler {
	return this_.handler.Group(mid, midWares...)
}

func (this_ *NoRoleService) Start(f func(interface{}), syncStats bool) {
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
		if v != "role_state_write" {
			subj := v + ".>"
			if this_.serverId != 0 {
				subj = v + "." + strconv.Itoa(int(this_.serverId)) + ".>"
			}
			this_.log.Info("nats sub", zap.String("subj", subj))
			this_.Subscribe(subj)
		}
	}
	go func() {
		isLeader := false
		v := "role_state_write"
		subj := v + ".>"
		if this_.serverId != 0 {
			subj = v + "." + strconv.Itoa(int(this_.serverId)) + ".>"
		}
		for beLeader := range this_.leaderNotifyCh {
			if beLeader && !isLeader {
				this_.log.Info("nats sub", zap.String("subj", subj))
				this_.Subscribe(subj)
				isLeader = true
				continue
			}
			if !beLeader && isLeader {
				this_.log.Info("nats unsub", zap.String("subj", subj))
				this_.UnSub(subj)
				isLeader = false
			}
		}
	}()
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
	this_.startStats()
}

var serverIdHeader = &models.ServerHeader{}

func (this_ *NoRoleService) Subscribe(subj string) {
	this_.nc.SubscribeHandler(subj, func(ctx *ctx.Context) {
		this_.el.PostEventQueue(ctx)
	})
}

func (this_ *NoRoleService) UnSub(subj string) {
	this_.nc.UnSub(subj)
}

func (this_ *NoRoleService) syncStats() {
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

func (this_ *NoRoleService) syncStatsOne(t time.Time) {
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

func (this_ *NoRoleService) GetRoleCount() int64 {
	return atomic.LoadInt64(&this_.roleCount)
}

func (this_ *NoRoleService) AddRoleCount(count int64) int64 {
	return atomic.AddInt64(&this_.roleCount, count)
}

func (this_ *NoRoleService) dispatchCtx(uc *ctx.Context) {
	roleId := uc.RoleId
	if roleId == "all" {
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
		hashIdx := rand.Intn(dispatchCnt)
		select {
		case this_.hashDispatch[hashIdx] <- uc:
		default:
			gopool.Submit(func() {
				if uc.TraceLogger == nil {
					uc.TraceLogger = logger.GetTraceLoggerWith(uc.TraceId, roleId)
				}
				uc.Error("hash queue full", zap.String("req", proto.MessageName(uc.Req)))
				handler.RespErr(uc, &errmsg.ErrMsg{
					ErrCode:         models.ErrorType_ErrorNormal,
					ErrMsg:          errmsg.InternalErrMsg,
					ErrInternalInfo: "queue full,maybe block",
				})
			})
		}
	}
}

// AfterFuncCtx 会继承ctx里面的ServerHeader
func (this_ *NoRoleService) AfterFuncCtx(c *ctx.Context, duration time.Duration, f func(ctx *ctx.Context)) {
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

func (this_ *NoRoleService) AfterFunc(duration time.Duration, f func(ctx *ctx.Context)) {
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
func (this_ *NoRoleService) UntilFuncCtx(c *ctx.Context, t time.Time, f func(ctx *ctx.Context)) {
	h := *c.ServerHeader
	h.ServerId = this_.serverId
	h.ServerType = this_.serverType
	ac := ctx.GetPoolContext()
	ac.F = f
	ac.ServerHeader = &h
	ac.TraceLogger = c.TraceLogger
	ac.TraceLogger.ResetInitFiledS(h.TraceId, c.RoleId)
	timer.UntilFunc(t, func() {
		this_.el.PostEventQueue(ac)
	})
}

func (this_ *NoRoleService) UntilFunc(t time.Time, f func(ctx *ctx.Context)) {
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

func (this_ *NoRoleService) TickFuncCtx(c *ctx.Context, d time.Duration, f func(ctx *ctx.Context) bool) {
	this_.AfterFuncCtx(c, d, func(ctx *ctx.Context) {
		if f(ctx) {
			this_.TickFuncCtx(ctx, d, f)
		}
	})
}

func (this_ *NoRoleService) TickFunc(d time.Duration, f func(ctx *ctx.Context) bool) {
	this_.AfterFunc(d, func(ctx *ctx.Context) {
		if f(ctx) {
			this_.TickFunc(d, f)
		}
	})
}

func (this_ *NoRoleService) GetNatsClient() *natsclient.ClusterClient {
	return this_.nc
}

func (this_ *NoRoleService) Close() {
	this_.nc.Close()
	this_.el.Stop()
	this_.nc.Shutdown()
	if this_.serverId > 0 {
		startcheck.StopCheck(this_.serverType, this_.serverId)
	}
}
func (this_ *NoRoleService) GetEventLoop() *eventloop.EventLoop {
	return this_.el
}

func hash(roleId values.RoleId) int64 {
	return int64(utils.Base34DecodeString(roleId)) & dispatchMask
}
