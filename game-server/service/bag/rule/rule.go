package rule

import (
	"time"

	"coin-server/common/ctx"
	"coin-server/common/values"
	"coin-server/game-server/util"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

func GetEquipById(ctx *ctx.Context, id values.ItemId) (*rulemodel.Equip, bool) {
	return rule.MustGetReader(ctx).Equip.GetEquipById(id)
}

func GetItemById(ctx *ctx.Context, id values.ItemId) (*rulemodel.Item, bool) {
	return rule.MustGetReader(ctx).Item.GetItemById(id)
}

func GetEquipEntryById(ctx *ctx.Context, id values.Integer) (*rulemodel.EquipEntry, bool) {
	return rule.MustGetReader(ctx).EquipEntry.GetEquipEntryById(id)
}

func GetUnlockEquipAffix(ctx *ctx.Context, star values.Level) map[int]struct{} {
	list, ok := rule.MustGetReader(ctx).KeyValue.GetIntegerArray("EquipUpStars")
	if !ok {
		return nil
	}
	ret := make(map[int]struct{})
	for i, v := range list {
		if star >= v {
			ret[i] = struct{}{}
		}
	}
	return ret
}

func UnlockEquipAffixCfg(ctx *ctx.Context) []values.Integer {
	list, ok := rule.MustGetReader(ctx).KeyValue.GetIntegerArray("EquipUpStars")
	if !ok {
		list = make([]values.Integer, 0)
	}
	return list
}

func GetEquipEntryCfgByAttribute(id values.Integer) []rulemodel.EquipEntry {
	return rule.MustGetReader(ctx.GetContext()).Equip.GetEquipEntry(id)
}

func GetAttrById(ctx *ctx.Context, id values.AttrId) (*rulemodel.Attr, bool) {
	return rule.MustGetReader(ctx).Attr.GetAttrById(id)
}

func GetAttrTransConfigById(ctx *ctx.Context, id values.AttrId) []rulemodel.AttrTrans {
	return rule.MustGetReader(ctx).AttrTrans.GetAttrTransByAttrId(id)
}

func GetRoleLvConfigByLv(ctx *ctx.Context, lv values.Level) (*rulemodel.RoleLv, bool) {
	item, ok := rule.MustGetReader(ctx).RoleLv.GetRoleLvById(lv)
	return item, ok
}

func GetHeroBuildIdList(ctx *ctx.Context, id values.HeroId) []values.Integer {
	return rule.MustGetReader(ctx).CustomParse.DeriveHeroMap(id)
}

func GetTalentList(ctx *ctx.Context, buildId []values.Integer, level values.Level) map[values.HeroBuildId][]*util.EquipTalentBonus {
	talentList := rule.MustGetReader(ctx).Talent.List()
	ret := make(map[values.HeroBuildId][]*util.EquipTalentBonus)
	for _, id := range buildId {
		for i := 0; i < len(talentList); i++ {
			talent := &talentList[i]
			if talent.BuildId == id {
				if _, ok := ret[id]; !ok {
					ret[id] = make([]*util.EquipTalentBonus, 0)
				}
				ret[id] = append(ret[id], &util.EquipTalentBonus{
					Talent: talent,
					Level:  level,
				})
			}
		}
	}
	return ret
}

func GetTalentById(ctx *ctx.Context, talentId values.Integer) (*rulemodel.Talent, bool) {
	talent, ok := rule.MustGetReader(ctx).Talent.GetTalentById(talentId)
	return talent, ok
}

func GetBuffById(ctx *ctx.Context, buffId values.HeroBuffId) (*rulemodel.Buff, bool) {
	buff, ok := rule.MustGetReader(ctx).Buff.GetBuffById(buffId)
	return buff, ok
}

func GetBagInitCap(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("DefaultBagNum")
	return v
}

func GetBagMaxCap(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("deBagCapacity")
	return v
}

func GetBagOnceUnlockLatticeCount(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("AddBagNum")
	return v
}

func GetExpandCapacitySpeedUpCost(ctx *ctx.Context) []values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetIntegerArray("AddBagTimeCost")
	return v
}

func GetUnlockTime(ctx *ctx.Context, id values.Integer) time.Duration {
	reader := rule.MustGetReader(ctx)
	cfg, ok := reader.Bag.GetBagById(id)
	if !ok {
		return time.Duration(reader.Bag.List()[reader.Bag.Len()-1].AddBagTime) * time.Second
	}
	return time.Duration(cfg.AddBagTime) * time.Second
}

func GetMailConfigTextId(ctx *ctx.Context, id values.Integer) (*rulemodel.Mail, bool) {
	item, ok := rule.MustGetReader(ctx).Mail.GetMailById(id)
	return item, ok
}
