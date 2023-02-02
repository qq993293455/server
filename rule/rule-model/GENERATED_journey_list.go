// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type JourneyList struct {
	Id        values.Integer                    `mapstructure:"id" json:"id"`
	SystemId  values.Integer                    `mapstructure:"system_id" json:"system_id"`
	Banner    string                            `mapstructure:"banner" json:"banner"`
	Language1 string                            `mapstructure:"language1" json:"language1"`
	AddItem   map[values.Integer]values.Integer `mapstructure:"add_item" json:"add_item"`
}

// parse func
func ParseJourneyList(data *Data) {
	if err := data.UnmarshalKey("journey_list", &h.journeyList); err != nil {
		panic(errors.New("parse table JourneyList err:\n" + err.Error()))
	}
	for i, el := range h.journeyList {
		h.journeyListMap[el.Id] = i
	}
}

func (i *JourneyList) Len() int {
	return len(h.journeyList)
}

func (i *JourneyList) List() []JourneyList {
	return h.journeyList
}

func (i *JourneyList) GetJourneyListById(id values.Integer) (*JourneyList, bool) {
	index, ok := h.journeyListMap[id]
	if !ok {
		return nil, false
	}
	return &h.journeyList[index], true
}
