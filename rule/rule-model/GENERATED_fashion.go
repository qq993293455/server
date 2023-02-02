// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type Fashion struct {
	Id                           values.Integer                    `mapstructure:"id" json:"id"`
	Language1                    string                            `mapstructure:"language1" json:"language1"`
	Quality                      values.Integer                    `mapstructure:"quality" json:"quality"`
	Icon                         string                            `mapstructure:"icon" json:"icon"`
	IsDefault                    values.Integer                    `mapstructure:"is_default" json:"is_default"`
	Language2                    string                            `mapstructure:"language2" json:"language2"`
	Attr                         map[values.Integer]values.Integer `mapstructure:"attr" json:"attr"`
	Hero                         values.Integer                    `mapstructure:"hero" json:"hero"`
	HeroImg                      string                            `mapstructure:"hero_img" json:"hero_img"`
	HeroSpine                    string                            `mapstructure:"hero_spine" json:"hero_spine"`
	HeroSpineOffsetInAvatarPanel []values.Integer                  `mapstructure:"hero_spine_offset_in_avatar_panel" json:"hero_spine_offset_in_avatar_panel"`
	HeroSpineSize                []values.Integer                  `mapstructure:"hero_spine_size" json:"hero_spine_size"`
	HeroImgSize                  []values.Integer                  `mapstructure:"hero_img_size" json:"hero_img_size"`
	HeroImgOffsetInAvatarPanel   []values.Integer                  `mapstructure:"hero_img_offset_in_avatar_panel" json:"hero_img_offset_in_avatar_panel"`
	HeroImgOffsetInNewCareer     []values.Integer                  `mapstructure:"hero_img_offset_in_new_career" json:"hero_img_offset_in_new_career"`
	ImageDisplayLocation         values.Integer                    `mapstructure:"image_display_location" json:"image_display_location"`
	NewModelId                   values.Integer                    `mapstructure:"new_model_id" json:"new_model_id"`
	ModleId                      string                            `mapstructure:"modle_id" json:"modle_id"`
	Modle3dSize                  values.Integer                    `mapstructure:"modle3d_size" json:"modle3d_size"`
	ModelProportion              values.Integer                    `mapstructure:"model_proportion" json:"model_proportion"`
}

// parse func
func ParseFashion(data *Data) {
	if err := data.UnmarshalKey("fashion", &h.fashion); err != nil {
		panic(errors.New("parse table Fashion err:\n" + err.Error()))
	}
	for i, el := range h.fashion {
		h.fashionMap[el.Id] = i
	}
}

func (i *Fashion) Len() int {
	return len(h.fashion)
}

func (i *Fashion) List() []Fashion {
	return h.fashion
}

func (i *Fashion) GetFashionById(id values.Integer) (*Fashion, bool) {
	index, ok := h.fashionMap[id]
	if !ok {
		return nil, false
	}
	return &h.fashion[index], true
}