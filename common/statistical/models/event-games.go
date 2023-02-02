package models

import (
	"strconv"
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 小游戏

type EventGames struct {
	// 必须字段
	IggId         int64     `json:"igg_id"`
	EventTime     time.Time `json:"event_time"`     // event_time 事件时间
	GwId          int64     `json:"gw_id"`          // 所属大世界
	EventType     string    `json:"event_type"`     // 为event_category事件分类里面的小类
	EventCategory string    `json:"event_category"` // 事件定义时填写的分类信息

	// 其它字段
	RoleId values.RoleId `json:"f4"` // 角色id
	GameId int64         `json:"f5"` // 游戏id
	IsWin  int64         `json:"f6"` // 是否成功
}

func (l *EventGames) HashKey() (dst []byte) {
	return strconv.AppendInt(dst, l.IggId, 10)
}

func (l *EventGames) ToJson() []byte {
	ret, err := json.Marshal(l)
	utils.Must(err)
	return ret
}

func (l *EventGames) GetEventType() string {
	return EventGamesEventType
}

func (l *EventGames) Preset() {
	if l.EventType == "" {
		l.EventType = l.GetEventType()
	}
	if l.EventCategory == "" {
		l.EventCategory = CommonEventCategory
	}
}
