package db

import (
	"coin-server/common/ctx"
	"coin-server/common/redisclient"
)

func RegisterCountIncr(ctx *ctx.Context) {
	ctx.NewOrm().Incr(redisclient.GetDefaultRedis(), redisclient.RegisterCount)
}
