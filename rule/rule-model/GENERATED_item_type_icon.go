// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ItemTypeIcon struct {
	Id       values.Integer `mapstructure:"id" json:"id"`
	ItemType values.Integer `mapstructure:"item_type" json:"item_type"`
	IconName string         `mapstructure:"icon_name" json:"icon_name"`
}

// parse func
func ParseItemTypeIcon(data *Data) {
	if err := data.UnmarshalKey("item_type_icon", &h.itemTypeIcon); err != nil {
		panic(errors.New("parse table ItemTypeIcon err:\n" + err.Error()))
	}
	for i, el := range h.itemTypeIcon {
		h.itemTypeIconMap[el.Id] = i
	}
}

func (i *ItemTypeIcon) Len() int {
	return len(h.itemTypeIcon)
}

func (i *ItemTypeIcon) List() []ItemTypeIcon {
	return h.itemTypeIcon
}

func (i *ItemTypeIcon) GetItemTypeIconById(id values.Integer) (*ItemTypeIcon, bool) {
	index, ok := h.itemTypeIconMap[id]
	if !ok {
		return nil, false
	}
	return &h.itemTypeIcon[index], true
}
