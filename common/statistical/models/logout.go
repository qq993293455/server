package models

import (
	"strconv"
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

type Logout struct {
	// 必须字段
	IggId         int64     `json:"igg_id"`
	EventType     string    `json:"event_type"`     // 为event_category事件分类里面的小类
	EventTime     time.Time `json:"event_time"`     // event_time 事件时间
	GwId          int64     `json:"gw_id"`          // 所属大世界
	EventCategory string    `json:"event_category"` // 事件定义时填写的分类信息

	// 其它字段
	RoleId        values.RoleId `json:"f4"`  // 角色id
	UserId        string        `json:"f5"`  // 用户id
	DeviceId      string        `json:"f6"`  // 设备id
	IP            string        `json:"f7"`  // 用户ip地址
	RuleVersion   string        `json:"f8"`  // 规则版本
	OnlineSeconds int64         `json:"f9"`  // 单次登录时长
	ClientVersion string        `json:"f10"` // 客户端版本
}

func (l *Logout) HashKey() (dst []byte) {
	return strconv.AppendInt(dst, l.IggId, 10)
}

func (l *Logout) ToJson() []byte {
	ret, err := json.Marshal(l)
	utils.Must(err)
	return ret
}

func (l *Logout) GetEventType() string {
	return LogoutEventType
}

func (l *Logout) Preset() {
	if l.EventType == "" {
		l.EventType = l.GetEventType()
	}
	if l.EventCategory == "" {
		l.EventCategory = CommonEventCategory
	}
}
