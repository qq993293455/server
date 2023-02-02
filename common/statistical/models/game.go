package models

import (
	"strconv"
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 自增主键不能指定类型

type Game struct {
	// 必须字段
	IggId         int64     `json:"igg_id"`
	EventTime     time.Time `json:"event_time"`     // event_time 事件时间
	GwId          int64     `json:"gw_id"`          // 所属大世界
	EventType     string    `json:"event_type"`     // 为event_category事件分类里面的小类
	EventCategory string    `json:"event_category"` // 事件定义时填写的分类信息

	// 其它字段
	RoleId           values.RoleId `json:"f4"` // 角色id
	Memory           string        `json:"f5"` // 设备内存大小
	StoryChapter     string        `json:"f6"` // 对话所属章节
	SkipChapter      string        `json:"f7"` // 对话所属章节跳过
	TutorialStep     string        `json:"f8"` // 教程成功操作
	TutorialSkipStep string        `json:"f9"` // 教程跳过操作
}

func (l *Game) HashKey() (dst []byte) {
	return strconv.AppendInt(dst, l.IggId, 10)
}

func (l *Game) ToJson() []byte {
	ret, err := json.Marshal(l)
	utils.Must(err)
	return ret
}

func (l *Game) GetEventType() string {
	return GameEventType
}

func (l *Game) Preset() {
	if l.EventType == "" {
		l.EventType = l.GetEventType()
	}
	if l.EventCategory == "" {
		l.EventCategory = CommonEventCategory
	}
}
