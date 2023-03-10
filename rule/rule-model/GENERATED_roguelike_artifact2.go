// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type RoguelikeArtifact2 struct {
	Id       values.Integer   `mapstructure:"id" json:"id"`
	CardIcon string           `mapstructure:"card_icon" json:"card_icon"`
	CardText string           `mapstructure:"card_text" json:"card_text"`
	ItemName string           `mapstructure:"item_name" json:"item_name"`
	CardId   []values.Integer `mapstructure:"card_id" json:"card_id"`
}

// parse func
func ParseRoguelikeArtifact2(data *Data) {
	if err := data.UnmarshalKey("roguelike_artifact2", &h.roguelikeArtifact2); err != nil {
		panic(errors.New("parse table RoguelikeArtifact2 err:\n" + err.Error()))
	}
	for i, el := range h.roguelikeArtifact2 {
		h.roguelikeArtifact2Map[el.Id] = i
	}
}

func (i *RoguelikeArtifact2) Len() int {
	return len(h.roguelikeArtifact2)
}

func (i *RoguelikeArtifact2) List() []RoguelikeArtifact2 {
	return h.roguelikeArtifact2
}

func (i *RoguelikeArtifact2) GetRoguelikeArtifact2ById(id values.Integer) (*RoguelikeArtifact2, bool) {
	index, ok := h.roguelikeArtifact2Map[id]
	if !ok {
		return nil, false
	}
	return &h.roguelikeArtifact2[index], true
}
