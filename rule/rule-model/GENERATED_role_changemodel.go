// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type RoleChangemodel struct {
	Id                values.Integer   `mapstructure:"id" json:"id"`
	ModleId           string           `mapstructure:"modle_id" json:"modle_id"`
	WeaponType        values.Integer   `mapstructure:"weapon_type" json:"weapon_type"`
	Modle3dSize       values.Integer   `mapstructure:"modle3d_size" json:"modle3d_size"`
	ModelProportion   values.Integer   `mapstructure:"model_proportion" json:"model_proportion"`
	BoxColliderCenter []values.Integer `mapstructure:"box_collider_center" json:"box_collider_center"`
	BoxColliderSize   []values.Integer `mapstructure:"box_collider_size" json:"box_collider_size"`
}

// parse func
func ParseRoleChangemodel(data *Data) {
	if err := data.UnmarshalKey("role_changemodel", &h.roleChangemodel); err != nil {
		panic(errors.New("parse table RoleChangemodel err:\n" + err.Error()))
	}
	for i, el := range h.roleChangemodel {
		h.roleChangemodelMap[el.Id] = i
	}
}

func (i *RoleChangemodel) Len() int {
	return len(h.roleChangemodel)
}

func (i *RoleChangemodel) List() []RoleChangemodel {
	return h.roleChangemodel
}

func (i *RoleChangemodel) GetRoleChangemodelById(id values.Integer) (*RoleChangemodel, bool) {
	index, ok := h.roleChangemodelMap[id]
	if !ok {
		return nil, false
	}
	return &h.roleChangemodel[index], true
}