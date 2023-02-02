package db

import (
	client_version "coin-server/common/client-version"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
)

func GetClientVersion() (*dao.ClientVersion, *errmsg.ErrMsg) {
	data := &dao.ClientVersion{Key: client_version.ClientVersionKey}
	_, err := orm.GetOrm(ctx.GetContext()).GetPB(client_version.GetClientVersionRedis(), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
