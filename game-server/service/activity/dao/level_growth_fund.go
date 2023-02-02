package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
)

func GetLevelGrowthFundData(ctx *ctx.Context) (*dao.LevelGrowthFund, *errmsg.ErrMsg) {
	data := &dao.LevelGrowthFund{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		data = &dao.LevelGrowthFund{
			RoleId: ctx.RoleId,
			Buy:    false,
			Info:   nil,
		}
	}
	if data.Info == nil {
		data.Info = map[int64]*models.LevelGrowthFundItem{}
	}
	return data, nil
}

func SaveLevelGrowthFundData(ctx *ctx.Context, data *dao.LevelGrowthFund) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}
