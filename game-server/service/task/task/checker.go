package task

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type ConditionChecker struct {
	svc *Service
}

func NewConditionChecker(svc *Service) *ConditionChecker {
	checker := &ConditionChecker{svc: svc}
	return checker
}

// 当前值，是否完成
func (c *ConditionChecker) CheckCondition(ctx *ctx.Context, typ models.TaskType, targetId values.Integer, targetCnt values.Integer) (values.Integer, bool, *errmsg.ErrMsg) {
	switch typ {
	case models.TaskType_TaskLevel:
		return c.CheckLevel(ctx, targetCnt)
	}
	return 0, false, nil
}

func (c *ConditionChecker) CheckLevel(ctx *ctx.Context, level values.Level) (values.Integer, bool, *errmsg.ErrMsg) {
	l, err := c.svc.UserService.GetLevel(ctx, ctx.RoleId)
	if err != nil {
		return 0, false, err
	}
	if l >= level {
		return level, true, nil
	}
	return l, false, nil
}
