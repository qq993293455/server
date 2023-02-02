package hero

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

// 英雄属性
const (
	levelAttr        = 1  // 等级
	equipAttr        = 2  // 装备
	equipSetAttr     = 3  // 装备套装
	talentAttr       = 4  // 天赋
	relicsAttr       = 5  // 遗物
	atlasAttr        = 6  // 图鉴
	titleAttr        = 7  // 头衔
	taskAttr         = 8  // 任务
	soulContractAttr = 9  // 魂契
	fashionAttr      = 10 // 时装
	relicsFuncAttr   = 11 // 遗物功能属性
	guildBlessAttr   = 12 // 遗物功能属性
	talentSkillAttr  = 13 // 天赋技能

	// 最终的属性里只有具体的值，没有百分比
	primaryAttr   = 100 // 最终一级属性
	secondaryAttr = 101 // 最终二级属性

)

// 英雄技能
const (
	selfSkillType   = 1 // 自身技能
	talentSkillType = 2 // 天赋技能
	roleSkillType   = 3 // 来自主角的技能
)

// 英雄buff
const (
	equipBuffType    = 1 // 来自装备的buff
	equipSetBuffType = 2 // 来自装备套装的buff
)

// 属性转换类型
const (
	ttPercent = 1 // 百分比
	ttFixed   = 2 // 固定值
)

type equipStarToValue struct {
	Range []values.Integer
	Value []values.Integer
}

type WeightItem struct {
	Quality values.Quality
	Weight  values.Integer
}

func AttrType2CombatValueType(ac values.Integer) values.Integer {
	switch ac {
	case levelAttr:
		return values.Integer(models.CombatValueType_CVTLevel)
	case equipAttr:
		return values.Integer(models.CombatValueType_CVTEquip)
	case equipSetAttr:
		return values.Integer(models.CombatValueType_CVTEquip)
	case talentAttr, talentSkillAttr:
		return values.Integer(models.CombatValueType_CVTSkill)
	case relicsAttr:
		return values.Integer(models.CombatValueType_CVTRelics)
	case atlasAttr:
		return values.Integer(models.CombatValueType_CVTAtlas)
	case titleAttr:
		return values.Integer(models.CombatValueType_CVTTitle)
	case taskAttr:
		return values.Integer(models.CombatValueType_CVTTask)
	case soulContractAttr:
		return values.Integer(models.CombatValueType_CVTSoulContract)
	case fashionAttr:
		return values.Integer(models.CombatValueType_CVTFashion)
	case relicsFuncAttr:
		return values.Integer(models.CombatValueType_CVTRelics)
	case guildBlessAttr:
		return values.Integer(models.CombatValueType_CVTGuildBless)
	default:
		return 0
	}
}

func SkillType2CombatValueType(st values.Integer) values.Integer {
	switch st {
	case selfSkillType:
		return values.Integer(models.CombatValueType_CVTSkill)
	case talentSkillType:
		return values.Integer(models.CombatValueType_CVTSkill)
	case roleSkillType:
		return values.Integer(models.CombatValueType_CVTSkill)
	default:
		return 0
	}
}

func BonusType2AttrType(bt models.AttrBonusType) values.Integer {
	switch bt {
	case models.AttrBonusType_TypeAtlas:
		return atlasAttr
	case models.AttrBonusType_TypeRelics:
		return relicsAttr
	case models.AttrBonusType_TypeLoopTask:
		return taskAttr
	case models.AttrBonusType_TypeTitle:
		return titleAttr
	case models.AttrBonusType_TypeRelicsFunc:
		return relicsFuncAttr
	default:
		return 0
	}
}

type UpgradeData struct {
	Cost     map[values.ItemId]values.Integer
	Upgraded map[values.EquipSlot]*models.HeroEquipSlot
	Count    values.Integer
}
