package dao

import "coin-server/common/values"

type Fashion struct {
	Id        values.Integer  `db:"id" json:"id"`
	RoleId    values.RoleId   `db:"role_id" json:"role_id"`
	UserId    string          `db:"user_id" json:"user_id"`
	ServerId  values.ServerId `db:"server_id" json:"server_id"`
	HeroId    values.Integer  `db:"hero_id" json:"hero_id"`
	FashionId values.Integer  `db:"fashion_id" json:"fashion_id"`
	ExpiredAt values.Integer  `db:"expired_at" json:"expired_at"`
	CreatedAt values.Integer  `db:"created_at" json:"created_at"`
}
