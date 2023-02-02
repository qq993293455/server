package rule

import (
	"time"

	"coin-server/common/ctx"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

func GetConfigByTitle(ctx *ctx.Context, title values.Integer) (map[values.Integer]values.Integer, bool) {
	cfg, ok := rule.MustGetReader(ctx).RoleLvTitle.GetRoleLvTitleById(title)
	if !ok {
		return nil, false
	}
	return cfg.CombatBattle, true
}

// GetRacingRankDuration 返回的是秒
func GetRacingRankDuration(ctx *ctx.Context) time.Duration {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt("CombatBattleDay") // 这里名字叫"day"但实际配置的是分钟
	return time.Duration(v) * time.Minute
}

// GetRankCount 获取排行榜上总人数（玩家自己除外）
func GetRankCount(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("CombatBattleNum")
	return v
}

func GetNextRefresh(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt("CombatBattleRefreshCD") // 秒
	now := timer.StartTime(ctx.StartTime)
	// 若配置不正确，默认10分钟cd
	if v <= 0 {
		return now.Add(time.Minute * 10).UnixMilli()
	}
	return now.Add(time.Second * time.Duration(v)).UnixMilli()
}

func GetMailId(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("CombatBattleMail")
	return v
}

func GetMailExpire(ctx *ctx.Context, id values.Integer) values.Integer {
	item, ok := rule.MustGetReader(ctx).Mail.GetMailById(id)
	if !ok {
		return 0
	}
	return item.Overdue
}

func GetRankReward(ctx *ctx.Context, rank values.Integer) *rulemodel.CombatBattle {
	list := rule.MustGetReader(ctx).CombatBattle.List()
	var target *rulemodel.CombatBattle
	for i := 0; i < len(list); i++ {
		item := list[i]
		min := item.SubsectionRank[0]
		max := item.SubsectionRank[1]
		if rank >= min && rank <= max {
			target = &item
			break
		}
	}
	if target == nil {
		target = &list[len(list)-1]
	}
	return target
}

func GetMaxRanking(ctx *ctx.Context) values.Integer {
	return rule.MustGetReader(ctx).CombatBattle.GetMaxRacingRanking()
}

func GetRankRewardById(ctx *ctx.Context, id values.Integer) (*rulemodel.CombatBattle, bool) {
	return rule.MustGetReader(ctx).CombatBattle.GetCombatBattleById(id)
}
