package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	commonmail "coin-server/common/mail"
	"coin-server/common/proto/dao"
)

func GetEntireMail(ctx *ctx.Context) (*dao.EntireMail, *errmsg.ErrMsg) {
	mail := &dao.EntireMail{Key: commonmail.EntireMailKey}
	_, err := ctx.NewOrm().GetPB(commonmail.GetMailRedis(), mail)
	if err != nil {
		return nil, err
	}
	return mail, nil
}

func SaveEntireMail(ctx *ctx.Context, mail *dao.EntireMail) {
	ctx.NewOrm().SetPB(commonmail.GetMailRedis(), mail)
}
