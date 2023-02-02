package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetDiscordInfo(ctx *ctx.Context, roleId values.RoleId) *dao.DiscordData {
	ret := &dao.DiscordData{
		RoleId: roleId,
	}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		return nil
	}
	return ret
}

func SaveDiscordInfo(ctx *ctx.Context, info *dao.DiscordData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), info)
}
