package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 自增主键不能指定类型

type Arena struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid         string `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId      string `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	OtherRoleId string `json:"other_role_id" gorm:"column:other_role_id;type:varchar(20);not null" desc:"挑战角色id"`
	Type        int64  `json:"type" gorm:"column:type;type:int(11);not null" desc:"竞技场类型"`
	StartTime   int64  `json:"start_time" gorm:"column:start_time;type:int(32);not null" desc:"开始时间"`
	EndTime     int64  `json:"end_time" gorm:"column:end_time;type:int(32);not null" desc:"结束时间"`
	IsOverTime  int64  `json:"is_over_time" gorm:"column:is_over_time;type:int(11);not null" desc:"是否超时"`
	IsWin       int64  `json:"is_win" gorm:"column:is_win;type:int(11);not null" desc:"是否成功"`
}

func (i *Arena) TableName() string {
	t := i.Time.UTC().Format("20060102")
	return i.Topic() + "_" + t
}

func (i *Arena) GetRoleId() []byte {
	return utils.StringToBytes(i.RoleId)
}

func (i *Arena) Topic() string {
	return ArenaTopic
}

func (i *Arena) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Arena) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		i.Time,
		i.IggId,
		i.ServerId,

		i.Xid,
		i.RoleId,
		i.OtherRoleId,
		i.Type,
		i.StartTime,
		i.EndTime,
		i.IsOverTime,
		i.IsWin,
	}
}

func (i *Arena) NewModel() Model {
	return &Arena{}
}

func (i *Arena) Desc() string {
	return "竞技场"
}
