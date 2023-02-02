package rule

import (
	"time"

	"coin-server/common/ctx"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

func GetAllAvailableActivity(ctx *ctx.Context) ([]*rulemodel.ActivityReward, []*rulemodel.ActivityReward) {
	list := rule.MustGetReader(ctx).ActivityReward.List()
	now := timer.StartTime(ctx.StartTime).UnixMilli()
	activated := make([]*rulemodel.ActivityReward, 0)
	inactive := make([]*rulemodel.ActivityReward, 0)
	for i := 0; i < len(list); i++ {
		item := &list[i]
		begin, err := time.ParseInLocation("2006-01-02 15:04:05", item.ActivityOpenTime, time.Local)
		if err != nil {
			continue
		}
		end := begin.Add(time.Duration(item.DurationTime) * time.Second)
		if begin.UnixMilli() <= now && (item.DurationTime == -1 || end.UnixMilli() > now) {
			activated = append(activated, item)
		}
		if begin.UnixMilli() > now {
			inactive = append(inactive, item)
		}
	}
	return activated, inactive
}

func GetEventById(ctx *ctx.Context, id values.Integer) (*rulemodel.Activity, bool) {
	cfg, ok := rule.MustGetReader(ctx).Activity.GetActivityById(id)
	return cfg, ok
}
