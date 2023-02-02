package guild

import (
	"fmt"
	"testing"

	pbdao "coin-server/common/proto/dao"
	"coin-server/common/utils/test"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/guild/rule"
)

var tm *test.ServerTestMain

func TestMain(m *testing.M) {
	tm = test.NewServerTestMain()
	m.Run()
}

func Test_addExp(t *testing.T) {
	tm.NewAstAndReq(t)
	svc := NewGuildService(0, 0, nil, &module.Module{}, nil)
	guild := &pbdao.Guild{
		Level: 33,
		Exp:   0,
	}
	cfg, _ := rule.GetGuildConfigByLevel(tm.Ctx, 1)
	svc.addGuildExp(tm.Ctx, guild, cfg, 30012333330, 0)
}

func Test_getAvailableBlessFromConfigByTarget(t *testing.T) {
	tm.NewAstAndReq(t)
	svc := NewGuildService(0, 0, nil, &module.Module{}, nil)

	target, _ := rule.GetBlessById(tm.Ctx, 1)
	list, _ := svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("1:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 2)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("2:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 3)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("3:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 4)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("4:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 4}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 5)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("5:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 5}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 6)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("6:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 6}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 7)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("7:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 4, 7}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 8)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("8:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 5, 8}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 9)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("9:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 4, 5, 6, 7, 8, 9}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 10)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("10:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 11)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("11:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 12)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("12:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 13)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("13:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 13}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 14)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("14:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 14}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 15)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("15:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 15}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 16)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("16:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 13, 16}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 17)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("17:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, list)

	target, _ = rule.GetBlessById(tm.Ctx, 18)
	list, _ = svc.getAvailableBlessFromConfigByTarget(tm.Ctx, target)
	fmt.Printf("18:%v\n", list)
	tm.Ast.EqualValues([]values.Integer{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, list)
}
