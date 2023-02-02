package equip_forge

import (
	"fmt"
	"testing"

	"coin-server/common/utils/test"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/equip-forge/rule"
	values2 "coin-server/game-server/service/equip-forge/values"
)

var tm *test.ServerTestMain

const totalWeight = values.Integer(10000)

func TestMain(m *testing.M) {
	tm = test.NewServerTestMain()
	m.Run()
}

func Test_handleBoxProb(t *testing.T) {
	tm.NewAstAndReq(t)
	svc := NewEquipForgeService(0, 0, nil, &module.Module{}, nil)

	t.Run("只有一个且概率不足100%", func(t *testing.T) {
		data := map[values.Quality]*values2.Product{
			1: {BoxId: 1, Quality: 1, Weight: 0},
			2: {BoxId: 2, Quality: 2, Weight: 0},
			3: {BoxId: 3, Quality: 3, Weight: 0},
			4: {BoxId: 4, Quality: 4, Weight: 20},
			5: {BoxId: 5, Quality: 5, Weight: 0},
			6: {BoxId: 6, Quality: 6, Weight: 0},
		}
		list := svc.handleBoxProb(data)
		tm.Req.Len(list, 1)
		var total values.Integer
		for _, product := range list {
			tm.Req.EqualValues(product.BoxId, 4)
			tm.Req.EqualValues(product.Quality, 4)
			tm.Req.EqualValues(product.Weight, 10000)
			total += product.Weight
		}
		tm.Req.EqualValues(total, totalWeight)
	})
	t.Run("单个概率超过100%", func(t *testing.T) {
		data := map[values.Quality]*values2.Product{
			1: {BoxId: 1, Quality: 1, Weight: 0},
			2: {BoxId: 2, Quality: 2, Weight: 1},
			3: {BoxId: 3, Quality: 3, Weight: 0},
			4: {BoxId: 4, Quality: 4, Weight: 20},
			5: {BoxId: 5, Quality: 5, Weight: 10000},
			6: {BoxId: 6, Quality: 6, Weight: 0},
		}
		list := svc.handleBoxProb(data)
		tm.Req.Len(list, 1)
		var total values.Integer
		for _, product := range list {
			tm.Req.EqualValues(product.BoxId, 5)
			tm.Req.EqualValues(product.Quality, 5)
			tm.Req.EqualValues(product.Weight, 10000)
			total += product.Weight
		}
		tm.Req.EqualValues(total, totalWeight)
	})
	t.Run("列表概率总和超过100%-不需要修正", func(t *testing.T) {
		data := map[values.Quality]*values2.Product{
			1: {BoxId: 1, Quality: 1, Weight: 0},
			2: {BoxId: 2, Quality: 2, Weight: 1000},
			3: {BoxId: 3, Quality: 3, Weight: 3000},
			4: {BoxId: 4, Quality: 4, Weight: 2000},
			5: {BoxId: 5, Quality: 5, Weight: 5000},
			6: {BoxId: 6, Quality: 6, Weight: 0},
		}
		list := svc.handleBoxProb(data)
		tm.Req.Len(list, 3)

		tm.Req.EqualValues(list[0].BoxId, 4)
		tm.Req.EqualValues(list[0].Weight, 2000)
		tm.Req.EqualValues(list[1].BoxId, 3)
		tm.Req.EqualValues(list[1].Weight, 3000)
		tm.Req.EqualValues(list[2].BoxId, 5)
		tm.Req.EqualValues(list[2].Weight, 5000)
		var total values.Integer
		for _, product := range list {
			total += product.Weight
		}
		tm.Req.EqualValues(total, totalWeight)
	})
	t.Run("列表概率总和超过100%-需要修正", func(t *testing.T) {
		data := map[values.Quality]*values2.Product{
			1: {BoxId: 1, Quality: 1, Weight: 0},
			2: {BoxId: 2, Quality: 2, Weight: 1000},
			3: {BoxId: 3, Quality: 3, Weight: 3000},
			4: {BoxId: 4, Quality: 4, Weight: 2000},
			5: {BoxId: 5, Quality: 5, Weight: 6000},
			6: {BoxId: 6, Quality: 6, Weight: 0},
		}
		list := svc.handleBoxProb(data)
		tm.Req.Len(list, 3)

		tm.Req.EqualValues(list[0].BoxId, 3)
		tm.Req.EqualValues(list[0].Weight, 2000)
		tm.Req.EqualValues(list[1].BoxId, 4)
		tm.Req.EqualValues(list[1].Weight, 2000)
		tm.Req.EqualValues(list[2].BoxId, 5)
		tm.Req.EqualValues(list[2].Weight, 6000)
		var total values.Integer
		for _, product := range list {
			total += product.Weight
		}
		tm.Req.EqualValues(total, totalWeight)
	})
	t.Run("列表概率总和超过100%-需要修正", func(t *testing.T) {
		data := map[values.Quality]*values2.Product{
			1: {BoxId: 1, Quality: 1, Weight: 0},
			2: {BoxId: 2, Quality: 2, Weight: 1000},
			3: {BoxId: 3, Quality: 3, Weight: 3000},
			4: {BoxId: 4, Quality: 4, Weight: 2000},
			5: {BoxId: 5, Quality: 5, Weight: 6000},
			6: {BoxId: 6, Quality: 6, Weight: 3000},
		}
		list := svc.handleBoxProb(data)
		tm.Req.Len(list, 3)

		tm.Req.EqualValues(list[0].BoxId, 4)
		tm.Req.EqualValues(list[0].Weight, 1000)
		tm.Req.EqualValues(list[1].BoxId, 6)
		tm.Req.EqualValues(list[1].Weight, 3000)
		tm.Req.EqualValues(list[2].BoxId, 5)
		tm.Req.EqualValues(list[2].Weight, 6000)
		var total values.Integer
		for _, product := range list {
			total += product.Weight
		}
		tm.Req.EqualValues(total, totalWeight)
	})
	t.Run("列表概率总和不足100%", func(t *testing.T) {
		data := map[values.Quality]*values2.Product{
			1: {BoxId: 1, Quality: 1, Weight: 0},
			2: {BoxId: 2, Quality: 2, Weight: 1000},
			3: {BoxId: 3, Quality: 3, Weight: 1000},
			4: {BoxId: 4, Quality: 4, Weight: 4000},
			5: {BoxId: 5, Quality: 5, Weight: 1000},
			6: {BoxId: 6, Quality: 6, Weight: 0},
		}
		list := svc.handleBoxProb(data)
		tm.Req.Len(list, 4)

		tm.Req.EqualValues(list[0].BoxId, 3)
		tm.Req.EqualValues(list[0].Weight, 1000)
		tm.Req.EqualValues(list[1].BoxId, 5)
		tm.Req.EqualValues(list[1].Weight, 1000)
		tm.Req.EqualValues(list[2].BoxId, 2)
		tm.Req.EqualValues(list[2].Weight, 4000)
		tm.Req.EqualValues(list[3].BoxId, 4)
		tm.Req.EqualValues(list[3].Weight, 4000)
		var total values.Integer
		for _, product := range list {
			total += product.Weight
		}
		tm.Req.EqualValues(total, totalWeight)
	})
	t.Run("概率修正测试", func(t *testing.T) {
		data := map[values.Quality]*values2.Product{
			1: {BoxId: 1, Quality: 1, Weight: 0},
			2: {BoxId: 2, Quality: 2, Weight: 1},
			3: {BoxId: 3, Quality: 3, Weight: 0},
			4: {BoxId: 4, Quality: 4, Weight: 20},
			5: {BoxId: 5, Quality: 5, Weight: 0},
			6: {BoxId: 6, Quality: 6, Weight: 0},
		}
		list := svc.handleBoxProb(data)
		tm.Req.Len(list, 2)

		tm.Req.EqualValues(list[0].BoxId, 4)
		tm.Req.EqualValues(list[0].Weight, 20)
		tm.Req.EqualValues(list[1].BoxId, 2)
		tm.Req.EqualValues(list[1].Weight, 9980)
		var total values.Integer
		for _, product := range list {
			total += product.Weight
		}
		tm.Req.EqualValues(total, totalWeight)
	})
	t.Run("策划给的正式数据测试", func(t *testing.T) {
		data := map[values.Quality]*values2.Product{
			1: {BoxId: 1, Quality: 1, Weight: 4000},
			2: {BoxId: 2, Quality: 2, Weight: 2000},
			3: {BoxId: 3, Quality: 3, Weight: 2000},
			4: {BoxId: 4, Quality: 4, Weight: 2000},
		}
		ret := make([]values.Integer, 0)
		for i := 0; i < 10; i++ {
			list := svc.handleBoxProb(data)
			var total values.Integer
			for _, product := range list {
				total += product.Weight
			}
			tm.Req.EqualValues(total, totalWeight)
			boxId := svc.randomBoxId(list)
			ret = append(ret, boxId)
		}
		fmt.Println(ret)
	})
}

func Test_randomBoxId(t *testing.T) {
	tm.NewAstAndReq(t)
	svc := NewEquipForgeService(0, 0, nil, &module.Module{}, nil)
	data := map[values.Quality]*values2.Product{
		1: {BoxId: 1, Quality: 1, Weight: 0},
		2: {BoxId: 2, Quality: 2, Weight: 1000},
		3: {BoxId: 3, Quality: 3, Weight: 1000},
		4: {BoxId: 4, Quality: 4, Weight: 4000},
		5: {BoxId: 5, Quality: 5, Weight: 1000},
		6: {BoxId: 6, Quality: 6, Weight: 0},
	}
	list := svc.handleBoxProb(data)
	var total values.Integer
	for _, product := range list {
		total += product.Weight
	}
	tm.Req.EqualValues(total, totalWeight)
	ret := make([]values.Integer, 0)
	for i := 0; i < 10; i++ {
		ret = append(ret, svc.randomBoxId(list))
	}
	fmt.Println(ret)
}

func Test_GetForgeLevelBonus(t *testing.T) {
	tm.NewAstAndReq(t)
	list := rule.GetForgeLevelBonus(tm.Ctx, 20, 100)
	fmt.Println(list)
}
