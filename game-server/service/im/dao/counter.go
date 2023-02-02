package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/timer"
	"coin-server/common/values"
)

func GetDrawCounter(ctx *ctx.Context, roleId values.RoleId) (*dao.DrawCounter, *errmsg.ErrMsg) {
	ret := &dao.DrawCounter{RoleId: roleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), ret)

	// reset
	if ret.DrawTime < timer.BeginOfDay(timer.Now()).UnixMilli() {
		ret.Count = 0
	}

	return ret, err
}

func SaveDrawCounter(ctx *ctx.Context, dc *dao.DrawCounter) {
	dc.DrawTime = timer.Now().UnixMilli()
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), dc)
}
