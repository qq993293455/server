package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetRelicsSuit(ctx *ctx.Context, roleId values.RoleId) ([]*dao.RelicsSuit, *errmsg.ErrMsg) {
	suit := make([]*dao.RelicsSuit, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getRelicsSuitKey(roleId), &suit)
	if err != nil {
		return nil, err
	}
	return suit, nil
}

func GetRelicsSuitById(ctx *ctx.Context, roleId values.RoleId, suitId values.RelicsSuitId) (*dao.RelicsSuit, *errmsg.ErrMsg) {
	suit := &dao.RelicsSuit{SuitId: suitId}
	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getRelicsSuitKey(roleId), suit)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, err
	}
	return suit, err
}

func SaveRelicsSuit(ctx *ctx.Context, roleId values.RoleId, suit *dao.RelicsSuit) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getRelicsSuitKey(roleId), suit)
	return
}

func getRelicsSuitKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.RelicsSuit, values.Hash, roleId)
}
