package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetGacha(ctx *ctx.Context, roleId values.RoleId) ([]*dao.Gacha, *errmsg.ErrMsg) {
	gacha := make([]*dao.Gacha, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getGachaKey(roleId), &gacha)
	if err != nil {
		return nil, err
	}
	return gacha, nil
}

func GetGachaById(ctx *ctx.Context, roleId values.RoleId, gachaId values.GachaId) (*dao.Gacha, *errmsg.ErrMsg) {
	gacha := &dao.Gacha{GachaId: gachaId}
	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getGachaKey(roleId), gacha)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return gacha, nil
}

func SaveGacha(ctx *ctx.Context, roleId values.RoleId, gacha *dao.Gacha) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getGachaKey(roleId), gacha)
	return
}

func SaveMultiGacha(ctx *ctx.Context, roleId values.RoleId, gacha []*dao.Gacha) {
	if len(gacha) == 0 {
		return
	}
	add := make([]orm.RedisInterface, len(gacha))
	for idx := range add {
		add[idx] = gacha[idx]
	}
	ctx.NewOrm().HMSetPB(redisclient.GetUserRedis(), getGachaKey(roleId), add)
	return
}

func getGachaKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.Gacha, values.Hash, roleId)
}

//func GetGachaUnlock(ctx *ctx.Context, roleId values.RoleId) ([]*dao.GachaUnlock, *errmsg.ErrMsg) {
//	gacha := make([]*dao.GachaUnlock, 0)
//	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getGachaUnlockKey(roleId), &gacha)
//	if err != nil {
//		return nil, err
//	}
//	return gacha, nil
//}
//
//func GetGachaUnlockById(ctx *ctx.Context, roleId values.RoleId, gachaId values.GachaId) (*dao.GachaUnlock, *errmsg.ErrMsg) {
//	gacha := &dao.GachaUnlock{GachaId: gachaId}
//	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getGachaUnlockKey(roleId), gacha)
//	if err != nil {
//		return nil, err
//	}
//	if !has {
//		return nil, nil
//	}
//	return gacha, nil
//}
//
//func SaveGachaUnlock(ctx *ctx.Context, roleId values.RoleId, gacha *dao.GachaUnlock) {
//	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getGachaUnlockKey(roleId), gacha)
//	return
//}
//
//func getGachaUnlockKey(roleId values.RoleId) string {
//	return utils.GenDefaultRedisKey(values.GachaUnlock, values.Hash, roleId)
//}
