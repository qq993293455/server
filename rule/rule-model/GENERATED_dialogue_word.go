// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type DialogueWord struct {
	NpcDialogueId values.Integer   `mapstructure:"npc_dialogue_id" json:"npc_dialogue_id"`
	Id            values.Integer   `mapstructure:"id" json:"id"`
	SpeakerId     values.Integer   `mapstructure:"speaker_id" json:"speaker_id"`
	MapNpcId      values.Integer   `mapstructure:"map_npc_id" json:"map_npc_id"`
	SpeakerSide   values.Integer   `mapstructure:"speaker_side" json:"speaker_side"`
	Word          string           `mapstructure:"word" json:"word"`
	RoleAudio     string           `mapstructure:"role_audio" json:"role_audio"`
	Bgm           string           `mapstructure:"bgm" json:"bgm"`
	ShockParam    []values.Integer `mapstructure:"shock_param" json:"shock_param"`
}

// parse func
func ParseDialogueWord(data *Data) {
	if err := data.UnmarshalKey("dialogue_word", &h.dialogueWord); err != nil {
		panic(errors.New("parse table DialogueWord err:\n" + err.Error()))
	}
	for i, el := range h.dialogueWord {
		if _, ok := h.dialogueWordMap[el.NpcDialogueId]; !ok {
			h.dialogueWordMap[el.NpcDialogueId] = map[values.Integer]int{el.Id: i}
		} else {
			h.dialogueWordMap[el.NpcDialogueId][el.Id] = i
		}
	}
}

func (i *DialogueWord) Len() int {
	return len(h.dialogueWord)
}

func (i *DialogueWord) List() []DialogueWord {
	return h.dialogueWord
}

func (i *DialogueWord) GetDialogueWordById(parentId, id values.Integer) (*DialogueWord, bool) {
	item, ok := h.dialogueWordMap[parentId]
	if !ok {
		return nil, false
	}
	index, ok := item[id]
	if !ok {
		return nil, false
	}
	return &h.dialogueWord[index], true
}
