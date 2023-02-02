package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetQuestRewardInfo(ctx *ctx.Context, roleId values.RoleId) *dao.QuestRewardData {
	ret := &dao.QuestRewardData{
		RoleId: roleId,
	}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		return nil
	}
	return ret
}

func SaveQuestRewardInfo(ctx *ctx.Context, info *dao.QuestRewardData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), info)
}
