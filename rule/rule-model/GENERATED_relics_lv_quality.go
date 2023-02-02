// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type RelicsLvQuality struct {
	RelicsLvId values.Integer                    `mapstructure:"relics_lv_id" json:"relics_lv_id"`
	Id         values.Integer                    `mapstructure:"id" json:"id"`
	UnlockLv   values.Integer                    `mapstructure:"unlock_lv" json:"unlock_lv"`
	LvCost     map[values.Integer]values.Integer `mapstructure:"lv_cost" json:"lv_cost"`
}

// parse func
func ParseRelicsLvQuality(data *Data) {
	if err := data.UnmarshalKey("relics_lv_quality", &h.relicsLvQuality); err != nil {
		panic(errors.New("parse table RelicsLvQuality err:\n" + err.Error()))
	}
	for i, el := range h.relicsLvQuality {
		if _, ok := h.relicsLvQualityMap[el.RelicsLvId]; !ok {
			h.relicsLvQualityMap[el.RelicsLvId] = map[values.Integer]int{el.Id: i}
		} else {
			h.relicsLvQualityMap[el.RelicsLvId][el.Id] = i
		}
	}
}

func (i *RelicsLvQuality) Len() int {
	return len(h.relicsLvQuality)
}

func (i *RelicsLvQuality) List() []RelicsLvQuality {
	return h.relicsLvQuality
}

func (i *RelicsLvQuality) GetRelicsLvQualityById(parentId, id values.Integer) (*RelicsLvQuality, bool) {
	item, ok := h.relicsLvQualityMap[parentId]
	if !ok {
		return nil, false
	}
	index, ok := item[id]
	if !ok {
		return nil, false
	}
	return &h.relicsLvQuality[index], true
}