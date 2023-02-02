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

type GuildUserApply struct {
	roleId values.RoleId
}

func NewGuildUserApply(roleId values.RoleId) *GuildUserApply {
	return &GuildUserApply{roleId: roleId}
}

func (gua *GuildUserApply) Get(ctx *ctx.Context) ([]*dao.GuildUserApply, *errmsg.ErrMsg) {
	list := make([]*dao.GuildUserApply, 0)
	if err := ctx.NewOrm().HGetAll(redisclient.GetGuildRedis(), gua.getGuildUserApplyKey(), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (gua *GuildUserApply) Save(ctx *ctx.Context, update []*dao.GuildUserApply) *errmsg.ErrMsg {
	val := make([]orm.RedisInterface, 0, len(update))
	for _, item := range update {
		val = append(val, item)
	}
	if len(val) > 0 {
		ctx.NewOrm().HMSetPB(redisclient.GetGuildRedis(), gua.getGuildUserApplyKey(), val)
	}
	return nil
}

func (gua *GuildUserApply) SaveOne(ctx *ctx.Context, update *dao.GuildUserApply) *errmsg.ErrMsg {
	ctx.NewOrm().HSetPB(redisclient.GetGuildRedis(), gua.getGuildUserApplyKey(), update)
	return nil
}

func (gua *GuildUserApply) DeleteOne(ctx *ctx.Context, del *dao.GuildUserApply) *errmsg.ErrMsg {
	ctx.NewOrm().HDelPB(redisclient.GetGuildRedis(), gua.getGuildUserApplyKey(), []orm.RedisInterface{del})
	return nil
}

func (gua *GuildUserApply) Delete(ctx *ctx.Context, del []*dao.GuildUserApply) *errmsg.ErrMsg {
	val := make([]orm.RedisInterface, 0, len(del))
	for _, apply := range del {
		val = append(val, apply)
	}
	if len(val) > 0 {
		ctx.NewOrm().HDelPB(redisclient.GetGuildRedis(), gua.getGuildUserApplyKey(), val)
	}
	return nil
}

func (gua *GuildUserApply) DeleteAll(ctx *ctx.Context) *errmsg.ErrMsg {
	ctx.NewOrm().Del(redisclient.GetGuildRedis(), gua.getGuildUserApplyKey())
	return nil
}

func (gua *GuildUserApply) getGuildUserApplyKey() string {
	return utils.GenDefaultRedisKey(values.GuildUserApply, values.Hash, gua.roleId)
}
