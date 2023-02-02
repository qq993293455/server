package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetPlaneDungeon(ctx *ctx.Context) (*daopb.PlaneDungeon, *errmsg.ErrMsg) {
	data := &daopb.PlaneDungeon{RoleId: ctx.RoleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}

	if data.Finished == nil {
		data.Finished = map[int64]bool{}
	}
	return data, nil
}

func SavePlaneDungeon(ctx *ctx.Context, data *daopb.PlaneDungeon) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}
