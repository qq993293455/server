// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type GuideProcess struct {
	Id            values.Integer   `mapstructure:"id" json:"id"`
	ProccessMark  string           `mapstructure:"proccess_mark" json:"proccess_mark"`
	Condition     []values.Integer `mapstructure:"condition" json:"condition"`
	TriggerType   values.Integer   `mapstructure:"trigger_type" json:"trigger_type"`
	StartSystem   values.Integer   `mapstructure:"start_system" json:"start_system"`
	ExtraGuide    values.Integer   `mapstructure:"extra_guide" json:"extra_guide"`
	CompleteStage values.Integer   `mapstructure:"complete_stage" json:"complete_stage"`
	ReplayStage   values.Integer   `mapstructure:"replay_stage" json:"replay_stage"`
}

// parse func
func ParseGuideProcess(data *Data) {
	if err := data.UnmarshalKey("guide_process", &h.guideProcess); err != nil {
		panic(errors.New("parse table GuideProcess err:\n" + err.Error()))
	}
	for i, el := range h.guideProcess {
		h.guideProcessMap[el.Id] = i
	}
}

func (i *GuideProcess) Len() int {
	return len(h.guideProcess)
}

func (i *GuideProcess) List() []GuideProcess {
	return h.guideProcess
}

func (i *GuideProcess) GetGuideProcessById(id values.Integer) (*GuideProcess, bool) {
	index, ok := h.guideProcessMap[id]
	if !ok {
		return nil, false
	}
	return &h.guideProcess[index], true
}
