package maintask

import (
	modelsPb "coin-server/common/proto/models"
	pbSvr "coin-server/common/proto/service"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"
)

// 每个用户对应单独的一个service实例
type service struct {
	task *modelsPb.MainTask
}

func New(ctx *core.RoleContext) core.ILoadTestModule {
	return &service{}
}

func (s *service) Process(ctx *core.RoleContext) {
	s.reset(ctx)
	s.acceptMainTask(ctx)
	s.cheatFinishMainTask(ctx)
	s.finishMainTask(ctx)
}

func (s *service) reset(ctx *core.RoleContext) {
	req := &pbSvr.MainTask_CheatResetMainTaskRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.MainTask_CheatResetMainTaskResponse)
}

func (s *service) getMainTask(ctx *core.RoleContext) {
	req := &pbSvr.MainTask_GetMainTaskRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	s.task = res.(*pbSvr.MainTask_GetMainTaskResponse).Task
}

func (s *service) acceptMainTask(ctx *core.RoleContext) {
	req := &pbSvr.MainTask_AcceptMainTaskRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.MainTask_AcceptMainTaskResponse)
}

func (s *service) cheatFinishMainTask(ctx *core.RoleContext) {
	req := &pbSvr.MainTask_CheatFinishMainTaskRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.MainTask_CheatFinishMainTaskResponse)
}

func (s *service) finishMainTask(ctx *core.RoleContext) {
	req := &pbSvr.MainTask_FinishMainTaskRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.MainTask_FinishMainTaskResponse)
}
