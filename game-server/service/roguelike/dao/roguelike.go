package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetRlDone(ctx *ctx.Context) (*dao.RoguelikeDone, *errmsg.ErrMsg) {
	res := &dao.RoguelikeDone{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.RoguelikeDone{
			RoleId:  ctx.RoleId,
			DoneMap: map[int64]bool{},
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	}
	if res.DoneMap == nil {
		res.DoneMap = map[int64]bool{}
	}
	return res, nil
}

func SaveRlDone(ctx *ctx.Context, data *dao.RoguelikeDone) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}
