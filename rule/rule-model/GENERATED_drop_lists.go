// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type DropLists struct {
	Id values.Integer `mapstructure:"id" json:"id"`
}

// parse func
func ParseDropLists(data *Data) {
	if err := data.UnmarshalKey("drop_lists", &h.dropLists); err != nil {
		panic(errors.New("parse table DropLists err:\n" + err.Error()))
	}
	for i, el := range h.dropLists {
		h.dropListsMap[el.Id] = i
	}
}

func (i *DropLists) Len() int {
	return len(h.dropLists)
}

func (i *DropLists) List() []DropLists {
	return h.dropLists
}

func (i *DropLists) GetDropListsById(id values.Integer) (*DropLists, bool) {
	index, ok := h.dropListsMap[id]
	if !ok {
		return nil, false
	}
	return &h.dropLists[index], true
}