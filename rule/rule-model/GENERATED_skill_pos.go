// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type SkillPos struct {
	Id       values.Integer   `mapstructure:"id" json:"id"`
	SkillPos []values.Integer `mapstructure:"skill_pos" json:"skill_pos"`
}

// parse func
func ParseSkillPos(data *Data) {
	if err := data.UnmarshalKey("skill_pos", &h.skillPos); err != nil {
		panic(errors.New("parse table SkillPos err:\n" + err.Error()))
	}
	for i, el := range h.skillPos {
		h.skillPosMap[el.Id] = i
	}
}

func (i *SkillPos) Len() int {
	return len(h.skillPos)
}

func (i *SkillPos) List() []SkillPos {
	return h.skillPos
}

func (i *SkillPos) GetSkillPosById(id values.Integer) (*SkillPos, bool) {
	index, ok := h.skillPosMap[id]
	if !ok {
		return nil, false
	}
	return &h.skillPos[index], true
}