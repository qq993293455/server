package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetBiography(ctx *ctx.Context) (*dao.HeroBiography, *errmsg.ErrMsg) {
	b := &dao.HeroBiography{RoleId: ctx.RoleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func SaveBiography(ctx *ctx.Context, biography *dao.HeroBiography) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), biography)
}
