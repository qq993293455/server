// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type GuildFunction struct {
	Id     values.Integer `mapstructure:"id" json:"id"`
	TextId string         `mapstructure:"text_id" json:"text_id"`
}

// parse func
func ParseGuildFunction(data *Data) {
	if err := data.UnmarshalKey("guild_function", &h.guildFunction); err != nil {
		panic(errors.New("parse table GuildFunction err:\n" + err.Error()))
	}
	for i, el := range h.guildFunction {
		h.guildFunctionMap[el.Id] = i
	}
}

func (i *GuildFunction) Len() int {
	return len(h.guildFunction)
}

func (i *GuildFunction) List() []GuildFunction {
	return h.guildFunction
}

func (i *GuildFunction) GetGuildFunctionById(id values.Integer) (*GuildFunction, bool) {
	index, ok := h.guildFunctionMap[id]
	if !ok {
		return nil, false
	}
	return &h.guildFunction[index], true
}