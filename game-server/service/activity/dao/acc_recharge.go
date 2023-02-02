package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetAccRecharge(ctx *ctx.Context) (*dao.AccRecharge, *errmsg.ErrMsg) {
	ds := &dao.AccRecharge{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), ds)
	if err != nil {
		return nil, err
	}
	if !ok {
		ds = &dao.AccRecharge{
			RoleId: ctx.RoleId,
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), ds)
	}
	if ds.DrawList == nil {
		ds.DrawList = map[int64]bool{}
	}
	return ds, nil
}

func SaveAccRecharge(ctx *ctx.Context, ds *dao.AccRecharge) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), ds)
}
