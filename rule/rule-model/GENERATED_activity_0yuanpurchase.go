// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type Activity0yuanpurchase struct {
	Id                values.Integer   `mapstructure:"id" json:"id"`
	TypId             values.Integer   `mapstructure:"typ_id" json:"typ_id"`
	Day               values.Integer   `mapstructure:"day" json:"day"`
	RewardImmediately []values.Integer `mapstructure:"reward_immediately" json:"reward_immediately"`
	RewardTomo        []values.Integer `mapstructure:"reward_tomo" json:"reward_tomo"`
	CostItem          []values.Integer `mapstructure:"cost_item" json:"cost_item"`
	PictureId         []string         `mapstructure:"picture_id" json:"picture_id"`
	MailId            values.Integer   `mapstructure:"mail_id" json:"mail_id"`
}

// parse func
func ParseActivity0yuanpurchase(data *Data) {
	if err := data.UnmarshalKey("activity_0yuanpurchase", &h.activity0yuanpurchase); err != nil {
		panic(errors.New("parse table Activity0yuanpurchase err:\n" + err.Error()))
	}
	for i, el := range h.activity0yuanpurchase {
		h.activity0yuanpurchaseMap[el.Id] = i
	}
}

func (i *Activity0yuanpurchase) Len() int {
	return len(h.activity0yuanpurchase)
}

func (i *Activity0yuanpurchase) List() []Activity0yuanpurchase {
	return h.activity0yuanpurchase
}

func (i *Activity0yuanpurchase) GetActivity0yuanpurchaseById(id values.Integer) (*Activity0yuanpurchase, bool) {
	index, ok := h.activity0yuanpurchaseMap[id]
	if !ok {
		return nil, false
	}
	return &h.activity0yuanpurchase[index], true
}
