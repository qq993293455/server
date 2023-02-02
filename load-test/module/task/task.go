package task

import (
	pbSvr "coin-server/common/proto/service"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"
)

// 每个用户对应单独的一个service实例
type service struct {
}

func New(ctx *core.RoleContext) core.ILoadTestModule {
	s := &service{}
	s.registerPush(ctx)
	return s
}

func (s *service) Process(ctx *core.RoleContext) {
	s.reset(ctx)
	s.getAllTask(ctx)
	s.getNPCTask(ctx)
	s.cheatUpdate(ctx)
}

func (s *service) reset(ctx *core.RoleContext) {
}

func (s *service) registerPush(ctx *core.RoleContext) {
}

func (s *service) getAllTask(ctx *core.RoleContext) {
	req := &pbSvr.LoopTask_GetTaskRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.LoopTask_GetTaskResponse)
}

func (s *service) getNPCTask(ctx *core.RoleContext) {
	req := &pbSvr.LoopTask_GetNPCTaskRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.LoopTask_GetNPCTaskResponse)
}

func (s *service) taskProcess(ctx *core.RoleContext) {
	req := &pbSvr.LoopTask_CheatAddNPCTaskRequest{NpcId: 1, TaskId: 10100}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.LoopTask_CheatAddNPCTaskResponse)

	req2 := &pbSvr.LoopTask_ReceiveNPCTaskRequest{NpcId: 1, TaskId: 10100}
	_, res2, err2 := ctx.Request(req2)
	assert.Nil(ctx, err2)
	assert.NotNil(ctx, res2)
	_ = res2.(*pbSvr.LoopTask_ReceiveNPCTaskResponse)

	req3 := &pbSvr.LoopTask_CheatUpdateTaskRequest{TaskId: 10100}
	_, res3, err3 := ctx.Request(req3)
	assert.Nil(ctx, err3)
	assert.NotNil(ctx, res3)
	_ = res3.(*pbSvr.LoopTask_CheatUpdateTaskResponse)

	req4 := &pbSvr.LoopTask_ClearNPCTaskRequest{NpcId: 1, TaskId: 10100}
	_, res4, err4 := ctx.Request(req4)
	assert.Nil(ctx, err4)
	assert.NotNil(ctx, res4)
	_ = res2.(*pbSvr.LoopTask_ClearNPCTaskResponse)
}

func (s *service) cheatUpdate(ctx *core.RoleContext) {
	req := &pbSvr.LoopTask_CheatUpdateTaskRequest{TaskId: 10051}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.LoopTask_CheatUpdateTaskResponse)
}
