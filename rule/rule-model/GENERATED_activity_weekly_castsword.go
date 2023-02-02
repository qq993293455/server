// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ActivityWeeklyCastsword struct {
	Id           values.Integer   `mapstructure:"id" json:"id"`
	ActivityId   values.Integer   `mapstructure:"activity_id" json:"activity_id"`
	Sequence     values.Integer   `mapstructure:"sequence" json:"sequence"`
	ActivityItem []values.Integer `mapstructure:"activity_item" json:"activity_item"`
	Reward       []values.Integer `mapstructure:"reward" json:"reward"`
}

// parse func
func ParseActivityWeeklyCastsword(data *Data) {
	if err := data.UnmarshalKey("activity_weekly_castsword", &h.activityWeeklyCastsword); err != nil {
		panic(errors.New("parse table ActivityWeeklyCastsword err:\n" + err.Error()))
	}
	for i, el := range h.activityWeeklyCastsword {
		h.activityWeeklyCastswordMap[el.Id] = i
	}
}

func (i *ActivityWeeklyCastsword) Len() int {
	return len(h.activityWeeklyCastsword)
}

func (i *ActivityWeeklyCastsword) List() []ActivityWeeklyCastsword {
	return h.activityWeeklyCastsword
}

func (i *ActivityWeeklyCastsword) GetActivityWeeklyCastswordById(id values.Integer) (*ActivityWeeklyCastsword, bool) {
	index, ok := h.activityWeeklyCastswordMap[id]
	if !ok {
		return nil, false
	}
	return &h.activityWeeklyCastsword[index], true
}
