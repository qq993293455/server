package event

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type PlaneDungeonFinish struct {
	RoleId    values.RoleId
	PlaneId   values.Integer  // 位面副本id
	IsSuccess bool            // 是否成功
	TaskType  models.TaskType // 任务类型
}
