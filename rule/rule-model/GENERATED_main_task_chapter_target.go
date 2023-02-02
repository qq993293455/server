// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type MainTaskChapterTarget struct {
	Id                 values.Integer                    `mapstructure:"id" json:"id"`
	TargetName         string                            `mapstructure:"target_name" json:"target_name"`
	ShowType           values.Integer                    `mapstructure:"show_type" json:"show_type"`
	EntranceButtonIcon string                            `mapstructure:"entrance_button_icon" json:"entrance_button_icon"`
	TargetDes          string                            `mapstructure:"target_des" json:"target_des"`
	SubmitDes          string                            `mapstructure:"submit_des" json:"submit_des"`
	TaskStage          []values.Integer                  `mapstructure:"task_stage" json:"task_stage"`
	FinalReward        map[values.Integer]values.Integer `mapstructure:"final_reward" json:"final_reward"`
	FinalRewardShow    []values.Integer                  `mapstructure:"final_reward_show" json:"final_reward_show"`
	DesLanguage        string                            `mapstructure:"des_language" json:"des_language"`
}

// parse func
func ParseMainTaskChapterTarget(data *Data) {
	if err := data.UnmarshalKey("main_task_chapter_target", &h.mainTaskChapterTarget); err != nil {
		panic(errors.New("parse table MainTaskChapterTarget err:\n" + err.Error()))
	}
	for i, el := range h.mainTaskChapterTarget {
		h.mainTaskChapterTargetMap[el.Id] = i
	}
}

func (i *MainTaskChapterTarget) Len() int {
	return len(h.mainTaskChapterTarget)
}

func (i *MainTaskChapterTarget) List() []MainTaskChapterTarget {
	return h.mainTaskChapterTarget
}

func (i *MainTaskChapterTarget) GetMainTaskChapterTargetById(id values.Integer) (*MainTaskChapterTarget, bool) {
	index, ok := h.mainTaskChapterTargetMap[id]
	if !ok {
		return nil, false
	}
	return &h.mainTaskChapterTarget[index], true
}
