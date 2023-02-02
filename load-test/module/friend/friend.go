package friend

import (
	"strings"

	pbSvr "coin-server/common/proto/less_service"
	_ "coin-server/common/proto/service"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"
)

type service struct {
	agreeUser  map[*testUser]struct{}
	rejectUser map[*testUser]struct{}
	blackUser  map[*testUser]struct{}
}

func New(ctx *core.RoleContext) core.ILoadTestModule {
	s := &service{
		agreeUser:  map[*testUser]struct{}{},
		rejectUser: map[*testUser]struct{}{},
		blackUser:  map[*testUser]struct{}{},
	}
	for i := 0; i < 5; i++ {
		s.agreeUser[newTestUser()] = struct{}{}
	}
	for i := 0; i < 5; i++ {
		s.rejectUser[newTestUser()] = struct{}{}
	}
	for i := 0; i < 5; i++ {
		s.blackUser[newTestUser()] = struct{}{}
	}

	return s
}

func (s *service) Process(ctx *core.RoleContext) {
	s.addRequest(ctx)
	s.agreeRequest(ctx)

	s.rejectRequest(ctx)

	s.addBlack(ctx)
	s.sendPoint(ctx)
	s.recvPoint(ctx)
	s.getList(ctx)
	s.clearFriend(ctx)
}

func (s *service) getList(ctx *core.RoleContext) {
	req := &pbSvr.Friend_GetFriendListRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Friend_GetFriendListResponse)
}

func (s *service) addRequest(ctx *core.RoleContext) {
	for role := range s.agreeUser {
		req := &pbSvr.Friend_AddRequestRequest{
			RoleId: role.ctx.RoleId,
		}
		_, res, err := ctx.Request(req)
		assert.Nil(ctx, err)
		assert.NotNil(ctx, res)
		_ = res.(*pbSvr.Friend_AddRequestResponse)
	}
	for role := range s.rejectUser {
		req := &pbSvr.Friend_AddRequestRequest{
			RoleId: role.ctx.RoleId,
		}
		_, res, err := ctx.Request(req)
		assert.Nil(ctx, err)
		assert.NotNil(ctx, res)
		_ = res.(*pbSvr.Friend_AddRequestResponse)
	}
	for role := range s.blackUser {
		req := &pbSvr.Friend_AddRequestRequest{
			RoleId: role.ctx.RoleId,
		}
		_, res, err := ctx.Request(req)
		assert.Nil(ctx, err)
		assert.NotNil(ctx, res)
		_ = res.(*pbSvr.Friend_AddRequestResponse)
	}
}

func (s *service) agreeRequest(ctx *core.RoleContext) {
	for tar := range s.agreeUser {
		req := &pbSvr.Friend_ConfirmRequestRequest{
			RoleId:    ctx.RoleId,
			IsConform: true,
		}
		_, res, err := tar.ctx.Request(req)
		assert.Nil(tar.ctx, err)
		assert.NotNil(tar.ctx, res)
		_ = res.(*pbSvr.Friend_ConfirmRequestResponse)
	}
	for tar := range s.blackUser {
		req := &pbSvr.Friend_ConfirmRequestRequest{
			RoleId:    ctx.RoleId,
			IsConform: true,
		}
		_, res, err := tar.ctx.Request(req)
		assert.Nil(tar.ctx, err)
		assert.NotNil(tar.ctx, res)
		_ = res.(*pbSvr.Friend_ConfirmRequestResponse)
	}
}

func (s *service) rejectRequest(ctx *core.RoleContext) {
	for tar := range s.rejectUser {
		req := &pbSvr.Friend_ConfirmRequestRequest{
			RoleId:    ctx.RoleId,
			IsConform: false,
		}
		_, res, err := tar.ctx.Request(req)
		assert.Nil(tar.ctx, err)
		assert.NotNil(tar.ctx, res)
		_ = res.(*pbSvr.Friend_ConfirmRequestResponse)
	}
}

func (s *service) addBlack(ctx *core.RoleContext) {
	for tar := range s.blackUser {
		req := &pbSvr.Friend_AddBlackRequest{
			RoleId: tar.ctx.RoleId,
		}
		_, res, err := ctx.Request(req)
		assert.Nil(ctx, err)
		assert.NotNil(ctx, res)
		_ = res.(*pbSvr.Friend_AddBlackResponse)
	}
}

func (s *service) sendPoint(ctx *core.RoleContext) {
	roleIds, idx := make([]string, len(s.agreeUser)), 0
	for tar := range s.agreeUser {
		roleIds[idx] = tar.ctx.RoleId
		idx++
	}
	req := &pbSvr.Friend_SendPointRequest{
		RoleIds: strings.Join(roleIds, ","),
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	assert.NotNil(ctx, res)
	_ = res.(*pbSvr.Friend_SendPointResponse)
}

func (s *service) recvPoint(ctx *core.RoleContext) {
	for tar := range s.agreeUser {
		req := &pbSvr.Friend_RecvPointRequest{
			RoleIds: strings.Join([]string{ctx.RoleId}, ","),
		}
		_, res, err := tar.ctx.Request(req)
		assert.Nil(tar.ctx, err)
		assert.NotNil(tar.ctx, res)
		_ = res.(*pbSvr.Friend_RecvPointResponse)
	}
}

func (s *service) clearFriend(ctx *core.RoleContext) {
	for role := range s.agreeUser {
		req := &pbSvr.Friend_DeleteRequest{
			RoleId: role.ctx.RoleId,
		}
		_, res, err := ctx.Request(req)
		assert.Nil(ctx, err)
		assert.NotNil(ctx, res)
		_ = res.(*pbSvr.Friend_DeleteResponse)
	}
	for role := range s.blackUser {
		req := &pbSvr.Friend_RemoveBlackRequest{
			RoleId: role.ctx.RoleId,
		}
		_, res, err := ctx.Request(req)
		assert.Nil(ctx, err)
		assert.NotNil(ctx, res)
		_ = res.(*pbSvr.Friend_RemoveBlackResponse)
	}
}