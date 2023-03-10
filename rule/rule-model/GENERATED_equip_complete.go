// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type EquipComplete struct {
	Id         values.Integer     `mapstructure:"id" json:"id"`
	TextId     string             `mapstructure:"text_id" json:"text_id"`
	Complete   []values.Integer   `mapstructure:"complete" json:"complete"`
	SkillTwo   [][]values.Integer `mapstructure:"skill_two" json:"skill_two"`
	SkillText2 string             `mapstructure:"skill_text2" json:"skill_text2"`
	SkillFour  [][]values.Integer `mapstructure:"skill_four" json:"skill_four"`
	SkillText4 string             `mapstructure:"skill_text4" json:"skill_text4"`
	SkillSix   [][]values.Integer `mapstructure:"skill_six" json:"skill_six"`
	SkillText6 string             `mapstructure:"skill_text6" json:"skill_text6"`
}

// parse func
func ParseEquipComplete(data *Data) {
	if err := data.UnmarshalKey("equip_complete", &h.equipComplete); err != nil {
		panic(errors.New("parse table EquipComplete err:\n" + err.Error()))
	}
	for i, el := range h.equipComplete {
		h.equipCompleteMap[el.Id] = i
	}
}

func (i *EquipComplete) Len() int {
	return len(h.equipComplete)
}

func (i *EquipComplete) List() []EquipComplete {
	return h.equipComplete
}

func (i *EquipComplete) GetEquipCompleteById(id values.Integer) (*EquipComplete, bool) {
	index, ok := h.equipCompleteMap[id]
	if !ok {
		return nil, false
	}
	return &h.equipComplete[index], true
}
