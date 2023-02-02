package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetDailySale(ctx *ctx.Context) (*dao.DailySale, *errmsg.ErrMsg) {
	ds := &dao.DailySale{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), ds)
	if err != nil {
		return nil, err
	}
	if !ok {
		ds = &dao.DailySale{
			RoleId:   ctx.RoleId,
			BuyTimes: map[int64]int64{},
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), ds)
	}
	if ds.BuyTimes == nil {
		ds.BuyTimes = map[int64]int64{}
	}
	if ds.CanBuyIds == nil {
		ds.CanBuyIds = make([]int64, 6)
	}
	return ds, nil
}

func SaveDailySale(ctx *ctx.Context, ds *dao.DailySale) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), ds)
}
