package service

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/eventlocal"
	"coin-server/common/eventloop"
	"coin-server/common/gopool"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/mysqlclient"
	"coin-server/common/natsclient"
	"coin-server/common/proto/models"
	"coin-server/common/protocol"
	"coin-server/common/safego"
	"coin-server/common/startcheck"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values"
	env2 "coin-server/common/values/env"
	"coin-server/guild-gvg-server/env"
	"coin-server/rule"

	"github.com/jmoiron/sqlx"

	"github.com/rs/xid"
	"go.uber.org/zap"
)

type GVGStatus int64

const (
	StatusSignup     GVGStatus = 0 // 报名中
	StatusMatch      GVGStatus = 1 // 匹配
	StatusFighting   GVGStatus = 2 // 战斗中
	StatusSettlement GVGStatus = 3 // 结算中
)

func (s GVGStatus) String() string {
	switch s {
	case StatusSignup:
		return "Signup"
	case StatusMatch:
		return "Match"
	case StatusFighting:
		return "Fighting"
	case StatusSettlement:
		return "Settlement"
	}
	return "Unknown"
}

type signupInfo struct {
	NickName string
	SignTime time.Time
	IsMatch  bool
}

type Service struct {
	nc         *natsclient.ClusterClient
	mysql      *mysqlclient.Client
	log        *logger.Logger
	serverId   values.ServerId
	serverType models.ServerType
	el         *eventloop.EventLoop
	handler    *handler.Handler
	gvgStatus  GVGStatus

	signupQueue     *eventloop.ChanEventLoop
	signupMap       map[string]signupInfo
	groupMap        map[int64]*GroupInfo
	userGroupInfo   map[string]*BuildRole // 映射用户->分组ID // 因为可能某个工会成员在匹配时未能加入活动
	guildGroupInfo  map[string]int64      //映射工会ID->分组ID
	activeStartTime int64
	activeEndTime   int64

	activeId int64 // 当前活动ID， 以周为单位算数出来的

	fightIndex int64

	saveDBChan            chan []interface{}
	closeSaveDBChan       chan struct{}
	saveWait              sync.WaitGroup
	maxAttackCount        int64
	addAttackCountSeconds int64
	addScorePerHour       int64 // 工会每分钟增加的积分
	initScore             int64 // 工会初始积分

	gameServerId int64 // 工会祝福加速  处理的GameServer
}

func (this_ *Service) GetGVGStats() GVGStatus {
	return GVGStatus(atomic.LoadInt64((*int64)(&this_.gvgStatus)))
}

func (this_ *Service) SetGVGStatus(status GVGStatus) {
	atomic.StoreInt64((*int64)(&this_.gvgStatus), int64(status))
}

func NewService(
	mysql *mysqlclient.Client,
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
	midS = append(midS, handler.Recover, handler.Tracing, handler.UnLocker, handler.DoWriteDB)
	midS = append(midS, wares...)
	r := rule.MustGetReader(nil)
	maxAttackCount, ok := r.KeyValue.GetInt64("GuildContendChallengeNumMax")
	utils.MustTrue(ok)
	guildContendChallengeTime, ok := r.KeyValue.GetInt64("GuildContendChallengeTime")
	utils.MustTrue(ok)
	guildContendEveryDayIntegral, ok := r.KeyValue.GetInt64("GuildContendEveryDayIntegral")
	utils.MustTrue(ok)
	guildContendEveryIntegral, ok := r.KeyValue.GetInt64("GuildContendEveryIntegral")
	utils.MustTrue(ok)
	gameServerId := env2.GetInteger(env.DEFAULT_GAME_SERVER_ID)
	s := &Service{
		nc:                    nc,
		log:                   log,
		serverId:              serverId,
		serverType:            serverType,
		el:                    el,
		handler:               handler.NewHandler(el, isDebug, env2.GetIsOpenMiddleError(), midS...),
		signupQueue:           eventloop.NewChanEventLoop(log),
		mysql:                 mysql,
		signupMap:             map[string]signupInfo{},
		groupMap:              map[int64]*GroupInfo{},
		userGroupInfo:         map[string]*BuildRole{},
		guildGroupInfo:        map[string]int64{},
		saveDBChan:            make(chan []interface{}, 10000),
		closeSaveDBChan:       make(chan struct{}),
		fightIndex:            time.Now().UTC().UnixMilli(),
		maxAttackCount:        maxAttackCount,
		addAttackCountSeconds: guildContendChallengeTime,
		addScorePerHour:       guildContendEveryDayIntegral,
		initScore:             guildContendEveryIntegral,
		gameServerId:          gameServerId,
	}
	s.startSave()
	s.StartDropOldTable()
	return s
}

func (this_ *Service) AsyncSaveDB(data []interface{}) {
	select {
	case <-this_.closeSaveDBChan:
		return
	default:
		this_.saveDBChan <- data
	}
}

func (this_ *Service) startSave() {
	this_.saveWait.Add(1)
	safego.GO(func(i interface{}) {
		this_.log.Error("startSave panic", zap.Any("panic info", i))
		this_.startSave()
	}, func() {
		defer this_.saveWait.Done()
		var closeFor bool
		for !closeFor {
			select {
			case saveData, ok := <-this_.saveDBChan:
				if ok {
					this_.saveData(saveData)
				}
			case <-this_.closeSaveDBChan:
				closeFor = true
			}
		}
		for saveData := range this_.saveDBChan {
			this_.saveData(saveData)
		}
	})
}

func (this_ *Service) saveData(data []interface{}) {
	err := this_.mysql.WithTxx(func(tx *sqlx.Tx) error {
		var err error
		for _, v := range data {
			switch t := v.(type) {
			case *BuildInfoSave:
				_, err = tx.Exec("UPDATE gvg.build SET build_info=? WHERE group_id=? AND guild_id=? AND build_id=?", t.Data, t.GroupId, t.GuildId, t.BuildId)
			case *GuildInfoSave:
				_, err = tx.Exec("UPDATE gvg.guild SET guild_data=? WHERE group_id=? AND guild_id=?", t.Data, t.GroupId, t.GuildId)
			case *FightingSave:
				_, err = tx.Exec(`INSERT INTO gvg.fighting(id,group_id,attack_guild_id,attacker,defend_guild_id,defender,is_builder,blood,is_win,create_time,personal_score_add,guild_score_add,build_id) 
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, t.Id, t.GroupId, t.AttackGuildId, t.Attacker, t.DefendGuildId, t.Defender, t.IsBuilder, t.Blood, t.IsWin, t.CreateTime, t.PersonalScoreAdd, t.GuildScoreAdd, t.BuildId)
			}
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		this_.log.Error("save data error", zap.Error(err), zap.Any("data", data))
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

func (this_ *Service) Start(f func(interface{})) {
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
	this_.signupQueue.Start(func(e interface{}) {
		this_.log.Warn("signupQueue unknown event", zap.Any("event", e))
	})
	this_.el.Start(func(event interface{}) {
		switch t := event.(type) {
		case *ctx.Context:
			this_.dispatchCtx(t)
		case GVGStatus:
			this_.SetGVGStatus(t)
		default:
			if f != nil {
				f(event)
			}
		}
	})
}

func (this_ *Service) Subscribe(subj string) {
	this_.nc.SubscribeHandler(subj, func(ctx *ctx.Context) {
		this_.el.PostEventQueue(ctx)
	})
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

func (this_ *Service) dispatchCtx(uc *ctx.Context) {
	if uc.TraceLogger == nil {
		uc.TraceLogger = this_.log.WithTrace(uc.TraceId, uc.RoleId)
	}
	if uc.F != nil {
		uc.StartTime = timer.Now().UnixNano()
	}
	gopool.Submit(func() {
		this_.handler.Handle(uc)
	})
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
	this_.signupQueue.Stop()
	this_.nc.Shutdown()
	close(this_.closeSaveDBChan)
	close(this_.saveDBChan)
	this_.saveWait.Wait()
}

func (this_ *Service) GetEventLoop() *eventloop.EventLoop {
	return this_.el
}
