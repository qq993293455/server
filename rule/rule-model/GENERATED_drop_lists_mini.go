// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type DropListsMini struct {
	DropListsId values.Integer `mapstructure:"drop_lists_id" json:"drop_lists_id"`
	Id          values.Integer `mapstructure:"id" json:"id"`
	DropId      values.Integer `mapstructure:"drop_id" json:"drop_id"`
	DropProb    values.Integer `mapstructure:"drop_prob" json:"drop_prob"`
}

// parse func
func ParseDropListsMini(data *Data) {
	if err := data.UnmarshalKey("drop_lists_mini", &h.dropListsMini); err != nil {
		panic(errors.New("parse table DropListsMini err:\n" + err.Error()))
	}
	for i, el := range h.dropListsMini {
		if _, ok := h.dropListsMiniMap[el.DropListsId]; !ok {
			h.dropListsMiniMap[el.DropListsId] = map[values.Integer]int{el.Id: i}
		} else {
			h.dropListsMiniMap[el.DropListsId][el.Id] = i
		}
	}
}

func (i *DropListsMini) Len() int {
	return len(h.dropListsMini)
}

func (i *DropListsMini) List() []DropListsMini {
	return h.dropListsMini
}

func (i *DropListsMini) GetDropListsMiniById(parentId, id values.Integer) (*DropListsMini, bool) {
	item, ok := h.dropListsMiniMap[parentId]
	if !ok {
		return nil, false
	}
	index, ok := item[id]
	if !ok {
		return nil, false
	}
	return &h.dropListsMini[index], true
}
