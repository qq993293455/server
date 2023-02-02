package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	commonmail "coin-server/common/mail"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/values"
)

type Mail struct {
	roleId values.RoleId
}

func NewMail(roleId values.RoleId) *Mail {
	return &Mail{roleId: roleId}
}

func (m *Mail) GetAll(ctx *ctx.Context) ([]*dao.MailItem, *errmsg.ErrMsg) {
	mails := make([]*dao.MailItem, 0)
	if err := ctx.NewOrm().HGetAll(commonmail.GetMailRedis(), getKey(m.roleId), &mails); err != nil {
		return nil, err
	}
	return mails, nil
}

func (m *Mail) Add(ctx *ctx.Context, mail *dao.MailItem) *errmsg.ErrMsg {
	ctx.NewOrm().HSetPB(commonmail.GetMailRedis(), getKey(m.roleId), mail)
	return nil
}

func (m *Mail) BatchAdd(ctx *ctx.Context, mailList []*dao.MailItem) *errmsg.ErrMsg {
	val := make([]orm.RedisInterface, 0, len(mailList))
	for _, item := range mailList {
		val = append(val, item)
	}
	if len(val) > 0 {
		ctx.NewOrm().HMSetPB(commonmail.GetMailRedis(), getKey(m.roleId), val)
	}
	return nil
}

func (m *Mail) UpdateMail(ctx *ctx.Context, mail *dao.MailItem) *errmsg.ErrMsg {
	ctx.NewOrm().HSetPB(commonmail.GetMailRedis(), getKey(m.roleId), mail)

	return nil
}

func (m *Mail) UpdateMails(ctx *ctx.Context, mails []*dao.MailItem) *errmsg.ErrMsg {
	val := make([]orm.RedisInterface, 0, len(mails))
	for _, item := range mails {
		val = append(val, item)
	}
	if len(val) > 0 {
		ctx.NewOrm().HMSetPB(commonmail.GetMailRedis(), getKey(m.roleId), val)
	}

	return nil
}

func (m *Mail) DeleteMail(ctx *ctx.Context, mail *dao.MailItem) *errmsg.ErrMsg {
	ctx.NewOrm().HDel(commonmail.GetMailRedis(), getKey(m.roleId), mail.Id)

	return nil
}

func (m *Mail) DeleteMails(ctx *ctx.Context, mails []*dao.MailItem) {
	val := make([]string, 0, len(mails))
	for _, mail := range mails {
		val = append(val, mail.Id)
	}
	if len(val) > 0 {
		ctx.NewOrm().HDel(commonmail.GetMailRedis(), getKey(m.roleId), val...)
	}
}

func getKey(roleId values.RoleId) string {
	return commonmail.GetMailKey(roleId)
}
