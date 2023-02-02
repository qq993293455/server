package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func Get(ctx *ctx.Context, roleId string) (*daopb.Journey, *errmsg.ErrMsg) {
	j := &daopb.Journey{RoleId: roleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), j)
	if err != nil {
		return nil, err
	}
	if !ok {
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), j)
	}
	return j, nil
}

func Save(ctx *ctx.Context, v *daopb.Journey) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), v)
}
