package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
	"coin-server/game-server/util"
)

func Get(ctx *ctx.Context, id values.RoleId) (*dao.Divination, *errmsg.ErrMsg) {
	data := &dao.Divination{RoleId: id}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetGuildRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	nextReset := util.DefaultNextRefreshTime().UnixMilli()
	if data.ResetAt < nextReset {
		data.AvailableCount = data.TotalCount
		data.ResetAt = nextReset
	}
	return data, nil
}

func Save(ctx *ctx.Context, data *dao.Divination) {
	ctx.NewOrm().SetPB(redisclient.GetGuildRedis(), data)
}
