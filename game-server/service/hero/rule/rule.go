package rule

import (
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

func GetHero(ctx *ctx.Context, id values.HeroId) (*rulemodel.RowHero, bool) {
	return rule.MustGetReader(ctx).RowHero.GetRowHeroById(id)
}

func GetHeroSkillByLevel(ctx *ctx.Context, heroId values.HeroId, level values.Level) []values.HeroSkillId {
	// 技能已单独配置
	// data := rule.MustGetReader(ctx).RowHero.HeroLvUp()
	// list, ok := data[heroId]
	// if !ok {
	//	return nil
	// }
	skills := make([]values.HeroSkillId, 0)
	// for _, item := range list {
	//	if item.Id <= level {
	//		skills = append(skills, item.SkillId()...)
	//	}
	// }
	return skills
}

func GetMaxEquipStar(ctx *ctx.Context) values.Level {
	return rule.MustGetReader(ctx).EquipStar.MaxEquipStar()
}

func GetUpgradeEquipStarCost(ctx *ctx.Context, star values.Level) (*rulemodel.EquipStar, bool) {
	item, ok := rule.MustGetReader(ctx).EquipStar.GetEquipStarById(star)
	return item, ok
}

func GetEquipSlot(ctx *ctx.Context, slot values.EquipSlot) (*rulemodel.EquipSlot, bool) {
	return rule.MustGetReader(ctx).EquipSlot.GetEquipSlotById(slot)
}

func GetEquipByItemId(ctx *ctx.Context, id values.ItemId) (*rulemodel.Equip, bool) {
	return rule.MustGetReader(ctx).Equip.GetEquipById(id)
}

func GetItemById(ctx *ctx.Context, id values.ItemId) (*rulemodel.Item, bool) {
	return rule.MustGetReader(ctx).Item.GetItemById(id)
}

func GetEquipLevelLimit(ctx *ctx.Context, q values.Integer) values.Level {
	item, ok := rule.MustGetReader(ctx).EquipQuality.GetEquipQualityById(q)
	if !ok {
		return 0
	}
	return item.HeroLv
}

func GetRoleLvConfigByLv(ctx *ctx.Context, lv values.Level) (*rulemodel.RoleLv, bool) {
	item, ok := rule.MustGetReader(ctx).RoleLv.GetRoleLvById(lv)
	return item, ok
}

func GetAttrTransConfigById(ctx *ctx.Context, id values.AttrId) []rulemodel.AttrTrans {
	return rule.MustGetReader(ctx).AttrTrans.GetAttrTransByAttrId(id)
}

// func GetEquipMelt(ctx *ctx.Context, lv values.Level) (*rulemodel.EquipMelting, bool) {
// 	return rule.MustGetReader(ctx).EquipMelting.GetEquipMeltingById(lv)
// }

func GetEquipRefine(ctx *ctx.Context, slot values.EquipSlot, level values.Level) (*rulemodel.EquipRefine, bool) {
	return rule.MustGetReader(ctx).EquipRefine.GetEquipRefine(slot, level)
}

func GetMaxEquipMeltLevel(ctx *ctx.Context, slot values.EquipSlot) values.Level {
	return rule.MustGetReader(ctx).EquipRefine.GetMaxEquipMeltLevel(slot)
}

func GetAttrById(ctx *ctx.Context, id values.AttrId) (*rulemodel.Attr, bool) {
	return rule.MustGetReader(ctx).Attr.GetAttrById(id)
}

func GetSkillById(ctx *ctx.Context, id values.HeroSkillId) (*rulemodel.Skill, bool) {
	return rule.MustGetReader(ctx).Skill.GetSkillById(id)
}

func GetMaxSkillId(ctx *ctx.Context, id values.HeroSkillId) values.HeroSkillId {
	return rule.MustGetReader(ctx).Skill.GetMaxSkillId(id)
}

func GetEnchant(ctx *ctx.Context, id values.ItemId) (*rulemodel.Enchantments, bool) {
	return rule.MustGetReader(ctx).Enchantments.GetEnchantmentsById(id)
}

func GetEquipEntryByGroup(ctx *ctx.Context, group values.Integer) []rulemodel.EquipEntry {
	return rule.MustGetReader(ctx).Equip.GetEquipEntry(group)
}

func GetEquipEntryById(ctx *ctx.Context, id values.Integer) (*rulemodel.EquipEntry, bool) {
	cfg, ok := rule.MustGetReader(ctx).EquipEntry.GetEquipEntryById(id)
	return cfg, ok
}

func IsInitTalent(ctx *ctx.Context, id values.Integer) bool {
	v := rule.MustGetReader(ctx).RowHero.InitTalent(id)
	return v == id
}

type Biography struct {
	Id        values.Integer
	TaskType  models.TaskType
	NeedCount values.Integer
}

func GetHeroBiography(ctx *ctx.Context, heroId []values.HeroId, notContainType models.TaskType) ([]values.Integer, []*Biography, []models.TaskType) {
	list := make([]rulemodel.Biography, 0)
	for _, id := range heroId {
		temp := rule.MustGetReader(ctx).Biography.GetBiographyByHero(id)
		list = append(list, temp...)
	}

	ids := make([]values.Integer, 0) // 默认解锁且有奖励的传记id
	ret := make([]*Biography, 0)
	types := make([]models.TaskType, 0)
	typeMap := map[models.TaskType]struct{}{notContainType: {}}
	for _, biography := range list {
		unlockCondition := biography.UnlockCondition
		if len(unlockCondition) == 0 && len(biography.Reward) > 0 {
			ids = append(ids, biography.Id)
		}
		if len(unlockCondition) == 3 {
			ret = append(ret, &Biography{
				Id:        biography.Id,
				TaskType:  models.TaskType(unlockCondition[0]),
				NeedCount: unlockCondition[2],
			})
			typ := models.TaskType(unlockCondition[0])
			if _, ok := typeMap[typ]; !ok {
				types = append(types, typ)
				typeMap[typ] = struct{}{}
			}
		}
	}
	return ids, ret, types
}

func GetHeroBiographyById(ctx *ctx.Context, id values.Integer) (*rulemodel.Biography, bool) {
	cfg, ok := rule.MustGetReader(ctx).Biography.GetBiographyById(id)
	if !ok {
		return nil, false
	}
	return cfg, true
}

func GetBiographyByTaskType(ctx *ctx.Context, typ models.TaskType) []*Biography {
	list := rule.MustGetReader(ctx).Biography.GetBiographyByTaskType(typ)
	ret := make([]*Biography, 0)
	for _, biography := range list {
		if len(biography.Reward) > 0 {
			var (
				taskType  models.TaskType
				needCount values.Integer
			)
			// 包含无解锁条件和有解锁的条件的
			if len(biography.UnlockCondition) == 3 {
				taskType = models.TaskType(biography.UnlockCondition[0])
				needCount = biography.UnlockCondition[2]
			}
			ret = append(ret, &Biography{
				Id:        biography.Id,
				TaskType:  taskType,
				NeedCount: needCount,
			})
		}
	}
	return ret
}

func GetMaxSoulContract(ctx *ctx.Context, heroId values.HeroId) (rulemodel.MaxSoulContract, bool) {
	return rule.MustGetReader(ctx).SoulContract.GetMaxSoulContractByHero(heroId)
}

func GetSoulContract(ctx *ctx.Context, heroId values.HeroId, rank, level values.Integer) (rulemodel.SoulContract, bool) {
	list := rule.MustGetReader(ctx).SoulContract.GetSoulContractByHero(heroId)
	var (
		item rulemodel.SoulContract
		find bool
	)
	for _, contract := range list {
		if contract.Rank == rank && contract.Level == level {
			item = contract
			find = true
			break
		}
	}
	return item, find
}

func GetNextSoulContract(ctx *ctx.Context, heroId values.HeroId, rank, level values.Integer) (rulemodel.SoulContract, bool) {
	list := rule.MustGetReader(ctx).SoulContract.GetSoulContractByHero(heroId)
	var (
		item rulemodel.SoulContract
		find bool
	)
	// 先找level+1的
	level++
	for _, contract := range list {
		if contract.Rank == rank && contract.Level == level {
			item = contract
			find = true
			break
		}
	}
	if !find {
		rank++
		level = 1
		for _, contract := range list {
			if contract.Rank == rank && contract.Level == level {
				item = contract
				find = true
				break
			}
		}
	}
	return item, find
}

func GetEquipStarLimit(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("EquipStarLimit")
	return v
}

func GetEquipEnchantingLimit(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("EquipEnchantingLimit")
	return v
}

func GetBaceCritDam(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetString("BaceCritDam")
	dam, _ := strconv.ParseFloat(v, 64)
	return values.Integer(dam * 10000)
}

func GetBuffById(ctx *ctx.Context, id values.HeroBuffId) (*rulemodel.Buff, bool) {
	buff, ok := rule.MustGetReader(ctx).Buff.GetBuffById(id)
	return buff, ok
}

func GetEquipResonanceByHero(ctx *ctx.Context, id values.HeroId) map[values.Level]rulemodel.EquipResonance {
	return rule.MustGetReader(ctx).EquipResonance.GetEquipResonanceByHero(id)
}

func GetEquipResonance(ctx *ctx.Context, id values.Integer) (*rulemodel.EquipResonance, bool) {
	er, ok := rule.MustGetReader(ctx).EquipResonance.GetEquipResonanceById(id)
	return er, ok
}

func GetFashionActivate(ctx *ctx.Context, id values.ItemId) (*rulemodel.FashionActivate, bool) {
	return rule.MustGetReader(ctx).FashionActivate.GetFashionActivateById(id)
}

func GetFashion(ctx *ctx.Context, id values.FashionId) (*rulemodel.Fashion, bool) {
	return rule.MustGetReader(ctx).Fashion.GetFashionById(id)
}

func GetDefaultFashion(ctx *ctx.Context, id values.HeroId) values.FashionId {
	return rule.MustGetReader(ctx).Fashion.GetDefaultFashion(id)
}

func GetMailConfigTextId(ctx *ctx.Context, id values.Integer) (*rulemodel.Mail, bool) {
	item, ok := rule.MustGetReader(ctx).Mail.GetMailById(id)
	return item, ok
}

func GetUpgradeEquipAllCostItemId(ctx *ctx.Context) []values.ItemId {
	return rule.MustGetReader(ctx).EquipStar.GetAllCostItemId()
}

func GetLanguageBackend(ctx *ctx.Context, id values.Integer) (*rulemodel.LanguageBackend, bool) {
	return rule.MustGetReader(ctx).LanguageBackend.GetLanguageBackendById(id)
}

func GetBlessById(ctx *ctx.Context, id values.Integer) (*rulemodel.GuildBlessing, bool) {
	return rule.MustGetReader(ctx).GuildBlessing.GetGuildBlessingById(id)
}

func GetAllBless(ctx *ctx.Context) []rulemodel.GuildBlessing {
	return rule.MustGetReader(ctx).GuildBlessing.List()
}

func GetSystemById(ctx *ctx.Context, systemId values.SystemId) (*rulemodel.System, bool) {
	return rule.MustGetReader(ctx).System.GetSystemById(values.Integer(systemId))
}

func GetMaxBlessPage(ctx *ctx.Context) values.Integer {
	return rule.MustGetReader(ctx).GuildBlessing.GetMaxGuildBlessingPage()
}
