// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ActivityGrowthfund struct {
	Id                values.Integer   `mapstructure:"id" json:"id"`
	TypId             values.Integer   `mapstructure:"typ_id" json:"typ_id"`
	Level             values.Integer   `mapstructure:"level" json:"level"`
	ActivityReward    []values.Integer `mapstructure:"activity_reward" json:"activity_reward"`
	ActivityPayReward []values.Integer `mapstructure:"activity_pay_reward" json:"activity_pay_reward"`
}

// parse func
func ParseActivityGrowthfund(data *Data) {
	if err := data.UnmarshalKey("activity_growthfund", &h.activityGrowthfund); err != nil {
		panic(errors.New("parse table ActivityGrowthfund err:\n" + err.Error()))
	}
	for i, el := range h.activityGrowthfund {
		h.activityGrowthfundMap[el.Id] = i
	}
}

func (i *ActivityGrowthfund) Len() int {
	return len(h.activityGrowthfund)
}

func (i *ActivityGrowthfund) List() []ActivityGrowthfund {
	return h.activityGrowthfund
}

func (i *ActivityGrowthfund) GetActivityGrowthfundById(id values.Integer) (*ActivityGrowthfund, bool) {
	index, ok := h.activityGrowthfundMap[id]
	if !ok {
		return nil, false
	}
	return &h.activityGrowthfund[index], true
}