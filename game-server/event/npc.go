package event

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type NpcDungeonFinish struct {
	RoleId    values.RoleId
	DungeonId values.Integer  // 副本id
	IsSuccess bool            // 是否成功
	TaskType  models.TaskType // 任务类型
}
