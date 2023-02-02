package state

import (
	"strconv"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/gopool"
	"coin-server/common/logger"
	"coin-server/common/proto/gatewaytcp"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	nostatesvc "coin-server/role-state-server/no-role-service"
	"coin-server/role-state-server/service/state/cache"
	"coin-server/role-state-server/service/state/raft"

	readpb "coin-server/common/proto/role_state_read"
	writepb "coin-server/common/proto/role_state_write"

	hraft "github.com/hashicorp/raft"
	"go.uber.org/zap"
)

var serverIdHeader = &models.ServerHeader{}

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *nostatesvc.NoRoleService
	cache      *cache.Mgr
	raftC      *hraft.Raft
	log        *logger.Logger
}

func NewStateService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *nostatesvc.NoRoleService,
	log *logger.Logger,
	raftOpt *raft.Options,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		cache:      cache.NewSlot(),
		log:        log,
	}
	serverIdHeader.ServerId = serverId
	serverIdHeader.ServerType = serverType
	var has bool
	has, s.raftC = raft.Init(raftOpt, s.svc.GetLeaderNotifyCh(), s.cache)
	if !has && raftOpt.Join {
		s.join(raftOpt)
	}
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取玩家状态", svc.GetRoleState)
	svc.svc.RegisterFunc("加入集群", svc.JoinCluster)

	svc.svc.RegisterEvent("登录通知", svc.LoginNotify)
	svc.svc.RegisterEvent("登出通知", svc.LogoutNotify)
	svc.svc.RegisterEvent("推送到客户端", svc.PushToClient)
	svc.svc.RegisterEvent("推送到所有客户端", svc.PushToAllClient)
}

func (svc *Service) LoginNotify(c *ctx.Context, _ *writepb.RoleStateRW_LoginNotifyEvent) {
	data := c.RoleId + ":" + strconv.FormatInt(c.GateId, 10)
	applyFuture := svc.raftC.Apply([]byte(data), 5*time.Second)
	if err := applyFuture.Error(); err != nil {
		svc.log.Error("Raft LoginNotify fail", zap.String("role_id", c.RoleId), zap.Int64("gate_id", c.GateId))
		return
	}
	svc.log.Info("Raft LoginNotify succ", zap.String("role_id", c.RoleId), zap.Int64("gate_id", c.GateId))
	return
}

func (svc *Service) LogoutNotify(c *ctx.Context, _ *writepb.RoleStateRW_LogoutNotifyEvent) {
	data := c.RoleId + ":" + strconv.FormatInt(-1, 10)
	applyFuture := svc.raftC.Apply([]byte(data), 5*time.Second)
	if err := applyFuture.Error(); err != nil {
		svc.log.Error("Raft LogoutNotify fail", zap.String("role_id", c.RoleId))
		return
	}
	svc.log.Info("Raft LogoutNotify succ", zap.String("role_id", c.RoleId))
}

func (svc *Service) JoinCluster(c *ctx.Context, req *writepb.RoleStateRW_JoinClusterRequest) (*writepb.RoleStateRW_JoinClusterResponse, *errmsg.ErrMsg) {
	addPeerFuture := svc.raftC.AddVoter(hraft.ServerID(req.ClusterServerId), hraft.ServerAddress(req.ClusterTcpAddr), 0, 0)
	if err := addPeerFuture.Error(); err != nil {
		svc.log.Error("Raft Node join fail", zap.String("id", req.ClusterServerId), zap.String("addr", req.ClusterTcpAddr))
		return nil, errmsg.NewErrInvalidRequestParam()
	}
	svc.log.Info("Raft Node join succ", zap.String("id", req.ClusterServerId), zap.String("addr", req.ClusterTcpAddr))
	return &writepb.RoleStateRW_JoinClusterResponse{}, nil
}

func (svc *Service) GetRoleState(c *ctx.Context, req *readpb.RoleStateROnly_GetRoleStateRequest) (*readpb.RoleStateROnly_GetRoleStateResponse, *errmsg.ErrMsg) {
	state := make([]*readpb.RoleStateROnly_RoleState, len(req.RoleIds))
	for idx, roleId := range req.RoleIds {
		st := &readpb.RoleStateROnly_RoleState{
			RoleId: roleId,
		}
		ok, gateIdStr := svc.cache.Get(roleId)
		if !ok {
			st.GateId = -1
		} else {
			st.GateId, _ = strconv.ParseInt(gateIdStr, 10, 64)
		}
		state[idx] = st
	}
	return &readpb.RoleStateROnly_GetRoleStateResponse{
		State: state,
	}, nil
}

func (svc *Service) PushToClient(c *ctx.Context, req *readpb.RoleStateROnly_PushManyToClient) {
	svc.log.Info("push msg to client", zap.Any("data", req.Pcs))
	m := map[values.GatewayId][]*gatewaytcp.GatewayStdTcp_PushToClient{}
	for _, v := range req.Pcs {
		m1 := map[values.GatewayId]*gatewaytcp.GatewayStdTcp_PushToClient{}
		for _, role := range v.Roles {
			ok, gateIdStr := svc.cache.Get(role)
			if ok {
				gateId, _ := strconv.ParseInt(gateIdStr, 10, 64)
				cpm, ok1 := m1[gateId]
				if !ok1 {
					cpm = &gatewaytcp.GatewayStdTcp_PushToClient{}
					m1[gateId] = cpm
					cpm.Messages = v.Messages
				}
				cpm.Roles = append(cpm.Roles, role)
			}
		}
		for k, v1 := range m1 {
			m[k] = append(m[k], v1)
		}
	}
	if len(m) > 0 {
		gopool.Submit(func() {
			for gateId, pms := range m {
				err := svc.svc.GetNatsClient().Publish(gateId, serverIdHeader, &gatewaytcp.GatewayStdTcp_PushManyToClient{Pcs: pms})
				if err != nil {
					svc.log.Error("push msg to client error", zap.Int64("gate_id", gateId), zap.Any("data", pms))
				}
			}
		})
	}
}

func (svc *Service) PushToAllClient(c *ctx.Context, req *readpb.RoleStateROnly_PushToAllClient) {
	svc.log.Info("push msg to all client")
	slotNum := svc.cache.GetSlotNum()
	for idx := 0; idx < slotNum; idx++ {
		m := svc.cache.GetOneSlot(idx)
		m1 := map[values.GatewayId][]*gatewaytcp.GatewayStdTcp_PushToClient{}
		for gatewayStr, roleIds := range m {
			gateId, _ := strconv.ParseInt(gatewayStr, 10, 64)
			//TODO: 拆roleIds的大小
			m1[gateId] = []*gatewaytcp.GatewayStdTcp_PushToClient{
				{
					Roles:    roleIds,
					Messages: req.Messages,
				},
			}
		}
		if len(m1) > 0 {
			gopool.Submit(func() {
				for gateId, pms := range m1 {
					err := svc.svc.GetNatsClient().Publish(gateId, serverIdHeader, &gatewaytcp.GatewayStdTcp_PushManyToClient{Pcs: pms})
					if err != nil {
						svc.log.Error("push msg to client error", zap.Int64("gate_id", gateId), zap.Any("data", pms))
					}
				}
			})
		}
	}
}

func (svc *Service) join(opt *raft.Options) {
	svc.svc.AfterFunc(1*time.Millisecond, func(c *ctx.Context) {
		tryCnt := 0
		for tryCnt < 5 {
			resp := &writepb.RoleStateRW_JoinClusterResponse{}
			err := svc.svc.GetNatsClient().RequestWithOut(c, svc.serverId, &writepb.RoleStateRW_JoinClusterRequest{
				ClusterServerId: opt.ServerId,
				ClusterTcpAddr:  opt.TcpAddr,
			}, resp)
			if err != nil {
				svc.log.Info("Raft Node join fail", zap.String("err", err.ErrMsg), zap.String("id", opt.ServerId), zap.String("addr", opt.TcpAddr), zap.Int("tryTime", tryCnt))
				tryCnt++
				time.Sleep(time.Second)
				continue
			}
			svc.log.Info("Raft Node join succ", zap.String("id", opt.ServerId), zap.String("addr", opt.TcpAddr))
			return
		}
		svc.log.Error("Raft Node join err", zap.String("id", opt.ServerId), zap.String("addr", opt.TcpAddr))
	})
}
