package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	event2 "coin-server/common/values/event"
)

type TaskService interface {
	// 获取所有目标条件进度
	GetCounter(ctx *ctx.Context) (map[string]*pbdao.CondCounter, *errmsg.ErrMsg)
	// 根据typ获取目标条件进度
	GetCounterByType(ctx *ctx.Context, taskType models.TaskType) (map[values.Integer]values.Integer, *errmsg.ErrMsg)
	GetCounterByTypeList(ctx *ctx.Context, taskTypes []models.TaskType) (map[models.TaskType]map[values.Integer]values.Integer, *errmsg.ErrMsg)
	// 更新目标条件
	UpdateTarget(ctx *ctx.Context, roleId values.RoleId, Typ models.TaskType, id, cnt values.Integer, replace ...bool)
	UpdateTargets(ctx *ctx.Context, roleId values.RoleId, tasks map[models.TaskType]*models.TaskUpdate)
	// 根据条件type、条件id、条件数量注册感兴趣的事件，若对所有数量感兴趣，则v=event2.allcond
	RegisterCondHandler(taskType models.TaskType, targetId, targetCnt values.Integer, handler event2.CondHandler, args any)
	// 检查任务条件是否达成
	CheckCondition(ctx *ctx.Context, taskType models.TaskType, targetId, targetCnt values.Integer) (values.Integer, bool, *errmsg.ErrMsg)
}
