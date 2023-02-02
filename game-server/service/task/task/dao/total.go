package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetCondByType(ctx *ctx.Context, roleId values.RoleId, typ models.TaskType) (*dao.CondCounter, *errmsg.ErrMsg) {
	cond := &dao.CondCounter{Typ: int64(typ)}
	_, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getTaskCondKey(roleId), cond)
	if err != nil {
		return nil, err
	}
	if cond.Count == nil {
		cond.Count = map[int64]int64{}
	}
	return cond, err
}

func GetCond(ctx *ctx.Context, roleId values.RoleId) (map[string]*dao.CondCounter, *errmsg.ErrMsg) {
	cond := map[string]*dao.CondCounter{}
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getTaskCondKey(roleId), &cond)
	if err != nil {
		return nil, err
	}
	return cond, nil
}

func SaveCond(ctx *ctx.Context, roleId values.RoleId, cond *dao.CondCounter) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getTaskCondKey(roleId), cond)
	return
}

func SaveManyCond(ctx *ctx.Context, roleId values.RoleId, cond []*dao.CondCounter) {
	if len(cond) == 0 {
		return
	}
	add := make([]orm.RedisInterface, len(cond))
	for idx := range add {
		add[idx] = cond[idx]
	}
	ctx.NewOrm().HMSetPB(redisclient.GetUserRedis(), getTaskCondKey(roleId), add)
	return
}

func getTaskCondKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.Cond, values.Hash, roleId)
}
