package model

import (
	client_version "coin-server/common/client-version"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
)

type Version struct {
}

func NewVersion() *Version {
	return &Version{}
}
func (v *Version) GetPB(key string) (*dao.ClientVersion, *errmsg.ErrMsg) {
	data := &dao.ClientVersion{Key: key}
	ok, err := orm.GetOrm(ctx.GetContext()).GetPB(client_version.GetClientVersionRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return data, err
	}
	return data, nil
}

func (v *Version) SavePB(data *dao.ClientVersion) *errmsg.ErrMsg {
	db := orm.GetOrm(ctx.GetContext())
	db.SetPB(client_version.GetClientVersionRedis(), data)
	return db.Do()
}
