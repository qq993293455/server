package rule

import (
	"time"

	"coin-server/common/ctx"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

func GetAllExpeditionQuantity(ctx *ctx.Context) map[models.TaskType][]rulemodel.ExpeditionQuantity {
	return rule.MustGetReader(ctx).ExpeditionQuantity.GetAllExpeditionQuantityGroupByTaskType()
}

func GetAllExpedition(ctx *ctx.Context) map[models.TaskType]map[values.Quality][]rulemodel.Expedition {
	return rule.MustGetReader(ctx).Expedition.GetAllExpedition()
}

func GetExpeditionById(ctx *ctx.Context, id values.Integer) (*rulemodel.Expedition, bool) {
	item, ok := rule.MustGetReader(ctx).Expedition.GetExpeditionById(id)
	return item, ok
}

func GetExpeditionQualityWeight(ctx *ctx.Context) (map[values.Integer]values.Integer, bool) {
	w, ok := rule.MustGetReader(ctx).KeyValue.GetMapInt64Int64("ExpeditionQualityWeight")
	return w, ok
}

func GetExpeditionRefreshFloors(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("ExpediotnRefreshFloors")
	return v
}

func GetExpeditionAccelerateCost(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("ExpeditionAccelerateCost")
	return v
}

func GetDefaultRefreshTime(ctx *ctx.Context) time.Duration {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("DefaultRefreshTime")
	return time.Duration(v)
}

func GetExpeditionFreeRefresh(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("ExpeditionFreeRefresh")
	return v
}

func GetExpeditionRefreshItem(ctx *ctx.Context) values.ItemId {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("ExpeditionRefreshItem")
	return v
}

func GetExpeditionRefreshCost(ctx *ctx.Context) (*models.Item, bool) {
	v, ok := rule.MustGetReader(ctx).KeyValue.GetItem("ExpeditionRefreshCost")
	return v, ok
}

func GetExpeditionCostRecovery(ctx *ctx.Context) (values.Float, values.Integer) {
	// 服务启动的时候判断了ExpeditionCostRecovery一定有值且一定为2个元素
	v, _ := rule.MustGetReader(ctx).KeyValue.GetIntegerArray("ExpeditionCostRecovery")
	return values.Float(v[0]), v[1] // 分别对应 恢复间隔（分钟）和恢复数量
}

func GetExpeditionTaskCostItemId(ctx *ctx.Context) values.ItemId {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("ExpeditionTaskCost")
	return v
}

func GetExpeditionCostLimit(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("ExpeditionCostLimit")
	return v
}

func GetRelicsSkill(ctx *ctx.Context, id values.Integer) (*rulemodel.RelicsSkill, bool) {
	rs, ok := rule.MustGetReader(ctx).RelicsSkill.GetRelicsSkillById(id)
	return rs, ok
}

func GetExpeditionRewardByLevel(ctx *ctx.Context, expeditionId values.Integer, level values.Level) (rulemodel.ExpeditionReward, bool) {
	list := rule.MustGetReader(ctx).ExpeditionReward.GetExpeditionRewardByExpeditionId(expeditionId)
	if len(list) <= 0 {
		return rulemodel.ExpeditionReward{}, false
	}
	maxCfg := list[0]
	minCfg := list[len(list)-1]
	if level >= maxCfg.MaxLv {
		return maxCfg, true
	}
	if level <= minCfg.MinLv || level <= minCfg.MaxLv {
		return minCfg, true
	}

	// list 是按从大到小的顺序排列
	for _, item := range list {
		if level >= item.MinLv && level <= item.MaxLv {
			return item, true
		}
	}
	return rulemodel.ExpeditionReward{}, false
}
