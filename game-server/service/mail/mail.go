package mail

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/redisclient"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	dao2 "coin-server/game-server/service/mail/dao"
	"coin-server/game-server/service/mail/rule"

	"go.uber.org/zap"

	"github.com/rs/xid"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func (svc *Service) Add(ctx *ctx.Context, roleId values.RoleId, mail *models.Mail, gm ...bool) *errmsg.ErrMsg {
	var fromGM bool
	if len(gm) > 0 {
		fromGM = gm[0]
	}
	if fromGM {
		role, err := svc.GetRoleByRoleId(ctx, roleId)
		if err != nil {
			return err
		}
		contentMap := make(map[int64]string)
		if err := json.Unmarshal([]byte(mail.Content), &contentMap); err != nil {
			return errmsg.NewInternalErr(err.Error())
		}
		titleMap := make(map[int64]string)
		if err := json.Unmarshal([]byte(mail.Title), &titleMap); err != nil {
			return errmsg.NewInternalErr(err.Error())
		}
		content, ok := contentMap[role.Language]
		if !ok {
			content = contentMap[enum.DefaultLanguage]
		}
		title, ok := titleMap[role.Language]
		if !ok {
			title = titleMap[enum.DefaultLanguage]
		}
		mail.Title = title
		mail.Content = content
	}
	svc.formatMail(ctx, mail)
	if err := ctx.DRLock(redisclient.GetLocker(), mailLock+roleId); err != nil {
		return err
	}
	if err := dao2.NewMail(roleId).Add(ctx, model2daoMailItem(mail, false)); err != nil {
		return nil
	}
	if err := svc.afterAdd(ctx); err != nil {
		return err
	}

	ctx.PublishEventLocal(&event.RedPointAdd{
		RoleId: roleId,
		Key:    enum.RedPointMailKey,
		Val:    1,
	})
	return nil
}

func (svc *Service) BatchAdd(ctx *ctx.Context, roleId values.RoleId, mailList []*models.Mail) *errmsg.ErrMsg {
	if err := ctx.DRLock(redisclient.GetLocker(), mailLock+roleId); err != nil {
		return err
	}
	newMailList := make([]*dao.MailItem, 0, len(mailList))
	for _, item := range mailList {
		svc.formatMail(ctx, item)
		mail := model2daoMailItem(item, false)
		newMailList = append(newMailList, mail)
	}

	if err := dao2.NewMail(roleId).BatchAdd(ctx, newMailList); err != nil {
		return nil
	}
	if err := svc.afterAdd(ctx); err != nil {
		return err
	}
	ctx.PublishEventLocal(&event.RedPointAdd{
		RoleId: roleId,
		Key:    enum.RedPointMailKey,
		Val:    values.Integer(len(newMailList)),
	})
	return nil
}

func (svc *Service) formatMail(ctx *ctx.Context, mail *models.Mail) {
	mail.Id = xid.New().String()
	if mail.Type == 0 {
		mail.Type = models.MailType_MailTypeSystem
	}
	now := timer.StartTime(ctx.StartTime)
	if mail.ExpiredAt <= 0 {
		mail.ExpiredAt = now.AddDate(0, 1, 0).UnixMilli()
	}
	if mail.CreatedAt <= 0 {
		mail.CreatedAt = now.UnixMilli()
	}
	if mail.ActivatedAt <= 0 {
		mail.ActivatedAt = now.UnixMilli()
	}
}

func NewMailService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		log:        log,
		Module:     module,
	}
	module.MailService = s
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取邮件列表", svc.List)
	svc.svc.RegisterFunc("阅读单封邮件", svc.Read)
	svc.svc.RegisterFunc("一键阅读所有邮件", svc.ReadAll)
	svc.svc.RegisterFunc("删除单封邮件", svc.Delete)
	svc.svc.RegisterFunc("一键删除所有已读邮件", svc.DeleteAll)

	svc.svc.RegisterFunc("cheat添加一封邮件", svc.CheatAdd)

	svc.svc.RegisterEvent("其他服务推送邮件", svc.SendMailFromOtherServer)

	h := svc.svc.Group(GMAuth)
	h.RegisterEvent("gm后台发送邮件", svc.GMAddMail)
	h.RegisterEvent("gm后台删除邮件", svc.GMDeleteMail)
}

// List 获取邮件列表
func (svc *Service) List(ctx *ctx.Context, req *servicepb.Mail_MailListRequest) (*servicepb.Mail_MailListResponse, *errmsg.ErrMsg) {
	if err := svc.updateRewardMail(ctx, req.Version); err != nil {
		return nil, err
	}
	unread, read, err := svc.mailList(ctx)
	if err != nil {
		return nil, err
	}
	return &servicepb.Mail_MailListResponse{
		Unread: unread,
		Read:   read,
	}, nil
}

// Read 阅读单封邮件
func (svc *Service) Read(ctx *ctx.Context, req *servicepb.Mail_ReadMailRequest) (*servicepb.Mail_ReadMailResponse, *errmsg.ErrMsg) {
	unread, _, err := svc.mailList(ctx)
	if err != nil {
		return nil, err
	}
	mail, ok := unread[req.Id]
	if !ok {
		return nil, errmsg.NewErrMailNotFound()
	}
	var hasAttachment bool
	if mail.Attachment != nil && len(mail.Attachment) > 0 {
		items := make(map[values.ItemId]values.Integer, len(mail.Attachment))
		for _, item := range mail.Attachment {
			items[item.ItemId] += item.Count
		}
		if err := svc.isBagEnough(ctx, items); err != nil {
			return nil, err
		}
		if _, err := svc.BagService.AddManyItem(ctx, ctx.RoleId, items); err != nil {
			return nil, err
		}
		hasAttachment = true
	}
	mailItem := model2daoMailItem(mail, true)

	svc.updateMailExpiredAt(ctx, mailItem, hasAttachment)

	if err := dao2.NewMail(ctx.RoleId).UpdateMail(ctx, mailItem); err != nil {
		return nil, err
	}
	ctx.PublishEventLocal(&event.RedPointChange{
		RoleId: ctx.RoleId,
		Key:    enum.RedPointMailKey,
		Val:    values.Integer(len(unread) - 1),
	})
	return &servicepb.Mail_ReadMailResponse{}, nil
}

// ReadAll 一键阅读所有邮件
func (svc *Service) ReadAll(ctx *ctx.Context, _ *servicepb.Mail_ReadAllMailRequest) (*servicepb.Mail_ReadAllMailResponse, *errmsg.ErrMsg) {
	unread, _, err := svc.mailList(ctx)
	if err != nil {
		return nil, err
	}
	items := make(map[values.ItemId]values.Integer)
	mailItems := make([]*dao.MailItem, 0, len(unread))
	for _, item := range unread {
		var hasAttachment bool
		if item.Attachment != nil && len(item.Attachment) > 0 {
			for _, attachmentItem := range item.Attachment {
				items[attachmentItem.ItemId] += attachmentItem.Count
			}
			hasAttachment = true
		}
		mailItem := model2daoMailItem(item, true)
		svc.updateMailExpiredAt(ctx, mailItem, hasAttachment)
		mailItems = append(mailItems, mailItem)
	}

	if len(items) > 0 {
		if err := svc.isBagEnough(ctx, items); err != nil {
			return nil, err
		}
		if _, err := svc.BagService.AddManyItem(ctx, ctx.RoleId, items); err != nil {
			return nil, err
		}
	}

	if err := dao2.NewMail(ctx.RoleId).UpdateMails(ctx, mailItems); err != nil {
		return nil, err
	}
	ctx.PublishEventLocal(&event.RedPointChange{
		RoleId: ctx.RoleId,
		Key:    enum.RedPointMailKey,
		Val:    values.Integer(0),
	})
	return &servicepb.Mail_ReadAllMailResponse{}, nil
}

// Delete 删除单封邮件
func (svc *Service) Delete(ctx *ctx.Context, req *servicepb.Mail_DeleteMailRequest) (*servicepb.Mail_DeleteMailResponse, *errmsg.ErrMsg) {
	unread, read, err := svc.mailList(ctx)
	if err != nil {
		return nil, err
	}
	mail, ok := unread[req.Id]
	if ok {
		return nil, errmsg.NewErrReadMailFirst()
	}
	mail, ok = read[req.Id]
	if !ok {
		return nil, errmsg.NewErrMailNotFound()
	}
	daoMail := model2daoMailItem(mail, true)
	if _, ok := svc.isEntireServerMail(mail.Id); ok {
		daoMail.DeletedAt = time.Now().UnixMilli()
		if err := dao2.NewMail(ctx.RoleId).UpdateMail(ctx, daoMail); err != nil {
			return nil, err
		}
	} else {
		if err := dao2.NewMail(ctx.RoleId).DeleteMail(ctx, daoMail); err != nil {
			return nil, err
		}
	}

	return &servicepb.Mail_DeleteMailResponse{}, nil
}

// DeleteAll 一键删除所有已读邮件
func (svc *Service) DeleteAll(ctx *ctx.Context, _ *servicepb.Mail_DeleteAllMailRequest) (*servicepb.Mail_DeleteAllMailResponse, *errmsg.ErrMsg) {
	_, read, err := svc.mailList(ctx)
	if err != nil {
		return nil, err
	}

	readMails := make([]*dao.MailItem, 0)
	entireMails := make([]*dao.MailItem, 0)
	for _, item := range read {
		daoMail := model2daoMailItem(item, true)
		if _, ok := svc.isEntireServerMail(daoMail.Id); ok {
			daoMail.DeletedAt = time.Now().UnixMilli()
			entireMails = append(entireMails, daoMail)
		} else {
			readMails = append(readMails, daoMail)
		}
	}
	if err := dao2.NewMail(ctx.RoleId).UpdateMails(ctx, entireMails); err != nil {
		return nil, err
	}
	dao2.NewMail(ctx.RoleId).DeleteMails(ctx, readMails)

	return &servicepb.Mail_DeleteAllMailResponse{}, nil
}

func (svc *Service) SendMailFromOtherServer(ctx *ctx.Context, req *servicepb.Mail_SendMailFromOtherServer) {
	svc.formatMail(ctx, req.Mail)
	err := svc.MailService.Add(ctx, ctx.RoleId, req.Mail, true)
	if err != nil {
		svc.log.Error("send other server mail err", zap.String("role_id", ctx.RoleId), zap.Any("mail", req.Mail))
	}
}

func (svc *Service) mailList(ctx *ctx.Context) (map[string]*models.Mail, map[string]*models.Mail, *errmsg.ErrMsg) {
	mailDao := dao2.NewMail(ctx.RoleId)
	list, err := mailDao.GetAll(ctx)
	if err != nil {
		return nil, nil, err
	}

	list, err = svc.entireServerMail(ctx, list)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().Unix() * 1000
	expired := make([]*dao.MailItem, 0)
	unread := make(map[string]*models.Mail, 0)
	read := make(map[string]*models.Mail, 0)
	for _, item := range list {
		if item.ExpiredAt <= now {
			expired = append(expired, item)
			continue
		}
		// 全服邮件被玩家删除了
		if item.ActivatedAt > now || item.DeletedAt > 0 {
			continue
		}
		mail := daoItem2model(item)
		if item.Read {
			read[item.Id] = mail
		} else {
			unread[item.Id] = mail
		}
	}
	mailDao.DeleteMails(ctx, expired)
	return unread, read, nil
}

func (svc *Service) entireServerMail(ctx *ctx.Context, list []*dao.MailItem) ([]*dao.MailItem, *errmsg.ErrMsg) {
	// if err := ctx.DRLock(redisclient.GetLocker(), entireServerMailLock); err != nil {
	// 	return nil, err
	// }
	mail, err := dao2.GetEntireMail(ctx)
	if err != nil {
		return list, err
	}
	if len(mail.Mails) <= 0 {
		return list, nil
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	// newEntireMails := make([]*dao.MailItem, 0)
	now := timer.StartTime(ctx.StartTime).UnixMilli()
	newMails := make([]*dao.MailItem, 0)
	for _, item := range mail.Mails {
		// 已过期
		if item.ExpiredAt <= now {
			continue
		}
		// 未激活
		if item.ActivatedAt > now {
			// newEntireMails = append(newEntireMails, item)
			continue
		}
		if !item.ForAll && role.CreateTime > item.CreatedAt {
			continue
		}
		var exist bool
		for _, mailItem := range list {
			id, ok := svc.isEntireServerMail(mailItem.Id)
			if ok && id == item.Id {
				exist = true
				break
			}
		}
		if !exist {
			tempMail := entire2user(item, role.Language)
			list = append(list, tempMail)
			newMails = append(newMails, tempMail)
		}
		// newEntireMails = append(newEntireMails, item)
	}
	// if len(newEntireMails) != len(mail.Mails) {
	// 	mail.Mails = newEntireMails
	// 	dao2.SaveEntireMail(ctx, mail)
	// }
	if len(newMails) > 0 {
		if err := dao2.NewMail(ctx.RoleId).BatchAdd(ctx, newMails); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (svc *Service) updateMailExpiredAt(ctx *ctx.Context, mail *dao.MailItem, hasAttachment bool) {
	// 全服邮件不能修改过期时间，否则可能会导致玩家再次获得该全服邮件
	if _, entire := svc.isEntireServerMail(mail.Id); entire {
		return
	}

	// 带附件的阅读之后1天过期，不带附件的阅读后3天过期
	now := timer.StartTime(ctx.StartTime)
	var expiredAt values.Integer
	if hasAttachment {
		expiredAt = now.AddDate(0, 0, 1).UnixMilli()
	} else {
		expiredAt = now.AddDate(0, 0, 3).UnixMilli()
	}
	if mail.ExpiredAt > expiredAt {
		mail.ExpiredAt = expiredAt
	}
}

func (svc *Service) isEntireServerMail(mailId string) (string, bool) {
	temp := strings.Split(mailId, "@")
	if len(temp) == 2 {
		return temp[1], true
	}
	return "", false
}

func (svc *Service) updateRewardMail(ctx *ctx.Context, v string) *errmsg.ErrMsg {
	// TODO 策划说先屏蔽掉该功能
	return nil
	if v == "" {
		return nil
	}
	reward, ok := rule.GetUpdateReward(ctx)
	if !ok {
		return nil
	}
	attachment := make([]*models.Item, 0)
	for id, count := range reward {
		attachment = append(attachment, &models.Item{
			ItemId: id,
			Count:  count,
		})
	}
	if err := ctx.DRLock(redisclient.GetLocker(), updateRewardMailLock+ctx.RoleId); err != nil {
		return err
	}
	ok, err := dao2.GetByVersion(ctx, v)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	id := values.Integer(enum.UpdateRewardId)
	var expiredAt values.Integer
	cfg, ok := rule.GetMailConfigTextId(ctx, id)
	if ok {
		expiredAt = timer.StartTime(ctx.StartTime).Add(time.Second * time.Duration(cfg.Overdue)).UnixMilli()
	}
	if err := svc.Add(ctx, ctx.RoleId, &models.Mail{
		Id:         xid.New().String(),
		Type:       models.MailType_MailTypeSystem,
		TextId:     id,
		ExpiredAt:  expiredAt,
		Attachment: attachment,
	}); err != nil {
		return err
	}
	dao2.SaveByVersion(ctx, v)
	return nil
}

func (svc *Service) isBagEnough(ctx *ctx.Context, items map[values.ItemId]values.Integer) *errmsg.ErrMsg {
	ok, err := svc.IsBagEnough(ctx, items)
	if err != nil {
		return err
	}
	if !ok {
		return errmsg.NewErrBagCapLimit()
	}
	return nil
}

func (svc *Service) afterAdd(ctx *ctx.Context) *errmsg.ErrMsg {
	unread, read, err := svc.mailList(ctx)
	if err != nil {
		return err
	}
	total := len(unread) + len(read)
	max := rule.GetMailMax(ctx)
	d := max - total
	if d >= 0 {
		return nil
	}
	del := make([]*dao.MailItem, 0)
	// 超过上限，需要删除部分邮件
	// 先删除已读的
	for _, mail := range read {
		del = append(del, &dao.MailItem{
			Id: mail.Id,
		})
		d++
		if d >= 0 {
			break
		}
	}
	// 需要删除未读邮件，按先获得先删除的逻辑删
	if d < 0 {
		temp := make([]*models.Mail, 0, len(unread))
		for _, mail := range unread {
			temp = append(temp, mail)
		}
		sort.Slice(temp, func(i, j int) bool {
			return temp[i].CreatedAt < temp[j].CreatedAt
		})
		for i := 0; i < len(temp); i++ {
			del = append(del, &dao.MailItem{
				Id: temp[i].Id,
			})
			d++
			if d >= 0 {
				break
			}
		}
	}
	if len(del) > 0 {
		dao2.NewMail(ctx.RoleId).DeleteMails(ctx, del)
	}
	return nil
}

func (svc *Service) CheatAdd(ctx *ctx.Context, req *servicepb.Mail_CheatAddMailRequest) (*servicepb.Mail_CheatAddMailResponse, *errmsg.ErrMsg) {
	newMail := &models.Mail{
		Id:         xid.New().String(),
		Type:       models.MailType_MailTypeSystem,
		Sender:     req.Sender,
		TextId:     req.TextId,
		Title:      req.Title,
		Content:    req.Content,
		Hi:         req.Hi,
		ExpiredAt:  time.Now().AddDate(0, 1, 0).Unix() * 1000,
		Args:       req.Args,
		Attachment: req.Attachment,
		CreatedAt:  time.Now().Unix() * 1000,
	}
	if req.TextId > 0 {
		cfg, ok := rule.GetMailConfigTextId(ctx, req.TextId)
		if ok {
			newMail.ExpiredAt = time.Now().Add(time.Duration(cfg.Overdue)*time.Second).Unix() * 1000
		}
	}
	if req.Expire > 0 {
		newMail.ExpiredAt = time.Now().Add(time.Duration(req.Expire)*time.Second).Unix() * 1000
	}
	if err := svc.Add(ctx, ctx.RoleId, newMail); err != nil {
		return nil, err
	}
	return &servicepb.Mail_CheatAddMailResponse{}, nil
}
