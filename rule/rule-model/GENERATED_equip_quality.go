// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type EquipQuality struct {
	Id     values.Integer `mapstructure:"id" json:"id"`
	HeroLv values.Integer `mapstructure:"hero_lv" json:"hero_lv"`
}

// parse func
func ParseEquipQuality(data *Data) {
	if err := data.UnmarshalKey("equip_quality", &h.equipQuality); err != nil {
		panic(errors.New("parse table EquipQuality err:\n" + err.Error()))
	}
	for i, el := range h.equipQuality {
		h.equipQualityMap[el.Id] = i
	}
}

func (i *EquipQuality) Len() int {
	return len(h.equipQuality)
}

func (i *EquipQuality) List() []EquipQuality {
	return h.equipQuality
}

func (i *EquipQuality) GetEquipQualityById(id values.Integer) (*EquipQuality, bool) {
	index, ok := h.equipQualityMap[id]
	if !ok {
		return nil, false
	}
	return &h.equipQuality[index], true
}
