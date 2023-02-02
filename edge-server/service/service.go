package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventloop"
	"coin-server/common/iggsdk"
	"coin-server/common/logger"
	"coin-server/common/natsclient"
	"coin-server/common/netstat"
	"coin-server/common/network/stdtcp"
	"coin-server/common/proto/cppbattle"
	"coin-server/common/proto/edge"
	"coin-server/common/proto/models"
	modelsPb "coin-server/common/proto/models"
	centerpb "coin-server/common/proto/newcenter"
	"coin-server/common/safego"
	self_ip "coin-server/common/self-ip"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/utils"
	env2 "coin-server/common/values/env"
	"coin-server/edge-server/children"
	"coin-server/edge-server/env"
	ipdomain "coin-server/edge-server/ip-domain"

	"github.com/gogo/protobuf/proto"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type Pos struct {
	X float32
	Y float32
}

type BossInfo struct {
	BossId      int64
	BattleId    int64
	MapId       int64
	RefreshTime int64
	DeadTime    int64
	killer      string
	IsDead      bool
	TransPos    Pos
}

type WaitListenPortMap struct {
	mu   sync.Mutex
	wMap map[int64]*children.WaitListenPort
}

func NewWaitListenPortMap() *WaitListenPortMap {
	return &WaitListenPortMap{wMap: map[int64]*children.WaitListenPort{}}
}

func (this_ *WaitListenPortMap) Add(w *children.WaitListenPort) {
	this_.mu.Lock()
	defer this_.mu.Unlock()
	this_.wMap[w.BattleId] = w
}

func (this_ *WaitListenPortMap) Get(battleId int64) (*children.WaitListenPort, bool) {
	this_.mu.Lock()
	v, ok := this_.wMap[battleId]
	this_.mu.Unlock()
	return v, ok
}

func (this_ *WaitListenPortMap) Del(battleId int64) {
	this_.mu.Lock()
	v, ok := this_.wMap[battleId]
	this_.mu.Unlock()
	if ok {
		v.Close(0)
		delete(this_.wMap, battleId)
	}
}

type Service struct {
	reconnectCenterCount int8
	center               atomic.Value
	log                  *logger.Logger
	centerAddr           string
	natsUrls             []string
	natsClient           *natsclient.ClusterClient
	serverType           modelsPb.ServerType
	svc                  *service.Service
	cel                  *eventloop.ChanEventLoop

	childrenMap    map[int64]*children.Child
	bossInfo       map[int64]map[int64]*BossInfo
	totalCpuWeight int64
	usedCpuWeight  int64
	program        string

	workDir string

	pingPongOnce sync.Once

	edgeType  models.EdgeType
	edgePorts struct { // 范围端口的开始和结束
		start uint16
		end   uint16
	}
	portIndex          int32
	usedPortMap        map[int32]struct{}
	netstatUsedPortMap atomic.Value
	localSelfIp        []string

	WaitListenPortMap *WaitListenPortMap
	ServerId          int64
}

func NewEdgeService(
	log *logger.Logger,
	natsUrls []string,
	program string,
	serverId int64,
) *Service {
	s := &Service{
		log:               log,
		natsUrls:          natsUrls,
		natsClient:        nil,
		serverType:        modelsPb.ServerType_EdgeServer,
		svc:               nil,
		childrenMap:       map[int64]*children.Child{},
		bossInfo:          make(map[int64]map[int64]*BossInfo),
		totalCpuWeight:    int64(float64(runtime.NumCPU()*100) * 0.8),
		program:           program,
		usedPortMap:       map[int32]struct{}{},
		cel:               eventloop.NewChanEventLoop(log),
		WaitListenPortMap: NewWaitListenPortMap(),
		ServerId:          serverId,
	}
	evnWeight := env.GetInt64(env.EDGE_TOTAL_WEIGHT)
	if evnWeight > 0 {
		s.totalCpuWeight = evnWeight
	}
	s.netstatUsedPortMap.Store(map[uint16]netstat.SkState{})
	s.edgeType = models.EdgeType(env2.GetInteger(env.EDGE_TYPE))
	s.edgePorts.start = env.GetUint16(env.EDGE_PORTS_START)
	s.edgePorts.end = env.GetUint16(env.EDGE_PORTS_END)
	s.portIndex = int32(s.edgePorts.start)
	s.localSelfIp = append(s.localSelfIp, "127.0.0.1", self_ip.SelfIpLan, self_ip.SelfIPWan)
	log.Info("edgePorts", zap.Uint16("start", s.edgePorts.start), zap.Uint16("end", s.edgePorts.end))
	s.svc = service.NewService(natsUrls, log, serverId, s.serverType, true, true)
	return s
}

func (this_ *Service) netstat() {
	out, err := netstat.UsedPorts(this_.edgePorts.start, this_.edgePorts.end)
	if err != nil {
		this_.log.Error("netstat error", zap.Error(err))
	}

	this_.netstatUsedPortMap.Store(out)
}

func (this_ *Service) GetOnePort() (uint16, bool) {
	return 0, false
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("创建战斗", this_.HandleCreate)
	this_.svc.RegisterFunc("设置edge类型，静态或者动态", this_.HandleSetEdgeType)
	this_.svc.RegisterFunc("杀掉进程", this_.HandleKillBattle)
	this_.svc.RegisterFunc("清退玩家", this_.HandleKillPlayer)
	this_.svc.RegisterEvent("战斗服报告端口", this_.HandleBattleNotifyListenPush)
}

func (this_ *Service) HandleBattleNotifyListenPush(c *ctx.Context, msg *edge.Edge_BattleNotifyListenPush) {
	w, ok := this_.WaitListenPortMap.Get(msg.BattleId)
	if ok {
		w.Close(uint16(msg.Port))
		if msg.OldPort != msg.Port {
			_, _ = eventloop.CallChanEventLoop[*edge.Edge_BattleNotifyListenPush, struct{}](this_.cel, msg, func(r *edge.Edge_BattleNotifyListenPush) (struct{}, *errmsg.ErrMsg) {
				delete(this_.usedPortMap, int32(r.OldPort))
				this_.usedPortMap[int32(r.Port)] = struct{}{}
				return struct{}{}, nil
			})
		}
	}
}

func (this_ *Service) HandleKillPlayer(c *ctx.Context, req *edge.Edge_KillPlayerRequest) (*edge.Edge_KillPlayerResponse, *errmsg.ErrMsg) {
	if len(req.Players) == 0 {
		return &edge.Edge_KillPlayerResponse{}, nil
	}
	children, err := eventloop.CallChanEventLoop[*edge.Edge_KillPlayerRequest, []*children.Child](this_.cel, req, func(r *edge.Edge_KillPlayerRequest) ([]*children.Child, *errmsg.ErrMsg) {
		temp := r.Players
		r.Players = map[int64]*edge.Edge_KillPlayers{}
		out := make([]*children.Child, 0, len(temp))
		for k, v := range temp {
			child, ok := this_.childrenMap[k]
			if ok {
				r.Players[k] = &edge.Edge_KillPlayers{}
				out = append(out, child)
				for _, roleId := range v.Roles {
					_, ok := child.PlayerInfo[roleId]
					if ok {
						r.Players[k].Roles = append(r.Players[k].Roles, roleId)
					}
				}
			}
		}

		return out, nil
	})

	if err != nil {
		return nil, err
	}
	if len(children) == 0 {
		return &edge.Edge_KillPlayerResponse{}, nil
	}
	for _, child := range children {
		ps, ok := req.Players[child.BattleId]
		if ok && len(ps.Roles) > 0 {
			s, ok := child.GetSession()
			if !ok {
				continue
			}
			err = s.Send(c.ServerHeader, &cppbattle.NSNB_EdgeKillPlayer{RoleId: ps.Roles})
			if err != nil {
				this_.log.Error("child kill player failed", zap.Int64("battle_id", child.BattleId), zap.Strings("roles", ps.Roles))
			}
		}
	}

	return &edge.Edge_KillPlayerResponse{}, nil
}

func (this_ *Service) HandleKillBattle(c *ctx.Context, req *edge.Edge_KillBattleRequest) (*edge.Edge_KillBattleResponse, *errmsg.ErrMsg) {
	child, err := eventloop.CallChanEventLoop[int64, *children.Child](this_.cel, req.BattleId, func(battleId int64) (*children.Child, *errmsg.ErrMsg) {
		child, ok := this_.childrenMap[battleId]
		if ok {
			return child, nil
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	if child == nil {
		return &edge.Edge_KillBattleResponse{Success: false}, nil
	}
	child.SetCenterKill()
	er := child.Process.Kill()
	if er != nil {
		this_.log.Error("kill children process error", zap.Error(er), zap.Any("child", child))
	}
	return &edge.Edge_KillBattleResponse{Success: true}, nil
}

type SetEdgeTypeResp = edge.Edge_SetEdgeTypeResponse

func (this_ *Service) HandleSetEdgeType(c *ctx.Context, req *edge.Edge_SetEdgeTypeRequest) (*edge.Edge_SetEdgeTypeResponse, *errmsg.ErrMsg) {
	return eventloop.CallChanEventLoop[models.EdgeType, *SetEdgeTypeResp](this_.cel, req.Typ, func(typ models.EdgeType) (*SetEdgeTypeResp, *errmsg.ErrMsg) {
		if typ != models.EdgeType_StatelessServer {
			if this_.edgeType != models.EdgeType_StatelessServer {
				return &SetEdgeTypeResp{
					Typ:     this_.edgeType,
					Success: false,
				}, nil
			}
			this_.edgeType = typ
			return &SetEdgeTypeResp{
				Typ:     this_.edgeType,
				Success: true,
			}, nil
		} else {
			childCount := len(this_.childrenMap)
			if childCount != 0 {
				return &SetEdgeTypeResp{
					Typ:     this_.edgeType,
					Success: false,
				}, nil
			}
			this_.edgeType = typ
			return &SetEdgeTypeResp{
				Typ:     this_.edgeType,
				Success: true,
			}, nil
		}
	})
}

func (this_ *Service) Run() error {
	this_.Router()
	this_.log.Info("Edge Server Start Run.......")
	this_.workDir = os.Getenv(env.WORK_DIR)
	this_.log.Info("got work dir from env", zap.String("dir", this_.workDir))
	err := utils.CheckAndCreate(this_.workDir)
	if err != nil {
		this_.log.Error("check or create work dir error", zap.Error(err))
		return err
	}

	err = os.Chdir(this_.workDir)
	if err != nil {
		this_.log.Error("chdir work dir error", zap.Error(err))
		return err
	}

	err = utils.CheckAndCreate(filepath.Join(this_.workDir, "log"))
	if err != nil {
		this_.log.Error("check or create work log dir error", zap.Error(err))
		return err
	}

	file := strings.TrimSpace(os.Getenv(env2.BATTLE_SERVER_PATH))
	ok := utils.CheckPathExists(file)
	if !ok {
		this_.log.Error("battle exec file not found")
		return errors.New(file + ":not found")
	}
	this_.svc.Start(func(i interface{}) {
		this_.log.Warn("unknown event", zap.Any("event", i), zap.Any("eventName", reflect.TypeOf(i).String()))
	}, false)

	this_.cel.Start(func(i interface{}) {
		switch e := i.(type) {
		case *children.Child:
			this_.childChange(e)
		case *edge.Edge_PlayerChange:
			this_.playerChange(e)
		case connectedCenter:
			this_.connectedCenterSync()
		case *edge.Edge_NotCanEnterPush:
			this_.notifyCenterCannotEnter(e)
		default:
			this_.log.Warn("cel unknown event", zap.Any("event", i))
		}
	})
	go this_.KeepCenterAlive()
	go this_.StartPingPong()
	return nil
}

func (this_ *Service) notifyCenterCannotEnter(push *edge.Edge_NotCanEnterPush) {
	this_.log.Info("notifyCenterCannotEnter", zap.Any("msg", push))
	child, ok := this_.childrenMap[push.BattleId]
	if !ok {
		this_.log.Error("notifyCenterCannotEnter not found battle", zap.Any("msg", push))
		return
	}
	center := this_.GetCenterSession()
	if center == nil || center.IsClose() {
		return
	}
	child.CannotEnter = true
	err := center.Send(nil, &centerpb.NewCenter_NotCanEnterPush{
		BattleId: push.BattleId,
		BossId:   push.BossId,
	})
	if err != nil {
		center.Close(err)
		this_.log.Warn("center send Center_SyncEdgeInfoPush failed", zap.Error(err))
	}
}

func (this_ *Service) Close() {
	this_.svc.Close()
	this_.cel.Stop()
	for _, v := range this_.childrenMap {
		err := v.Process.Kill()
		if err != nil {
			this_.log.Error("kill children process error", zap.Error(err), zap.Any("child", v))
		}
	}
}

func (this_ *Service) connectedCenterSync() {
	center := this_.GetCenterSession()
	if center == nil || center.IsClose() {
		this_.log.Warn("connectedCenterSync failed,center==nil || center.IsClose()")
		return
	}
	push := &centerpb.NewCenter_SyncEdgeInfoPush{
		IsAll:           true,
		RemainCpuWeight: this_.totalCpuWeight - this_.usedCpuWeight,
		Typ:             this_.edgeType,
	}
	for _, v := range this_.childrenMap {
		ei := &centerpb.NewCenter_EdgeInfo{
			BattleId:    v.BattleId,
			MapId:       v.MapId,
			GuildBossId: v.CreateInfo.GuildDayId,
			RoomId:      v.CreateInfo.RoomId,
			BossHallId:  v.CreateInfo.BossId,
			LineId:      v.CreateInfo.LineId,
			CanNotEnter: v.CannotEnter,
		}
		for k, v := range v.PlayerInfo {
			ei.Roles = append(ei.Roles, &centerpb.NewCenter_EdgeUser{
				RoleId:  k,
				AddTime: v.AddTime,
			})
		}
		push.Edges = append(push.Edges, ei)
	}
	err := center.Send(nil, push)
	if err != nil {
		center.Close(err)
		this_.log.Warn("center send Center_SyncEdgeInfoPush failed", zap.Error(err))
	}
	this_.log.Info("connectedCenterSync success", zap.Any("push", push))
}

func (this_ *Service) playerChangeSyncCenter(msg *edge.Edge_PlayerChange, child *children.Child) {
	center := this_.GetCenterSession()
	if center == nil || center.IsClose() {
		return
	}
	push := &centerpb.NewCenter_SyncEdgeInfoPush{
		RemainCpuWeight: this_.totalCpuWeight - this_.usedCpuWeight,
		Typ:             this_.edgeType,
	}

	ei := &centerpb.NewCenter_EdgeInfo{
		BattleId:    msg.BattleId,
		MapId:       child.MapId,
		IsAdd:       msg.IsAdd,
		GuildBossId: child.CreateInfo.GuildDayId,
		RoomId:      child.CreateInfo.RoomId,
		BossHallId:  child.CreateInfo.BossId,
		LineId:      child.CreateInfo.LineId,
		CanNotEnter: child.CannotEnter,
	}
	if msg.IsAdd {
		for _, v := range msg.Roles {
			pi, ok := child.PlayerInfo[v]
			if ok {
				ei.Roles = append(ei.Roles, &centerpb.NewCenter_EdgeUser{
					RoleId:  v,
					AddTime: pi.AddTime,
				})
			}
		}
	} else {
		for _, v := range msg.Roles {
			ei.Roles = append(ei.Roles, &centerpb.NewCenter_EdgeUser{
				RoleId: v,
			})
		}
	}

	push.Edges = append(push.Edges, ei)
	err := center.Send(nil, push)
	if err != nil {
		center.Close(err)
		this_.log.Warn("center send playerChangeSyncCenter failed", zap.Error(err))
	}
}

func (this_ *Service) battleEndSyncCenter(child *children.Child) {
	center := this_.GetCenterSession()
	if center == nil || center.IsClose() {
		return
	}
	push := &centerpb.NewCenter_SyncEdgeInfoPush{
		RemainCpuWeight: this_.totalCpuWeight - this_.usedCpuWeight,
		Typ:             this_.edgeType,
	}

	ei := &centerpb.NewCenter_EdgeInfo{
		BattleId:    child.BattleId,
		MapId:       child.MapId,
		IsEnd:       true,
		GuildBossId: child.CreateInfo.GuildDayId,
		RoomId:      child.CreateInfo.RoomId,
		BossHallId:  child.CreateInfo.BossId,
		LineId:      child.CreateInfo.LineId,
		CanNotEnter: child.CannotEnter,
	}
	push.Edges = append(push.Edges, ei)
	err := center.Send(nil, push)
	if err != nil {
		center.Close(err)
		this_.log.Warn("center send battleEndSyncCenter failed", zap.Error(err))
	}
}

func (this_ *Service) playerChange(msg *edge.Edge_PlayerChange) {
	child, ok := this_.childrenMap[msg.BattleId]
	if !ok {
		this_.log.Error("player change not found battle child", zap.Any("msg", msg))
		return
	}

	if msg.IsAdd {
		for _, v := range msg.Roles {
			_, ok := child.PlayerInfo[v]
			if !ok {
				child.PlayerInfo[v] = &children.PlayerInfo{AddTime: timer.Now().Unix()}
				this_.log.Info("player enter battle", zap.Int64("battleId", msg.BattleId), zap.String("roleId", v))
			}
		}
	} else {
		for _, v := range msg.Roles {
			delete(child.PlayerInfo, v)
			this_.log.Info("player leave battle", zap.Int64("battleId", msg.BattleId), zap.String("roleId", v))
		}
	}
	this_.playerChangeSyncCenter(msg, child)
}

func (this_ *Service) childChange(child *children.Child) {
	if child.Error == nil && child.ProcessState == nil {
		this_.childrenMap[child.BattleId] = child
		this_.log.Info("create child success", zap.Any("child", child))
	} else {
		this_.usedCpuWeight -= child.CpuWeight
		_, ok := this_.childrenMap[child.BattleId]
		if ok {
			this_.battleEndSyncCenter(child)
		}
		delete(this_.childrenMap, child.BattleId)
		delete(this_.usedPortMap, int32(child.Port))

		if child.ProcessState != nil {
			status := child.ProcessState.Sys().(syscall.WaitStatus)
			isCoreDump := status.CoreDump()
			isSignaled := status.Signaled()
			this_.log.Info("delete child success", zap.Any("child", child),
				zap.Int("exit_state", child.ProcessState.ExitCode()), zap.Bool("core_dump", isCoreDump),
				zap.Bool("signaled", isSignaled),
			)

			if isCoreDump {
				str := fmt.Sprintf("[%s]\n", child.IP)
				if child.CreateInfo.RoomId != 0 {
					str += fmt.Sprintf("多人副本,RoomID[%d],BossEffects[%v],MonsterEffects[%v]", child.CreateInfo.RoomId, child.CreateInfo.BossEffects, child.CreateInfo.MonsterEffects)
				} else if child.CreateInfo.GuildDayId != "" {
					str += fmt.Sprintf("工会Boss,GuildId[%s],GuildDayId[%s]", child.CreateInfo.GuildId, child.CreateInfo.GuildDayId)
				} else {
					str += fmt.Sprintf("挂机副本")
				}

				str += fmt.Sprintf("\n程序端口：%d", child.Port)
				tag := fmt.Sprintf("battleId[%d],mapSceneId[%d] 出现崩溃(codedump)", child.BattleId, child.MapId)
				safego.GOWithLogger(this_.log, func() {
					sendICErr := send2IC(tag, map[string]string{
						"content": str,
						"at_user": "all",
					})
					if sendICErr != nil {
						this_.log.Warn("coredump Send IC error", zap.String("str", str), zap.Error(sendICErr))
					}
				})
				iggsdk.GetAlarmIns().Send(tag + str)
			}

			if child.CreateInfo.Typ == models.EdgeType_StaticServer && !child.IsCenterKill() {
				safego.GOWithLogger(this_.log, func() {
					tag := fmt.Sprintf("[%s]battleId[%d],mapSceneId[%d] 挂机副本程序退出", child.IP, child.BattleId, child.MapId)
					str := "挂机副本程序退出，请联系程序查看原因"
					sendICErr := send2IC(tag, map[string]string{
						"content": str,
						"at_user": "all",
					})
					if sendICErr != nil {
						this_.log.Warn("exit Send IC error", zap.String("tag", tag), zap.String("str", str), zap.Error(sendICErr))
					}
				})
			}
		} else {
			this_.log.Info("delete child success", zap.Any("child", child))
		}
	}
}

type CallCreateResp struct {
	BattleId   int64
	MapSceneId int64
	Ip         string
	Port       int64
}

func (this_ *Service) HandleCreate(c *ctx.Context, req *edge.Edge_CreateServerRequest) (*edge.Edge_CreateServerResponse, *errmsg.ErrMsg) {
	if req.Damages == nil {
		req.Damages = map[string]int64{}
	}
	if req.Bots == nil {
		req.Bots = []*models.Bot{}
	}
	if req.MonsterEffects == nil {
		req.MonsterEffects = []int64{}
	}
	if req.BossEffects == nil {
		req.BossEffects = []int64{}
	}
	if req.CardMap == nil {
		req.CardMap = map[string]int64{}
	}
	this_.netstat()
	netstatPortsMap := this_.netstatUsedPortMap.Load().(map[uint16]netstat.SkState)
	out1, err := eventloop.CallChanEventLoop[*edge.Edge_CreateServerRequest, *CallCreateResp](this_.cel, req,
		func(reqTemp *edge.Edge_CreateServerRequest) (*CallCreateResp, *errmsg.ErrMsg) {
			if this_.edgeType == models.EdgeType_StatelessServer {
				return nil, errmsg.NewNormalErr("err_edge_type_is_stateless", "edge is stateless")
			}

			if this_.edgeType != reqTemp.Typ {
				return nil, errmsg.NewNormalErr("err_edge_type_not_equal", "edge type error")
			}

			child, ok := this_.childrenMap[reqTemp.BattleId]
			if ok {
				return &CallCreateResp{
					BattleId:   child.BattleId,
					MapSceneId: child.MapId,
					Ip:         child.IP,
					Port:       int64(child.Port),
				}, nil
			}
			start := this_.portIndex
			upToMax := false
			found := false
			for {
				if this_.portIndex > int32(this_.edgePorts.end) {
					this_.portIndex = int32(this_.edgePorts.start)
					upToMax = true
				}
				if upToMax && this_.portIndex == start {
					break
				}
				_, ok := this_.usedPortMap[this_.portIndex]
				if ok {
					this_.portIndex++
					continue
				}
				_, ok = netstatPortsMap[uint16(this_.portIndex)]
				if ok {
					this_.portIndex++
					continue
				}
				found = true
				break
			}

			if !found && upToMax {
				this_.log.Warn("CreateChild failed, not found port")
				return nil, errmsg.NewNormalErr("err_not_found_port", "CreateChild failed, not found port")
			}

			this_.usedPortMap[this_.portIndex] = struct{}{} // 减少listen的次数
			port := this_.portIndex
			this_.portIndex++
			return &CallCreateResp{Port: int64(port)}, nil
		})

	if err != nil {
		return nil, err
	}
	if out1.BattleId != 0 {
		return &edge.Edge_CreateServerResponse{
			BattleId:   out1.BattleId,
			MapSceneId: out1.MapSceneId,
			Ip:         out1.Ip,
			Port:       out1.Port,
		}, nil
	}
	out, err := this_.CreateChild(c, uint16(out1.Port), req)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (this_ *Service) CreateChild(
	ctx *ctx.Context,
	port uint16,
	req *edge.Edge_CreateServerRequest,
) (*edge.Edge_CreateServerResponse, *errmsg.ErrMsg) {
	_, err := eventloop.CallChanEventLoop[int64, struct{}](this_.cel, req.CpuWeight, func(cw int64) (struct{}, *errmsg.ErrMsg) {
		if this_.usedCpuWeight+cw > this_.totalCpuWeight {
			return struct{}{}, errmsg.NewNormalErr("err_please_retry", "cup weight too large")
		}
		this_.usedCpuWeight += cw
		return struct{}{}, nil
	})
	if err != nil {
		return nil, err
	}
	wait := make(chan error)
	c := &children.Child{
		EdgeId:       strconv.Itoa(int(this_.ServerId)),
		NatsUrls:     this_.natsUrls,
		CreateInfo:   req,
		BattleId:     req.BattleId,
		MapId:        req.MapSceneId,
		Program:      this_.program,
		ProcAttr:     os.ProcAttr{},
		Process:      nil,
		ProcessState: nil,
		CpuWeight:    req.CpuWeight,
		Port:         port,
		IP:           ipdomain.IPToDomain(self_ip.SelfIPWan),
		OnStarted: func(child *children.Child) {
			wait <- child.Error
		},
		OnStart: func(child *children.Child) {
			this_.cel.PostEventQueue(child)
		},
		OnExit: func(child *children.Child) {
			this_.cel.PostEventQueue(child)
		},
		OnPlayerChange: func(child *children.Child, change *edge.Edge_PlayerChange) {
			change.BattleId = child.BattleId
			this_.cel.PostEventQueue(change)
		},
		OnNotCanEnterPush: func(child *children.Child, push *edge.Edge_NotCanEnterPush) {
			child.CannotEnter = true
			this_.cel.PostEventQueue(push)
		},

		Log: this_.log,
		Wlp: &children.WaitListenPort{
			BattleId: req.BattleId,
			PortChan: make(chan uint16, 1),
		},
	}
	safego.GOWithLogger(this_.log, c.StartAndWait)
	this_.WaitListenPortMap.Add(c.Wlp)
	time.AfterFunc(time.Second*20, func() {
		this_.WaitListenPortMap.Del(req.BattleId)
	})
	er := <-wait
	if er != nil {
		return nil, errmsg.NewNormalErr("err_fork_battle_failed", er.Error())
	}

	out := &cppbattle.NSNB_KeepAliveResponse{}
	for i := 0; i < 10; i++ {
		err := this_.svc.GetNatsClient().RequestWithOut(ctx, req.BattleId, &cppbattle.NSNB_KeepAliveRequest{}, out, time.Millisecond*200)
		if err == nil {
			break
		}
	}
	return &edge.Edge_CreateServerResponse{
		BattleId:   c.BattleId,
		MapSceneId: c.MapId,
		Ip:         c.IP,
		Port:       int64(c.Port),
	}, nil
}

// KeepCenterAlive 保持和中心服务器之间的连接，保活
func (this_ *Service) KeepCenterAlive() {
	// 得到center的地址
	this_.getCenterAddr()
	// 开始连接
	this_.log.Info("start connect center server", zap.String("addr", this_.centerAddr))
	stdtcp.Connect(this_.centerAddr, time.Second*3, true, this_, this_.log, true)
}

func (this_ *Service) getCenterAddr() {

	centerAddr := strings.TrimSpace(os.Getenv(env.CENTER_ADDR))
	if centerAddr != "" {
		this_.centerAddr = centerAddr
		return
	}

	if this_.natsClient == nil {
		this_.natsClient = natsclient.NewClusterClient(modelsPb.ServerType_EdgeServer, 0, this_.natsUrls, this_.log)
		this_.log.Info("create nats cluster client", zap.Strings("urls", this_.natsUrls))
	}
reqLoop:
	out := &centerpb.NewCenter_SelfAddrResponse{}
	err := this_.RequestWithOut(&centerpb.NewCenter_SelfAddrRequest{}, out)
	if err != nil {
		this_.log.Warn("getCenterAddr error", zap.Error(err))
		time.Sleep(time.Second)
		goto reqLoop
	}
	this_.centerAddr = out.Addr
	this_.log.Info("got center server addr", zap.String("addr", this_.centerAddr))
}

func (this_ *Service) RequestWithOut(msg proto.Message, out proto.Message) *errmsg.ErrMsg {
	header := &modelsPb.ServerHeader{
		StartTime:  timer.Now().UnixNano(),
		ServerType: this_.serverType,
		TraceId:    xid.New().String(),
	}
	c := &ctx.Context{ServerHeader: header}
	return this_.natsClient.RequestWithOut(c, env.CenterServerId(), msg, out)
}

func (this_ *Service) TCPRequestWithOut(session *stdtcp.Session, msg proto.Message, out proto.Message) *errmsg.ErrMsg {
	header := &modelsPb.ServerHeader{
		StartTime:  timer.Now().UnixNano(),
		ServerType: this_.serverType,
		TraceId:    xid.New().String(),
	}
	return session.RPCRequestOut(header, msg, out)
}

func (this_ *Service) GetCenterSession() *stdtcp.Session {
	v := this_.center.Load()
	if v == nil {
		return nil
	}
	return v.(*stdtcp.Session)
}

func (this_ *Service) SendPING() *errmsg.ErrMsg {
	center := this_.GetCenterSession()
	if center == nil || center.IsClose() {
		return nil
	}
	ping := &models.PING{Now: timer.UnixNano()}
	return center.Send(nil, ping)
}

func (this_ *Service) StartPingPong() {
	this_.pingPongOnce.Do(func() {
		timer.Ticker(time.Second, func() bool {
			center := this_.GetCenterSession()
			if center == nil || center.IsClose() {
				return true
			}

			err := this_.SendPING()
			if err != nil {
				center.Close(err)
			}
			//			this_.log.Debug("send ping success", zap.Time("now", timer.Now()))
			return true
		})
	})
}

type connectedCenter struct {
}

func (this_ *Service) OnConnected(session *stdtcp.Session) {
	this_.reconnectCenterCount = 0
	this_.log.Info("connect center success", zap.String("center", session.RemoteAddr()))
	this_.log.Info("start send auth msg")
	out := &centerpb.NewCenter_AuthEdgeResponse{}
	now := timer.Now().Unix()
	token := utils.MD5String("edge-center-" + strconv.Itoa(int(now)))
	err := this_.TCPRequestWithOut(session, &centerpb.NewCenter_AuthEdgeRequest{
		Token: token,
		Now:   now,
	}, out)
	if err != nil {
		this_.log.Info("send auth msg error, close center server connection", zap.Error(err))
		session.Close(err)
	}
	token = utils.MD5String("center-edge-" + strconv.Itoa(int(out.Now)))
	if token != out.Token {
		this_.log.Info("auth msg verify failed, close center server connection", zap.Error(err))
		session.Close(err)
	}
	this_.center.Store(session)
	this_.log.Info("auth msg verify success,start receive center msg")
	this_.cel.PostEventQueue(connectedCenter{})
}

func (this_ *Service) OnDisconnected(session *stdtcp.Session, err error) {
	time.Sleep(time.Second)
	this_.log.Error("disconnect center success", zap.String("center", session.RemoteAddr()), zap.Error(err))
	this_.center.Store((*stdtcp.Session)(nil))
	this_.reconnectCenterCount++
	if this_.reconnectCenterCount > 3 {
		session.SetAbortReconnect()
		go this_.KeepCenterAlive()
	}
}

func (this_ *Service) OnRequest(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) {
	err := this_.svc.HandleTCPData(session, rpcIndex, msgName, frame)
	if err != nil {
		session.Close(err)
	}
}

var pongName = (&models.PONG{}).XXX_MessageName()

func (this_ *Service) OnMessage(session *stdtcp.Session, msgName string, frame []byte) {
	if msgName == pongName {
		return
	}
	err := this_.svc.HandleTCPData(session, 0, msgName, frame)
	if err != nil {
		session.Close(err)
	}
}
