package shop

import (
	pbSvr "coin-server/common/proto/service"
	"coin-server/common/values/enum"
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
	s.cheatAddDiamond(ctx, 30)
	s.getShopAndBuy(ctx)
	s.refresh(ctx)
	s.reqCnt++
	if s.reqCnt >= 10 {
		s.cheatClearRefreshCnt(ctx)
	}
}

func (s *service) getShopAndBuy(ctx *core.RoleContext) {
	req := &pbSvr.Shop_GetShopListRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	data := res.(*pbSvr.Shop_GetShopListResponse)
	for idx := range data.Detail {
		buyReq := &pbSvr.Shop_BuyRequest{
			DetailIdx: int64(idx),
		}
		_, buyRes, err := ctx.Request(buyReq)
		assert.Nil(ctx, err)
		assert.NotNil(ctx, res)
		_ = buyRes.(*pbSvr.Shop_BuyResponse)
	}
}

func (s *service) refresh(ctx *core.RoleContext) {
	req := &pbSvr.Shop_RefreshRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Shop_RefreshResponse)
}

func (s *service) cheatClearRefreshCnt(ctx *core.RoleContext) {
	req := &pbSvr.Shop_CheatClearRefreshCntRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Shop_CheatClearRefreshCntResponse)
}

func (s *service) cheatAddDiamond(ctx *core.RoleContext, cnt int64) {
	req := &pbSvr.Bag_CheatAddItemRequest{
		ItemId: enum.BoundDiamond,
		Count:  cnt,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Bag_CheatAddItemResponse)
}
