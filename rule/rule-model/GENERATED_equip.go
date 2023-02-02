// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type Equip struct {
	Id               values.Integer                    `mapstructure:"id" json:"id"`
	CompleteId       values.Integer                    `mapstructure:"complete_id" json:"complete_id"`
	Kind             values.Integer                    `mapstructure:"kind" json:"kind"`
	Typ              values.Integer                    `mapstructure:"typ" json:"typ"`
	SpecificType     values.Integer                    `mapstructure:"specific_type" json:"specific_type"`
	QualityNum       map[values.Integer]values.Integer `mapstructure:"quality_num" json:"quality_num"`
	QualityEffects   map[values.Integer]values.Integer `mapstructure:"quality_effects" json:"quality_effects"`
	EquipSlot        values.Integer                    `mapstructure:"equip_slot" json:"equip_slot"`
	EquipMod         map[values.Integer]string         `mapstructure:"equip_mod" json:"equip_mod"`
	EquipModHeight   map[values.Integer]string         `mapstructure:"equip_mod_height" json:"equip_mod_height"`
	EquipJob         []values.Integer                  `mapstructure:"equip_job" json:"equip_job"`
	ModleId          []string                          `mapstructure:"modle_id" json:"modle_id"`
	AttributeQuality values.Integer                    `mapstructure:"attribute_quality" json:"attribute_quality"`
	QualityEquip     values.Integer                    `mapstructure:"quality_equip" json:"quality_equip"`
	MeltingValue     values.Integer                    `mapstructure:"melting_value" json:"melting_value"`
	AttrId           []values.Integer                  `mapstructure:"attr_id" json:"attr_id"`
	AttrValue        []values.Integer                  `mapstructure:"attr_value" json:"attr_value"`
	AttrStarRange    [][]values.Integer                `mapstructure:"attr_star_range" json:"attr_star_range"`
	Attr             [][]values.Integer                `mapstructure:"attr" json:"attr"`
	AttributeNum     values.Integer                    `mapstructure:"attribute_num" json:"attribute_num"`
	Attribute        values.Integer                    `mapstructure:"attribute" json:"attribute"`
	EquipSkill       map[values.Integer]values.Integer `mapstructure:"equip_skill" json:"equip_skill"`
}

// parse func
func ParseEquip(data *Data) {
	if err := data.UnmarshalKey("equip", &h.equip); err != nil {
		panic(errors.New("parse table Equip err:\n" + err.Error()))
	}
	for i, el := range h.equip {
		h.equipMap[el.Id] = i
	}
}

func (i *Equip) Len() int {
	return len(h.equip)
}

func (i *Equip) List() []Equip {
	return h.equip
}

func (i *Equip) GetEquipById(id values.Integer) (*Equip, bool) {
	index, ok := h.equipMap[id]
	if !ok {
		return nil, false
	}
	return &h.equip[index], true
}