// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type AtlasItem struct {
	AtlasId    values.Integer                    `mapstructure:"atlas_id" json:"atlas_id"`
	Id         values.Integer                    `mapstructure:"id" json:"id"`
	Attrrate   map[values.Integer]values.Integer `mapstructure:"attrrate" json:"attrrate"`
	Attrnum    map[values.Integer]values.Integer `mapstructure:"attrnum" json:"attrnum"`
	ItemReward map[values.Integer]values.Integer `mapstructure:"item_reward" json:"item_reward"`
}

// parse func
func ParseAtlasItem(data *Data) {
	if err := data.UnmarshalKey("atlas_item", &h.atlasItem); err != nil {
		panic(errors.New("parse table AtlasItem err:\n" + err.Error()))
	}
	for i, el := range h.atlasItem {
		if _, ok := h.atlasItemMap[el.AtlasId]; !ok {
			h.atlasItemMap[el.AtlasId] = map[values.Integer]int{el.Id: i}
		} else {
			h.atlasItemMap[el.AtlasId][el.Id] = i
		}
	}
}

func (i *AtlasItem) Len() int {
	return len(h.atlasItem)
}

func (i *AtlasItem) List() []AtlasItem {
	return h.atlasItem
}

func (i *AtlasItem) GetAtlasItemById(parentId, id values.Integer) (*AtlasItem, bool) {
	item, ok := h.atlasItemMap[parentId]
	if !ok {
		return nil, false
	}
	index, ok := item[id]
	if !ok {
		return nil, false
	}
	return &h.atlasItem[index], true
}
