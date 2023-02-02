package model

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

const beta_key = "BETA_WHITE_LIST"

type BetaWhiteList struct{}

func NewBetaWhiteList() *BetaWhiteList {
	return &BetaWhiteList{}
}

func (w *BetaWhiteList) Save(data *dao.BetaWhiteList) *errmsg.ErrMsg {
	db := orm.GetOrm(ctx.GetContext())
	db.HMSetPB(redisclient.GetGuildRedis(), beta_key, []orm.RedisInterface{data})
	return db.Do()
}

func (w *BetaWhiteList) Del(device string) *errmsg.ErrMsg {
	data := &dao.BetaWhiteList{Device: device}
	db := orm.GetOrm(ctx.GetContext())
	db.HDelPB(redisclient.GetGuildRedis(), beta_key, []orm.RedisInterface{data})
	return db.Do()
}

func (w *BetaWhiteList) GetAll() ([]*dao.BetaWhiteList, *errmsg.ErrMsg) {
	list := make([]*dao.BetaWhiteList, 0)
	db := orm.GetOrm(ctx.GetContext())
	if err := db.HGetAll(redisclient.GetGuildRedis(), beta_key, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (w *BetaWhiteList) GetOne(deviceId string) (*dao.BetaWhiteList, bool, *errmsg.ErrMsg) {
	out := &dao.BetaWhiteList{
		Device: deviceId,
	}
	ok, err := orm.GetOrm(ctx.GetContext()).HGetPB(redisclient.GetGuildRedis(), beta_key, out)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	return out, true, nil
}
