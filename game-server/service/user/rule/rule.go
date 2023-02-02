package rule

import (
	"sort"

	"coin-server/common/ctx"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

func GetTitleSkill(c *ctx.Context, title values.Integer) map[int64]int64 {
	titleCfg, ok := rule.MustGetReader(c).RoleLvTitle.GetRoleLvTitleById(title)
	if !ok {
		return map[int64]int64{}
	}
	return titleCfg.TitleSkill
}

func GetPrevTitleSkill(c *ctx.Context, title values.Integer) map[int64]int64 {
	return GetTitleSkill(c, title-1)
}

type ExpSkipEffcient struct {
	Count values.Integer
	Rate  values.Float
}

func GetExpSkipCfg(ctx *ctx.Context, id values.ItemId) (*rulemodel.ExpSkip, bool) {
	item, ok := rule.MustGetReader(ctx).ExpSkip.GetExpSkipById(id)
	return item, ok
}

func GetExpSkipEffcient(ctx *ctx.Context, count values.Integer) *ExpSkipEffcient {
	data, ok := rule.MustGetReader(ctx).KeyValue.GetMapInt64Int64("ExpSkipEffcient")
	if !ok {
		return &ExpSkipEffcient{}
	}
	list := make([]*ExpSkipEffcient, 0, len(data))
	for _count, rate := range data {
		list = append(list, &ExpSkipEffcient{
			Count: _count,
			Rate:  values.Float(rate),
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Count < list[j].Count
	})
	if count <= list[0].Count {
		return list[0]
	}
	if count >= list[len(list)-1].Count {
		return list[len(list)-1]
	}
	var find *ExpSkipEffcient
	for i := 0; i < len(list); i++ {
		if count > list[i].Count && count <= list[i+1].Count {
			find = list[i+1]
			break
		}
	}
	return find
}

func GetExpSkipMaxCount(ctx *ctx.Context) values.Integer {
	return rule.MustGetReader(ctx).ExpSkip.GetMaxExpSkipCount()
}

func ExpGetUpperLimit(ctx *ctx.Context) values.Integer {
	v, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("ExpGetUpperLimit")
	if !ok {
		return 0
	}
	return v/10000 + 1
}

func GetExpGetTimeLimit(ctx *ctx.Context) values.Integer {
	v, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("ExpGetTimeLimit")
	if !ok {
		return 0
	}
	return v
}

func IsNormalPayItem(ctx *ctx.Context, pcId values.Integer) (bool, bool) {
	v, ok := rule.MustGetReader(ctx).Charge.GetChargeById(pcId)
	if !ok {
		return false, false
	}
	return v.Normal, ok
}

// MustGetInitialUseOfAvatar 获取初始使用的头像、头像框 [头像, 头像框]
func MustGetInitialUseOfAvatar(c *ctx.Context) []values.Integer {
	ret, ok := rule.MustGetReader(c).KeyValue.GetIntegerArray("InitialUseOfAvatar")
	if !ok {
		panic("key_value InitialUseOfAvatar not found")
	}
	if len(ret) != 2 {
		panic("key_value InitialUseOfAvatar len != 2")
	}
	return ret
}
