package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

const dbKey = "roguelike_boss_skill"

func GetBossSkill(ctx *ctx.Context) (*dao.RoguelikeBossSkill, *errmsg.ErrMsg) {
	res := &dao.RoguelikeBossSkill{Key: dbKey}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.RoguelikeBossSkill{
			Key: dbKey,
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	}
	return res, nil
}

func SaveBossSkill(ctx *ctx.Context, v *dao.RoguelikeBossSkill) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), v)
}
