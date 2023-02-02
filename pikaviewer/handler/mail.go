package handler

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"coin-server/common/dlock"
	"coin-server/common/logger"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/redisclient"
	utils2 "coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/pikaviewer/model"
	"coin-server/pikaviewer/utils"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type Mail struct {
}

type SendForm struct {
	SendWay     int      `json:"send_way" binding:"required"`
	IggId       []int    `json:"igg_id"`
	Sender      string   `json:"sender"`
	TextId      int64    `json:"text_id"`
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Hi          string   `json:"hi"`
	ActivatedAt int64    `json:"activated_at"`
	ExpiredAt   int64    `json:"expired_at"`
	Args        []string `json:"args"`
	Attachment  string   `json:"attachment"`
	ForAll      bool     `json:"for_all"`
}

type DeleteForm struct {
	Entire bool     `json:"entire"`
	UId    uint64   `json:"uid"`
	MailId []string `json:"mailId" binding:"required"`
}

func (h *Mail) Send(req *SendForm) ([]string, error) {
	if req.SendWay == 2 {
		return h.SendEntireMail(req)
	}
	if len(req.IggId) <= 0 {
		return nil, utils.NewDefaultErrorWithMsg("请输入接收邮件的玩家iggId")
	}
	// roleIds := make([]string, 0, len(req.UId))
	// for _, uid := range req.UId {
	// 	roleIds = append(roleIds, utils2.Base34EncodeToString(uid))
	// }
	p := model.NewPlayer()
	// roles, err := p.GetRoles(roleIds)
	// if err != nil {
	// 	return nil, err
	// }
	// if len(roles) <= 0 {
	// 	return roleIds, nil
	// }
	// if len(roles) > 100 {
	// 	return roleIds, utils.NewDefaultErrorWithMsg("一次最多支持100个玩家发放")
	// }
	userIds := make([]string, 0)
	// userRoleMap := make(map[string]values.RoleId)
	for _, iggId := range req.IggId {
		userIds = append(userIds, strconv.Itoa(iggId))
		// userRoleMap[role.UserId] = role.RoleId
	}
	data, err := p.GetServerId(userIds)
	if err != nil {
		return nil, err
	}
	successMap := make(map[values.RoleId]struct{})
	mail, err := h.buildMail(req)
	if err != nil {
		return nil, err
	}
	for roleId, serverId := range data {
		if err := utils.NATS.Publish(serverId, &models.ServerHeader{
			StartTime:  time.Now().UnixNano(),
			RoleId:     roleId,
			ServerId:   serverId,
			ServerType: models.ServerType_GMServer,
			InServerId: serverId,
		}, &servicepb.GM_SendMail{
			Mail: mail,
		}); err != nil {
			break
		}
		successMap[roleId] = struct{}{}
	}
	failed := make([]values.RoleId, 0)
	for roleId := range data {
		if _, ok := successMap[roleId]; ok {
			continue
		}
		failed = append(failed, roleId)
	}
	return failed, nil
}

const entireServerMailLock = "lock:entire:server"

func (h *Mail) SendEntireMail(req *SendForm) ([]string, error) {
	locker := dlock.GetLocker()
	if err := locker.Lock(redisclient.GetLocker(), time.Second*5, entireServerMailLock); err != nil {
		logger.DefaultLogger.Error("SendEntireMail lock err",
			zap.Any("req", req),
			zap.Error(err),
		)
		return nil, utils.NewDefaultErrorWithMsg(err.Error())
	}
	defer func() {
		if err := locker.Unlock(); err != nil {
			logger.DefaultLogger.Error("SendEntireMail unlock pay lock err",
				zap.Error(err),
			)
		}
		dlock.PutLocker(locker)
	}()
	p := model.NewPlayer()
	mails, err := p.EntireMail()
	if err != nil {
		return nil, err
	}
	mail, err := h.buildMail(req)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	newEntireMails := make([]*dao.MailItem, 0)
	for _, item := range mails {
		// 已过期
		if item.ExpiredAt <= now {
			continue
		}
		newEntireMails = append(newEntireMails, item)
	}
	newEntireMails = append(newEntireMails, &dao.MailItem{
		Id:          xid.New().String(),
		Type:        mail.Type,
		Sender:      mail.Sender,
		TextId:      mail.TextId,
		Title:       mail.Title,
		Content:     mail.Content,
		Hi:          mail.Hi,
		ExpiredAt:   mail.ExpiredAt,
		Args:        mail.Args,
		Attachment:  mail.Attachment,
		Read:        false,
		CreatedAt:   mail.CreatedAt,
		DeletedAt:   0,
		ActivatedAt: mail.ActivatedAt,
		ForAll:      req.ForAll,
	})
	if err := p.SaveEntireMail(newEntireMails); err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *Mail) buildMail(req *SendForm) (*models.Mail, error) {
	attachment := make([]*models.Item, 0)
	if req.Attachment != "" {
		attachmentMap := make(map[int64]int64)
		if err := json.Unmarshal([]byte(req.Attachment), &attachmentMap); err != nil {
			return nil, errors.New("附件数据格式有误" + err.Error())
		}
		for id, count := range attachmentMap {
			attachment = append(attachment, &models.Item{
				ItemId: id,
				Count:  count,
			})
		}
	}
	return &models.Mail{
		Id:          "",
		Type:        models.MailType_MailTypeGM,
		Sender:      req.Sender,
		TextId:      req.TextId,
		Title:       req.Title,
		Content:     req.Content,
		Hi:          req.Hi,
		ExpiredAt:   req.ExpiredAt * 1000,
		Args:        req.Args,
		Attachment:  attachment,
		CreatedAt:   time.Now().UnixMilli(),
		ActivatedAt: req.ActivatedAt * 1000,
	}, nil
}

func (h *Mail) Delete(req *DeleteForm) error {
	if len(req.MailId) <= 0 {
		return utils.NewDefaultErrorWithMsg("请选择您要删除的邮件")
	}

	if req.Entire {
		return h.deleteEntireMail(req.MailId)
	}

	return h.deletePlayerMail(utils2.Base34EncodeToString(req.UId), req.MailId)
}

func (h *Mail) deleteEntireMail(mailId []string) error {
	p := model.NewPlayer()
	mail, err := p.EntireMail()
	if err != nil {
		return nil
	}
	list := make([]*dao.MailItem, 0)
	for _, item := range mail {
		var find bool
		for _, id := range mailId {
			if id == item.Id {
				find = true
				break
			}
		}
		if !find {
			list = append(list, item)
		}
	}
	return p.SaveEntireMail(list)
}

func (h *Mail) deletePlayerMail(uid string, mailId []string) error {
	p := model.NewPlayer()
	role, err := p.GetRole(uid)
	if err != nil {
		return err
	}
	if role == nil {
		return utils.NewDefaultErrorWithMsg("玩家不存在")
	}
	serverId, err := p.GetOneServerId(role.UserId)
	if err != nil {
		return err
	}
	if serverId == 0 {
		return utils.NewDefaultErrorWithMsg("未获取到当前玩家正确的server id，请联系后端程序")
	}
	return utils.NATS.Publish(serverId, &models.ServerHeader{
		StartTime:  time.Now().UnixNano(),
		RoleId:     role.RoleId,
		ServerId:   serverId,
		ServerType: models.ServerType_GMServer,
		InServerId: serverId,
	}, &servicepb.GM_DeleteMail{
		MailId: mailId,
	})
}

func (h *Mail) GetEntireMail() ([]*dao.MailItem, error) {
	return model.NewPlayer().EntireMail()
}
