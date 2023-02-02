// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type Relics struct {
	Id                   values.Integer     `mapstructure:"id" json:"id"`
	FragmentRelics       []values.Integer   `mapstructure:"fragment_relics" json:"fragment_relics"`
	SuitId               values.Integer     `mapstructure:"suit_id" json:"suit_id"`
	NameSuit             string             `mapstructure:"name_suit" json:"name_suit"`
	SuitIcon             string             `mapstructure:"suit_icon" json:"suit_icon"`
	SuitPos              []values.Integer   `mapstructure:"suit_pos" json:"suit_pos"`
	SuitSkill            [][]values.Integer `mapstructure:"suit_skill" json:"suit_skill"`
	HeroId               values.Integer     `mapstructure:"hero_id" json:"hero_id"`
	HeroIcon             string             `mapstructure:"hero_icon" json:"hero_icon"`
	NameRoom             string             `mapstructure:"name_room" json:"name_room"`
	MapId                values.Integer     `mapstructure:"map_id" json:"map_id"`
	LvMax                values.Integer     `mapstructure:"lv_max" json:"lv_max"`
	AttrId               []values.Integer   `mapstructure:"attr_id" json:"attr_id"`
	AttrValue            []values.Integer   `mapstructure:"attr_value" json:"attr_value"`
	AttrPer              []values.Integer   `mapstructure:"attr_per" json:"attr_per"`
	AttrStarRange        [][]values.Integer `mapstructure:"attr_star_range" json:"attr_star_range"`
	Attr                 [][]values.Integer `mapstructure:"attr" json:"attr"`
	StageMax             values.Integer     `mapstructure:"stage_max" json:"stage_max"`
	RelicsSkill          []values.Integer   `mapstructure:"relics_skill" json:"relics_skill"`
	StarsCost            [][]values.Integer `mapstructure:"stars_cost" json:"stars_cost"`
	RelicsFunctionAttrId values.Integer     `mapstructure:"relics_function_attr_id" json:"relics_function_attr_id"`
	MapName              string             `mapstructure:"map_name" json:"map_name"`
}

// parse func
func ParseRelics(data *Data) {
	if err := data.UnmarshalKey("relics", &h.relics); err != nil {
		panic(errors.New("parse table Relics err:\n" + err.Error()))
	}
	for i, el := range h.relics {
		h.relicsMap[el.Id] = i
	}
}

func (i *Relics) Len() int {
	return len(h.relics)
}

func (i *Relics) List() []Relics {
	return h.relics
}

func (i *Relics) GetRelicsById(id values.Integer) (*Relics, bool) {
	index, ok := h.relicsMap[id]
	if !ok {
		return nil, false
	}
	return &h.relics[index], true
}