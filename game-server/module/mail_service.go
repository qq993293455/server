package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type MailService interface {
	Add(ctx *ctx.Context, roleId values.RoleId, mail *models.Mail, gm ...bool) *errmsg.ErrMsg
	BatchAdd(ctx *ctx.Context, roleId values.RoleId, mailList []*models.Mail) *errmsg.ErrMsg
}
