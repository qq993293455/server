// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ForgeFixedRecipe struct {
	Id                          values.Integer                    `mapstructure:"id" json:"id"`
	Level                       values.Integer                    `mapstructure:"level" json:"level"`
	EquipKind                   values.Integer                    `mapstructure:"equip_kind" json:"equip_kind"`
	EquipId                     values.Integer                    `mapstructure:"equip_id" json:"equip_id"`
	Gold                        values.Integer                    `mapstructure:"gold" json:"gold"`
	MaterialsList               map[values.Integer]values.Integer `mapstructure:"materials_list" json:"materials_list"`
	Products                    [][]values.Integer                `mapstructure:"products" json:"products"`
	MinimumGuaranteeProbability map[values.Integer]values.Integer `mapstructure:"minimum_guarantee_probability" json:"minimum_guarantee_probability"`
}

// parse func
func ParseForgeFixedRecipe(data *Data) {
	if err := data.UnmarshalKey("forge_fixed_recipe", &h.forgeFixedRecipe); err != nil {
		panic(errors.New("parse table ForgeFixedRecipe err:\n" + err.Error()))
	}
	for i, el := range h.forgeFixedRecipe {
		h.forgeFixedRecipeMap[el.Id] = i
	}
}

func (i *ForgeFixedRecipe) Len() int {
	return len(h.forgeFixedRecipe)
}

func (i *ForgeFixedRecipe) List() []ForgeFixedRecipe {
	return h.forgeFixedRecipe
}

func (i *ForgeFixedRecipe) GetForgeFixedRecipeById(id values.Integer) (*ForgeFixedRecipe, bool) {
	index, ok := h.forgeFixedRecipeMap[id]
	if !ok {
		return nil, false
	}
	return &h.forgeFixedRecipe[index], true
}