package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetExpSkip(ctx *ctx.Context) (*dao.ExpSkip, bool, *errmsg.ErrMsg) {
	es := &dao.ExpSkip{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), es)
	if err != nil {
		return nil, false, err
	}
	if ok && es.RateDuration == nil {
		es.RateDuration = map[int64]int64{}
	}
	return es, ok, nil
}

func SaveExpSkip(ctx *ctx.Context, es *dao.ExpSkip) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), es)
}
