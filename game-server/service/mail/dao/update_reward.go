package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func getUpdateRewardKey(v string) string {
	return utils.GenDefaultRedisKey(values.UpdateReward, values.Hash, v)
}

func GetByVersion(ctx *ctx.Context, v string) (bool, *errmsg.ErrMsg) {
	out := &dao.UpdateReward{Version: v}
	return ctx.NewOrm().HGetPB(redisclient.GetDefaultRedis(), getUpdateRewardKey(v), out)
}

func SaveByVersion(ctx *ctx.Context, v string) {
	ctx.NewOrm().HSetPB(redisclient.GetDefaultRedis(), getUpdateRewardKey(v), &dao.UpdateReward{Version: v})
}
