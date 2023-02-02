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

func GetSkillStone(ctx *ctx.Context, roleId values.RoleId, stoneId values.ItemId) (*dao.SkillStone, *errmsg.ErrMsg) {
	stone := &dao.SkillStone{ItemId: stoneId}
	_, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getSkillStoneKey(roleId), stone)
	if err != nil {
		return nil, err
	}
	return stone, err
}

func GetSkillStones(ctx *ctx.Context, roleId values.RoleId, stones []*dao.SkillStone) *errmsg.ErrMsg {
	list := make([]orm.RedisInterface, len(stones))
	for idx := range list {
		list[idx] = stones[idx]
	}
	_, err := ctx.NewOrm().HMGetPB(redisclient.GetUserRedis(), getSkillStoneKey(roleId), list)
	if err != nil {
		return err
	}
	return nil
}

func GetSkillStoneBag(ctx *ctx.Context, roleId values.RoleId) ([]*dao.SkillStone, *errmsg.ErrMsg) {
	bag := make([]*dao.SkillStone, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getSkillStoneKey(roleId), &bag)
	if err != nil {
		return nil, err
	}
	return bag, nil
}

func SaveSkillStone(ctx *ctx.Context, roleId values.RoleId, stone *dao.SkillStone) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getSkillStoneKey(roleId), stone)
	return
}

// func DelSkillStone(ctx *ctx.Context, roleId values.RoleId, stone *dao.SkillStone) *errmsg.ErrMsg {
// 	stones := []orm.RedisInterface{stone}
// 	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), getSkillStoneKey(roleId), stones)
// 	bagLen, err := GetBagLen(ctx, roleId)
// 	if err != nil {
// 		return err
// 	}
// 	bagLen.Length--
// 	return UpdateBagLen(ctx, bagLen)
// }

func SaveManySkillStone(ctx *ctx.Context, roleId values.RoleId, stones []*dao.SkillStone) {
	if len(stones) == 0 {
		return
	}
	add := make([]orm.RedisInterface, len(stones))
	for idx := range add {
		add[idx] = stones[idx]
	}
	ctx.NewOrm().HMSetPB(redisclient.GetUserRedis(), getSkillStoneKey(roleId), add)
	return
}

func DelManySkillStone(ctx *ctx.Context, roleId values.RoleId, stones []*dao.SkillStone) {
	if len(stones) == 0 {
		return
	}
	del := make([]orm.RedisInterface, len(stones))
	for idx := range del {
		del[idx] = stones[idx]
	}
	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), getSkillStoneKey(roleId), del)
}

func getSkillStoneKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.SkillStone, values.Hash, roleId)
}
