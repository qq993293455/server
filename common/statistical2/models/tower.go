package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 自增主键不能指定类型

type Tower struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid     string `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId  string `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	Type    int64  `json:"type" gorm:"column:type;type:int(11);not null" desc:"塔类型"`
	Level   int64  `json:"level" gorm:"column:level;type:int(11);not null" desc:"层数"`
	UseTime int64  `json:"use_time" gorm:"column:use_time;type:int(11);not null" desc:"用时"`
	IsWin   int64  `json:"is_win" gorm:"column:is_win;type:int(11);not null" desc:"0 失败 1 成功"`
}

func (i *Tower) TableName() string {
	t := i.Time.UTC().Format("20060102")
	return i.Topic() + "_" + t
}

func (i *Tower) GetRoleId() []byte {
	return utils.StringToBytes(i.RoleId)
}

func (i *Tower) Topic() string {
	return TowerTopic
}

func (i *Tower) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Tower) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		i.Time,
		i.IggId,
		i.ServerId,

		i.Xid,
		i.RoleId,
		i.Type,
		i.Level,
		i.UseTime,
		i.IsWin,
	}
}

func (i *Tower) NewModel() Model {
	return &Tower{}
}

func (i *Tower) Desc() string {
	return "爬塔"
}
