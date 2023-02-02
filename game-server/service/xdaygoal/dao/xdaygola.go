package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetXDayGoalInfo(ctx *ctx.Context, roleId values.RoleId) *dao.XDayGoalData {
	ret := &dao.XDayGoalData{
		RoleId: roleId,
	}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		return nil
	}
	return ret
}

func SaveSevenDaysInfo(ctx *ctx.Context, info *dao.XDayGoalData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), info)
}
