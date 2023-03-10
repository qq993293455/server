// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ShieldedWord struct {
	Id           values.Integer `mapstructure:"id" json:"id"`
	ShieldedWord string         `mapstructure:"shielded_word" json:"shielded_word"`
}

// parse func
func ParseShieldedWord(data *Data) {
	if err := data.UnmarshalKey("shielded_word", &h.shieldedWord); err != nil {
		panic(errors.New("parse table ShieldedWord err:\n" + err.Error()))
	}
	for i, el := range h.shieldedWord {
		h.shieldedWordMap[el.Id] = i
	}
}

func (i *ShieldedWord) Len() int {
	return len(h.shieldedWord)
}

func (i *ShieldedWord) List() []ShieldedWord {
	return h.shieldedWord
}

func (i *ShieldedWord) GetShieldedWordById(id values.Integer) (*ShieldedWord, bool) {
	index, ok := h.shieldedWordMap[id]
	if !ok {
		return nil, false
	}
	return &h.shieldedWord[index], true
}
