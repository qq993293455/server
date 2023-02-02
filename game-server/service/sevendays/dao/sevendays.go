package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetSevenDaysInfo(ctx *ctx.Context, roleId values.RoleId) *dao.SevenDayData {
	ret := &dao.SevenDayData{
		RoleId: roleId,
	}
	ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	return ret
}

func SaveSevenDaysInfo(ctx *ctx.Context, data *dao.SevenDayData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}
