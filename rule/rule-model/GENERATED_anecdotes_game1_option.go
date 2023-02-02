// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type AnecdotesGame1Option struct {
	AnecdotesGame1Id values.Integer                    `mapstructure:"anecdotes_game1_id" json:"anecdotes_game1_id"`
	Id               values.Integer                    `mapstructure:"id" json:"id"`
	OptionType       values.Integer                    `mapstructure:"option_type" json:"option_type"`
	OptionName       string                            `mapstructure:"option_name" json:"option_name"`
	OptionIcon       string                            `mapstructure:"option_icon" json:"option_icon"`
	DropList         map[values.Integer]values.Integer `mapstructure:"drop_list" json:"drop_list"`
	Item             map[values.Integer]values.Integer `mapstructure:"item" json:"item"`
	SkillNum         values.Integer                    `mapstructure:"skill_num" json:"skill_num"`
	SkillId          values.Integer                    `mapstructure:"skill_id" json:"skill_id"`
}

// parse func
func ParseAnecdotesGame1Option(data *Data) {
	if err := data.UnmarshalKey("anecdotes_game1_option", &h.anecdotesGame1Option); err != nil {
		panic(errors.New("parse table AnecdotesGame1Option err:\n" + err.Error()))
	}
	for i, el := range h.anecdotesGame1Option {
		if _, ok := h.anecdotesGame1OptionMap[el.AnecdotesGame1Id]; !ok {
			h.anecdotesGame1OptionMap[el.AnecdotesGame1Id] = map[values.Integer]int{el.Id: i}
		} else {
			h.anecdotesGame1OptionMap[el.AnecdotesGame1Id][el.Id] = i
		}
	}
}

func (i *AnecdotesGame1Option) Len() int {
	return len(h.anecdotesGame1Option)
}

func (i *AnecdotesGame1Option) List() []AnecdotesGame1Option {
	return h.anecdotesGame1Option
}

func (i *AnecdotesGame1Option) GetAnecdotesGame1OptionById(parentId, id values.Integer) (*AnecdotesGame1Option, bool) {
	item, ok := h.anecdotesGame1OptionMap[parentId]
	if !ok {
		return nil, false
	}
	index, ok := item[id]
	if !ok {
		return nil, false
	}
	return &h.anecdotesGame1Option[index], true
}