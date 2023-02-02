package models

import (
	"strconv"
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 自增主键不能指定类型

type Arena struct {
	// 必须字段
	IggId         int64     `json:"igg_id"`
	EventTime     time.Time `json:"event_time"`     // event_time 事件时间
	GwId          int64     `json:"gw_id"`          // 所属大世界
	EventType     string    `json:"event_type"`     // 为event_category事件分类里面的小类
	EventCategory string    `json:"event_category"` // 事件定义时填写的分类信息

	// 其它字段
	RoleId      values.RoleId `json:"f4"`  // 角色id
	OtherRoleId values.RoleId `json:"f5"`  // 挑战角色id
	Type        int64         `json:"f6"`  // 竞技场类型
	StartTime   int64         `json:"f7"`  // 开始时间
	EndTime     int64         `json:"f8"`  // 结束时间
	IsOverTime  int64         `json:"f9"`  // 是否超时
	IsWin       int64         `json:"f10"` // 是否成功
}

func (l *Arena) HashKey() (dst []byte) {
	return strconv.AppendInt(dst, l.IggId, 10)
}

func (l *Arena) ToJson() []byte {
	ret, err := json.Marshal(l)
	utils.Must(err)
	return ret
}

func (l *Arena) GetEventType() string {
	return ArenaEventType
}

func (l *Arena) Preset() {
	if l.EventType == "" {
		l.EventType = l.GetEventType()
	}
	if l.EventCategory == "" {
		l.EventCategory = CommonEventCategory
	}
}
