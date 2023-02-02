// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type Activity struct {
	Id                         values.Integer `mapstructure:"id" json:"id"`
	TimeType                   values.Integer `mapstructure:"time_type" json:"time_type"`
	ActivityOpenTime           string         `mapstructure:"activity_open_time" json:"activity_open_time"`
	DurationTime               string         `mapstructure:"duration_time" json:"duration_time"`
	ActivityVersion            values.Integer `mapstructure:"activity_version" json:"activity_version"`
	SystemId                   values.Integer `mapstructure:"system_id" json:"system_id"`
	ChargeId                   []string       `mapstructure:"charge_id" json:"charge_id"`
	ActivityDescribeLanguageId []string       `mapstructure:"activity_describe_language_id" json:"activity_describe_language_id"`
	ActivityPictureId          []string       `mapstructure:"activity_picture_id" json:"activity_picture_id"`
	Banner                     string         `mapstructure:"banner" json:"banner"`
	ActivitySort               values.Integer `mapstructure:"activity_sort" json:"activity_sort"`
}

// parse func
func ParseActivity(data *Data) {
	if err := data.UnmarshalKey("activity", &h.activity); err != nil {
		panic(errors.New("parse table Activity err:\n" + err.Error()))
	}
	for i, el := range h.activity {
		h.activityMap[el.Id] = i
	}
}

func (i *Activity) Len() int {
	return len(h.activity)
}

func (i *Activity) List() []Activity {
	return h.activity
}

func (i *Activity) GetActivityById(id values.Integer) (*Activity, bool) {
	index, ok := h.activityMap[id]
	if !ok {
		return nil, false
	}
	return &h.activity[index], true
}
