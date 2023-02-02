package mail

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	dao2 "coin-server/game-server/service/mail/dao"

	"go.uber.org/zap"
)

func GMAuth(next handler.HandleFunc) handler.HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		if ctx.ServerType != models.ServerType_GMServer {
			return errmsg.NewProtocolErrorInfo("invalid server type request")
		}
		return next(ctx)
	}
}

func (svc *Service) GMAddMail(ctx *ctx.Context, req *servicepb.GM_SendMail) {
	svc.formatMail(ctx, req.Mail)
	err := svc.MailService.Add(ctx, ctx.RoleId, req.Mail, true)
	if err != nil {
		svc.log.Error("send gm mail err", zap.Error(err), zap.String("role_id", ctx.RoleId), zap.Any("mail", req.Mail))
	}
}

func (svc *Service) GMDeleteMail(ctx *ctx.Context, req *servicepb.GM_DeleteMail) {
	mailDao := dao2.NewMail(ctx.RoleId)
	list, err := mailDao.GetAll(ctx)
	if err != nil {
		svc.log.Error("gm delete mail err", zap.String("role_id", ctx.RoleId), zap.Any("mail_id", req.MailId))
		return
	}
	deleteList := make([]*dao.MailItem, 0)
	for _, id := range req.MailId {
		for _, item := range list {
			if id == item.Id {
				deleteList = append(deleteList, item)
				break
			}
		}
	}
	if len(deleteList) <= 0 {
		return
	}
	mailDao.DeleteMails(ctx, deleteList)
}
