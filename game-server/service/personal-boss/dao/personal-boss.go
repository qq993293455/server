package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

const GlobalPersonalBossInfoKey = "global_personal_boss_info_key"

func GetGlobalPersonalBossInfo(ctx *ctx.Context) (*daopb.GlobalPersonalBossInfo, *errmsg.ErrMsg) {
	data := &daopb.GlobalPersonalBossInfo{Key: GlobalPersonalBossInfoKey}
	_, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SaveGlobalPersonalBossInfo(ctx *ctx.Context, data *daopb.GlobalPersonalBossInfo) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}

func GetPersonalBossInfo(ctx *ctx.Context, roleId values.RoleId) (*daopb.PersonalBossInfo, *errmsg.ErrMsg) {
	data := &daopb.PersonalBossInfo{RoleId: roleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SavePersonalBossInfo(ctx *ctx.Context, data *daopb.PersonalBossInfo) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}

func GetPersonalBossHelperInfo(ctx *ctx.Context, roleId values.RoleId) (*daopb.PersonalBossHelperInfo, *errmsg.ErrMsg) {
	data := &daopb.PersonalBossHelperInfo{RoleId: roleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SavePersonalBossHelperInfo(ctx *ctx.Context, data *daopb.PersonalBossHelperInfo) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}
