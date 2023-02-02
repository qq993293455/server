package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetWeeklySale(ctx *ctx.Context) (*dao.WeeklySale, *errmsg.ErrMsg) {
	ds := &dao.WeeklySale{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), ds)
	if err != nil {
		return nil, err
	}
	if !ok {
		ds = &dao.WeeklySale{
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

func SaveWeeklySale(ctx *ctx.Context, ds *dao.WeeklySale) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), ds)
}
