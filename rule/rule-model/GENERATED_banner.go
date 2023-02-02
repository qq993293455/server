// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type Banner struct {
	Id          values.Integer   `mapstructure:"id" json:"id"`
	Img         string           `mapstructure:"img" json:"img"`
	InWorldInfo []values.Integer `mapstructure:"in_world_info" json:"in_world_info"`
}

// parse func
func ParseBanner(data *Data) {
	if err := data.UnmarshalKey("banner", &h.banner); err != nil {
		panic(errors.New("parse table Banner err:\n" + err.Error()))
	}
	for i, el := range h.banner {
		h.bannerMap[el.Id] = i
	}
}

func (i *Banner) Len() int {
	return len(h.banner)
}

func (i *Banner) List() []Banner {
	return h.banner
}

func (i *Banner) GetBannerById(id values.Integer) (*Banner, bool) {
	index, ok := h.bannerMap[id]
	if !ok {
		return nil, false
	}
	return &h.banner[index], true
}