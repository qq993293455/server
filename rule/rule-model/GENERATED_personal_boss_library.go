// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type PersonalBossLibrary struct {
	Id                      values.Integer   `mapstructure:"id" json:"id"`
	BossId                  values.Integer   `mapstructure:"boss_id" json:"boss_id"`
	HpCoefficient           values.Integer   `mapstructure:"hp_coefficient" json:"hp_coefficient"`
	AtkCoefficient          values.Integer   `mapstructure:"atk_coefficient" json:"atk_coefficient"`
	DefCoefficient          values.Integer   `mapstructure:"def_coefficient" json:"def_coefficient"`
	BossOffsetInAvatarPanel []values.Integer `mapstructure:"boss_offset_in_avatar_panel" json:"boss_offset_in_avatar_panel"`
	BossSpineSize           values.Integer   `mapstructure:"boss_spine_size" json:"boss_spine_size"`
	WhetherToUseBoss        values.Integer   `mapstructure:"whether_to_use_boss" json:"whether_to_use_boss"`
}

// parse func
func ParsePersonalBossLibrary(data *Data) {
	if err := data.UnmarshalKey("personal_boss_library", &h.personalBossLibrary); err != nil {
		panic(errors.New("parse table PersonalBossLibrary err:\n" + err.Error()))
	}
	for i, el := range h.personalBossLibrary {
		h.personalBossLibraryMap[el.Id] = i
	}
}

func (i *PersonalBossLibrary) Len() int {
	return len(h.personalBossLibrary)
}

func (i *PersonalBossLibrary) List() []PersonalBossLibrary {
	return h.personalBossLibrary
}

func (i *PersonalBossLibrary) GetPersonalBossLibraryById(id values.Integer) (*PersonalBossLibrary, bool) {
	index, ok := h.personalBossLibraryMap[id]
	if !ok {
		return nil, false
	}
	return &h.personalBossLibrary[index], true
}
