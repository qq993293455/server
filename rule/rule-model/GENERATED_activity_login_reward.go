// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ActivityLoginReward struct {
	Id             values.Integer   `mapstructure:"id" json:"id"`
	TypId          values.Integer   `mapstructure:"typ_id" json:"typ_id"`
	Day            values.Integer   `mapstructure:"day" json:"day"`
	ActivityReward []values.Integer `mapstructure:"activity_reward" json:"activity_reward"`
}

// parse func
func ParseActivityLoginReward(data *Data) {
	if err := data.UnmarshalKey("activity_login_reward", &h.activityLoginReward); err != nil {
		panic(errors.New("parse table ActivityLoginReward err:\n" + err.Error()))
	}
	for i, el := range h.activityLoginReward {
		h.activityLoginRewardMap[el.Id] = i
	}
}

func (i *ActivityLoginReward) Len() int {
	return len(h.activityLoginReward)
}

func (i *ActivityLoginReward) List() []ActivityLoginReward {
	return h.activityLoginReward
}

func (i *ActivityLoginReward) GetActivityLoginRewardById(id values.Integer) (*ActivityLoginReward, bool) {
	index, ok := h.activityLoginRewardMap[id]
	if !ok {
		return nil, false
	}
	return &h.activityLoginReward[index], true
}
