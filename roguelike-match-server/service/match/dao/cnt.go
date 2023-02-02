package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetCnt(ctx *ctx.Context, roleId values.RoleId) (*dao.RoguelikeJoinCnt, *errmsg.ErrMsg) {
	res := &dao.RoguelikeJoinCnt{RoleId: roleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.RoguelikeJoinCnt{
			RoleId: roleId,
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	}
	return res, nil
}

func SaveCnt(ctx *ctx.Context, v *dao.RoguelikeJoinCnt) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), v)
}
