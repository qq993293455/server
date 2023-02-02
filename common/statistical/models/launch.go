package models

import (
	"strconv"
	"time"

	"coin-server/common/utils"

	json "github.com/json-iterator/go"
)

// 登录前打点

type Launch struct {
	// 必须字段
	IggId         int64     `json:"igg_id"`
	EventTime     time.Time `json:"event_time"`     // event_time 事件时间
	GwId          int64     `json:"gw_id"`          // 所属大世界
	EventType     string    `json:"event_type"`     // 为event_category事件分类里面的小类
	EventCategory string    `json:"event_category"` // 事件定义时填写的分类信息

	// 其它字段
	UDId     string `json:"f4"` // udid
	GameId   string `json:"f5"` // gameid
	Device   string `json:"f6"` // 设备号
	Progress string `json:"f7"` // 进度
}

func (l *Launch) HashKey() (dst []byte) {
	return strconv.AppendInt(dst, l.IggId, 10)
}

func (l *Launch) ToJson() []byte {
	ret, err := json.Marshal(l)
	utils.Must(err)
	return ret
}

func (l *Launch) GetEventType() string {
	return LaunchEventType
}

func (l *Launch) Preset() {
	if l.EventType == "" {
		l.EventType = l.GetEventType()
	}
	if l.EventCategory == "" {
		l.EventCategory = CommonEventCategory
	}
}
