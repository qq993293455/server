// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type HeadSculpture struct {
	Id                      values.Integer `mapstructure:"id" json:"id"`
	HeadSculptureType       values.Integer `mapstructure:"head_sculpture_type" json:"head_sculpture_type"`
	HeadName                string         `mapstructure:"head_name" json:"head_name"`
	UnlockConditionDescribe string         `mapstructure:"unlock_condition_describe" json:"unlock_condition_describe"`
	HeroImg                 string         `mapstructure:"hero_img" json:"hero_img"`
}

// parse func
func ParseHeadSculpture(data *Data) {
	if err := data.UnmarshalKey("head_sculpture", &h.headSculpture); err != nil {
		panic(errors.New("parse table HeadSculpture err:\n" + err.Error()))
	}
	for i, el := range h.headSculpture {
		h.headSculptureMap[el.Id] = i
	}
}

func (i *HeadSculpture) Len() int {
	return len(h.headSculpture)
}

func (i *HeadSculpture) List() []HeadSculpture {
	return h.headSculpture
}

func (i *HeadSculpture) GetHeadSculptureById(id values.Integer) (*HeadSculpture, bool) {
	index, ok := h.headSculptureMap[id]
	if !ok {
		return nil, false
	}
	return &h.headSculpture[index], true
}
