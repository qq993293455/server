package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 自增主键不能指定类型

type SubItem struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid    string `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId string `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	ItemId int64  `json:"item_id" gorm:"column:item_id;type:int(11);not null" desc:"道具id"`
	Num    int64  `json:"num" gorm:"column:num;type:int(11);not null" desc:"数量"`
	Msg    string `json:"msg" gorm:"column:msg;type:varchar(64);not null" desc:"消息id"`
}

func (i *SubItem) TableName() string {
	t := i.Time.UTC().Format("20060102")
	return i.Topic() + "_" + t
}

func (i *SubItem) GetRoleId() []byte {
	return utils.StringToBytes(i.RoleId)
}

func (i *SubItem) Topic() string {
	return SubItemTopic
}

func (i *SubItem) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (i *SubItem) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		i.Time,
		i.IggId,
		i.ServerId,

		i.Xid,
		i.RoleId,
		i.ItemId,
		i.Num,
		i.Msg,
	}
}

func (i *SubItem) NewModel() Model {
	return &SubItem{}
}

func (i *SubItem) Desc() string {
	return "道具消耗"
}
