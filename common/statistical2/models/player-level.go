package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

type PlayerLevel struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid    string        `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId values.RoleId `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	Level  values.Level  `json:"level" gorm:"column:level;type:int(11);not null" desc:"等级"`
}

func (p *PlayerLevel) TableName() string {
	t := p.Time.UTC().Format("20060102")
	return p.Topic() + "_" + t
}

func (p *PlayerLevel) GetRoleId() []byte {
	return utils.StringToBytes(p.RoleId)
}

func (p *PlayerLevel) Topic() string {
	return PlayerLevelTopic
}

func (p *PlayerLevel) ToJson() ([]byte, error) {
	return json.Marshal(p)
}

func (p *PlayerLevel) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致
	return []interface{}{
		p.Time,
		p.IggId,
		p.ServerId,

		p.Xid,
		p.RoleId,
		p.Level,
	}
}
func (p *PlayerLevel) NewModel() Model {
	return &PlayerLevel{}
}

func (p *PlayerLevel) Desc() string {
	return "玩家等级提升"
}
