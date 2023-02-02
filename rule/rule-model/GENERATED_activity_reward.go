// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ActivityReward struct {
	Id                         values.Integer     `mapstructure:"id" json:"id"`
	ActivityTimeType           values.Integer     `mapstructure:"activity_time_type" json:"activity_time_type"`
	ActivityOpenTime           string             `mapstructure:"activity_open_time" json:"activity_open_time"`
	DurationTime               values.Integer     `mapstructure:"duration_time" json:"duration_time"`
	ResetFrequency             values.Integer     `mapstructure:"reset_frequency" json:"reset_frequency"`
	SyttemId                   values.Integer     `mapstructure:"syttem_id" json:"syttem_id"`
	ActivityReward             [][]values.Integer `mapstructure:"activity_reward" json:"activity_reward"`
	ItemPictureId              []string           `mapstructure:"item_picture_id" json:"item_picture_id"`
	SpecialFieldId             []string           `mapstructure:"special_field_id" json:"special_field_id"`
	ActivityDescribeLanguageId []string           `mapstructure:"activity_describe_language_id" json:"activity_describe_language_id"`
	ActivityIcon               string             `mapstructure:"activity_icon" json:"activity_icon"`
	ActivitySort               values.Integer     `mapstructure:"activity_sort" json:"activity_sort"`
}

// parse func
func ParseActivityReward(data *Data) {
	if err := data.UnmarshalKey("activity_reward", &h.activityReward); err != nil {
		panic(errors.New("parse table ActivityReward err:\n" + err.Error()))
	}
	for i, el := range h.activityReward {
		h.activityRewardMap[el.Id] = i
	}
}

func (i *ActivityReward) Len() int {
	return len(h.activityReward)
}

func (i *ActivityReward) List() []ActivityReward {
	return h.activityReward
}

func (i *ActivityReward) GetActivityRewardById(id values.Integer) (*ActivityReward, bool) {
	index, ok := h.activityRewardMap[id]
	if !ok {
		return nil, false
	}
	return &h.activityReward[index], true
}