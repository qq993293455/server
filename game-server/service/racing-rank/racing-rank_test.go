package racing_rank

import (
	"testing"

	pbdao "coin-server/common/proto/dao"
	"coin-server/common/utils/test"
	"coin-server/game-server/service/racing-rank/rule"
)

var tm *test.ServerTestMain

func TestMain(m *testing.M) {
	tm = test.NewServerTestMain()
	m.Run()
}

func Test_GetRankReward(t *testing.T) {
	tm.NewAstAndReq(t)
	cfg := rule.GetRankReward(tm.Ctx, 1)
	tm.Ast.EqualValues(1, cfg.Id)

	cfg = rule.GetRankReward(tm.Ctx, 2)
	tm.Ast.EqualValues(2, cfg.Id)

	cfg = rule.GetRankReward(tm.Ctx, 3)
	tm.Ast.EqualValues(3, cfg.Id)

	cfg = rule.GetRankReward(tm.Ctx, 4)
	tm.Ast.EqualValues(4, cfg.Id)

	cfg = rule.GetRankReward(tm.Ctx, 5)
	tm.Ast.EqualValues(5, cfg.Id)

	cfg = rule.GetRankReward(tm.Ctx, 8)
	tm.Ast.EqualValues(5, cfg.Id)

	cfg = rule.GetRankReward(tm.Ctx, 10)
	tm.Ast.EqualValues(5, cfg.Id)

	cfg = rule.GetRankReward(tm.Ctx, 33)
	tm.Ast.EqualValues(7, cfg.Id)

	cfg = rule.GetRankReward(tm.Ctx, 37)
	tm.Ast.EqualValues(8, cfg.Id)

	cfg = rule.GetRankReward(tm.Ctx, 51)
	tm.Ast.EqualValues(8, cfg.Id)
}

func Test_isGetAllReward(t *testing.T) {
	tm.NewAstAndReq(t)
	svc := NewRacingRank(0, 0, nil, nil, nil)
	v := svc.isGetAllReward(tm.Ctx, &pbdao.RacingRankStatus{
		HighestRank:  1,
		RewardedRank: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	})
	tm.Req.True(v)
}
