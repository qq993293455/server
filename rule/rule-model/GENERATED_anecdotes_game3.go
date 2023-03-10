// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type AnecdotesGame3 struct {
	Id     values.Integer `mapstructure:"id" json:"id"`
	GameLv values.Integer `mapstructure:"game_lv" json:"game_lv"`
	Weight values.Integer `mapstructure:"weight" json:"weight"`
}

// parse func
func ParseAnecdotesGame3(data *Data) {
	if err := data.UnmarshalKey("anecdotes_game3", &h.anecdotesGame3); err != nil {
		panic(errors.New("parse table AnecdotesGame3 err:\n" + err.Error()))
	}
	for i, el := range h.anecdotesGame3 {
		h.anecdotesGame3Map[el.Id] = i
	}
}

func (i *AnecdotesGame3) Len() int {
	return len(h.anecdotesGame3)
}

func (i *AnecdotesGame3) List() []AnecdotesGame3 {
	return h.anecdotesGame3
}

func (i *AnecdotesGame3) GetAnecdotesGame3ById(id values.Integer) (*AnecdotesGame3, bool) {
	index, ok := h.anecdotesGame3Map[id]
	if !ok {
		return nil, false
	}
	return &h.anecdotesGame3[index], true
}
