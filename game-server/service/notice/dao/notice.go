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

func getKey(id values.RoleId) string {
	return utils.GenDefaultRedisKey(values.Notice, values.Hash, id)
}

func Get(ctx *ctx.Context) ([]*dao.NoticeRead, *errmsg.ErrMsg) {
	list := make([]*dao.NoticeRead, 0)
	if err := ctx.NewOrm().HGetAll(redisclient.GetDefaultRedis(), getKey(ctx.RoleId), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func GetOne(ctx *ctx.Context, id string) (bool, *errmsg.ErrMsg) {
	out := &dao.NoticeRead{NoticeId: id}
	ok, err := ctx.NewOrm().HGetPB(redisclient.GetDefaultRedis(), getKey(ctx.RoleId), out)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func Save(ctx *ctx.Context, data *dao.NoticeRead) {
	ctx.NewOrm().HSetPB(redisclient.GetDefaultRedis(), getKey(ctx.RoleId), data)
}

func Del(ctx *ctx.Context, list ...*dao.NoticeRead) {
	if len(list) <= 0 {
		return
	}
	ins := make([]orm.RedisInterface, 0, len(list))
	for _, item := range list {
		ins = append(ins, item)
	}
	ctx.NewOrm().HDelPB(redisclient.GetDefaultRedis(), getKey(ctx.RoleId), ins)
}
