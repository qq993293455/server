// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type JourneyChest struct {
	Id           values.Integer                    `mapstructure:"id" json:"id"`
	ChestType    values.Integer                    `mapstructure:"chest_type" json:"chest_type"`
	MinimumLevel values.Integer                    `mapstructure:"minimum_level" json:"minimum_level"`
	HighestLevel values.Integer                    `mapstructure:"highest_level" json:"highest_level"`
	FixedReward  map[values.Integer]values.Integer `mapstructure:"fixed_reward" json:"fixed_reward"`
	BoxId        values.Integer                    `mapstructure:"box_id" json:"box_id"`
}

// parse func
func ParseJourneyChest(data *Data) {
	if err := data.UnmarshalKey("journey_chest", &h.journeyChest); err != nil {
		panic(errors.New("parse table JourneyChest err:\n" + err.Error()))
	}
	for i, el := range h.journeyChest {
		h.journeyChestMap[el.Id] = i
	}
}

func (i *JourneyChest) Len() int {
	return len(h.journeyChest)
}

func (i *JourneyChest) List() []JourneyChest {
	return h.journeyChest
}

func (i *JourneyChest) GetJourneyChestById(id values.Integer) (*JourneyChest, bool) {
	index, ok := h.journeyChestMap[id]
	if !ok {
		return nil, false
	}
	return &h.journeyChest[index], true
}
