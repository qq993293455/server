package event

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

//type KillMonster struct {
//	RoleId    values.RoleId
//	MonsterId values.Integer
//	Num       values.Integer
//}

type Submit struct {
	TaskId values.TaskId
	ItemId values.ItemId
	Count  values.Integer
	Kind   models.TaskKind
	Typ    models.TaskType
}

type Gather struct {
	TaskId values.TaskId
	Kind   models.TaskKind
}

type Move struct {
	TaskId values.TaskId
	Id     values.Integer
	Count  values.Integer
	Kind   models.TaskKind
}

type Collect struct {
	ItemId values.ItemId
	Count  values.Integer
}

type Talk struct {
	DialogId     values.Integer
	HeadDialogId values.Integer // 头部对话id
	IsEnd        bool           // 是否是结束选项
	OptIdx       values.Integer
	TaskId       values.TaskId
	Kind         models.TaskKind
	Typ          models.TaskType
}

type KillMonster struct {
	TaskId    values.TaskId
	MonsterId values.Integer
	Count     values.Integer
	Kind      models.TaskKind
	Typ       models.TaskType
}

type Vector2 struct {
	X, Y float64
}

// Teleport 玩家传送到指定位置
type Teleport struct {
	RoleId values.RoleId
	MapId  int64
	Pos    Vector2
}

// UpdateTarget 更新到目标条件, 增量更新
type UpdateTarget struct {
	RoleId  values.RoleId
	Typ     models.TaskType
	Id      values.Integer
	Count   values.Integer
	Replace bool
}

// TargetUpdate 目标条件更新发出事件（根据任务类型看是增量还是累计）
type TargetUpdate struct {
	Typ          models.TaskType
	Id           values.Integer
	Count        values.Integer
	Incr         values.Integer
	IsAccumulate bool // 增量还是累计
	IsReplace    bool
}

type TaskUpdate struct {
	Sys  values.SystemId
	Task *models.Task
}
