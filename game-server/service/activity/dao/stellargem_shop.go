package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetStellargemShop(ctx *ctx.Context) (*dao.StellargemShop, *errmsg.ErrMsg) {
	ds := &dao.StellargemShop{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), ds)
	if err != nil {
		return nil, err
	}
	if !ok {
		ds = &dao.StellargemShop{
			RoleId: ctx.RoleId,
			BuyCnt: map[int64]int64{},
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), ds)
	}
	if ds.BuyCnt == nil {
		ds.BuyCnt = map[int64]int64{}
	}
	return ds, nil
}

func SaveStellargemShop(ctx *ctx.Context, ds *dao.StellargemShop) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), ds)
}
