package event

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type ItemUpdate struct {
	RoleId values.RoleId
	Items  []*models.Item
	Incr   []values.Integer
}

type EquipUpdate struct {
	New    bool // 是否新获得的装备
	RoleId values.RoleId
	Items  map[values.ItemId]values.Integer
	Equips []*models.Equipment
}

type EquipDestroyed struct {
	RoleId  values.RoleId
	EquipId []values.EquipId
}

type RelicsUpdate struct {
	IsNewRelics bool
	Relics      []*models.Relics
}

type SkillStoneUpdate struct {
	RoleId values.RoleId
	Stones []*models.SkillStone
	Incr   []values.Integer
}

type TalentRuneUpdate struct {
	RoleId values.RoleId
	Runes  []*models.TalentRune
}

type TalentRuneDestroyed struct {
	RoleId  values.RoleId
	RuneIds []values.RuneId
}
