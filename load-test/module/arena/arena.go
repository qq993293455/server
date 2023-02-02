package Arena

import (
	"coin-server/common/proto/models"
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
	s.UnlockSystem(ctx)
	s.ArenaSelfRankingRequest(ctx)
	s.ArenaArenaRankingRequest(ctx)
	s.ArenaChallenge(ctx)
}

func (s service) UnlockSystem(ctx *core.RoleContext) {
	req := &pbSvr.SystemUnlock_CheatUnlockSystemRequest{
		SystemId: models.SystemType_SystemArena,
	}
	_, _, err := ctx.Request(req)
	assert.Nil(ctx, err)
}

func (s service) ArenaSelfRankingRequest(ctx *core.RoleContext) int32 {
	req := &pbSvr.Arena_ArenaSelfRankingRequest{
		Type: models.ArenaType_ArenaType_Default,
	}
	_, res, err := ctx.Request(req)
	assert.Nil(ctx, err)
	ret := res.(*pbSvr.Arena_ArenaSelfRankingResponse)
	return ret.RankingId
}

func (s service) ArenaArenaRankingRequest(ctx *core.RoleContext) {
	startIndex := 1
	count := 20

	for {
		req := &pbSvr.Arena_ArenaRankingRequest{
			Type:       models.ArenaType_ArenaType_Default,
			StartIndex: int32(startIndex),
			Count:      int32(count),
		}

		_, res, err := ctx.Request(req)
		assert.Nil(ctx, err)

		ret := res.(*pbSvr.Arena_ArenaRankingResponse)
		if len(ret.RankingInfo) >= count {
			startIndex += len(ret.RankingInfo)
		} else {
			break
		}
	}
}

func (s service) ArenaChallenge(ctx *core.RoleContext) {
	req1 := &pbSvr.Arena_ArenaGetChallengeRequest{
		Type: models.ArenaType_ArenaType_Default,
	}

	_, res1, err := ctx.Request(req1)
	assert.Nil(ctx, err)

	ret1 := res1.(*pbSvr.Arena_ArenaGetChallengeResponse)
	assert.True(ctx, len(ret1.RankingInfo) > 0)
	challengeInfo := ret1.RankingInfo[0]

	req2 := &pbSvr.Arena_ArenaChallengeRequest{
		Type:          models.ArenaType_ArenaType_Default,
		RoleId:        challengeInfo.RoleId,
		RankingId:     challengeInfo.RankingId,
		SelfRankingId: s.ArenaSelfRankingRequest(ctx),
	}

	_, _, err = ctx.Request(req2)
	if err != nil {
		return
	}

	req3 := &pbSvr.Arena_ArenaChallengeResultPrcRequest{
		Type:     models.ArenaType_ArenaType_Default,
		IsWin:    true,
		PlayerId: challengeInfo.RoleId,
	}

	_, _, err = ctx.Request(req3)
	assert.Nil(ctx, err)

	//fmt.Println(s.ArenaSelfRankingRequest(ctx))
}
