package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 自增主键不能指定类型

type EventGames struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid    string `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId string `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	GameId int64  `json:"game_id" gorm:"column:game_id;type:int(11);not null" desc:"游戏id"`
	IsWin  int64  `json:"is_win" gorm:"column:is_win;type:int(11);not null" desc:"是否成功"`
}

func (i *EventGames) TableName() string {
	t := i.Time.UTC().Format("20060102")
	return i.Topic() + "_" + t
}

func (i *EventGames) GetRoleId() []byte {
	return utils.StringToBytes(i.RoleId)
}

func (i *EventGames) Topic() string {
	return EventGamesTopic
}

func (i *EventGames) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (i *EventGames) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		i.Time,
		i.IggId,
		i.ServerId,

		i.Xid,
		i.RoleId,
		i.GameId,
		i.IsWin,
	}
}

func (i *EventGames) NewModel() Model {
	return &EventGames{}
}

func (i *EventGames) Desc() string {
	return "小游戏"
}
