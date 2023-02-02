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

func GetRelics(ctx *ctx.Context, roleId values.RoleId) ([]*dao.Relics, *errmsg.ErrMsg) {
	relics := make([]*dao.Relics, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getRelicsKey(roleId), &relics)
	if err != nil {
		return nil, err
	}
	return relics, nil
}

func GetMultiRelics(ctx *ctx.Context, roleId values.RoleId, relics []*dao.Relics) ([]int, *errmsg.ErrMsg) {
	ri := make([]orm.RedisInterface, len(relics))
	for idx := range ri {
		ri[idx] = relics[idx]
	}
	nfi, err := ctx.NewOrm().HMGetPB(redisclient.GetUserRedis(), getRelicsKey(roleId), ri)
	if err != nil {
		return nfi, err
	}
	return nfi, nil
}

func GetRelicsById(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) (*dao.Relics, *errmsg.ErrMsg) {
	relics := &dao.Relics{RelicsId: itemId}
	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getRelicsKey(roleId), relics)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, err
	}
	return relics, err
}

func SaveRelics(ctx *ctx.Context, roleId values.RoleId, relics *dao.Relics) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getRelicsKey(roleId), relics)
	return
}

func DeleteRelic(ctx *ctx.Context, roleId values.RoleId, r *dao.Relics) {
	relics := []orm.RedisInterface{r}
	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), getRelicsKey(roleId), relics)
}

func SaveMultiRelics(ctx *ctx.Context, roleId values.RoleId, relics []*dao.Relics) {
	if len(relics) == 0 {
		return
	}
	add := make([]orm.RedisInterface, len(relics))
	for idx := range add {
		add[idx] = relics[idx]
	}
	ctx.NewOrm().HMSetPB(redisclient.GetUserRedis(), getRelicsKey(roleId), add)
	return
}

func getRelicsKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.Relics, values.Hash, roleId)
}

func getRelicsSuitKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.RelicsSuit, values.Hash, roleId)
}
