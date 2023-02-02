// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type MapRelation struct {
	Id             values.Integer `mapstructure:"id" json:"id"`
	RoleReachMapId values.Integer `mapstructure:"role_reach_map_id" json:"role_reach_map_id"`
	Tower          values.Integer `mapstructure:"tower" json:"tower"`
	Meet           values.Integer `mapstructure:"meet" json:"meet"`
	Arena          values.Integer `mapstructure:"arena" json:"arena"`
	Duel           values.Integer `mapstructure:"duel" json:"duel"`
	Plane          values.Integer `mapstructure:"plane" json:"plane"`
	NpcBattle      values.Integer `mapstructure:"npc_battle" json:"npc_battle"`
	PersonalBoss   values.Integer `mapstructure:"personal_boss" json:"personal_boss"`
	TestBattle     values.Integer `mapstructure:"test_battle" json:"test_battle"`
	GvgBattle      values.Integer `mapstructure:"gvg_battle" json:"gvg_battle"`
}

// parse func
func ParseMapRelation(data *Data) {
	if err := data.UnmarshalKey("map_relation", &h.mapRelation); err != nil {
		panic(errors.New("parse table MapRelation err:\n" + err.Error()))
	}
	for i, el := range h.mapRelation {
		h.mapRelationMap[el.Id] = i
	}
}

func (i *MapRelation) Len() int {
	return len(h.mapRelation)
}

func (i *MapRelation) List() []MapRelation {
	return h.mapRelation
}

func (i *MapRelation) GetMapRelationById(id values.Integer) (*MapRelation, bool) {
	index, ok := h.mapRelationMap[id]
	if !ok {
		return nil, false
	}
	return &h.mapRelation[index], true
}