package util

import (
	"coin-server/common/values"
	rulemodel "coin-server/rule/rule-model"
)

type EquipTalentBonus struct {
	Talent *rulemodel.Talent
	Level  values.Level
}
