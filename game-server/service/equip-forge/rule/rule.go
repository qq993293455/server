package rule

import (
	"coin-server/common/ctx"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

func GetMaxForgeLevel(ctx *ctx.Context) values.Level {
	return rule.MustGetReader(ctx).ForgeLevel.GetMaxForgeLevel()
}

func GetLevelConfig(ctx *ctx.Context, level values.Level) (*rulemodel.ForgeLevel, bool) {
	return rule.MustGetReader(ctx).ForgeLevel.GetForgeLevelById(level)
}

func GetFixedRecipeInfo(ctx *ctx.Context, id values.Integer) (*rulemodel.ForgeFixedRecipe, bool) {
	return rule.MustGetReader(ctx).ForgeFixedRecipe.GetForgeFixedRecipeById(id)
}

func GetSupplement(ctx *ctx.Context, id values.Integer) (*rulemodel.ForgeSupplement, bool) {
	return rule.MustGetReader(ctx).ForgeSupplement.GetForgeSupplementById(id)
}

// GetForgeLevelBonus 获取打造等级对打造品质的加成
func GetForgeLevelBonus(ctx *ctx.Context, level values.Level, recipeLv values.Level) map[values.Quality]values.Integer {
	// 先找到对应装备等级下的所有数据，再根据当前打造等级在列表里找最大的记录
	list := rule.MustGetReader(ctx).ForgeLevel.GetForgeLevelByEquipLevel(recipeLv)
	if len(list) <= 0 {
		for recipeLv > 1 {
			recipeLv--
			list = rule.MustGetReader(ctx).ForgeLevel.GetForgeLevelByEquipLevel(recipeLv)
			if len(list) > 0 {
				break
			}
		}
	}
	var find map[values.Quality]values.Integer
	for i := 0; i < len(list); i++ {
		if level >= list[i].Id {
			// if i < len(list)-1 {
			// 	if list[i].Id >= recipeLv && list[i+1].Id < recipeLv {
			// 		find = list[i].FixedQualityProbability
			// 		break
			// 	}
			// } else if recipeLv >= list[i].Id {
			find = list[i].FixedQualityProbability
			// break
			// }
		}
	}
	return find
}

func GetEquipQuality(ctx *ctx.Context, itemId values.ItemId) values.Quality {
	item, ok := rule.MustGetReader(ctx).Item.GetItemById(itemId)
	if !ok {
		return 0
	}
	return item.Quality
}

func GetFixedNormalMinimumGuaranteeCount(ctx *ctx.Context) (values.Integer, bool) {
	return rule.MustGetReader(ctx).KeyValue.GetInt64("FixedNormalEquipMinimumGuarantee")
}

func GetForgeOnceExp(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("ForgeExp")
	return v
}

func GetEquipForgeMaxLevel(ctx *ctx.Context) values.Integer {
	return rule.MustGetReader(ctx).ForgeLevel.GetEquipForgeMaxLevel()
}

func GetMailConfigTextId(ctx *ctx.Context, id values.Integer) (*rulemodel.Mail, bool) {
	item, ok := rule.MustGetReader(ctx).Mail.GetMailById(id)
	return item, ok
}
