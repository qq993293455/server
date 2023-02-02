package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 自增主键不能指定类型

type Roguelike struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid         string `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId      string `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	RoguelikeId int64  `json:"roguelike_id" gorm:"column:roguelike_id;type:int(11);not null" desc:"关卡ID"`
	Duration    int64  `json:"duration" gorm:"column:duration;type:int(11);not null" desc:"持续时间"`
	IsSucc      int64  `json:"is_succ" gorm:"column:is_succ;type:int(11);not null" desc:"是否成功0失败1成功"`
}

func (i *Roguelike) TableName() string {
	t := i.Time.UTC().Format("20060102")
	return i.Topic() + "_" + t
}

func (i *Roguelike) GetRoleId() []byte {
	return utils.StringToBytes(i.RoleId)
}

func (i *Roguelike) Topic() string {
	return BattleTopic
}

func (i *Roguelike) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Roguelike) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		i.Time,
		i.IggId,
		i.ServerId,

		i.Xid,
		i.RoleId,
		i.RoguelikeId,
		i.Duration,
		i.IsSucc,
	}
}

func (i *Roguelike) NewModel() Model {
	return &Roguelike{}
}

func (i *Roguelike) Desc() string {
	return "多人副本"
}
