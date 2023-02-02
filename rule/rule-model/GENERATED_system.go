// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type System struct {
	Id                       values.Integer   `mapstructure:"id" json:"id"`
	Name                     string           `mapstructure:"name" json:"name"`
	NameId                   string           `mapstructure:"name_id" json:"name_id"`
	Desc                     string           `mapstructure:"desc" json:"desc"`
	UnlockCondition          []values.Integer `mapstructure:"unlock_condition" json:"unlock_condition"`
	SysType                  values.Integer   `mapstructure:"sys_type" json:"sys_type"`
	EventType                values.Integer   `mapstructure:"event_type" json:"event_type"`
	Sort                     values.Integer   `mapstructure:"sort" json:"sort"`
	ParentSysId              values.Integer   `mapstructure:"parent_sys_id" json:"parent_sys_id"`
	UnlockPopup              bool             `mapstructure:"unlock_popup" json:"unlock_popup"`
	UnlockAni                bool             `mapstructure:"unlock_ani" json:"unlock_ani"`
	PanelName                string           `mapstructure:"panel_name" json:"panel_name"`
	MainDataModel            string           `mapstructure:"main_data_model" json:"main_data_model"`
	SysIcon                  string           `mapstructure:"sys_icon" json:"sys_icon"`
	SystemIconName           string           `mapstructure:"system_icon_name" json:"system_icon_name"`
	MapBuildingResponseRange []values.Integer `mapstructure:"map_building_response_range" json:"map_building_response_range"`
	IsPreview                values.Integer   `mapstructure:"is_preview" json:"is_preview"`
	PreviewSequence          values.Integer   `mapstructure:"preview_sequence" json:"preview_sequence"`
	ModuleName               []string         `mapstructure:"module_name" json:"module_name"`
}

// parse func
func ParseSystem(data *Data) {
	if err := data.UnmarshalKey("system", &h.system); err != nil {
		panic(errors.New("parse table System err:\n" + err.Error()))
	}
	for i, el := range h.system {
		h.systemMap[el.Id] = i
	}
}

func (i *System) Len() int {
	return len(h.system)
}

func (i *System) List() []System {
	return h.system
}

func (i *System) GetSystemById(id values.Integer) (*System, bool) {
	index, ok := h.systemMap[id]
	if !ok {
		return nil, false
	}
	return &h.system[index], true
}