package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/values"
	rule2 "coin-server/game-server/service/bag/rule"
)

func InitBagConfig(ctx *ctx.Context, cap values.Integer) *dao.BagConfig {
	data := &dao.BagConfig{
		RoleId: ctx.RoleId,
		Config: &models.BagConfig{
			Quality:          0,
			Capacity:         cap,
			CapacityOccupied: 0,
			BuyCount:         0,
			UnlockTime:       0,
		},
	}
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
	return data
}

// func GetBagLen(ctx *ctx.Context, roleId values.RoleId) (*dao.BagLen, *errmsg.ErrMsg) {
// 	ret := &dao.BagLen{RoleId: roleId}
// 	has, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
// 	if !has {
// 		ctx.NewOrm().SetPB(redisclient.GetUserRedis(), &dao.BagLen{
// 			RoleId: roleId,
// 			Length: 0,
// 		})
// 	}
// 	return ret, err
// }
//
// func UpdateBagLen(ctx *ctx.Context, bagLen *dao.BagLen) *errmsg.ErrMsg {
// 	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), bagLen)
// 	return nil
// }

func GetBagConfig(ctx *ctx.Context, roleId values.RoleId) (*dao.BagConfig, *errmsg.ErrMsg) {
	cfg := &dao.BagConfig{RoleId: roleId}
	has, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), cfg)
	if err != nil {
		return nil, err
	}
	if !has {
		return InitBagConfig(ctx, rule2.GetBagInitCap(ctx)), nil
	}
	return cfg, nil
}

func SaveBagConfig(ctx *ctx.Context, config *dao.BagConfig) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), config)
	return
}
