// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type Atlas struct {
	Id values.Integer `mapstructure:"id" json:"id"`
}

// parse func
func ParseAtlas(data *Data) {
	if err := data.UnmarshalKey("atlas", &h.atlas); err != nil {
		panic(errors.New("parse table Atlas err:\n" + err.Error()))
	}
	for i, el := range h.atlas {
		h.atlasMap[el.Id] = i
	}
}

func (i *Atlas) Len() int {
	return len(h.atlas)
}

func (i *Atlas) List() []Atlas {
	return h.atlas
}

func (i *Atlas) GetAtlasById(id values.Integer) (*Atlas, bool) {
	index, ok := h.atlasMap[id]
	if !ok {
		return nil, false
	}
	return &h.atlas[index], true
}