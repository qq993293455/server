package model

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

const key = "WHITE_LIST"

type WhiteList struct{}

func NewWhiteList() *WhiteList {
	return &WhiteList{}
}

func (w *WhiteList) Save(data *dao.WhiteList) *errmsg.ErrMsg {
	db := orm.GetOrm(ctx.GetContext())
	db.HMSetPB(redisclient.GetGuildRedis(), key, []orm.RedisInterface{data})
	return db.Do()
}

func (w *WhiteList) Del(device string) *errmsg.ErrMsg {
	data := &dao.WhiteList{Device: device}
	db := orm.GetOrm(ctx.GetContext())
	db.HDelPB(redisclient.GetGuildRedis(), key, []orm.RedisInterface{data})
	return db.Do()
}

func (w *WhiteList) GetAll() ([]*dao.WhiteList, *errmsg.ErrMsg) {
	list := make([]*dao.WhiteList, 0)
	db := orm.GetOrm(ctx.GetContext())
	if err := db.HGetAll(redisclient.GetGuildRedis(), key, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (w *WhiteList) GetOne(deviceId string) (*dao.WhiteList, bool, *errmsg.ErrMsg) {
	out := &dao.WhiteList{
		Device: deviceId,
	}
	ok, err := orm.GetOrm(ctx.GetContext()).HGetPB(redisclient.GetGuildRedis(), key, out)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	return out, true, nil
}
