// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ActivityMonthlycard struct {
	Id              values.Integer   `mapstructure:"id" json:"id"`
	Language1       string           `mapstructure:"language1" json:"language1"`
	ActivityId      values.Integer   `mapstructure:"activity_id" json:"activity_id"`
	UnlockCondition []values.Integer `mapstructure:"unlock_condition" json:"unlock_condition"`
	Banner          string           `mapstructure:"banner" json:"banner"`
	Timeliness      values.Integer   `mapstructure:"timeliness" json:"timeliness"`
	Duration        values.Integer   `mapstructure:"duration" json:"duration"`
	ActivateType    values.Integer   `mapstructure:"activate_type" json:"activate_type"`
	PurchaseOptions []values.Integer `mapstructure:"purchase_options" json:"purchase_options"`
	RepeatPurchase  values.Integer   `mapstructure:"repeat_purchase" json:"repeat_purchase"`
	Price           []values.Integer `mapstructure:"price" json:"price"`
	Price2          []values.Integer `mapstructure:"price2" json:"price2"`
	FirstPrice      []values.Integer `mapstructure:"first_price" json:"first_price"`
	ShowPrice       []values.Integer `mapstructure:"show_price" json:"show_price"`
	AveragePrice    values.Integer   `mapstructure:"average_price" json:"average_price"`
	Discount        values.Integer   `mapstructure:"discount" json:"discount"`
	DailyReward     []values.Integer `mapstructure:"daily_reward" json:"daily_reward"`
	PurchaseReward  []values.Integer `mapstructure:"purchase_reward" json:"purchase_reward"`
	RenewalReward   []values.Integer `mapstructure:"renewal_reward" json:"renewal_reward"`
	Value           values.Integer   `mapstructure:"value" json:"value"`
	Mail            values.Integer   `mapstructure:"mail" json:"mail"`
	Mail2           values.Integer   `mapstructure:"mail2" json:"mail2"`
	Mail3           values.Integer   `mapstructure:"mail3" json:"mail3"`
}

// parse func
func ParseActivityMonthlycard(data *Data) {
	if err := data.UnmarshalKey("activity_monthlycard", &h.activityMonthlycard); err != nil {
		panic(errors.New("parse table ActivityMonthlycard err:\n" + err.Error()))
	}
	for i, el := range h.activityMonthlycard {
		h.activityMonthlycardMap[el.Id] = i
	}
}

func (i *ActivityMonthlycard) Len() int {
	return len(h.activityMonthlycard)
}

func (i *ActivityMonthlycard) List() []ActivityMonthlycard {
	return h.activityMonthlycard
}

func (i *ActivityMonthlycard) GetActivityMonthlycardById(id values.Integer) (*ActivityMonthlycard, bool) {
	index, ok := h.activityMonthlycardMap[id]
	if !ok {
		return nil, false
	}
	return &h.activityMonthlycard[index], true
}
