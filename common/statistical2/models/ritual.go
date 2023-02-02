package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

type Ritual struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid    string        `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId values.RoleId `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	UserId string        `json:"user_id" gorm:"column:user_id;type:varchar(20);not null" desc:"用户id"`
	HeroId int64         `json:"hero_id" gorm:"column:hero_id;type:int(11);not null" desc:"救治的英雄id"`
}

func (m *Ritual) TableName() string {
	t := m.Time.UTC().Format("20060102")
	return m.Topic() + "_" + t
}

func (m *Ritual) GetRoleId() []byte {
	return utils.StringToBytes(m.RoleId)
}

func (m *Ritual) Topic() string {
	return RitualTopic
}

func (m *Ritual) ToJson() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Ritual) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		m.Time,
		m.IggId,
		m.ServerId,

		m.Xid,
		m.RoleId,
		m.UserId,
		m.HeroId,
	}
}

func (m *Ritual) NewModel() Model {
	return &Ritual{}
}

func (m *Ritual) Desc() string {
	return "救治仪式解锁角色"
}
