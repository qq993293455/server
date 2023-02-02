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

type GuildUser struct {
	roleId values.RoleId
}

func NewGuildUser(roleId values.RoleId) *GuildUser {
	return &GuildUser{roleId: roleId}
}

func (gu *GuildUser) Get(ctx *ctx.Context) (*dao.GuildUser, *errmsg.ErrMsg) {
	user := &dao.GuildUser{RoleId: gu.roleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetGuildRedis(), user)
	if err != nil {
		return nil, err
	}
	if !ok {
		user = &dao.GuildUser{
			RoleId:           gu.roleId,
			GuildId:          "",
			Cd:               0,
			Build:            &models.GuildBuild{},
			FirstJoinReward:  false,
			ActiveValue:      map[int64]int64{},
			TotalActiveValue: 0,
		}
	}
	if user.ActiveValue == nil {
		user.ActiveValue = map[int64]int64{}
	}
	return user, nil
}

func (gu *GuildUser) GetMulti(ctx *ctx.Context, roleIds []values.RoleId) (map[values.RoleId]*dao.GuildUser, *errmsg.ErrMsg) {
	data := make([]orm.RedisInterface, 0, len(roleIds))
	for _, roleId := range roleIds {
		data = append(data, &dao.GuildUser{RoleId: roleId})
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
		res := make(map[values.RoleId]*dao.GuildUser, 0)
		for idx, v := range data {
			if _, exist := notFoundId[idx]; !exist {
				temp := v.(*dao.GuildUser)
				res[temp.RoleId] = temp
			} else {
				res[roleIds[idx]] = &dao.GuildUser{
					RoleId: roleIds[idx],
				}
			}
		}
		return res, nil
	}
	res := make(map[values.RoleId]*dao.GuildUser, len(data))
	for _, v := range data {
		temp := v.(*dao.GuildUser)
		res[temp.RoleId] = temp
	}
	return res, nil
}

func (gu *GuildUser) Save(ctx *ctx.Context, user *dao.GuildUser) *errmsg.ErrMsg {
	ctx.NewOrm().SetPB(redisclient.GetGuildRedis(), user)
	return nil
}

func (gu *GuildUser) BatchSave(ctx *ctx.Context, list []*dao.GuildUser) *errmsg.ErrMsg {
	val := make([]orm.RedisInterface, 0, len(list))
	for _, item := range list {
		val = append(val, item)
	}
	if len(val) > 0 {
		ctx.NewOrm().MSetPB(redisclient.GetGuildRedis(), val)
	}
	return nil
}

func getGuildUserKey(id values.GuildId) string {
	return utils.GenDefaultRedisKey(values.GuildUser, values.Hash, id)
}
