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

type GuildApply struct {
	id values.GuildId
}

func NewGuildApply(id values.GuildId) *GuildApply {
	return &GuildApply{id: id}
}

func (ga *GuildApply) Get(ctx *ctx.Context) ([]*dao.GuildApply, *errmsg.ErrMsg) {
	list := make([]*dao.GuildApply, 0)
	if err := ctx.NewOrm().HGetAll(redisclient.GetGuildRedis(), ga.getGuildApplyKey(), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (ga *GuildApply) GetOne(ctx *ctx.Context, id values.RoleId) (*dao.GuildApply, bool, *errmsg.ErrMsg) {
	list, err := ga.Get(ctx)
	if err != nil {
		return nil, false, err
	}
	for _, member := range list {
		if member.RoleId == id {
			return member, true, nil
		}
	}
	return nil, false, nil
}

func (ga *GuildApply) Save(ctx *ctx.Context, update []*dao.GuildApply) error {
	val := make([]orm.RedisInterface, 0, len(update))
	for _, item := range update {
		val = append(val, item)
	}
	if len(val) > 0 {
		ctx.NewOrm().HMSetPB(redisclient.GetGuildRedis(), ga.getGuildApplyKey(), val)
	}
	return nil
}

func (ga *GuildApply) SaveOne(ctx *ctx.Context, apply *dao.GuildApply) *errmsg.ErrMsg {
	ctx.NewOrm().HSetPB(redisclient.GetGuildRedis(), ga.getGuildApplyKey(), apply)
	return nil
}

func (ga *GuildApply) Delete(ctx *ctx.Context, del []*dao.GuildApply) *errmsg.ErrMsg {
	val := make([]orm.RedisInterface, 0, len(del))
	for _, apply := range del {
		val = append(val, apply)
	}
	if len(val) > 0 {
		ctx.NewOrm().HDelPB(redisclient.GetGuildRedis(), ga.getGuildApplyKey(), val)
	}
	return nil
}

func (ga *GuildApply) DeleteKey(ctx *ctx.Context) {
	ctx.NewOrm().Del(redisclient.GetGuildRedis(), ga.getGuildApplyKey())
}

func (ga *GuildApply) DeleteAll(ctx *ctx.Context) *errmsg.ErrMsg {
	ctx.NewOrm().Del(redisclient.GetGuildRedis(), ga.getGuildApplyKey())
	return nil
}

func (ga *GuildApply) getGuildApplyKey() string {
	return utils.GenDefaultRedisKey(values.GuildApply, values.Hash, ga.id)
}
