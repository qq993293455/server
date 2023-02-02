package guild

import (
	"fmt"

	"coin-server/common/proto/guild_filter_service"
	"coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"

	"github.com/rs/xid"
)

type service struct {
	guild1AutoJoinOff  *models.Guild
	guild2AutoJoinOff  *models.Guild
	guild3WithAutoJoin *models.Guild

	roleGuildMap map[values.RoleId]values.GuildId

	guild2Leader              *testUser
	guild3Leader              *testUser
	joinApplyUser1            *testUser // 申请不开启自动加入的公会的用户
	joinApplyUser2            *testUser // 申请不开启自动加入的公会的用户
	joinApplySuccessUser      *testUser // 申请开启自动加入的公会的用户
	joinApplyUserForRejectAll *testUser

	memberList     []*models.GuildMember
	guildApplyList map[values.GuildId][]*models.GuildApply

	agreeInviteMap    map[values.RoleId][]*models.GuildInvite
	rejectInviteMap   map[values.RoleId][]*models.GuildInvite
	agreeInviteUsers  map[values.RoleId]*testUser
	rejectInviteUsers map[values.RoleId]*testUser
}

func New(ctx *core.RoleContext) core.ILoadTestModule {
	s := &service{
		agreeInviteUsers:  map[string]*testUser{},
		rejectInviteUsers: map[string]*testUser{},
	}

	s.guild2Leader = newTestUser()
	s.guild3Leader = newTestUser()
	s.joinApplyUser1 = newTestUser()
	s.joinApplyUser2 = newTestUser()
	s.joinApplySuccessUser = newTestUser()
	s.joinApplyUserForRejectAll = newTestUser()

	for i := 0; i < 5; i++ {
		u1 := newTestUser()
		u2 := newTestUser()
		s.agreeInviteUsers[u1.ctx.UserId] = u1
		s.rejectInviteUsers[u2.ctx.UserId] = u2
	}

	return s
}

func (s *service) Process(ctx *core.RoleContext) {
	s.reset()
	// 这两个接口需要启动guild-filter-server
	s.find(ctx)
	s.recommend(ctx)

	//s.createGuild1(ctx)
	//s.createGuild2()
	//s.createGuild3()
	//
	//s.findById(ctx)
	//s.enter(ctx)
	//s.modify(ctx)
	//s.members(ctx)
	//
	//s.invite(ctx)
	//s.getInviteList()
	//s.agreeInvite()
	//s.rejectInvite()
	//s.joinApplyToGuild1(s.joinApplyUser1)
	//s.joinApplyToGuild1(s.joinApplyUser2)
	//s.joinApplyToGuild2(s.joinApplyUser2)
	//s.joinApplyJoinSuccess()
	//s.getApplyList(ctx)
	//s.getApplyList(s.guild2Leader.ctx)
	//s.getApplyList(s.guild3Leader.ctx)
	//
	//s.agreeApply(ctx)
	//s.rejectApply(ctx)
	//s.rejectAllApply(s.guild2Leader.ctx)
	//
	//s.promotion(ctx)
	//s.demotion(ctx)
	//s.removeFromGuild1(ctx)
	//s.removeOne(ctx)
	//s.leaderChange(s.guild3Leader.ctx)
	//s.exit(s.guild3Leader.ctx)
	//
	//s.dissolve(ctx)
	//s.dissolve(s.guild2Leader.ctx)
	//s.dissolve(s.joinApplySuccessUser.ctx)
}

func (s *service) reset() {
	s.guild1AutoJoinOff = nil
	s.guild2AutoJoinOff = nil
	s.guild3WithAutoJoin = nil

	s.roleGuildMap = make(map[values.RoleId]values.GuildId)

	s.memberList = nil
	s.guildApplyList = make(map[values.GuildId][]*models.GuildApply)

	s.agreeInviteMap = make(map[values.RoleId][]*models.GuildInvite)
	s.rejectInviteMap = make(map[values.RoleId][]*models.GuildInvite)
	s.agreeInviteUsers = make(map[values.RoleId]*testUser)
	s.rejectInviteUsers = make(map[values.RoleId]*testUser)
}

func (s *service) find(ctx *core.RoleContext) {
	req := &guild_filter_service.Guild_FindRequest{
		Name: "",
		Lang: "en",
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*guild_filter_service.Guild_FindResponse)
}

func (s *service) recommend(ctx *core.RoleContext) {
	req := &guild_filter_service.Guild_RecommendRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*guild_filter_service.Guild_RecommendResponse)
}

func (s *service) findById(ctx *core.RoleContext) {
	req := &less_service.Guild_FindRequest{
		Id: []values.GuildId{
			s.guild1AutoJoinOff.Id,
			s.guild2AutoJoinOff.Id,
			s.guild3WithAutoJoin.Id,
		},
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	result := res.(*less_service.Guild_FindResponse)
	if len(result.List) != 3 {
		panic(fmt.Errorf("find guild by id error: %d,should be 3", len(result.List)))
	}
}

func (s *service) createGuild1(ctx *core.RoleContext) {
	v := xid.New().String()
	req := &less_service.Guild_CreateRequest{
		Name:     "LT_" + v,
		Flag:     "1",
		Lang:     "en",
		Intro:    v,
		Notice:   v,
		AutoJoin: false,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	result := res.(*less_service.Guild_CreateResponse)
	s.guild1AutoJoinOff = result.Info
	s.roleGuildMap[ctx.RoleId] = s.guild1AutoJoinOff.Id
}

func (s *service) createGuild2() {
	v := xid.New().String()
	req := &less_service.Guild_CreateRequest{
		Name:     "LT_" + v,
		Flag:     "1",
		Lang:     "en",
		Intro:    v,
		Notice:   v,
		AutoJoin: false,
	}
	_, res, err := s.guild2Leader.ctx.Request(req)
	assert.Nil(s.guild2Leader.ctx, err)
	result := res.(*less_service.Guild_CreateResponse)
	s.guild2AutoJoinOff = result.Info
	s.roleGuildMap[s.guild2Leader.ctx.RoleId] = s.guild2AutoJoinOff.Id
}

func (s *service) createGuild3() {
	v := xid.New().String()
	req := &less_service.Guild_CreateRequest{
		Name:     "LT_" + v,
		Flag:     "1",
		Lang:     "en",
		Intro:    v,
		Notice:   v,
		AutoJoin: true,
	}
	_, res, err := s.guild3Leader.ctx.Request(req)
	assert.Nil(s.guild3Leader.ctx, err)
	result := res.(*less_service.Guild_CreateResponse)
	s.guild3WithAutoJoin = result.Info
	s.roleGuildMap[s.guild3Leader.ctx.RoleId] = s.guild3WithAutoJoin.Id
}

func (s *service) enter(ctx *core.RoleContext) {
	req := &less_service.Guild_EnterRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	result := res.(*less_service.Guild_EnterResponse)
	s.guild1AutoJoinOff = result.Info
}

func (s *service) modify(ctx *core.RoleContext) {
	newV := xid.New().String()
	req := &less_service.Guild_ModifyRequest{
		Name:     newV,
		Flag:     newV,
		Lang:     newV,
		Intro:    newV,
		Notice:   newV,
		AutoJoin: 2,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	result := res.(*less_service.Guild_ModifyResponse)
	s.guild1AutoJoinOff = result.Info
}

func (s *service) members(ctx *core.RoleContext) {
	req := &less_service.Guild_MembersRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	result := res.(*less_service.Guild_MembersResponse)
	s.memberList = result.Members
}

func (s *service) invite(ctx *core.RoleContext) {
	for _, v := range s.agreeInviteUsers {
		req := &less_service.Guild_InviteJoinRequest{
			RoleId: v.ctx.RoleId,
		}
		_, res, err := ctx.Request(req)
		assert.Nil(ctx, err)
		_ = res.(*less_service.Guild_InviteJoinResponse)
	}
	for _, v := range s.rejectInviteUsers {
		req := &less_service.Guild_InviteJoinRequest{
			RoleId: v.ctx.RoleId,
		}
		_, res, err := ctx.Request(req)
		assert.Nil(ctx, err)
		_ = res.(*less_service.Guild_InviteJoinResponse)
	}
}

func (s *service) getInviteList() {
	for _, v := range s.agreeInviteUsers {
		req := &less_service.Guild_InviteListRequest{}
		_, res, err := v.ctx.Request(req)
		assert.Nil(v.ctx, err)
		result := res.(*less_service.Guild_InviteListResponse)
		s.agreeInviteMap[v.ctx.RoleId] = result.List
	}
	for _, v := range s.rejectInviteUsers {
		req := &less_service.Guild_InviteListRequest{}
		_, res, err := v.ctx.Request(req)
		assert.Nil(v.ctx, err)
		result := res.(*less_service.Guild_InviteListResponse)
		s.rejectInviteMap[v.ctx.RoleId] = result.List
	}
}

func (s *service) agreeInvite() {
	for _, user := range s.agreeInviteUsers {
		list := s.agreeInviteMap[user.ctx.RoleId]
		if len(list) <= 0 {
			continue
		}
		for _, invite := range list {
			req := &less_service.Guild_HandleInviteRequest{
				Uuid:  invite.Uuid,
				Agree: true,
			}
			_, res, err := user.ctx.Request(req)
			assert.Nil(user.ctx, err)
			result := res.(*less_service.Guild_HandleInviteResponse)
			if result.Info.Id != invite.GuildId {
				panic("guild id inconsistent")
			}
		}
	}
}

func (s *service) rejectInvite() {
	for _, user := range s.rejectInviteUsers {
		list := s.rejectInviteMap[user.ctx.RoleId]
		if len(list) <= 0 {
			continue
		}
		for _, invite := range list {
			req := &less_service.Guild_HandleInviteRequest{
				Uuid:  invite.Uuid,
				Agree: false,
			}
			_, res, err := user.ctx.Request(req)
			assert.Nil(user.ctx, err)
			result := res.(*less_service.Guild_HandleInviteResponse)
			if result.Info != nil {
				panic("guild id inconsistent")
			}
		}
	}
}

func (s *service) joinApplyToGuild1(user *testUser) {
	req := &less_service.Guild_JoinApplyRequest{
		Id: s.guild1AutoJoinOff.Id,
	}
	_, res, err := user.ctx.Request(req)
	assert.Nil(user.ctx, err)
	result := res.(*less_service.Guild_JoinApplyResponse)
	if result.Info != nil {
		panic("joinApply result error, should be nil")
	}
}

func (s *service) joinApplyToGuild2(user *testUser) {
	req := &less_service.Guild_JoinApplyRequest{
		Id: s.guild2AutoJoinOff.Id,
	}
	_, res, err := user.ctx.Request(req)
	assert.Nil(user.ctx, err)
	result := res.(*less_service.Guild_JoinApplyResponse)
	if result.Info != nil {
		panic("joinApply result error, should be nil")
	}
}

func (s *service) joinApplyJoinSuccess() {
	req := &less_service.Guild_JoinApplyRequest{
		Id: s.guild3WithAutoJoin.Id,
	}
	_, res, err := s.joinApplySuccessUser.ctx.Request(req)
	assert.Nil(s.joinApplySuccessUser.ctx, err)
	result := res.(*less_service.Guild_JoinApplyResponse)
	if result.Info.Id != s.guild3WithAutoJoin.Id {
		panic(fmt.Errorf("joinApplyJoinSuccess error, guild id not equal, %s,%s", s.guild3WithAutoJoin.Id, result.Info.Id))
	}
}

func (s *service) getApplyList(ctx *core.RoleContext) {
	req := &less_service.Guild_ApplyListRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	result := res.(*less_service.Guild_ApplyListResponse)
	id, ok := s.roleGuildMap[ctx.RoleId]
	if !ok {
		panic(fmt.Errorf("%s not found in roleGuildMap", ctx.RoleId))
	}
	s.guildApplyList[id] = result.List
}

func (s *service) agreeApply(ctx *core.RoleContext) {
	req := &less_service.Guild_HandleApplyRequest{
		RoleId: s.joinApplyUser1.ctx.RoleId,
		Agree:  true,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*less_service.Guild_HandleApplyResponse)
}

func (s *service) rejectApply(ctx *core.RoleContext) {
	req := &less_service.Guild_HandleApplyRequest{
		RoleId: s.joinApplyUser2.ctx.RoleId,
		Agree:  false,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*less_service.Guild_HandleApplyResponse)
}

func (s *service) rejectAllApply(ctx *core.RoleContext) {
	req := &less_service.Guild_RejectAllRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*less_service.Guild_RejectAllResponse)
}

func (s *service) promotion(ctx *core.RoleContext) {
	req := &less_service.Guild_PromotionRequest{
		RoleId: s.joinApplyUser1.ctx.RoleId,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*less_service.Guild_PromotionResponse)
}

func (s *service) demotion(ctx *core.RoleContext) {
	req := &less_service.Guild_DemotionRequest{
		RoleId: s.joinApplyUser1.ctx.RoleId,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*less_service.Guild_DemotionResponse)
}

func (s *service) removeFromGuild1(ctx *core.RoleContext) {
	for id := range s.agreeInviteUsers {
		req := &less_service.Guild_RemoveRequest{
			RoleId: id,
		}
		_, res, err := ctx.Request(req)
		assert.Nil(ctx, err)
		_ = res.(*less_service.Guild_RemoveResponse)
	}
}

func (s *service) removeOne(ctx *core.RoleContext) {
	req := &less_service.Guild_RemoveRequest{
		RoleId: s.joinApplyUser1.ctx.RoleId,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*less_service.Guild_RemoveResponse)
}

func (s *service) leaderChange(ctx *core.RoleContext) {
	req := &less_service.Guild_LeaderChangeRequest{
		RoleId: s.joinApplySuccessUser.ctx.RoleId,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*less_service.Guild_LeaderChangeResponse)
}

func (s *service) exit(ctx *core.RoleContext) {
	req := &less_service.Guild_ExitRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*less_service.Guild_ExitResponse)
}

func (s *service) dissolve(ctx *core.RoleContext) {
	req := &less_service.Guild_DissolveRequest{}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	_ = res.(*less_service.Guild_DissolveResponse)
}
