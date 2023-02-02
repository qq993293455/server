package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetXDayGoalInfo(ctx *ctx.Context, roleId values.RoleId) *dao.MoonthlyCardData {
	ret := &dao.MoonthlyCardData{
		RoleId: roleId,
	}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		ret.Infos = make(map[int64]*models.MoonthlyCardActivityInfo)
	}

	for _, aData := range ret.Infos {
		if aData.Progress == nil {
			aData.Progress = make(map[int64]int64)
		}
		if aData.ActivationCards == nil {
			aData.ActivationCards = make(map[int64]*models.MoonthlyCardInfo)
		}
		if aData.LastLoginTime == nil {
			aData.LastLoginTime = make(map[int64]int64)
		}
		if aData.PurchaseTimes == nil {
			aData.PurchaseTimes = make(map[int64]*models.MoonthlyCardPurchaseTimes)
		}
	}

	return ret
}

func SaveSevenDaysInfo(ctx *ctx.Context, info *dao.MoonthlyCardData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), info)
}
