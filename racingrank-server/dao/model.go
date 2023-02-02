package dao

import "coin-server/common/values"

type Data struct {
	RoleId       values.RoleId  `db:"role_id" json:"role_id"`
	HighestPower values.Integer `db:"highest_power" json:"highest_power"`
	LoginTime    values.Integer `db:"login_time" json:"login_time"`
}

type EndTime struct {
	RoleId  string `db:"role_id" json:"role_id"`
	EndTime int64  `db:"end_time" json:"end_time"`
}
