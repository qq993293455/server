// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type Emoji struct {
	Id        values.Integer `mapstructure:"id" json:"id"`
	EmojiName string         `mapstructure:"emoji_name" json:"emoji_name"`
}

// parse func
func ParseEmoji(data *Data) {
	if err := data.UnmarshalKey("emoji", &h.emoji); err != nil {
		panic(errors.New("parse table Emoji err:\n" + err.Error()))
	}
	for i, el := range h.emoji {
		h.emojiMap[el.Id] = i
	}
}

func (i *Emoji) Len() int {
	return len(h.emoji)
}

func (i *Emoji) List() []Emoji {
	return h.emoji
}

func (i *Emoji) GetEmojiById(id values.Integer) (*Emoji, bool) {
	index, ok := h.emojiMap[id]
	if !ok {
		return nil, false
	}
	return &h.emoji[index], true
}
