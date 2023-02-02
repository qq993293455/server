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

func GetEquipBrief(ctx *ctx.Context, roleId values.RoleId) ([]*dao.EquipmentBrief, *errmsg.ErrMsg) {
	bag := make([]*dao.EquipmentBrief, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getEquipBriefKey(roleId), &bag)
	if err != nil {
		return nil, err
	}
	return bag, nil
}

func SaveEquipBrief(ctx *ctx.Context, roleId values.RoleId, equips ...*dao.EquipmentBrief) {
	if len(equips) == 0 {
		return
	}
	add := make([]orm.RedisInterface, len(equips))
	for idx := range add {
		add[idx] = equips[idx]
	}
	ctx.NewOrm().HMSetPB(redisclient.GetUserRedis(), getEquipBriefKey(roleId), add)
}

func DelEquipBrief(ctx *ctx.Context, roleId values.RoleId, equips ...values.EquipId) {
	// if len(equips) == 0 {
	// 	return
	// }
	// del := make([]orm.RedisInterface, 0)
	// for _, v := range equips {
	// 	del = append(del, &dao.EquipmentBrief{EquipId: v})
	// }
	// ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), getEquipBriefKey(roleId), del)
}

func getEquipBriefKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.EquipBriefBag, values.Hash, roleId)
}
