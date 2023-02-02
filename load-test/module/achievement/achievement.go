package achievement

import (
	pbSvr "coin-server/common/proto/service"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"
)

// 每个用户对应单独的一个service实例
type service struct {
	reqCnt int64
}

func New(ctx *core.RoleContext) core.ILoadTestModule {
	s := &service{}
	return s
}

func (s *service) Process(ctx *core.RoleContext) {
	s.cheatAddCnt(ctx)
	s.getList(ctx)
	s.getDetail(ctx)
	s.reqCnt++
	if s.reqCnt >= 50 {
		s.cheatClear(ctx)
		s.reqCnt = 0
	}
}

func (s *service) cheatClear(ctx *core.RoleContext) {
	req := &pbSvr.Achievement_CheatClearRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Achievement_CheatClearResponse)
}

func (s *service) cheatAddCnt(ctx *core.RoleContext) {
	req := &pbSvr.Achievement_CheatAddCounterRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Achievement_CheatAddCounterResponse)
}

func (s *service) getList(ctx *core.RoleContext) {
	req := &pbSvr.Achievement_GetAchievementListRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Achievement_GetAchievementListResponse)
}

func (s *service) getDetail(ctx *core.RoleContext) {
	req := &pbSvr.Achievement_GetAchievementDetailRequest{
		AchievementId: 1,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Achievement_GetAchievementDetailResponse)
}
