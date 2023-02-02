package event

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

// 获得常规天赋点
type GainTalentCommonPoint struct {
	Num values.Integer
}

// 获得职业天赋点
type GainTalentParticularPoint struct {
	ConfigId values.Integer
	Num      values.Integer
}

// 天赋改变
type TalentChange struct {
	ConfigId values.Integer
	Attr     *models.TalentAdvance
}

// 技能改变
type SkillChange struct {
	ConfigId values.Integer
	Skills   *models.SkillAdvance
}

// 装备对技能等级加成
type EquipBonusTalent struct {
	HeroId values.HeroId // OriginId
	Data   map[values.HeroSkillId]values.Level
}
