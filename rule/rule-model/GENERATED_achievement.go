// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type Achievement struct {
	Id                values.Integer `mapstructure:"id" json:"id"`
	TitleName         string         `mapstructure:"title_name" json:"title_name"`
	TitleNameLanguage values.Integer `mapstructure:"title_name_language" json:"title_name_language"`
	Typ               values.Integer `mapstructure:"typ" json:"typ"`
}

// parse func
func ParseAchievement(data *Data) {
	if err := data.UnmarshalKey("achievement", &h.achievement); err != nil {
		panic(errors.New("parse table Achievement err:\n" + err.Error()))
	}
	for i, el := range h.achievement {
		h.achievementMap[el.Id] = i
	}
}

func (i *Achievement) Len() int {
	return len(h.achievement)
}

func (i *Achievement) List() []Achievement {
	return h.achievement
}

func (i *Achievement) GetAchievementById(id values.Integer) (*Achievement, bool) {
	index, ok := h.achievementMap[id]
	if !ok {
		return nil, false
	}
	return &h.achievement[index], true
}