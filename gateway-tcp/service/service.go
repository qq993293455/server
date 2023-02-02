package service

import (
	"errors"
	"strings"
	"sync"
	"time"

	"coin-server/common/bytespool"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/metrics"
	"coin-server/common/natsclient"
	"coin-server/common/network/stdtcp"
	"coin-server/common/proto/broadcast"
	"coin-server/common/proto/gatewaytcp"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/protocol"
	"coin-server/common/serverids"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/env"

	"github.com/gogo/protobuf/proto"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type UserInfo struct {
	ServerId       int64
	RoleId         string
	UserId         string
	Session        *stdtcp.Session
	RuleVersion    string
	MapId          int64
	BattleServerId int64
	LoginTime      int64
	ClientVersion  string
	Language       int64
}

type Service struct {
	serverId int64
	nc       *natsclient.ClusterClient
	log      *logger.Logger
	isDebug  bool

	acceptor       *stdtcp.Acceptor
	um             sync.Map
	onlineUser     map[int64]int64
	onlineUserLock sync.Mutex

	svc              *service.Service
	gvgguildServerID int64
}

func (this_ *Service) onlineInc(language int64) {
	this_.onlineUserLock.Lock()
	this_.onlineUser[language]++
	this_.onlineUserLock.Unlock()
}

func (this_ *Service) onlineDec(language int64) {
	this_.onlineUserLock.Lock()
	this_.onlineUser[language]--
	this_.onlineUserLock.Unlock()
}

func (this_ *Service) onlineCount() map[int64]int64 {
	out := make(map[int64]int64, 16)
	this_.onlineUserLock.Lock()
	for k, v := range this_.onlineUser {
		out[k] = v
	}
	this_.onlineUserLock.Unlock()
	return out
}

func NewService(serverId int64, listen string, urls []string, log *logger.Logger, isDebug bool) *Service {
	svc := service.NewService(urls, log, serverId, models.ServerType_GatewayStdTcp, true, true)
	nc := svc.GetNatsClient()
	s := &Service{
		serverId:   serverId,
		nc:         nc,
		isDebug:    isDebug,
		log:        log,
		svc:        svc,
		onlineUser: map[int64]int64{},
	}
	gvgguildServerID := env.GetInteger(env.GUILD_GVG_SERVER_ID)
	s.gvgguildServerID = gvgguildServerID
	var err error
	s.acceptor, err = stdtcp.NewAcceptor(listen, log, isDebug, s, false)
	if err != nil {
		panic(err)
	}
	return s
}

func (this_ *Service) Online(roleId values.RoleId, ui *UserInfo) {
	this_.um.Store(roleId, ui)
	this_.onlineInc(ui.Language)
	metrics.OnlineUserTotal.Inc()
}

func (this_ *Service) Offline(roleId values.RoleId) {
	v, ok := this_.um.Load(roleId)
	if ok {
		ui := v.(*UserInfo)
		this_.um.Delete(roleId)
		this_.onlineDec(ui.Language)
		metrics.OnlineUserTotal.Dec()
	}
}

const PINGDuration = time.Second * 15

func (this_ *Service) OnConnected(session *stdtcp.Session) {
	err := this_.SendPING(session)
	if err != nil {
		session.Close(err)
		return
	}
	timer.Ticker(PINGDuration, func() bool {
		if timer.Now().Sub(session.LastReadTime()) >= PINGDuration*3 {
			session.Close(stdtcp.PINGPONGTimeout)
			return false
		}

		err := this_.SendPING(session)
		if err != nil {
			session.Close(err)
			return false
		}
		return true
	})

}

func (this_ *Service) SendPING(session *stdtcp.Session) *errmsg.ErrMsg {
	ping := &models.PING{Now: timer.UnixNano()}
	return session.Send(nil, ping)
}

var (
	BattlePrefix       = "newbattle"
	CPPBattlePrefix    = "cppbattle"
	CPPCopyPrefix      = "cppcopy"
	GuildFilter        = "guild_filter_service"
	RecommendPrefix    = "recommend"
	DungeonMatchPrefix = "dungeon_match"
	GenRankPrefix      = "gen_rank"
	RoguelikePrefix    = "roguelike_match"
	GVGGuildPrefix     = "gvgguild"
)

// AssignServerId 分配server_id
func (this_ *Service) AssignServerId(serverId values.ServerId) values.ServerId {
	if serverId != 0 {
		return serverId
	}
	return serverids.Assign()
}

func (this_ *Service) OnRequest(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) {
	now := timer.Now()

	var reqServerId, toServerId values.ServerId
	defer func() {
		this_.log.Debug("OnRequest", zap.String("msgName", msgName),
			zap.Duration("useTime", timer.Now().Sub(now)),
			zap.Int64("reqServerId", reqServerId), zap.Int64("toServerId", toServerId))
	}()

	var info *UserInfo
	if msgName == loginMsgID {
		login := &lessservicepb.User_RoleLoginRequest{}
		err := login.Unmarshal(protocol.GetMsgDataFromRaw(frame))
		if err != nil {
			e := session.RPCResponseError(rpcIndex, nil, errmsg.NewProtocolError(err))
			if e != nil {
				this_.log.Warn("session.RPCResponseError error", zap.Error(e))
			}
			return
		}
		reqServerId = login.ServerId

		info = &UserInfo{
			RuleVersion:   login.RuleVersion,
			ServerId:      this_.AssignServerId(reqServerId),
			UserId:        login.IggId,
			ClientVersion: login.ClientVersion,
			//Language:      login.Language,
		}
		if info.UserId == "" {
			info.UserId = login.UserId
		}
		if info.UserId == "" {
			this_.CloseSession(session, errorNoLogin)
			return
		}
	} else {
		meta := session.GetMeta()
		if meta == nil {
			this_.CloseSession(session, errorNoLogin)
			return
		}
		var ok bool
		info, ok = meta.(*UserInfo)
		if !ok {
			this_.CloseSession(session, errorNoLogin)
			return
		}
	}
	ip := session.RemoteIP()
	sh := &models.ServerHeader{}
	sh.Ip = ip
	sh.ServerId = this_.serverId
	sh.ServerType = models.ServerType_GatewayStdTcp
	sh.RoleId = info.RoleId
	sh.UserId = info.UserId
	sh.RuleVersion = info.RuleVersion
	sh.StartTime = timer.UnixNano()
	sh.TraceId = xid.NewWithTime(timer.Now()).String()
	sh.GateId = this_.serverId
	sh.InServerId = info.ServerId
	sh.BattleServerId = info.BattleServerId
	sh.BattleMapId = info.MapId

	toServerId = info.ServerId
	i := strings.IndexByte(msgName, '.')
	preMsgName := msgName[:i]
	switch preMsgName {
	case BattlePrefix, CPPBattlePrefix, CPPCopyPrefix:
		toServerId = info.BattleServerId
	case GuildFilter, RecommendPrefix, GenRankPrefix:
		toServerId = 0
	case RoguelikePrefix:
		toServerId = env.GetRoguelikeServerId()
	case GVGGuildPrefix:
		toServerId = this_.gvgguildServerID
	}
	outData, err := this_.nc.RequestData(toServerId, sh, msgName, protocol.GetMsgDataFromRaw(frame))
	if err != nil {
		goto requestResponse
	}
	if msgName == loginMsgID && err == nil {
		resp := &models.Resp{}
		err = protocol.DecodeInternal(outData, nil, resp)
		if err != nil {
			goto requestResponse
		}
		lr := &lessservicepb.User_RoleLoginResponse{}
		if resp.ErrCode != 0 {
			err = (*errmsg.ErrMsg)(resp)
			goto requestResponse
		} else {
			e := lr.Unmarshal(resp.Resp.Value)
			if e != nil {
				err = errmsg.NewProtocolError(e)
				goto requestResponse
			}
		}

		if lr.Status == models.Status_SUCCESS && err == nil {
			serverId := lr.ServerId
			if reqServerId != 0 {
				serverId = reqServerId
			}
			ui := &UserInfo{
				ServerId:       serverId,
				RoleId:         lr.RoleId,
				UserId:         info.UserId,
				Session:        session,
				RuleVersion:    info.RuleVersion,
				MapId:          lr.MapId,
				BattleServerId: lr.BattleServerId,
				LoginTime:      sh.StartTime,
				Language:       lr.Language,
				ClientVersion:  info.ClientVersion,
			}
			this_.NotifyClientLoginOther(lr.RoleId, sh.StartTime)
			this_.Online(lr.RoleId, ui)
			session.SetMeta(ui)
			this_.log.Info("login", zap.String("user_id", ui.UserId), zap.String("role_id", ui.RoleId),
				zap.Int64("server_id", ui.ServerId), zap.String("rule_version", ui.RuleVersion))

			err = this_.svc.GetNatsClient().Publish(0, nil, &broadcast.GatewayStdTcp_LoginCheck{
				RoleId:    lr.RoleId,
				LoginTime: sh.StartTime,
			})
		} else if lr.Status == models.Status_FREEZE {
			err = errmsg.NewErrUserIdLoginLimit()
			goto requestResponse
		}
	}

requestResponse:

	if err == nil {
		n := protocol.GetEncodeInternalDataLen(respMsgID, outData)
		d := bytespool.GetSample(n)
		err = protocol.EncodeTCPInternalDataFrom(d, uint8(models.MsgType_response), rpcIndex, respMsgID, outData)
		if err != nil {
			return
		}
		err = session.Write(d)
		if err != nil {
			this_.log.Error("OnRequest Write error", zap.Error(err))
		}
	} else {
		errMsg := (*models.Resp)(err)
		n := protocol.GetEncodeLen(nil, errMsg)
		d := bytespool.GetSample(n)
		err = protocol.EncodeTCPFrom(d, uint8(models.MsgType_response), rpcIndex, nil, errMsg)
		if err != nil {
			return
		}
		err = session.Write(d)
		if err != nil {
			this_.log.Error("OnRequest Write error", zap.Error(err))
		}
	}
}

func (this_ *Service) NotifyClientLoginOther(roleId string, loginTime int64) {
	v, ok := this_.um.Load(roleId)
	if ok {
		oldUI, ok := v.(*UserInfo)
		if ok {
			_ = oldUI.Session.Send(nil, &servicepb.User_OtherLoginPush{LoginTime: loginTime})
			timer.AfterFunc(time.Second*3, func() {
				oldUI.Session.Close(nil)
			})
			this_.Offline(roleId)
		}
	}
}

func (this_ *Service) OnDisconnected(session *stdtcp.Session, _ error) {
	meta := session.GetMeta()
	if meta == nil {
		return
	}
	info, ok := meta.(*UserInfo)
	if !ok || info == nil {
		return
	}
	v, ok := this_.um.Load(info.RoleId)
	if ok {
		ui, ok := v.(*UserInfo)
		if ok {
			if ui.LoginTime == info.LoginTime {
				this_.Offline(info.RoleId)
			} else {
				return
			}
		}
	}

	this_.log.Info("logout", zap.String("user_id", info.UserId), zap.String("role_id", info.RoleId), zap.Int64("server_id", info.ServerId))
	ip := session.RemoteIP()
	sh := &models.ServerHeader{}
	sh.Ip = ip
	sh.ServerId = this_.serverId
	sh.ServerType = models.ServerType_GatewayStdTcp
	sh.RoleId = info.RoleId
	sh.UserId = info.UserId
	sh.RuleVersion = info.RuleVersion
	sh.StartTime = timer.UnixNano()
	sh.TraceId = xid.NewWithTime(timer.Now()).String()
	sh.GateId = this_.serverId
	sh.InServerId = info.ServerId
	sh.BattleServerId = info.BattleServerId
	sh.BattleMapId = info.MapId
	e := this_.nc.Publish(info.ServerId, sh, &lessservicepb.User_RoleLogoutPush{ClientVersion: info.ClientVersion})
	if e != nil {
		this_.log.Warn("notify logout error", zap.Error(e), zap.String("role_id", info.RoleId))
	}

	// e = this_.nc.Publish(info.BattleServerId, sh, &cppbattle.CPPBattle_LeaveArea{})
	// if e != nil {
	//	this_.log.Warn("notify logout error: CPPBattle_LeaveArea", zap.Error(e), zap.String("role_id", info.RoleId))
	// }
	//
	// e = this_.nc.Publish(info.BattleServerId, sh, &newbattlepb.NewBattle_L5LeaveRequest{})
	// if e != nil {
	//	this_.log.Warn("notify logout error", zap.Error(e), zap.String("role_id", info.RoleId))
	// }
}

var (
	loginMsgID       = proto.MessageName(&lessservicepb.User_RoleLoginRequest{})
	pongMsgID        = proto.MessageName(&models.PONG{})
	roleLogoutPushID = proto.MessageName(&lessservicepb.User_RoleLogoutPush{})
	errorNoLogin     = errors.New("not_login_auth")
	respMsgID        = proto.MessageName(&models.Resp{})
)

func (this_ *Service) OnMessage(session *stdtcp.Session, msgName string, frame []byte) {
	if msgName == pongMsgID {
		return
	}
	if msgName == roleLogoutPushID {
		session.Close(nil)
		return
	}

	header := &models.ServerHeader{}
	msgData, err := protocol.DecodeHeaderInternal(frame, header)
	if err != nil {
		this_.CloseSession(session, err)
		return
	}

	meta := session.GetMeta()
	if meta == nil {
		this_.CloseSession(session, errorNoLogin)
		return
	}
	var ok bool
	info, ok := meta.(*UserInfo)
	if !ok {
		this_.CloseSession(session, errorNoLogin)
		return
	}
	ip := session.RemoteIP()
	sh := &models.ServerHeader{}
	sh.Ip = ip
	sh.ServerId = this_.serverId
	sh.ServerType = models.ServerType_GatewayStdTcp
	sh.RoleId = info.RoleId
	sh.UserId = info.UserId
	sh.RuleVersion = info.RuleVersion
	sh.StartTime = timer.UnixNano()
	sh.TraceId = xid.NewWithTime(timer.Now()).String()
	sh.GateId = this_.serverId
	sh.InServerId = info.ServerId
	sh.BattleServerId = info.BattleServerId
	toServerId := info.ServerId
	i := strings.IndexByte(msgName, '.')
	preMsgName := msgName[:i]
	switch preMsgName {
	case BattlePrefix, CPPBattlePrefix, CPPCopyPrefix:
		toServerId = info.BattleServerId
		if toServerId == 0 {
			this_.log.Warn("server header has no BattleServerId")
			return
		}
	case GuildFilter, RecommendPrefix, GenRankPrefix:
		toServerId = 0
	case RoguelikePrefix:
		toServerId = env.GetRoguelikeServerId()
	case GVGGuildPrefix:
		toServerId = this_.gvgguildServerID
	}
	err = this_.nc.PublishRawData(toServerId, sh, msgName, msgData)
	if err != nil {
		this_.log.Error("nats Publish error", zap.Error(err), zap.String("req", msgName))
	}
}

func (this_ *Service) CloseSession(session *stdtcp.Session, err error) {
	session.Close(err)
}

func (this_ *Service) Close() {
	this_.acceptor.Close()
	this_.nc.Close()
}

func (this_ *Service) Run() {
	this_.Router()
	this_.svc.Start(func(event interface{}) {
		this_.log.Warn("unknown event", zap.Any("event", event))
	}, false)
	this_.acceptor.Start()
}

func (this_ *Service) Router() {
	this_.svc.RegisterEvent("处理推送消息", this_.DealPushMessage)
	this_.svc.RegisterEvent("处理单个推送消息", this_.DealPushMessageOne)
	this_.svc.RegisterEvent("玩家战斗服改变推送", this_.ChangeBattleId)
	this_.svc.RegisterEvent("其他玩家顶号检查", this_.OtherLoginCheck)
	this_.svc.RegisterEvent("踢玩家下线", this_.KickOffUserPush)
	this_.svc.RegisterFunc("获取在线玩家数量", this_.GetOnlineCount)
}

func (this_ *Service) GetOnlineCount(_ *ctx.Context, in *gatewaytcp.GatewayStdTcp_GetOnlineCountRequest) (*gatewaytcp.GatewayStdTcp_GetOnlineCountResponse, *errmsg.ErrMsg) {
	oc := this_.onlineCount()
	count := int64(0)
	for _, v := range oc {
		count += v
	}
	return &gatewaytcp.GatewayStdTcp_GetOnlineCountResponse{
		Count:         count,
		LanguageCount: oc,
	}, nil
}

func (this_ *Service) KickOffUserPush(_ *ctx.Context, in *broadcast.GatewayStdTcp_KickOffUserPush) {
	v, ok := this_.um.Load(in.RoleId)
	if !ok {
		return
	}
	ui, ok := v.(*UserInfo)
	if !ok {
		return
	}
	_ = ui.Session.Send(nil, &servicepb.User_KickOffSelfPush{Status: in.Status})
	timer.AfterFunc(time.Second*5, func() {
		ui.Session.Close(errors.New("kick off"))
	})
	this_.Offline(in.RoleId)
}

func (this_ *Service) OtherLoginCheck(_ *ctx.Context, in *broadcast.GatewayStdTcp_LoginCheck) {
	v, ok := this_.um.Load(in.RoleId)
	if !ok {
		return
	}

	ui, ok := v.(*UserInfo)
	if !ok {
		return
	}

	if ui.LoginTime != in.LoginTime {
		_ = ui.Session.Send(nil, &servicepb.User_OtherLoginPush{LoginTime: in.LoginTime})
		timer.AfterFunc(time.Second*3, func() {
			ui.Session.Close(nil)
		})
		this_.Offline(in.RoleId)
	}
}

func (this_ *Service) ChangeBattleId(c *ctx.Context, in *gatewaytcp.GatewayStdTcp_UserChangeBattleId) {
	v, ok := this_.um.Load(c.RoleId)
	if ok {
		info, ok := v.(*UserInfo)
		if ok {
			info.MapId = in.MapId
			info.BattleServerId = in.BattleServerId
		}
	}
}

func (this_ *Service) DealPushMessageOne(_ *ctx.Context, in *gatewaytcp.GatewayStdTcp_PushToClient) {
	if len(in.Messages) == 0 {
		return
	}
	resp := &models.Resp{
		OtherMsg: in.Messages,
	}
	n := protocol.GetEncodeLen(nil, resp)
	d := bytespool.GetSample(n)
	err := protocol.EncodeTCPFrom(d, uint8(models.MsgType_push), 0, nil, resp)
	if err != nil {
		this_.log.Warn("protocol.EncodeTCPFrom error", zap.Error(err))
		bytespool.PutSample(d)
		return
	}
	for _, roleId := range in.Roles {
		vv, ok := this_.um.Load(roleId)
		if ok {
			ui, ok1 := vv.(*UserInfo)
			if ok1 {
				d1 := bytespool.GetSample(n)
				copy(d1, d)
				_ = ui.Session.Write(d1)
			}
		}
	}
	bytespool.PutSample(d)
}

func (this_ *Service) DealPushMessage(_ *ctx.Context, in *gatewaytcp.GatewayStdTcp_PushManyToClient) {
	if len(in.Pcs) == 0 {
		return
	}
	for _, v := range in.Pcs {
		resp := &models.Resp{
			OtherMsg: v.Messages,
		}
		n := protocol.GetEncodeLen(nil, resp)
		d := bytespool.GetSample(n)
		err := protocol.EncodeTCPFrom(d, uint8(models.MsgType_push), 0, nil, resp)
		if err != nil {
			this_.log.Warn("protocol.EncodeTCPFrom error", zap.Error(err))
			bytespool.PutSample(d)
			continue
		}
		for _, roleId := range v.Roles {
			vv, ok := this_.um.Load(roleId)
			if ok {
				ui, ok1 := vv.(*UserInfo)
				if ok1 {
					d1 := bytespool.GetSample(n)
					copy(d1, d)
					_ = ui.Session.Write(d1)
				}
			}
		}
		bytespool.PutSample(d)
	}
}
