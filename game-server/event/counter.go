package event

import (
	protosvc "coin-server/common/proto/service"
	"coin-server/common/values"
)

type CountTyp uint8

// 成就统计数据变化
type CounterCntChangeData struct {
	AchievementId values.AchievementId
	Val            values.Integer
	CountTyp       protosvc.CountTyp
}
