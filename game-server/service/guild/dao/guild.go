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

type Guild struct {
	id values.GuildId
}

func NewGuild(id values.GuildId) *Guild {
	return &Guild{id: id}
}

func (g *Guild) Get(ctx *ctx.Context) (*dao.Guild, *errmsg.ErrMsg) {
	data := &dao.Guild{Id: g.id}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetGuildRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return data, nil
}

func (g *Guild) GetMulti(ctx *ctx.Context, ids []values.GuildId) (map[values.GuildId]*dao.Guild, *errmsg.ErrMsg) {
	data := make([]orm.RedisInterface, 0, len(ids))
	for _, id := range ids {
		data = append(data, &dao.Guild{Id: id})
	}
	notFound, err := orm.GetOrm(ctx).MGetPB(redisclient.GetGuildRedis(), data...)
	if err != nil {
		return nil, err
	}
	if len(notFound) > 0 {
		notFoundId := make(map[int]struct{})
		for _, id := range notFound {
			notFoundId[id] = struct{}{}
		}
		res := make(map[values.GuildId]*dao.Guild, 0)
		for idx, v := range data {
			if _, exist := notFoundId[idx]; exist {
				continue
			}
			temp := v.(*dao.Guild)
			res[temp.Id] = temp
		}
		return res, nil
	}
	res := make(map[values.GuildId]*dao.Guild, len(data))
	for _, v := range data {
		temp := v.(*dao.Guild)
		res[temp.Id] = temp
	}
	return res, nil
}

func (g *Guild) Save(ctx *ctx.Context, guild *dao.Guild, dissolve ...bool) *errmsg.ErrMsg {
	ctx.NewOrm().SetPB(redisclient.GetGuildRedis(), guild)

	return nil
}

func getGuildKey(id values.GuildId) string {
	return utils.GenDefaultRedisKey(values.Guild, values.Hash, id)
}
