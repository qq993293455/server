package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetGift(ctx *ctx.Context, id values.GiftId) (*dao.Gift, *errmsg.ErrMsg) {
	ret := &dao.Gift{GiftId: id}
	_, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), ret)
	return ret, err
}

func SaveGift(ctx *ctx.Context, gift *dao.Gift) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), gift)
}
