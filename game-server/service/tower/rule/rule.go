package rule

import (
	"coin-server/common/ctx"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

type TowerPublicConfig struct {
	Challenge_limit       int        //基础爬塔每日挑战次数上限
	Change_hero_cd        int        //基础爬塔战斗切换角色冷却时间（秒）
	Accumulate_max        int        //基础爬塔挂机奖励最大累积时长（配置为分, 读取转秒）
	Settlement_interval   int        //基础爬塔挂机奖励结算间隔时长（配置为分, 读取转秒）
	Cost_meditation_times [][2]int64 //基础爬塔每日冥想次数/Cost_Item消耗
	Cost_Item             int64      //基础爬塔每日冥想 消耗物品ID
	Free_meditation_times int        //基础爬塔每日免费冥想次数
	Meditation_time       int        //基础爬塔冥想奖励结算时长（分）
	Tower_refresh_time    int        //基础爬塔重置时间（时）
}

func GetTowerDefalutById(ctx *ctx.Context, id int64) (*rulemodel.TowerDefault, bool) {
	config, ok := rule.MustGetReader(ctx).TowerDefault.GetTowerDefaultById(id)
	return config, ok
}

func GetTowerDefalutMaxLevel(ctx *ctx.Context) int32 {
	config_datas := rule.MustGetReader(ctx).TowerDefault.List()
	max_level := 1
	for _, config := range config_datas {
		if config.Id > int64(max_level) {
			max_level = int(config.Id)
		}
	}
	return int32(max_level)
}

func GetTowerDefalutPublicInfo(ctx *ctx.Context) *TowerPublicConfig {
	var ok bool = false
	var config *TowerPublicConfig = new(TowerPublicConfig)
	config.Challenge_limit, ok = rule.MustGetReader(ctx).KeyValue.GetInt("DefaultTowerBattleLimit")
	if !ok {
		return nil
	}
	config.Change_hero_cd, ok = rule.MustGetReader(ctx).KeyValue.GetInt("DefaultTowerMultiHeroCD")
	if !ok {
		return nil
	}
	config.Accumulate_max, ok = rule.MustGetReader(ctx).KeyValue.GetInt("DefaultTowerAccumulateDuration")
	if !ok {
		return nil
	}
	config.Accumulate_max *= 60
	config.Settlement_interval, ok = rule.MustGetReader(ctx).KeyValue.GetInt("DefaultTowerAccumulateSettlement")
	if !ok {
		return nil
	}
	config.Settlement_interval *= 60
	config.Cost_meditation_times = rule.MustGetReader(ctx).DefaultTowerCost()

	config.Free_meditation_times, ok = rule.MustGetReader(ctx).KeyValue.GetInt("DefaultTowerCostFree")
	if !ok {
		return nil
	}
	config.Meditation_time, ok = rule.MustGetReader(ctx).KeyValue.GetInt("DefaultTowerCostDuration")
	if !ok {
		return nil
	}
	config.Meditation_time *= 60
	config.Tower_refresh_time, ok = rule.MustGetReader(ctx).KeyValue.GetInt("DefaultRefreshTime")
	if !ok {
		return nil
	}
	config.Tower_refresh_time /= 3600
	config.Cost_Item, ok = rule.MustGetReader(ctx).KeyValue.GetInt64("DefaultTowerCostitem")
	if !ok {
		return nil
	}
	return config
}

func GetTestSingleBattleById(ctx *ctx.Context, id int64) (*rulemodel.TestBattle, bool) {
	config, ok := rule.MustGetReader(ctx).TestBattle.GetTestBattleById(id)
	return config, ok
}
