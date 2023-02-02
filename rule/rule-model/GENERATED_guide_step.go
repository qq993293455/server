// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type GuideStep struct {
	GuideProcessId       values.Integer   `mapstructure:"guide_process_id" json:"guide_process_id"`
	Id                   values.Integer   `mapstructure:"id" json:"id"`
	StepMark             string           `mapstructure:"step_mark" json:"step_mark"`
	RecordPoint          string           `mapstructure:"record_point" json:"record_point"`
	IsForce              bool             `mapstructure:"is_force" json:"is_force"`
	TopMostMask          bool             `mapstructure:"top_most_mask" json:"top_most_mask"`
	GuideTyp             values.Integer   `mapstructure:"guide_typ" json:"guide_typ"`
	ClickGuideTarget     string           `mapstructure:"click_guide_target" json:"click_guide_target"`
	ClickGuideDescParam  []values.Integer `mapstructure:"click_guide_desc_param" json:"click_guide_desc_param"`
	GuideArrowParam      []values.Integer `mapstructure:"guide_arrow_param" json:"guide_arrow_param"`
	GestureGuideTarget   string           `mapstructure:"gesture_guide_target" json:"gesture_guide_target"`
	TargetEntityType     values.Integer   `mapstructure:"target_entity_type" json:"target_entity_type"`
	TargetEntityId       values.Integer   `mapstructure:"target_entity_id" json:"target_entity_id"`
	FrameGuideParam      []values.Integer `mapstructure:"frame_guide_param" json:"frame_guide_param"`
	FrameGuideSizeOffset []values.Integer `mapstructure:"frame_guide_size_offset" json:"frame_guide_size_offset"`
	FrameGuideDescPos    []values.Integer `mapstructure:"frame_guide_desc_pos" json:"frame_guide_desc_pos"`
	PictureGuide         string           `mapstructure:"picture_guide" json:"picture_guide"`
	GuideCharacter       values.Integer   `mapstructure:"guide_character" json:"guide_character"`
	Text                 string           `mapstructure:"text" json:"text"`
}

// parse func
func ParseGuideStep(data *Data) {
	if err := data.UnmarshalKey("guide_step", &h.guideStep); err != nil {
		panic(errors.New("parse table GuideStep err:\n" + err.Error()))
	}
	for i, el := range h.guideStep {
		if _, ok := h.guideStepMap[el.GuideProcessId]; !ok {
			h.guideStepMap[el.GuideProcessId] = map[values.Integer]int{el.Id: i}
		} else {
			h.guideStepMap[el.GuideProcessId][el.Id] = i
		}
	}
}

func (i *GuideStep) Len() int {
	return len(h.guideStep)
}

func (i *GuideStep) List() []GuideStep {
	return h.guideStep
}

func (i *GuideStep) GetGuideStepById(parentId, id values.Integer) (*GuideStep, bool) {
	item, ok := h.guideStepMap[parentId]
	if !ok {
		return nil, false
	}
	index, ok := item[id]
	if !ok {
		return nil, false
	}
	return &h.guideStep[index], true
}