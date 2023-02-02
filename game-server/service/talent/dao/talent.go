package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/game-server/service/talent/values"
)

func GetTalent(ctx *ctx.Context) (values.TalentI, *errmsg.ErrMsg) {
	res := &dao.Talent{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.Talent{
			RoleId: ctx.RoleId,
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	}
	if res.LockStoneIds == nil {
		res.LockStoneIds = map[int64]bool{}
	}
	if res.FirstUpdateM == nil {
		res.FirstUpdateM = map[int64]int64{}
	}
	return values.NewTalent(res), nil
}

func SaveTalent(ctx *ctx.Context, i values.TalentI) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), i.ToDao())
}
