// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type EquipType struct {
	Id   values.Integer `mapstructure:"id" json:"id"`
	Name string         `mapstructure:"name" json:"name"`
}

// parse func
func ParseEquipType(data *Data) {
	if err := data.UnmarshalKey("equip_type", &h.equipType); err != nil {
		panic(errors.New("parse table EquipType err:\n" + err.Error()))
	}
	for i, el := range h.equipType {
		h.equipTypeMap[el.Id] = i
	}
}

func (i *EquipType) Len() int {
	return len(h.equipType)
}

func (i *EquipType) List() []EquipType {
	return h.equipType
}

func (i *EquipType) GetEquipTypeById(id values.Integer) (*EquipType, bool) {
	index, ok := h.equipTypeMap[id]
	if !ok {
		return nil, false
	}
	return &h.equipType[index], true
}
