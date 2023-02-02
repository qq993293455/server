package models

import (
	"strconv"
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

type TaskAction int64

const (
	TaskActionAccept = iota + 1 // 接取
	TaskActionFinish            // 完成
)

type MainTask struct {
	// 必须字段
	IggId         int64     `json:"igg_id"`
	EventTime     time.Time `json:"event_time"`     // event_time 事件时间
	GwId          int64     `json:"gw_id"`          // 所属大世界
	EventType     string    `json:"event_type"`     // 为event_category事件分类里面的小类
	EventCategory string    `json:"event_category"` // 事件定义时填写的分类信息

	// 其它字段
	RoleId  values.RoleId `json:"f4"` // 角色id
	UserId  string        `json:"f5"` // 用户id
	TaskId  int64         `json:"f6"` // 任务id
	Chapter int64         `json:"f7"` // 任务章节
	Index   int64         `json:"f8"` // 任务章节索引
	Action  TaskAction    `json:"f9"` // 动作 1:接取 2:完成
}

func (l *MainTask) HashKey() (dst []byte) {
	return strconv.AppendInt(dst, l.IggId, 10)
}

func (l *MainTask) ToJson() []byte {
	ret, err := json.Marshal(l)
	utils.Must(err)
	return ret
}

func (l *MainTask) GetEventType() string {
	return MainTaskEventType
}

func (l *MainTask) Preset() {
	if l.EventType == "" {
		l.EventType = l.GetEventType()
	}
	if l.EventCategory == "" {
		l.EventCategory = CommonEventCategory
	}
}

func (m *MainTask) Desc() string {
	return "主线任务"
}
