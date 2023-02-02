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

func GetEquip(ctx *ctx.Context, roleId values.RoleId, equipId values.EquipId) (*dao.Equipment, *errmsg.ErrMsg) {
	equip := &dao.Equipment{EquipId: equipId}
	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getEquipKey(roleId), equip)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return equip, nil
}

func GetManyEquip(ctx *ctx.Context, roleId values.RoleId, equips []*dao.Equipment) *errmsg.ErrMsg {
	equip := make([]orm.RedisInterface, len(equips))
	for idx := range equip {
		equip[idx] = equips[idx]
	}
	not, err := ctx.NewOrm().HMGetPB(redisclient.GetUserRedis(), getEquipKey(roleId), equip)
	if err != nil {
		return err
	}
	if len(not) != 0 {
		return errmsg.NewErrBagEquipNotExist()
	}
	return nil
}

func GetEquipBag(ctx *ctx.Context, roleId values.RoleId) ([]*dao.Equipment, *errmsg.ErrMsg) {
	bag := make([]*dao.Equipment, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getEquipKey(roleId), &bag)
	if err != nil {
		return nil, err
	}
	return bag, nil
}

func SaveEquip(ctx *ctx.Context, roleId values.RoleId, equip *dao.Equipment) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getEquipKey(roleId), equip)
	return
}

func SaveManyEquip(ctx *ctx.Context, roleId values.RoleId, equips []*dao.Equipment) {
	if len(equips) == 0 {
		return
	}
	add := make([]orm.RedisInterface, len(equips))
	for idx := range add {
		add[idx] = equips[idx]
	}
	ctx.NewOrm().HMSetPB(redisclient.GetUserRedis(), getEquipKey(roleId), add)
	return
}

func DelEquip(ctx *ctx.Context, roleId values.RoleId, equip *dao.Equipment) {
	equips := []orm.RedisInterface{equip}
	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), getEquipKey(roleId), equips)
}

func DelManyEquip(ctx *ctx.Context, roleId values.RoleId, equips ...values.EquipId) *errmsg.ErrMsg {
	if len(equips) == 0 {
		return nil
	}
	del := make([]orm.RedisInterface, 0)
	for _, v := range equips {
		del = append(del, &dao.Equipment{EquipId: v})
	}
	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), getEquipKey(roleId), del)
	return nil
}

func getEquipKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.EquipBag, values.Hash, roleId)
}

func GetEquipId(ctx *ctx.Context, roleId values.RoleId) (*dao.EquipId, *errmsg.ErrMsg) {
	data := &dao.EquipId{RoleId: roleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SaveEquipId(ctx *ctx.Context, data *dao.EquipId) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}
