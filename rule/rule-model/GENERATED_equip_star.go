// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type EquipStar struct {
	Id       values.Integer                    `mapstructure:"id" json:"id"`
	Cost     map[values.Integer]values.Integer `mapstructure:"cost" json:"cost"`
	HeroLv   values.Integer                    `mapstructure:"hero_lv" json:"hero_lv"`
	Advanced values.Integer                    `mapstructure:"advanced" json:"advanced"`
}

// parse func
func ParseEquipStar(data *Data) {
	if err := data.UnmarshalKey("equip_star", &h.equipStar); err != nil {
		panic(errors.New("parse table EquipStar err:\n" + err.Error()))
	}
	for i, el := range h.equipStar {
		h.equipStarMap[el.Id] = i
	}
}

func (i *EquipStar) Len() int {
	return len(h.equipStar)
}

func (i *EquipStar) List() []EquipStar {
	return h.equipStar
}

func (i *EquipStar) GetEquipStarById(id values.Integer) (*EquipStar, bool) {
	index, ok := h.equipStarMap[id]
	if !ok {
		return nil, false
	}
	return &h.equipStar[index], true
}