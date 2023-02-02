package bag

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/bag/dao"
	"coin-server/rule"
)

func (s *Service) addManyStones(ctx *ctx.Context, roleId values.RoleId, stoneMap map[values.ItemId]values.Integer) (*event.SkillStoneUpdate, *errmsg.ErrMsg) {
	stoneList := make([]*pbdao.SkillStone, 0, len(stoneMap))
	for stoneId := range stoneMap {
		stoneList = append(stoneList, &pbdao.SkillStone{
			ItemId: stoneId,
		})
	}
	err := dao.GetSkillStones(ctx, roleId, stoneList)
	if err != nil {
		return nil, err
	}
	i := 0
	itemIncr := make([]values.Integer, len(stoneList))
	for _, item := range stoneList {
		// if item.Count == 0 && isShow(ctx, item.ItemId) {
		// 	err = addBagLength(ctx, 1)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }
		cfg, _ := rule.MustGetReader(ctx).Item.GetItemById(item.ItemId)
		if cfg.ExpiredTime > 0 {
			item.Expire = cfg.ExpiredTime * 1000
		} else if cfg.ExpiredTime < 0 {
			item.Expire = timer.StartTime(ctx.StartTime).UnixMilli() + cfg.ExpiredTime*1000
		}
		item.Count += stoneMap[item.ItemId]
		itemIncr[i] = stoneMap[item.ItemId]
		i++
	}

	dao.SaveManySkillStone(ctx, roleId, stoneList)
	e := &event.SkillStoneUpdate{
		RoleId: roleId,
		Stones: SkillStoneDao2Models(stoneList),
		Incr:   itemIncr,
	}
	return e, nil
}

func (s *Service) subManyStones(ctx *ctx.Context, roleId values.RoleId, stoneMap map[values.ItemId]values.Integer) (*event.SkillStoneUpdate, *errmsg.ErrMsg) {
	var (
		del, save []*pbdao.SkillStone
	)
	stoneList := make([]*pbdao.SkillStone, 0, len(stoneMap))
	for stoneId := range stoneMap {
		stoneList = append(stoneList, &pbdao.SkillStone{
			ItemId: stoneId,
		})
	}
	err := dao.GetSkillStones(ctx, roleId, stoneList)
	if err != nil {
		return nil, err
	}
	itemIncr := make([]values.Integer, 0)
	for _, stone := range stoneList {
		if stoneMap[stone.ItemId] == 0 {
			continue
		}
		if stone.Lock {
			return nil, errmsg.NewErrSkillStoneIsLock()
		}
		if stone.Count-stoneMap[stone.ItemId] < 0 {
			return nil, errmsg.NewErrBagNotEnough()
		}
		stone.Count -= stoneMap[stone.ItemId]
		if stone.Count == 0 && isShow(ctx, stone.ItemId) {
			if del == nil {
				del = []*pbdao.SkillStone{stone}
			} else {
				del = append(del, stone)
			}
		} else {
			if save == nil {
				save = []*pbdao.SkillStone{stone}
			} else {
				save = append(save, stone)
			}
		}
		itemIncr = append(itemIncr, -stoneMap[stone.ItemId])
	}

	if len(del) != 0 {
		dao.DelManySkillStone(ctx, roleId, del)
	}
	if len(save) != 0 {
		dao.SaveManySkillStone(ctx, roleId, save)
	}

	e := &event.SkillStoneUpdate{
		RoleId: roleId,
		Stones: SkillStoneDao2Models(stoneList),
		Incr:   itemIncr,
	}
	return e, nil
}

func (s *Service) lockSkillStone(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) *errmsg.ErrMsg {
	stone, err := dao.GetSkillStone(ctx, roleId, itemId)
	if err != nil {
		return err
	}
	stone.Lock = true
	dao.SaveSkillStone(ctx, roleId, stone)
	ctx.PublishEventLocal(&event.SkillStoneUpdate{
		RoleId: roleId,
		Stones: SkillStoneDao2Models([]*pbdao.SkillStone{stone}),
	})
	return nil
}

func (s *Service) unlockSkillStone(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) *errmsg.ErrMsg {
	stone, err := dao.GetSkillStone(ctx, roleId, itemId)
	if err != nil {
		return err
	}
	stone.Lock = false
	dao.SaveSkillStone(ctx, roleId, stone)
	ctx.PublishEventLocal(&event.SkillStoneUpdate{
		RoleId: roleId,
		Stones: SkillStoneDao2Models([]*pbdao.SkillStone{stone}),
	})
	return nil
}
