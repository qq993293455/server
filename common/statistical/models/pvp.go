package models

import (
	"strconv"
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
	IggId         int64     `json:"igg_id"`
	EventTime     time.Time `json:"event_time"`     // event_time 事件时间
	GwId          int64     `json:"gw_id"`          // 所属大世界
	EventType     string    `json:"event_type"`     // 为event_category事件分类里面的小类
	EventCategory string    `json:"event_category"` // 事件定义时填写的分类信息

	// 其它字段
	RoleId   values.RoleId `json:"f4"` // 角色id
	Opponent string        `json:"f5"` // 对手Id
	Rank     int64         `json:"f6"` // 阶位Id
	Point    int64         `json:"f7"` // 积分变化
	EventTyp PvpEventTyp   `json:"f8"` // 事件类型 1:开始，2:胜利，3:失败，4:退出
}

func (l *Pvp) HashKey() (dst []byte) {
	return strconv.AppendInt(dst, l.IggId, 10)
}

func (l *Pvp) ToJson() []byte {
	ret, err := json.Marshal(l)
	utils.Must(err)
	return ret
}

func (l *Pvp) GetEventType() string {
	return PVPEventType
}

func (l *Pvp) Preset() {
	if l.EventType == "" {
		l.EventType = l.GetEventType()
	}
	if l.EventCategory == "" {
		l.EventCategory = CommonEventCategory
	}
}
