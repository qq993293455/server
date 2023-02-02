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

type GuildMember struct {
	id values.GuildId
}

func NewGuildMember(id values.GuildId) *GuildMember {
	return &GuildMember{id: id}
}

func (gm *GuildMember) Get(ctx *ctx.Context) ([]*dao.GuildMember, *errmsg.ErrMsg) {
	list := make([]*dao.GuildMember, 0)
	if err := ctx.NewOrm().HGetAll(redisclient.GetGuildRedis(), getGuildMemberKey(gm.id), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (gm *GuildMember) GetOne(ctx *ctx.Context, id values.RoleId) (*dao.GuildMember, bool, *errmsg.ErrMsg) {
	list, err := gm.Get(ctx)
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

func (gm *GuildMember) Save(ctx *ctx.Context, update []*dao.GuildMember) *errmsg.ErrMsg {
	val := make([]orm.RedisInterface, 0, len(update))
	for _, item := range update {
		val = append(val, item)
	}
	if len(val) > 0 {
		ctx.NewOrm().HMSetPB(redisclient.GetGuildRedis(), getGuildMemberKey(gm.id), val)
	}
	return nil
}

func (gm *GuildMember) SaveOne(ctx *ctx.Context, member *dao.GuildMember) *errmsg.ErrMsg {
	ctx.NewOrm().HMSetPB(redisclient.GetGuildRedis(), getGuildMemberKey(gm.id), []orm.RedisInterface{member})
	return nil
}

func (gm *GuildMember) Delete(ctx *ctx.Context, member *dao.GuildMember, dissolve ...bool) *errmsg.ErrMsg {
	ctx.NewOrm().HDelPB(redisclient.GetGuildRedis(), getGuildMemberKey(gm.id), []orm.RedisInterface{member})

	return nil
}

func (gm *GuildMember) GetMulti(ctx *ctx.Context, ids []values.GuildId) (map[values.GuildId][]*dao.GuildMember, *errmsg.ErrMsg) {
	data := make(map[values.GuildId][]*dao.GuildMember)
	for _, id := range ids {
		list := make([]*dao.GuildMember, 0)
		if err := ctx.NewOrm().HGetAll(redisclient.GetGuildRedis(), getGuildMemberKey(id), &list); err != nil {
			return nil, err
		}
		data[id] = list
	}
	return data, nil
}

func getGuildMemberKey(id values.GuildId) string {
	return utils.GenDefaultRedisKey(values.GuildMember, values.Hash, id)
}
