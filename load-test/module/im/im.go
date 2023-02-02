package im

import (
	pbSvr "coin-server/common/proto/service"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"
)

// 每个用户对应单独的一个service实例
type service struct {
	giftID string
}

func New(ctx *core.RoleContext) core.ILoadTestModule {
	return &service{}
}

func (s *service) Process(ctx *core.RoleContext) {
	s.reset(ctx)
	s.cheatAddGift(ctx)
	s.drawGift(ctx)
	s.drawGift(ctx)
}

func (s *service) reset(ctx *core.RoleContext) {
	//s.giftID = ""
}

func (s *service) cheatAddGift(ctx *core.RoleContext) {
	req := &pbSvr.Im_CheatSendGiftRequest{
		No: 1000,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	s.giftID = res.(*pbSvr.Im_CheatSendGiftResponse).Id
}

func (s *service) drawGift(ctx *core.RoleContext) {
	req := &pbSvr.Im_DrawGiftRequest{Id: s.giftID}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Im_DrawGiftResponse)
}
