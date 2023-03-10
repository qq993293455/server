// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type FightingCapacity struct {
	Id           values.Integer `mapstructure:"id" json:"id"`
	AdvancedType values.Integer `mapstructure:"advanced_type" json:"advanced_type"`
	CultureName  string         `mapstructure:"culture_name" json:"culture_name"`
}

// parse func
func ParseFightingCapacity(data *Data) {
	if err := data.UnmarshalKey("fighting_capacity", &h.fightingCapacity); err != nil {
		panic(errors.New("parse table FightingCapacity err:\n" + err.Error()))
	}
	for i, el := range h.fightingCapacity {
		h.fightingCapacityMap[el.Id] = i
	}
}

func (i *FightingCapacity) Len() int {
	return len(h.fightingCapacity)
}

func (i *FightingCapacity) List() []FightingCapacity {
	return h.fightingCapacity
}

func (i *FightingCapacity) GetFightingCapacityById(id values.Integer) (*FightingCapacity, bool) {
	index, ok := h.fightingCapacityMap[id]
	if !ok {
		return nil, false
	}
	return &h.fightingCapacity[index], true
}
