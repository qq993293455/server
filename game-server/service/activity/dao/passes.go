// 通行证

package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

const GlobalPassesResetKey = "global_passes_reset"

func GetPassesReset(ctx *ctx.Context) (*daopb.PassesReset, *errmsg.ErrMsg) {
	data := &daopb.PassesReset{Key: GlobalPassesResetKey}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return data, nil
}

func SavePassesReset(ctx *ctx.Context, data *daopb.PassesReset) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}

func GetPasses(ctx *ctx.Context) (*daopb.Passes, *errmsg.ErrMsg) {
	data := &daopb.Passes{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return data, nil
}

func SavePasses(ctx *ctx.Context, data *daopb.Passes) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}
