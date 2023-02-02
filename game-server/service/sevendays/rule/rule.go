package rule

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

const sevenDay = 10001

func GetActivityLoginRewardData(ctx *ctx.Context) ([]rulemodel.ActivityLoginReward, *errmsg.ErrMsg) {
	config := rule.MustGetReader(ctx).ActivityLoginReward.List()
	return config, nil
}

// func GetActivityRewardData(ctx *ctx.Context) (*rulemodel.ActivityReward, *errmsg.ErrMsg) {
// 	config, ok := rule.MustGetReader(ctx).ActivityReward.GetActivityRewardById(sevenDay)
// 	if !ok {
// 		return nil, errmsg.NewErrSevenDaysConfig()
// 	}
// 	return config, nil
// }

func GetSevenDayRefreshTime(ctx *ctx.Context, log *logger.Logger) (int64, *errmsg.ErrMsg) {
	ActivityEveryDayRefreshTime, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("ActivityEveryDayRefreshTime")
	if !ok {
		log.Error("ActivityEveryDayRefreshTime error")
		return 0, errmsg.NewErrSevenDaysConfig()
	}
	return ActivityEveryDayRefreshTime * 3600, nil
}

func GetGetBossHallCnf(ctx *ctx.Context, bossId int64) (*rulemodel.BossHall, *errmsg.ErrMsg) {
	conf, ok := rule.MustGetReader(ctx).BossHall.GetBossHallById(bossId)
	if !ok {
		return nil, errmsg.NewErrBossHallConfig()
	}
	return conf, nil
}
