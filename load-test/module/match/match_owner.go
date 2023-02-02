package match

import (
	"math/rand"
	"time"

	pbSvr "coin-server/common/proto/dungeon_match"
	"coin-server/common/proto/models"
	_ "coin-server/common/proto/service"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"
)

type ownerService struct {
}

func NewOwner(ctx *core.RoleContext) core.ILoadTestModule {
	return &ownerService{}
}

func (s *ownerService) Process(ctx *core.RoleContext) {
	origin := s.create(ctx)
	r1 := s.cetCurr(ctx)
	assert.True(ctx, r1.Room.RoomId == origin.Room.RoomId)
	assert.True(ctx, r1.Room.OwnerId == ctx.RoleId)
	s.getRoleRoom(ctx, r1.Room)
	time.Sleep(time.Duration(rand.Intn(10)+5) * time.Second)
	s.dismiss(ctx)
}

func (s *ownerService) create(ctx *core.RoleContext) *pbSvr.DungeonMatch_CreateRoomResponse {
	req := &pbSvr.DungeonMatch_CreateRoomRequest{
		DungeonId: 1,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	return res.(*pbSvr.DungeonMatch_CreateRoomResponse)
}

func (s *ownerService) cetCurr(ctx *core.RoleContext) *pbSvr.DungeonMatch_GetCurrRoomResponse {
	req := &pbSvr.DungeonMatch_GetCurrRoomRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	return res.(*pbSvr.DungeonMatch_GetCurrRoomResponse)
}

func (s *ownerService) getRoleRoom(ctx *core.RoleContext, r *models.Room) {
	req := &pbSvr.DungeonMatch_GetRoleRoomRequest{
		RoleId: r.OwnerId,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	resp := res.(*pbSvr.DungeonMatch_GetRoleRoomResponse)
	assert.True(ctx, resp.Room.OwnerId == r.OwnerId)
}

func (s *ownerService) dismiss(ctx *core.RoleContext) {
	req := &pbSvr.DungeonMatch_DismissRoomRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.DungeonMatch_DismissRoomResponse)
}
