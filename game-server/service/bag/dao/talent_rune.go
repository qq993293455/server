package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetTalentRune(ctx *ctx.Context, roleId values.RoleId, runeId values.RuneId) (*dao.TalentRune, *errmsg.ErrMsg) {
	stone := &dao.TalentRune{RuneId: runeId}
	_, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getTalentRuneKey(roleId), stone)
	if err != nil {
		return nil, err
	}
	return stone, err
}

func GetTalentRunes(ctx *ctx.Context, roleId values.RoleId, stones []*dao.TalentRune) *errmsg.ErrMsg {
	list := make([]orm.RedisInterface, len(stones))
	for idx := range list {
		list[idx] = stones[idx]
	}
	_, err := ctx.NewOrm().HMGetPB(redisclient.GetUserRedis(), getTalentRuneKey(roleId), list)
	if err != nil {
		return err
	}
	return nil
}

func GetTalentRuneBag(ctx *ctx.Context, roleId values.RoleId) ([]*dao.TalentRune, *errmsg.ErrMsg) {
	bag := make([]*dao.TalentRune, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getTalentRuneKey(roleId), &bag)
	if err != nil {
		return nil, err
	}
	return bag, nil
}

func SaveTalentRune(ctx *ctx.Context, roleId values.RoleId, stone *dao.TalentRune) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getTalentRuneKey(roleId), stone)
	return
}

// func DelTalentRune(ctx *ctx.Context, roleId values.RoleId, stone *dao.TalentRune) *errmsg.ErrMsg {
// 	stones := []orm.RedisInterface{stone}
// 	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), getTalentRuneKey(roleId), stones)
// 	bagLen, err := GetBagLen(ctx, roleId)
// 	if err != nil {
// 		return err
// 	}
// 	bagLen.Length--
// 	return UpdateBagLen(ctx, bagLen)
// }

func SaveManyTalentRune(ctx *ctx.Context, roleId values.RoleId, runes []*dao.TalentRune) {
	if len(runes) == 0 {
		return
	}
	add := make([]orm.RedisInterface, len(runes))
	for idx := range add {
		add[idx] = runes[idx]
	}
	ctx.NewOrm().HMSetPB(redisclient.GetUserRedis(), getTalentRuneKey(roleId), add)
	return
}

func DelManyTalentRune(ctx *ctx.Context, roleId values.RoleId, stones []*dao.TalentRune) {
	if len(stones) == 0 {
		return
	}
	del := make([]orm.RedisInterface, len(stones))
	for idx := range del {
		del[idx] = stones[idx]
	}
	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), getTalentRuneKey(roleId), del)
	// bagLen, err := GetBagLen(ctx, roleId)
	// if err != nil {
	// 	return err
	// }
	// bagLen.Length -= values.Integer(len(stones))
	// return UpdateBagLen(ctx, bagLen)
}

func getTalentRuneKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.TalentRune, values.Hash, roleId)
}
