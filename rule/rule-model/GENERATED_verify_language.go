// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type VerifyLanguage struct {
	Id          values.Integer `mapstructure:"id" json:"id"`
	TextId      string         `mapstructure:"text_id" json:"text_id"`
	Name        string         `mapstructure:"name" json:"name"`
	SysLanguage bool           `mapstructure:"sys_language" json:"sys_language"`
}

// parse func
func ParseVerifyLanguage(data *Data) {
	if err := data.UnmarshalKey("verify_language", &h.verifyLanguage); err != nil {
		panic(errors.New("parse table VerifyLanguage err:\n" + err.Error()))
	}
	for i, el := range h.verifyLanguage {
		h.verifyLanguageMap[el.Id] = i
	}
}

func (i *VerifyLanguage) Len() int {
	return len(h.verifyLanguage)
}

func (i *VerifyLanguage) List() []VerifyLanguage {
	return h.verifyLanguage
}

func (i *VerifyLanguage) GetVerifyLanguageById(id values.Integer) (*VerifyLanguage, bool) {
	index, ok := h.verifyLanguageMap[id]
	if !ok {
		return nil, false
	}
	return &h.verifyLanguage[index], true
}
