package user

import (
	"testing"

	"coin-server/common/utils/test"
	"coin-server/game-server/service/user/rule"
)

var tm *test.ServerTestMain

func TestMain(m *testing.M) {
	tm = test.NewServerTestMain()
	m.Run()
}

func Test_GetExpSkipEffcient(t *testing.T) {
	tm.NewAstAndReq(t)
	tm.Ast.EqualValues(12000, rule.GetExpSkipEffcient(tm.Ctx, 0).Rate)
	tm.Ast.EqualValues(12000, rule.GetExpSkipEffcient(tm.Ctx, 4).Rate)
	tm.Ast.EqualValues(12000, rule.GetExpSkipEffcient(tm.Ctx, 5).Rate)
	tm.Ast.EqualValues(10000, rule.GetExpSkipEffcient(tm.Ctx, 6).Rate)
	tm.Ast.EqualValues(10000, rule.GetExpSkipEffcient(tm.Ctx, 10).Rate)
	tm.Ast.EqualValues(8000, rule.GetExpSkipEffcient(tm.Ctx, 14).Rate)
	tm.Ast.EqualValues(8000, rule.GetExpSkipEffcient(tm.Ctx, 15).Rate)
	tm.Ast.EqualValues(6000, rule.GetExpSkipEffcient(tm.Ctx, 16).Rate)
	tm.Ast.EqualValues(5000, rule.GetExpSkipEffcient(tm.Ctx, 21).Rate)
	tm.Ast.EqualValues(5000, rule.GetExpSkipEffcient(tm.Ctx, 30).Rate)
	tm.Ast.EqualValues(5000, rule.GetExpSkipEffcient(tm.Ctx, 33).Rate)
}
