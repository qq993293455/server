package hero

import (
	"testing"

	"coin-server/common/utils/test"
)

var tm *test.ServerTestMain

func TestMain(m *testing.M) {
	tm = test.NewServerTestMain()
	m.Run()
}

func Test_transformAttr(t *testing.T) {
	// tm.NewAstAndReq(t)
	// svc := NewHeroService(0, 0, nil, &module.Module{}, nil)
	// hero := &pbdao.Hero{
	//	Id:    1,
	//	Skill: nil,
	//	Attrs: map[int64]*models.HeroAttrItem{
	//		1: {
	//			Attr: map[int64]int64{
	//				1: 3,
	//				2: 10,
	//				3: 20,
	//				5: 100000,
	//			},
	//		},
	//	},
	//	EquipSlot: nil,
	// }
	// svc.transformAttr(tm.Ctx, hero)
	// tm.Ast.EqualValues(3, hero.Attrs[1].Attr[1])
	// tm.Ast.EqualValues(10, hero.Attrs[1].Attr[2])
	// tm.Ast.EqualValues(20, hero.Attrs[1].Attr[3])
	// tm.Ast.EqualValues(1, hero.Attrs[1].Attr[101])
	// tm.Ast.EqualValues(101, hero.Attrs[1].Attr[103])
	// tm.Ast.EqualValues(3, hero.Attrs[1].Attr[111])
	// tm.Ast.EqualValues(20, hero.Attrs[1].Attr[106])
	// tm.Ast.EqualValues(10, hero.Attrs[1].Attr[108])
	// tm.Ast.EqualValues(60, hero.Attrs[1].Attr[105])
	// tm.Ast.EqualValues(100, hero.Attrs[1].Attr[102])
}

//
// func Test_oneEquipAttr(t *testing.T) {
// 	tm.NewAstAndReq(t)
// 	svc := NewHeroService(0, 0, nil, &module.Module{}, nil)
// 	data := svc.oneEquipAttr(tm.Ctx, &models.Equipment{
// 		EquipId: "",
// 		ItemId:  110001,
// 		Level:   0,
// 		Affix: []*models.Affix{
// 			{
// 				AffixId:    1,
// 				Quality:    0,
// 				AffixValue: 10,
// 				SkillId:    0,
// 				Active:     false,
// 				AttrId:     101,
// 				Bonus: map[int64]int64{
// 					1: 10,
// 				},
// 			},
// 		},
// 		HeroId: 0,
// 	}, 0)
// 	tm.Ast.EqualValues(map[int64]int64{101: 20, 102: 10}, data)
// 	data = svc.oneEquipAttr(tm.Ctx, &models.Equipment{
// 		EquipId: "",
// 		ItemId:  110001,
// 		Level:   0,
// 		Affix:   nil,
// 		HeroId:  0,
// 	}, 1)
// 	tm.Ast.EqualValues(map[int64]int64{101: 11, 102: 11}, data)
// 	data = svc.oneEquipAttr(tm.Ctx, &models.Equipment{
// 		EquipId: "",
// 		ItemId:  110001,
// 		Level:   0,
// 		Affix:   nil,
// 		HeroId:  0,
// 	}, 5)
// 	tm.Ast.EqualValues(map[int64]int64{101: 15, 102: 15}, data)
// 	data = svc.oneEquipAttr(tm.Ctx, &models.Equipment{
// 		EquipId: "",
// 		ItemId:  110001,
// 		Level:   0,
// 		Affix:   nil,
// 		HeroId:  0,
// 	}, 33)
// 	tm.Ast.EqualValues(map[int64]int64{101: 82, 102: 82}, data)
// 	data = svc.oneEquipAttr(tm.Ctx, &models.Equipment{
// 		EquipId: "",
// 		ItemId:  110001,
// 		Level:   0,
// 		Affix:   nil,
// 		HeroId:  0,
// 	}, 49)
// 	tm.Ast.EqualValues(map[int64]int64{101: 155, 102: 155}, data)
// 	data = svc.oneEquipAttr(tm.Ctx, &models.Equipment{
// 		EquipId: "",
// 		ItemId:  110001,
// 		Level:   0,
// 		Affix:   nil,
// 		HeroId:  0,
// 	}, 50)
// 	tm.Ast.EqualValues(map[int64]int64{101: 160, 102: 160}, data)
// 	data = svc.oneEquipAttr(tm.Ctx, &models.Equipment{
// 		EquipId: "",
// 		ItemId:  110001,
// 		Level:   0,
// 		Affix:   nil,
// 		HeroId:  0,
// 	}, 60)
// 	tm.Ast.EqualValues(map[int64]int64{101: 220, 102: 220}, data)
// 	data = svc.oneEquipAttr(tm.Ctx, &models.Equipment{
// 		EquipId: "",
// 		ItemId:  110001,
// 		Level:   0,
// 		Affix:   nil,
// 		HeroId:  0,
// 	}, 61)
// 	tm.Ast.EqualValues(map[int64]int64{101: 226, 102: 226}, data)
// }
//
// func Test_updateHeroCombatValue(t *testing.T) {
// 	tm.NewAstAndReq(t)
// 	svc := NewHeroService(0, 0, nil, &module.Module{}, nil)
// 	hero := &pbdao.Hero{
// 		Id: 1,
// 		Skill: []*models.HeroSkillItem{
// 			{10101001, 1},
// 			{10105001, 1},
// 			{10106001, 1},
// 			{10107001, 1},
// 			{10108001, 1},
// 		},
// 		Attrs: map[int64]*models.HeroAttrItem{
// 			1: {
// 				Attr: map[int64]int64{
// 					1: 10,
// 					2: 10,
// 					3: 30,
// 				},
// 			},
// 		},
// 		EquipSlot:   nil,
// 		CombatValue: nil,
// 	}
// 	svc.updateHeroCombatValue(tm.Ctx, hero, 1)
// 	fmt.Println(hero)
// }

func Test_updateHeroAttrs(t *testing.T) {
	// tm.NewAstAndReq(t)
	// svc := NewHeroService(0, 0, nil, &module.Module{}, nil)
	// hero := &pbdao.Hero{
	// 	Id: 1,
	// 	Skill: []*models.HeroSkillItem{
	// 		{10101001, 1},
	// 		{10105001, 1},
	// 		{10106001, 1},
	// 		{10107001, 1},
	// 		{10108001, 1},
	// 	},
	// 	Attrs: map[int64]*models.HeroAttrItem{
	// 		1: {
	// 			Attr: map[int64]int64{
	// 				1: 10,
	// 				2: 10,
	// 				3: 30,
	// 			},
	// 		},
	// 	},
	// 	EquipSlot:   nil,
	// 	CombatValue: nil,
	// }
	// fixed := []*models.OverAllAttr{
	// 	{
	// 		Typ:  1,
	// 		Attr: map[int64]int64{1: 10, 2: 20},
	// 	},
	// }
	// percent := []*models.OverAllAttr{
	// 	{
	// 		Typ:  1,
	// 		Attr: map[int64]int64{1: 10, 2: 20},
	// 	},
	// }
	// svc.updateHeroAttrs(tm.Ctx, hero, 1, false, false, fixed, percent)
	// b, _ := json.Marshal(hero)
	// var out bytes.Buffer
	// json.Indent(&out, b, "", "  ")
	// fmt.Println(out.String())
}

func Test_getOnSetInfo(t *testing.T) {
	// tm.NewAstAndReq(t)
	// svc := NewHeroService(0, 0, nil, &module.Module{}, nil)
	// svc.getEquipSetInfo(tm.Ctx, []values.ItemId{140002, 220002, 240002, 250002, 310002, 320002})
}
