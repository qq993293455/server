// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type AnecdotesGame8 struct {
	Id       values.Integer `mapstructure:"id" json:"id"`
	Name     string         `mapstructure:"name" json:"name"`
	ChestImg string         `mapstructure:"chest_img" json:"chest_img"`
	GameLv   values.Integer `mapstructure:"game_lv" json:"game_lv"`
	Weight   values.Integer `mapstructure:"weight" json:"weight"`
}

// parse func
func ParseAnecdotesGame8(data *Data) {
	if err := data.UnmarshalKey("anecdotes_game8", &h.anecdotesGame8); err != nil {
		panic(errors.New("parse table AnecdotesGame8 err:\n" + err.Error()))
	}
	for i, el := range h.anecdotesGame8 {
		h.anecdotesGame8Map[el.Id] = i
	}
}

func (i *AnecdotesGame8) Len() int {
	return len(h.anecdotesGame8)
}

func (i *AnecdotesGame8) List() []AnecdotesGame8 {
	return h.anecdotesGame8
}

func (i *AnecdotesGame8) GetAnecdotesGame8ById(id values.Integer) (*AnecdotesGame8, bool) {
	index, ok := h.anecdotesGame8Map[id]
	if !ok {
		return nil, false
	}
	return &h.anecdotesGame8[index], true
}
