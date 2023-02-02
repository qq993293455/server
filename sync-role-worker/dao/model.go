package dao

import (
	jsoniter "github.com/json-iterator/go"
)

type Role struct {
	RoleId       string `db:"role_id" json:"role_id"`
	IggId        string `db:"igg_id" json:"igg_id"`
	Nickname     string `db:"nickname" json:"nickname"`
	Level        int64  `db:"level" json:"level"`
	Avatar       int64  `db:"avatar" json:"avatar"`
	AvatarFrame  int64  `db:"avatar_frame" json:"avatar_frame"`
	Power        int64  `db:"power" json:"power"`
	HighestPower int64  `db:"highest_power" json:"highest_power"`
	Title        int64  `db:"title" json:"title"`
	Language     int64  `db:"language" json:"language"`
	LoginTime    int64  `db:"login_time" json:"login_time"`
	LogoutTime   int64  `db:"logout_time" json:"logout_time"`
	CreateTime   int64  `db:"create_time" json:"create_time"`
}

func (m *Role) ToJSON() []byte {
	ret, _ := jsoniter.Marshal(m)
	return ret
}

func (m *Role) Reset() {
	*m = Role{}
}
