package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/im"
	"coin-server/common/mail"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	gatewaypb "coin-server/common/proto/gatewaytcp"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/utils"
	"coin-server/common/values/enum"
	"coin-server/pikaviewer/global"
	"coin-server/pikaviewer/model"
	utils2 "coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"

	"github.com/rs/xid"
)

type Player struct {
	UID        uint64 `json:"uid"`
	Nickname   string `json:"nickname"`
	CreateTime int64  `json:"create_time"`
	LoginTime  int64  `json:"login_time"`
}

func (h *Player) GetPlayerInfo(name string) ([]*Player, int, error) {
	count, err := model.NewPlayer().CountByName(name)
	if err != nil {
		return nil, count, err
	}
	if count <= 0 {
		return nil, count, nil
	}
	data, err := model.NewPlayer().FindByName(name)
	if err != nil {
		return nil, count, err
	}
	list := make([]*Player, 0, len(data))
	for _, item := range data {
		list = append(list, &Player{
			UID:        utils.Base34DecodeString(item.RoleId),
			Nickname:   item.Nickname,
			CreateTime: item.CreateTime,
			LoginTime:  item.LoginTime,
		})
	}
	return list, count, nil
}

func (h *Player) MailInfo(uid uint64) ([]*dao.MailItem, error) {
	roleId := utils.Base34EncodeToString(uid)
	mails := make([]*dao.MailItem, 0)
	if err := orm.GetOrm(ctx.GetContext()).HGetAll(mail.GetMailRedis(), mail.GetMailKey(roleId), &mails); err != nil {
		return nil, err
	}
	list := make([]*dao.MailItem, 0)
	for _, item := range mails {
		if strings.Contains(item.Id, "entire@") && item.DeletedAt > 0 {
			continue
		}
		list = append(list, item)
	}
	return list, nil
}

func (h *Player) KickOffUser(iggId int, sec int64, status int64) error {
	user, ok, err := model.NewPlayer().GetUserData(strconv.Itoa(iggId))
	if err != nil {
		return err
	}
	if !ok {
		return utils2.NewDefaultErrorWithMsg("玩家不存在")
	}
	if _, err := utils2.NATS.RequestProto(user.ServerId, &models.ServerHeader{
		StartTime:  time.Now().UnixNano(),
		RoleId:     user.RoleId,
		ServerId:   user.ServerId,
		ServerType: models.ServerType_GMServer,
		InServerId: user.ServerId,
		TraceId:    xid.New().String(),
	}, &servicepb.User_KickOffUserRequest{KickoffSeconds: sec, Status: status}); err != nil {
		return utils2.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}

func (h *Player) GetPlayerInfoForSOP(val int, isIggId bool) (gin.H, error) {
	// TODO 限流

	var (
		err    error
		ok     bool
		roleId string
		uid    uint64
		user   *dao.User
		role   *dao.Role
	)
	if isIggId {
		user, ok, err = model.NewPlayer().GetUserData(strconv.Itoa(val))
		if err != nil {
			return nil, utils2.NewDefaultErrorWithMsg(err.Error())
		}
		if !ok {
			return nil, utils2.PlayerNotExist
		}

		roleId = user.RoleId
		uid = utils.Base34DecodeString(roleId)
		role, err = model.NewPlayer().GetRole(roleId)
		if err != nil {
			return nil, utils2.NewDefaultErrorWithMsg(err.Error())
		}
		if role == nil {
			return nil, utils2.PlayerNotExist
		}
	} else {
		uid = uint64(val)
		roleId = utils.Base34EncodeToString(uid)

		role, err = model.NewPlayer().GetRole(roleId)
		if err != nil {
			return nil, utils2.NewDefaultErrorWithMsg(err.Error())
		}
		if role == nil {
			return nil, utils2.PlayerNotExist
		}
		user, ok, err = model.NewPlayer().GetUserData(role.UserId)
		if err != nil {
			return nil, utils2.NewDefaultErrorWithMsg(err.Error())
		}
		if !ok {
			return nil, utils2.PlayerNotExist
		}
	}

	items, err := model.NewPlayer().GetItemCount(roleId, enum.CurrencyList...)
	if err != nil {
		return nil, err
	}

	return gin.H{
		"uid":         uid,
		"role_id":     roleId,
		"igg_id":      user.UserId,
		"device":      user.DeviceId,
		"ban":         user.FreezeTime,
		"server_id":   user.ServerId,
		"nick_name":   role.Nickname,
		"level":       role.Level,
		"power":       role.Power,
		"title":       role.Title,
		"language":    role.Language,
		"login":       role.Login,
		"logout":      role.Logout,
		"create_time": role.CreateTime,
		"recharge":    role.Recharge,
		"items":       items,
	}, nil
}

func (h *Player) BanPlayer(iggId int, d int64) error {
	user, ok, err := model.NewPlayer().GetUserData(strconv.Itoa(iggId))
	if err != nil {
		return utils2.NewDefaultErrorWithMsg(err.Error())
	}
	if !ok {
		return utils2.PlayerNotExist
	}

	var freezeTime int64
	if d > 0 {
		if err := h.KickOffUser(iggId, d, 1); err != nil {
			return err
		}
		freezeTime = time.Now().Add(time.Second * time.Duration(d)).Unix()
	}

	user.FreezeTime = freezeTime
	return model.NewPlayer().Ban(user)
}

func (h *Player) BanChat(iggId int, sec int) error {
	user, ok, err := model.NewPlayer().GetUserData(strconv.Itoa(iggId))
	if err != nil {
		return utils2.NewDefaultErrorWithMsg(err.Error())
	}
	if !ok {
		return utils2.PlayerNotExist
	}
	return im.DefaultClient.BanPost(context.Background(), user.RoleId, sec)
}

func (h *Player) UnBanChat(iggId int) error {
	user, ok, err := model.NewPlayer().GetUserData(strconv.Itoa(iggId))
	if err != nil {
		return utils2.NewDefaultErrorWithMsg(err.Error())
	}
	if !ok {
		return utils2.PlayerNotExist
	}
	return im.DefaultClient.UnBanPost(context.Background(), user.RoleId)
}

func (h *Player) OnlineCount() (map[int64]int64, error) {
	countMap := make(map[int64]int64)
	for _, serverId := range global.Config.GateWay {
		out := gatewaypb.GatewayStdTcp_GetOnlineCountResponse{}
		if err := utils2.NATS.RequestWithOut(ctx.GetContext(), serverId, &gatewaypb.GatewayStdTcp_GetOnlineCountRequest{}, &out); err != nil {
			return nil, utils2.NewDefaultErrorWithMsg(err.Error())
		}
		for l, c := range out.LanguageCount {
			countMap[l] += c
		}
	}

	return countMap, nil
}

func (h *Player) Total() (int64, error) {
	return model.NewPlayer().Total()
}

type UpdatePlayerCurrencyReq struct {
	IggId    int             `json:"igg_id" binding:"required"`
	Currency map[int64]int64 `json:"currency" binding:"required"`
}

func (h *Player) UpdatePlayerCurrency(req *UpdatePlayerCurrencyReq) error {
	user, ok, err := model.NewPlayer().GetUserData(strconv.Itoa(req.IggId))
	if err != nil {
		return utils2.NewDefaultErrorWithMsg(err.Error())
	}
	if !ok {
		return utils2.PlayerNotExist
	}
	role, err := model.NewPlayer().GetRole(user.RoleId)
	if err != nil {
		return utils2.NewDefaultErrorWithMsg(err.Error())
	}
	if role == nil {
		return utils2.PlayerNotExist
	}

	if _, err := utils2.NATS.RequestProto(user.ServerId, &models.ServerHeader{
		StartTime:  time.Now().UnixNano(),
		RoleId:     user.RoleId,
		ServerId:   user.ServerId,
		ServerType: models.ServerType_GMServer,
		InServerId: user.ServerId,
		TraceId:    xid.New().String(),
	}, &servicepb.User_UpdateUserCurrencyRequest{
		Currency: req.Currency,
	}); err != nil {
		return utils2.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}
