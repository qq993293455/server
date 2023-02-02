package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetAdvanceDungeon(ctx *ctx.Context, dungeonId values.Integer) (*dao.Dungeon, *errmsg.ErrMsg) {
	out := &dao.Dungeon{
		Id: dungeonId,
	}
	ok, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getAdvanceDungeonKey(ctx.RoleId), out)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return out, nil
}

func SaveAdvanceDungeon(ctx *ctx.Context, dungeon *dao.Dungeon) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getAdvanceDungeonKey(ctx.RoleId), dungeon)
}

func getAdvanceDungeonKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.AdvanceDungeon, values.Hash, roleId)
}
