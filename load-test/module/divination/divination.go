package divination

import (
	"fmt"

	pbSvr "coin-server/common/proto/service"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"
)

type service struct {
}

func New(ctx *core.RoleContext) core.ILoadTestModule {
	return &service{}
}

func (s service) Process(ctx *core.RoleContext) {
	s.getDivinationInfo(ctx)
	s.divinationOnce(ctx)
	s.cheatResetTimes(ctx)
}

func (s *service) getDivinationInfo(ctx *core.RoleContext) {
	req := &pbSvr.Divination_DivinationInfoRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*pbSvr.Divination_DivinationInfoResponse)
}

func (s *service) divinationOnce(ctx *core.RoleContext) {
	req := &pbSvr.Divination_DivinationOnceRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	result := res.(*pbSvr.Divination_DivinationOnceResponse)
	if result.Info.AvailableCount == result.Info.TotalCount {
		panic(fmt.Errorf("divination faild"))
	}
}

func (s *service) cheatResetTimes(ctx *core.RoleContext) {
	req := &pbSvr.Divination_CheatResetDivinationRequest{}
	_, _, err := ctx.Request(req)
	assert.Nil(ctx, err)
}
