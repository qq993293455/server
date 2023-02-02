package models

import (
	"strconv"
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 自增主键不能指定类型

type BattleEventTyp int64

const (
	BattleEventTypStart = iota + 1
	BattleEventTypWin
	BattleEventTypLose
	BattleEventTypExit
)

type Battle struct {
	// 必须字段
	IggId         int64     `json:"igg_id"`
	EventTime     time.Time `json:"event_time"`     // event_time 事件时间
	GwId          int64     `json:"gw_id"`          // 所属大世界
	EventType     string    `json:"event_type"`     // 为event_category事件分类里面的小类
	EventCategory string    `json:"event_category"` // 事件定义时填写的分类信息

	// 其它字段
	RoleId   values.RoleId  `json:"f4"` // 角色id
	Mission  int64          `json:"f5"` // 关卡
	EventTyp BattleEventTyp `json:"f6"` // 事件类型
}

func (l *Battle) HashKey() (dst []byte) {
	return strconv.AppendInt(dst, l.IggId, 10)
}

func (l *Battle) ToJson() []byte {
	ret, err := json.Marshal(l)
	utils.Must(err)
	return ret
}

func (l *Battle) GetEventType() string {
	return BattleEventType
}

func (l *Battle) Preset() {
	if l.EventType == "" {
		l.EventType = l.GetEventType()
	}
	if l.EventCategory == "" {
		l.EventCategory = CommonEventCategory
	}
}
