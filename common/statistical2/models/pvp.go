package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 自增主键不能指定类型

type PvpEventTyp int64

const (
	PvpEventTypStart = iota + 1
	PvpEventTypWin
	PvpEventTypLose
	PvpEventTypExit
)

type Pvp struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid      string      `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId   string      `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	Opponent string      `json:"opponent" gorm:"column:opponent;type:varchar(20);not null" desc:"对手Id"`
	Rank     int64       `json:"rank" gorm:"column:rank;type:int(11);not null" desc:"阶位Id"`
	Point    int64       `json:"point" gorm:"column:point;type:int(11);not null" desc:"积分变化"`
	EventTyp PvpEventTyp `json:"event_typ" gorm:"column:event_typ;type:int(11);not null" desc:"事件类型 1:开始，2:胜利，3:失败，4:退出"`
}

func (i *Pvp) TableName() string {
	t := i.Time.UTC().Format("20060102")
	return i.Topic() + "_" + t
}

func (i *Pvp) GetRoleId() []byte {
	return utils.StringToBytes(i.RoleId)
}

func (i *Pvp) Topic() string {
	return PVPTopic
}

func (i *Pvp) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Pvp) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		i.Time,
		i.IggId,
		i.ServerId,

		i.Xid,
		i.RoleId,
		i.Opponent,
		i.Rank,
		i.Point,
		i.EventTyp,
	}
}

func (i *Pvp) NewModel() Model {
	return &Pvp{}
}

func (i *Pvp) Desc() string {
	return "PVP"
}
