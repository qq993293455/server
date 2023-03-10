// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type PlaneDungeon struct {
	Id               values.Integer                    `mapstructure:"id" json:"id"`
	MapScene         values.Integer                    `mapstructure:"map_scene" json:"map_scene"`
	Bt               string                            `mapstructure:"bt" json:"bt"`
	MonsterGroupInfo map[values.Integer]values.Integer `mapstructure:"monster_group_info" json:"monster_group_info"`
	DropReward       map[values.Integer]values.Integer `mapstructure:"drop_reward" json:"drop_reward"`
	Duration         values.Integer                    `mapstructure:"duration" json:"duration"`
}

// parse func
func ParsePlaneDungeon(data *Data) {
	if err := data.UnmarshalKey("plane_dungeon", &h.planeDungeon); err != nil {
		panic(errors.New("parse table PlaneDungeon err:\n" + err.Error()))
	}
	for i, el := range h.planeDungeon {
		h.planeDungeonMap[el.Id] = i
	}
}

func (i *PlaneDungeon) Len() int {
	return len(h.planeDungeon)
}

func (i *PlaneDungeon) List() []PlaneDungeon {
	return h.planeDungeon
}

func (i *PlaneDungeon) GetPlaneDungeonById(id values.Integer) (*PlaneDungeon, bool) {
	index, ok := h.planeDungeonMap[id]
	if !ok {
		return nil, false
	}
	return &h.planeDungeon[index], true
}
