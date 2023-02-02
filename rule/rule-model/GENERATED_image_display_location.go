// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ImageDisplayLocation struct {
	Id                                    values.Integer   `mapstructure:"id" json:"id"`
	Gm                                    string           `mapstructure:"gm" json:"gm"`
	NameId                                string           `mapstructure:"name_id" json:"name_id"`
	HeadImg                               string           `mapstructure:"head_img" json:"head_img"`
	BattleHeadIcon                        string           `mapstructure:"battle_head_icon" json:"battle_head_icon"`
	FullBodyImg                           string           `mapstructure:"full_body_img" json:"full_body_img"`
	BattleLoadingPosUpper                 []values.Integer `mapstructure:"battle_loading_pos_upper" json:"battle_loading_pos_upper"`
	IsBattleLoadingPosUpperFlipHorizontal values.Integer   `mapstructure:"is_battle_loading_pos_upper_flip_horizontal" json:"is_battle_loading_pos_upper_flip_horizontal"`
	BattleLoadingPosBelow                 []values.Integer `mapstructure:"battle_loading_pos_below" json:"battle_loading_pos_below"`
	IsBattleLoadingPosBelowFlipHorizontal values.Integer   `mapstructure:"is_battle_loading_pos_below_flip_horizontal" json:"is_battle_loading_pos_below_flip_horizontal"`
	BattleLoadingScale                    []values.Integer `mapstructure:"battle_loading_scale" json:"battle_loading_scale"`
	TowerDefaultPos                       []values.Integer `mapstructure:"tower_default_pos" json:"tower_default_pos"`
	TowerDefaultScale                     []values.Integer `mapstructure:"tower_default_scale" json:"tower_default_scale"`
}

// parse func
func ParseImageDisplayLocation(data *Data) {
	if err := data.UnmarshalKey("image_display_location", &h.imageDisplayLocation); err != nil {
		panic(errors.New("parse table ImageDisplayLocation err:\n" + err.Error()))
	}
	for i, el := range h.imageDisplayLocation {
		h.imageDisplayLocationMap[el.Id] = i
	}
}

func (i *ImageDisplayLocation) Len() int {
	return len(h.imageDisplayLocation)
}

func (i *ImageDisplayLocation) List() []ImageDisplayLocation {
	return h.imageDisplayLocation
}

func (i *ImageDisplayLocation) GetImageDisplayLocationById(id values.Integer) (*ImageDisplayLocation, bool) {
	index, ok := h.imageDisplayLocationMap[id]
	if !ok {
		return nil, false
	}
	return &h.imageDisplayLocation[index], true
}