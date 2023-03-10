// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type GuildContendbuild struct {
	Id            values.Integer   `mapstructure:"id" json:"id"`
	TextId        values.Integer   `mapstructure:"text_id" json:"text_id"`
	Icon          string           `mapstructure:"icon" json:"icon"`
	BuildPriority values.Integer   `mapstructure:"build_priority" json:"build_priority"`
	BuildHp       values.Integer   `mapstructure:"build_hp" json:"build_hp"`
	BuildHpUp     values.Integer   `mapstructure:"build_hp_up" json:"build_hp_up"`
	BuildGeneral  values.Integer   `mapstructure:"build_general" json:"build_general"`
	BuildEff      []values.Integer `mapstructure:"build_eff" json:"build_eff"`
	EffTextId     values.Integer   `mapstructure:"eff_text_id" json:"eff_text_id"`
	DefenseNum    values.Integer   `mapstructure:"defense_num" json:"defense_num"`
	NumAdd        values.Integer   `mapstructure:"num_add" json:"num_add"`
	MapSceneId    values.Integer   `mapstructure:"map_scene_id" json:"map_scene_id"`
}

// parse func
func ParseGuildContendbuild(data *Data) {
	if err := data.UnmarshalKey("guild__contendbuild", &h.guildContendbuild); err != nil {
		panic(errors.New("parse table GuildContendbuild err:\n" + err.Error()))
	}
	for i, el := range h.guildContendbuild {
		h.guildContendbuildMap[el.Id] = i
	}
}

func (i *GuildContendbuild) Len() int {
	return len(h.guildContendbuild)
}

func (i *GuildContendbuild) List() []GuildContendbuild {
	return h.guildContendbuild
}

func (i *GuildContendbuild) GetGuildContendbuildById(id values.Integer) (*GuildContendbuild, bool) {
	index, ok := h.guildContendbuildMap[id]
	if !ok {
		return nil, false
	}
	return &h.guildContendbuild[index], true
}
