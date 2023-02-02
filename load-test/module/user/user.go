package user

import (
	"fmt"
	"sync"
	"time"

	"coin-server/common/errmsg"
	userSvrPb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	"coin-server/common/proto/recommend"
	pbSvr "coin-server/common/proto/service"
	"coin-server/common/values/enum"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"
	defaultEnv "coin-server/load-test/env"

	"github.com/gogo/protobuf/proto"
)

// 每个用户对应单独的一个service实例
type service struct {
	once sync.Once
}

func New(ctx *core.RoleContext) core.ILoadTestModule {
	s := &service{
		once: sync.Once{},
	}
	s.registerPush(ctx)
	return s
}

func (s *service) registerPush(ctx *core.RoleContext) {
	ping := &models.PING{}
	ctx.RegisterPush(ping.XXX_MessageName(), func(ctx *core.RoleContext, msg proto.Message) *errmsg.ErrMsg {
		_ = msg.(*models.PING)
		_ = ctx.AsyncSend(&models.PONG{})
		ctx.Debug("pong")
		return nil
	})
}

// 该模块的单次压力测试入口
// 一个用户一个goroutine，对应一个单独的uid，所有模块的压测都在同一个goroutine，该goroutine对应所有模块的ctx都相同
func (s *service) Process(ctx *core.RoleContext) {
	s.once.Do(func() {
		time.Sleep(10 * time.Second)
	})
	s.reset(ctx)
	//s.GetRecommend(ctx)
	//s.UserEntry(ctx)

	s.cheatAddItem(ctx)
	s.levelUpgrade(ctx)
	s.levelUpgrade(ctx)
	//s.advance(ctx)
	s.levelUpgradeMany(ctx)
	s.cheatSetLevel(ctx)
}

func (s *service) reset(ctx *core.RoleContext) {
}

func Login(ctx *core.RoleContext) *errmsg.ErrMsg {
	req := &userSvrPb.User_RoleLoginRequest{
		UserId:   ctx.UserId,
		ServerId: defaultEnv.GetTargetServerId(),
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	response := res.(*userSvrPb.User_RoleLoginResponse)
	//ctx.RoleInfo = response.Date.RoleInfo
	ctx.RoleId = response.RoleId
	return nil
}

func (s *service) GetRecommend(ctx *core.RoleContext) {
	req := &recommend.Recommend_RecommendRequest{
		Language: 1,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*recommend.Recommend_RecommendResponse)
	return
}

func (s *service) UserEntry(ctx *core.RoleContext) {
	req := &userSvrPb.User_CheatUserEntryRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*userSvrPb.User_CheatUserEntryResponse)
	return
}

func (s *service) cheatAddItem(ctx *core.RoleContext) {
	req := &pbSvr.Bag_CheatAddItemRequest{
		ItemId: enum.Gold,
		Count:  200,
	}
	_, _, err := ctx.Request(req)
	assert.Nil(ctx, err)

	req = &pbSvr.Bag_CheatAddItemRequest{
		ItemId: enum.BoundDiamond,
		Count:  400,
	}
	_, _, err = ctx.Request(req)
	assert.Nil(ctx, err)

	req = &pbSvr.Bag_CheatAddItemRequest{
		ItemId: enum.RoleExp,
		Count:  25600,
	}
	_, _, err = ctx.Request(req)
	assert.Nil(ctx, err)
}

func (s *service) levelUpgrade(ctx *core.RoleContext) {
	req := &userSvrPb.User_LevelUpgradeRequest{}
	_, _, err := ctx.Request(req)
	assert.Nil(ctx, err)
}

// func (s *service) advance(ctx *core.RoleContext) {
// 	req := &userSvrPb.User_AdvanceRequest{}
// 	_, res, err := ctx.Request(req)
// 	assert.Nil(ctx, err)
// 	result := res.(*userSvrPb.User_AdvanceResponse)
// 	if result.Level != 4 {
// 		panic(fmt.Sprintf("advance failed, advanceId: %d", result.Level))
// 	}
// }

func (s *service) levelUpgradeMany(ctx *core.RoleContext) {
	req := &userSvrPb.User_LevelUpgradeManyRequest{
		Count: 5,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	result := res.(*userSvrPb.User_LevelUpgradeManyResponse)
	if result.Level != 9 {
		panic(fmt.Sprintf("level upgrade failed, level: %d", result.Level))
	}
}

func (s *service) cheatSetLevel(ctx *core.RoleContext) {
	req := &userSvrPb.User_CheatSetLevelRequest{
		Level: 1,
	}
	_, _, err := ctx.Request(req)
	assert.Nil(ctx, err)
}
