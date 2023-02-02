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

type GuildInvite struct {
	roleId values.RoleId
}

func NewGuildInvite(roleId values.RoleId) *GuildInvite {
	return &GuildInvite{roleId: roleId}
}

func (gi *GuildInvite) Get(ctx *ctx.Context) ([]*dao.GuildInvite, *errmsg.ErrMsg) {
	list := make([]*dao.GuildInvite, 0)
	if err := ctx.NewOrm().HGetAll(redisclient.GetGuildRedis(), gi.getGuildInviteKey(), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (gi *GuildInvite) Save(ctx *ctx.Context, update []*dao.GuildInvite) *errmsg.ErrMsg {
	val := make([]orm.RedisInterface, 0, len(update))
	for _, item := range update {
		val = append(val, item)
	}
	if len(val) > 0 {
		ctx.NewOrm().HMSetPB(redisclient.GetGuildRedis(), gi.getGuildInviteKey(), val)
	}
	return nil
}

func (gi *GuildInvite) SaveOne(ctx *ctx.Context, update *dao.GuildInvite) *errmsg.ErrMsg {
	ctx.NewOrm().HSetPB(redisclient.GetGuildRedis(), gi.getGuildInviteKey(), update)
	return nil
}

func (gi *GuildInvite) DeleteOne(ctx *ctx.Context, del *dao.GuildInvite) *errmsg.ErrMsg {
	ctx.NewOrm().HDelPB(redisclient.GetGuildRedis(), gi.getGuildInviteKey(), []orm.RedisInterface{del})
	return nil
}

func (gi *GuildInvite) Delete(ctx *ctx.Context, del []*dao.GuildInvite) *errmsg.ErrMsg {
	val := make([]orm.RedisInterface, 0, len(del))
	for _, apply := range del {
		val = append(val, apply)
	}
	if len(val) > 0 {
		ctx.NewOrm().HDelPB(redisclient.GetGuildRedis(), gi.getGuildInviteKey(), val)
	}
	return nil
}

func (gi *GuildInvite) DeleteAll(ctx *ctx.Context) *errmsg.ErrMsg {
	ctx.NewOrm().Del(redisclient.GetGuildRedis(), gi.getGuildInviteKey())
	return nil
}

func (gi *GuildInvite) getGuildInviteKey() string {
	return utils.GenDefaultRedisKey(values.GuildInvite, values.Hash, gi.roleId)
}
