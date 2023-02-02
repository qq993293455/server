// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type AnecdotesGames struct {
	Id             values.Integer   `mapstructure:"id" json:"id"`
	Name           string           `mapstructure:"name" json:"name"`
	NameId         string           `mapstructure:"name_id" json:"name_id"`
	Typ            values.Integer   `mapstructure:"typ" json:"typ"`
	Subtype        values.Integer   `mapstructure:"subtype" json:"subtype"`
	TriggerType    []values.Integer `mapstructure:"trigger_type" json:"trigger_type"`
	TriggerTimes   values.Integer   `mapstructure:"trigger_times" json:"trigger_times"`
	Des            string           `mapstructure:"des" json:"des"`
	LockedDescId   string           `mapstructure:"locked_desc_id" json:"locked_desc_id"`
	Game           string           `mapstructure:"game" json:"game"`
	DropId         []values.Integer `mapstructure:"drop_id" json:"drop_id"`
	CardPro        values.Integer   `mapstructure:"card_pro" json:"card_pro"`
	ViewObject2D   string           `mapstructure:"view_object2_d" json:"view_object2_d"`
	ViewObject3D   string           `mapstructure:"view_object3_d" json:"view_object3_d"`
	ColliderParams []values.Integer `mapstructure:"collider_params" json:"collider_params"`
}

// parse func
func ParseAnecdotesGames(data *Data) {
	if err := data.UnmarshalKey("anecdotes_games", &h.anecdotesGames); err != nil {
		panic(errors.New("parse table AnecdotesGames err:\n" + err.Error()))
	}
	for i, el := range h.anecdotesGames {
		h.anecdotesGamesMap[el.Id] = i
	}
}

func (i *AnecdotesGames) Len() int {
	return len(h.anecdotesGames)
}

func (i *AnecdotesGames) List() []AnecdotesGames {
	return h.anecdotesGames
}

func (i *AnecdotesGames) GetAnecdotesGamesById(id values.Integer) (*AnecdotesGames, bool) {
	index, ok := h.anecdotesGamesMap[id]
	if !ok {
		return nil, false
	}
	return &h.anecdotesGames[index], true
}
