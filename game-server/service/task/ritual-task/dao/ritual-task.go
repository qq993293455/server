package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func Get(ctx *ctx.Context, id values.RoleId) (*dao.ChaosRitual, *errmsg.ErrMsg) {
	data := &dao.ChaosRitual{RoleId: id}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetGuildRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return data, nil
}

func Save(ctx *ctx.Context, data *dao.ChaosRitual) {
	ctx.NewOrm().SetPB(redisclient.GetGuildRedis(), data)
}
