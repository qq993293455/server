package redisclient

import (
	"coin-server/common/ctx"
	"coin-server/common/orm"
)

func NewOrm(ctx *ctx.Context) *orm.Orm {
	return ctx.NewOrm()
}
