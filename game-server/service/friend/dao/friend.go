package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func Get(ctx *ctx.Context, roleId values.RoleId) (*dao.Friend, *errmsg.ErrMsg) {
	res := &dao.Friend{RoleId: roleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		data := newDao(roleId)
		ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
		return data, nil
	}
	return handleNilMap(res), nil
}

func Gets(ctx *ctx.Context, roleIds []values.RoleId) ([]*dao.Friend, *errmsg.ErrMsg) {
	data := make([]orm.RedisInterface, len(roleIds))
	for idx, roleId := range roleIds {
		data[idx] = &dao.Friend{
			RoleId: roleId,
		}
	}
	ids, err := ctx.NewOrm().MGetPB(redisclient.GetUserRedis(), data...)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if len(ids) > 0 {
		notFoundId := make(map[int]bool)
		for _, id := range ids {
			notFoundId[id] = false
		}
		res := make([]*dao.Friend, 0)
		for idx, v := range data {
			if _, exist := notFoundId[idx]; !exist {
				res = append(res, v.(*dao.Friend))
			}
		}
		return res, nil
	}
	res := make([]*dao.Friend, len(data))
	for idx, v := range data {
		res[idx] = handleNilMap(v.(*dao.Friend))
	}
	return res, nil
}

func newDao(roleId values.RoleId) *dao.Friend {
	return &dao.Friend{
		RoleId:    roleId,
		Friends:   map[string]*dao.FriendValue{},
		Requests:  map[string]*dao.RequestValue{},
		Blacklist: map[string]*dao.BlackListValue{},
	}
}

func handleNilMap(data *dao.Friend) *dao.Friend {
	if data.Friends == nil {
		data.Friends = map[string]*dao.FriendValue{}
	}
	if data.Requests == nil {
		data.Requests = map[string]*dao.RequestValue{}
	}
	if data.Blacklist == nil {
		data.Blacklist = map[string]*dao.BlackListValue{}
	}
	return data
}

func Save(ctx *ctx.Context, data *dao.Friend) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}

func Saves(ctx *ctx.Context, list []*dao.Friend) *errmsg.ErrMsg {
	msetList := make([]orm.RedisInterface, len(list))
	for idx, val := range list {
		msetList[idx] = val
	}
	ctx.NewOrm().MSetPB(redisclient.GetUserRedis(), msetList)
	return nil
}
