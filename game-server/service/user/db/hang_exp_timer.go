package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetHangExpTimer(ctx *ctx.Context) (*dao.HangExpTimer, *errmsg.ErrMsg) {
	out := &dao.HangExpTimer{
		RoleId: ctx.RoleId,
		Timing: false,
		Time:   0,
		Pushed: false,
	}
	_, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func SaveHangExpTimer(ctx *ctx.Context, data *dao.HangExpTimer) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}
