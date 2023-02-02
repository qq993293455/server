package player

import (
	"encoding/json"
	"math/rand"
	"sync/atomic"
	"time"

	"coin-server/client-guild-gvg/db"
	"coin-server/common/logger"
	"coin-server/common/msgcreate"
	"coin-server/common/network/stdtcp"
	"coin-server/common/proto/gvgguild"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/protocol"
	"coin-server/common/utils"
	"coin-server/rule"

	"github.com/rs/xid"

	"go.uber.org/zap"
)

type User struct {
	UserId     string
	RoleId     string
	SetLevel   int64
	GuildId    string
	isLockAll  uint64
	EnterGuild bool
	IsCreator  bool
}

var (
	isMustOpenBossHall int64
	canContinueChan    = make(chan struct{})
	canContinue        bool
)

type Client struct {
	log  *logger.TraceLogger
	Sess *Connection
	User
	db *db.DB
}

func NewClient(userId string, log *logger.Logger, userLevel int64, sd *db.DB) *Client {
	ct := &Connection{
		Sess: nil,
	}

	c := &Client{
		log:  log.WithTrace(xid.New().String(), userId),
		Sess: ct,
		User: User{
			UserId:   userId,
			SetLevel: userLevel,
		},
	}
	ct.HandleOnConnected = c.LogicOnConnected
	ct.HandleOnDisconnected = c.LogicOnDisconnected
	ct.HandleOnMessage = c.LogicOnMessage
	ct.HandleOnRequest = c.LogicOnRequest

	return c
}

func (this_ *Client) Save() {
	value, err := json.Marshal(this_.User)
	if err != nil {
		panic(err)
	}
	this_.db.Save(&db.KV{
		Key:   this_.UserId,
		Value: string(value),
	})
}

func (this_ *Client) Connect(addr string, log *logger.Logger) {
	stdtcp.Connect(addr, time.Second*3, true, this_.Sess, log, true)
}

func (this_ *Client) LogicOnConnected(session *stdtcp.Session) {
	this_.log.Debug("logic connect success", zap.String("userId", this_.UserId))
	this_.log.Debug("start login", zap.String("userId", this_.UserId))
	rl := &lessservicepb.User_RoleLoginRequest{
		UserId:        this_.UserId,
		ServerId:      1,
		AppKey:        "",
		Language:      0,
		RuleVersion:   "",
		Version:       0,
		ClientVersion: "0.0.1",
	}
	session.SetMeta(rl.UserId)
	respLogin := &lessservicepb.User_RoleLoginResponse{}
	err := session.RPCRequestOut(nil, rl, respLogin)
	if err != nil {
		this_.log.Error("login failed", zap.String("status", respLogin.Status.String()), zap.Error(err))
		return
	}
	if respLogin.Status != models.Status_SUCCESS {
		this_.log.Error("login failed", zap.String("status", respLogin.Status.String()))
		return
	}
	this_.log.Debug("login success", zap.String("userId", this_.UserId), zap.String("roleId", respLogin.RoleId))
	this_.RoleId = respLogin.RoleId
	this_.Save()
	this_.DoLogic()
}

func RandFlag() int64 {
	list := rule.MustGetReader(nil).GuildSign.List()
	utils.MustTrue(len(list) > 0)
	return list[rand.Int()%len(list)].Id
}

func RandLanguage() int64 {
	list := rule.MustGetReader(nil).VerifyLanguage.List()
	utils.MustTrue(len(list) > 0)
	return list[rand.Int()%len(list)].Id
}

func (this_ *Client) DoLogic() error {
	{
		this_.log.Debug("开始修改玩家等级", zap.Int64("level", this_.SetLevel), zap.String("userId", this_.UserId), zap.String("roleId", this_.RoleId))
		csr := &lessservicepb.User_CheatSetLevelRequest{Level: this_.SetLevel}
		csrOut := &lessservicepb.User_CheatSetLevelResponse{}
		err := this_.Sess.Sess.RPCRequestOut(nil, csr, csrOut)
		if err != nil {
			this_.log.Error("修改玩家等级失败", zap.Int64("level", this_.SetLevel), zap.String("userId", this_.UserId), zap.String("roleId", this_.RoleId), zap.Error(err))
			return err
		}
		this_.log.Debug("修改玩家等级到成功", zap.Int64("level", this_.SetLevel), zap.String("userId", this_.UserId), zap.String("roleId", this_.RoleId))
	}
	{ //解锁所有系统
		if atomic.AddUint64(&this_.isLockAll, 1) == 1 {
			this_.log.Debug("解锁所有系统")
			mbo := &servicepb.BossHall_MustOpenRequest{Open: true}
			mboOut := &servicepb.BossHall_MustOpenResponse{}
			err := this_.Sess.Sess.RPCRequestOut(nil, mbo, mboOut)
			if err != nil {
				this_.log.Error("GM设置恶魔秘境强制开启失败", zap.String("userId", this_.UserId), zap.String("roleId", this_.RoleId), zap.Error(err))
				return err
			}
			this_.log.Debug("解锁所有系统成功")
		}
		// 给自己加钻石
		this_.log.Debug("加钻石1000000")
		addItem := &servicepb.Bag_CheatAddItemRequest{ItemId: 102, Count: 10000000}
		addItemOut := &servicepb.Bag_CheatAddItemResponse{}
		err := this_.Sess.Sess.RPCRequestOut(nil, addItem, addItemOut)
		if err != nil {
			this_.log.Error("加钻石1000000失败", zap.String("userId", this_.UserId), zap.String("roleId", this_.RoleId), zap.Error(err))
			return err
		}
		this_.log.Debug("加钻石1000000成功")
	}
	{ // 创建工会
		if this_.GuildId == "" {
			this_.log.Debug("创建工会")
			createGuild := &lessservicepb.Guild_GuildCreateRequest{
				Name:     this_.UserId,
				Flag:     RandFlag(),
				Lang:     RandLanguage(),
				Intro:    "guild-gvg-test-client:" + this_.UserId,
				Notice:   "guild-gvg-test-client:" + this_.UserId,
				AutoJoin: true,
			}
			createGuildOut := &lessservicepb.Guild_GuildCreateResponse{}
			err := this_.Sess.Sess.RPCRequestOut(nil, createGuild, createGuildOut)
			if err != nil {
				this_.log.Error("创建工会失败", zap.String("roleId", this_.RoleId), zap.Error(err))
				return err
			}
			this_.GuildId = createGuildOut.Info.Id
			this_.EnterGuild = true
			this_.IsCreator = true
			this_.Save()
			this_.log.Debug("创建工会成功")

			this_.log.Debug("提升工会等级")
			for {
				guildBuilding := &lessservicepb.Guild_GuildBuildRequest{}
				guildBuildingOut := &lessservicepb.Guild_GuildBuildResponse{}
				err := this_.Sess.Sess.RPCRequestOut(nil, guildBuilding, guildBuildingOut)
				if err != nil {
					this_.log.Error("提升工会等级失败", zap.String("roleId", this_.RoleId), zap.Error(err))
					return err
				}
				if guildBuildingOut.Level >= 3 {
					break
				}
			}
			this_.log.Debug("提升工会等级成功")
		} else if !this_.EnterGuild && this_.GuildId != "" { // 加入工会
			this_.log.Debug("加入工会")
			joinGuild := &lessservicepb.Guild_GuildJoinApplyRequest{Id: []string{this_.GuildId}}
			joinGuildOut := &lessservicepb.Guild_GuildJoinApplyResponse{}
			err := this_.Sess.Sess.RPCRequestOut(nil, joinGuild, joinGuildOut)
			if err != nil {
				this_.log.Error("加入工会失败", zap.String("guild", this_.GuildId), zap.String("roleId", this_.RoleId), zap.Error(err))
				return err
			}
			this_.EnterGuild = true
			this_.Save()
			this_.log.Debug("加入工会成功")
		}
	}
	return nil
}

func (this_ *Client) DoGVG() error {

	this_.log.Debug("查询是否可以报名")
	queryStatus := &gvgguild.GuildGVG_QueryStatusRequest{}
	queryStatusOut := &gvgguild.GuildGVG_QueryStatusResponse{}
	err := this_.Sess.Sess.RPCRequestOut(nil, queryStatus, queryStatusOut)
	if err != nil {
		this_.log.Error("查询是否可以报名失败", zap.String("userId", this_.UserId), zap.String("roleId", this_.RoleId), zap.Error(err))
		return err
	}
	this_.log.Debug("查询是否可以报名成功", zap.String("status", queryStatusOut.Status.String()))
	if queryStatusOut.Status != gvgguild.GuildGVG_CanSignup && queryStatusOut.Status != gvgguild.GuildGVG_Fighting {
		return nil
	}

	if this_.IsCreator {

		if queryStatusOut.Status == gvgguild.GuildGVG_CanSignup {

			this_.log.Debug("添加活跃度")
			addItem := &servicepb.Bag_CheatAddItemRequest{ItemId: 109, Count: 10000000}
			addItemOut := &servicepb.Bag_CheatAddItemResponse{}
			err = this_.Sess.Sess.RPCRequestOut(nil, addItem, addItemOut)
			if err != nil {
				this_.log.Error("添加活跃度失败", zap.String("userId", this_.UserId), zap.String("roleId", this_.RoleId), zap.Error(err))
				return err
			}
			this_.log.Debug("添加活跃度成功")

			this_.log.Debug("开始报名GVG")
			signup := &gvgguild.GuildGVG_SignupRequest{}
			signupOut := &gvgguild.GuildGVG_SignupResponse{}
			err = this_.Sess.Sess.RPCRequestOut(nil, signup, signupOut)
			if err != nil {
				this_.log.Error("开始报名GVG失败", zap.String("userId", this_.UserId), zap.String("roleId", this_.RoleId), zap.Error(err))
				return err
			}
			this_.log.Debug("开始报名GVG成功")
		}

	}

	if queryStatusOut.Status == gvgguild.GuildGVG_Fighting {
		this_.DoFight()
	}

	return nil
}

func (this_ *Client) DoFight() {

}

func (this_ *Client) LogicOnDisconnected(session *stdtcp.Session, err error) {
	this_.log.Debug("disconnect", zap.String("userId", this_.UserId), zap.String("roleId", this_.RoleId), zap.Error(err))
}

func (this_ *Client) LogicOnMessage(session *stdtcp.Session, msgName string, frame []byte) {
	h := &models.ServerHeader{}
	msg := msgcreate.NewMessage(msgName)
	err := protocol.DecodeInternal(frame, h, msg)
	if err != nil {
		panic(err)
	}

	if msgName == (&models.PING{}).XXX_MessageName() {
		_ = session.Send(nil, &models.PONG{})
	} else {
		this_.log.Debug("On Message", zap.String("msgName", msgcreate.MessageName(msg)), zap.Any("msg", msg))
	}
}

func (this_ *Client) LogicOnRequest(session *stdtcp.Session, rpcIndex uint32, msgName string, frame []byte) {

}
