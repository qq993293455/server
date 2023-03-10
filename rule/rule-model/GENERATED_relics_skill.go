// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type RelicsSkill struct {
	Id         values.Integer   `mapstructure:"id" json:"id"`
	Text       string           `mapstructure:"text" json:"text"`
	Typ        []values.Integer `mapstructure:"typ" json:"typ"`
	Value      values.Integer   `mapstructure:"value" json:"value"`
	AddTyp     values.Integer   `mapstructure:"add__typ" json:"add__typ"`
	StarsValue []values.Integer `mapstructure:"stars_value" json:"stars_value"`
}

// parse func
func ParseRelicsSkill(data *Data) {
	if err := data.UnmarshalKey("relics_skill", &h.relicsSkill); err != nil {
		panic(errors.New("parse table RelicsSkill err:\n" + err.Error()))
	}
	for i, el := range h.relicsSkill {
		h.relicsSkillMap[el.Id] = i
	}
}

func (i *RelicsSkill) Len() int {
	return len(h.relicsSkill)
}

func (i *RelicsSkill) List() []RelicsSkill {
	return h.relicsSkill
}

func (i *RelicsSkill) GetRelicsSkillById(id values.Integer) (*RelicsSkill, bool) {
	index, ok := h.relicsSkillMap[id]
	if !ok {
		return nil, false
	}
	return &h.relicsSkill[index], true
}
