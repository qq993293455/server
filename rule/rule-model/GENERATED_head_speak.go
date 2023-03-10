// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type HeadSpeak struct {
	Id                values.Integer     `mapstructure:"id" json:"id"`
	TriggerCond       values.Integer     `mapstructure:"trigger_cond" json:"trigger_cond"`
	TriggerCondParam  []values.Integer   `mapstructure:"trigger_cond_param" json:"trigger_cond_param"`
	TriggerCondParams [][]values.Integer `mapstructure:"trigger_cond_params" json:"trigger_cond_params"`
}

// parse func
func ParseHeadSpeak(data *Data) {
	if err := data.UnmarshalKey("head_speak", &h.headSpeak); err != nil {
		panic(errors.New("parse table HeadSpeak err:\n" + err.Error()))
	}
	for i, el := range h.headSpeak {
		h.headSpeakMap[el.Id] = i
	}
}

func (i *HeadSpeak) Len() int {
	return len(h.headSpeak)
}

func (i *HeadSpeak) List() []HeadSpeak {
	return h.headSpeak
}

func (i *HeadSpeak) GetHeadSpeakById(id values.Integer) (*HeadSpeak, bool) {
	index, ok := h.headSpeakMap[id]
	if !ok {
		return nil, false
	}
	return &h.headSpeak[index], true
}
