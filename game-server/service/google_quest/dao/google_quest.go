package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetGoogleQuestInfo(ctx *ctx.Context, roleId values.RoleId) *dao.GoogleQuestData {
	ret := &dao.GoogleQuestData{
		RoleId: roleId,
	}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		return nil
	}
	return ret
}

func SaveGoogleQuestInfo(ctx *ctx.Context, info *dao.GoogleQuestData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), info)
}
