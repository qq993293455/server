package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetZeroBuy(ctx *ctx.Context) (*dao.ZeroBuy, *errmsg.ErrMsg) {
	zb := &dao.ZeroBuy{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), zb)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, err
	}
	return zb, nil
}

func SaveZeroBuy(ctx *ctx.Context, ds *dao.ZeroBuy) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), ds)
}
