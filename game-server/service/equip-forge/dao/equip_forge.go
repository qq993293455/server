package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
	values2 "coin-server/game-server/service/equip-forge/values"
)

func Get(ctx *ctx.Context) (*dao.EquipForge, *errmsg.ErrMsg) {
	res := &dao.EquipForge{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		data := newDao(ctx.RoleId)
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
		return data, nil
	}
	return res, nil
}

func Save(ctx *ctx.Context, data *dao.EquipForge) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}

func newDao(roleId values.RoleId) *dao.EquipForge {
	return &dao.EquipForge{
		RoleId:          roleId,
		Level:           values2.InitEquipForgeLevel,
		Exp:             0,
		ForgeCount:      map[int64]int64{},
		TotalForgeCount: 0,
	}
}
