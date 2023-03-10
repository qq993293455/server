// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ImgLanguage struct {
	Id       values.Integer `mapstructure:"id" json:"id"`
	Group    values.Integer `mapstructure:"group" json:"group"`
	Language values.Integer `mapstructure:"language" json:"language"`
	ImgName  string         `mapstructure:"img_name" json:"img_name"`
}

// parse func
func ParseImgLanguage(data *Data) {
	if err := data.UnmarshalKey("img_language", &h.imgLanguage); err != nil {
		panic(errors.New("parse table ImgLanguage err:\n" + err.Error()))
	}
	for i, el := range h.imgLanguage {
		h.imgLanguageMap[el.Id] = i
	}
}

func (i *ImgLanguage) Len() int {
	return len(h.imgLanguage)
}

func (i *ImgLanguage) List() []ImgLanguage {
	return h.imgLanguage
}

func (i *ImgLanguage) GetImgLanguageById(id values.Integer) (*ImgLanguage, bool) {
	index, ok := h.imgLanguageMap[id]
	if !ok {
		return nil, false
	}
	return &h.imgLanguage[index], true
}
