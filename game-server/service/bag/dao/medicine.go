package dao

import (
	"fmt"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
	"coin-server/rule"
)

func GetMedicineInfo(ctx *ctx.Context, roleId values.RoleId) (*dao.MedicineInfo, *errmsg.ErrMsg) {
	cfg := &dao.MedicineInfo{RoleId: roleId}
	has, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), cfg)
	if err != nil {
		return nil, err
	}
	if !has {
		cfg = &dao.MedicineInfo{
			RoleId:   roleId,
			NextTake: map[int64]int64{},
			Open:     map[int64]int64{},
			AutoTake: getDefaultMedicine(ctx),
		}
		for k := range cfg.AutoTake {
			cfg.NextTake[k] = 0
			cfg.Open[k] = 0
		}
		ctx.NewOrm().SetPB(redisclient.GetUserRedis(), cfg)
	}
	return cfg, nil
}

func SaveMedicineInfo(ctx *ctx.Context, cd *dao.MedicineInfo) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), cd)
	return
}

func getDefaultMedicine(ctx *ctx.Context) map[values.Integer]values.Integer {
	r := rule.MustGetReader(ctx)
	hp, ok := r.KeyValue.GetInt64("TakeMedicineHPPercentage")
	if !ok {
		panic(fmt.Sprintf("TakeMedicineHPPercentage Key not found"))
	}
	mp, ok := r.KeyValue.GetInt64("TakeMedicineMPPercentage")
	if !ok {
		panic(fmt.Sprintf("TakeMedicineMPPercentage Key not found"))
	}
	return map[values.Integer]values.Integer{1: hp, 2: mp}
}
