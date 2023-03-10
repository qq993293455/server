// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type MapSelect struct {
	Id                  values.Integer                    `mapstructure:"id" json:"id"`
	Scene               values.Integer                    `mapstructure:"scene" json:"scene"`
	UseType             values.Integer                    `mapstructure:"use_type" json:"use_type"`
	Belong              values.Integer                    `mapstructure:"belong" json:"belong"`
	PresetMap           string                            `mapstructure:"preset_map" json:"preset_map"`
	Order               values.Integer                    `mapstructure:"order" json:"order"`
	Desc                string                            `mapstructure:"desc" json:"desc"`
	UnlockCondition     [][]values.Integer                `mapstructure:"unlock_condition" json:"unlock_condition"`
	UnlockConditionDesc []string                          `mapstructure:"unlock_condition_desc" json:"unlock_condition_desc"`
	Banner              values.Integer                    `mapstructure:"banner" json:"banner"`
	Production          []values.Integer                  `mapstructure:"production" json:"production"`
	ExploreDegree       []values.Integer                  `mapstructure:"explore_degree" json:"explore_degree"`
	Reward              map[values.Integer]values.Integer `mapstructure:"reward" json:"reward"`
}

// parse func
func ParseMapSelect(data *Data) {
	if err := data.UnmarshalKey("map_select", &h.mapSelect); err != nil {
		panic(errors.New("parse table MapSelect err:\n" + err.Error()))
	}
	for i, el := range h.mapSelect {
		h.mapSelectMap[el.Id] = i
	}
}

func (i *MapSelect) Len() int {
	return len(h.mapSelect)
}

func (i *MapSelect) List() []MapSelect {
	return h.mapSelect
}

func (i *MapSelect) GetMapSelectById(id values.Integer) (*MapSelect, bool) {
	index, ok := h.mapSelectMap[id]
	if !ok {
		return nil, false
	}
	return &h.mapSelect[index], true
}
