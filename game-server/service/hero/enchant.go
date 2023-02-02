package hero

import (
	"sort"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/hero/dao"
	"coin-server/game-server/service/hero/rule"
)

func (svc *Service) GetEnchantInfo(ctx *ctx.Context, _ *servicepb.Enchant_EnchantInfoRequest) (*servicepb.Enchant_EnchantInfoResponse, *errmsg.ErrMsg) {
	info, err := dao.GetEnchant(ctx)
	if err != nil {
		return nil, err
	}
	hero, ok, err := dao.NewHero(ctx.RoleId).GetOne(ctx, info.HeroId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &servicepb.Enchant_EnchantInfoResponse{}, nil
	}
	return &servicepb.Enchant_EnchantInfoResponse{
		HeroId: hero.BuildId,
		SlotId: info.SlotId,
		Info:   info.Affix,
	}, nil
}

func (svc *Service) EnchantGen(ctx *ctx.Context, req *servicepb.Enchant_EnchantGenRequest) (*servicepb.Enchant_EnchantGenResponse, *errmsg.ErrMsg) {
	realId, err := svc.getHeroRealId(ctx, req.HeroId)
	if err != nil {
		return nil, err
	}
	// role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	// if err != nil {
	// 	return nil, err
	// }
	hero, ok, err := dao.NewHero(ctx.RoleId).GetOne(ctx, realId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	if !svc.isEnchantUnlock(ctx, hero, req.SlotId) {
		return nil, errmsg.NewErrEquipEnchantNotUnlock()
	}

	info, err := dao.GetEnchant(ctx)
	if err != nil {
		return nil, err
	}

	if info.HeroId > 0 {
		return nil, errmsg.NewErrEnchantLastUnfinished()
	}
	material, ok := rule.GetEnchant(ctx, req.MaterialId)
	if !ok {
		return nil, errmsg.NewErrEnchantMaterialNotExist()
	}
	if material.EquipSlot != req.SlotId {
		return nil, errmsg.NewErrEnchantMaterialAndSlotNotMatch()
	}
	var find bool
	// 不限定职业
	if len(material.EquipJob) == 1 && material.EquipJob[0] == 0 {
		find = true
	} else {
		for _, v := range material.EquipJob {
			cfg, ok := rule.GetHero(ctx, req.HeroId)
			if !ok {
				continue
			}
			if v == cfg.OriginId {
				find = true
				break
			}
		}
	}
	if !find {
		return nil, errmsg.NewErrEnchantMaterialAndHeroNotMatch()
	}
	if err := svc.SubItem(ctx, ctx.RoleId, material.Id, 1); err != nil {
		return nil, errmsg.NewErrEnchantMaterialNotEnough()
	}
	cfgList := rule.GetEquipEntryByGroup(ctx, material.Group)
	if len(cfgList) == 0 {
		return nil, errmsg.NewErrEnchantMaterialNotExist()
	}
	// 先取第一条的数据，计算出品质
	q, err := svc.GenAffixQuality(svc.handleProb(cfgList[0].Qualityweight), 0)
	if err != nil {
		return nil, err
	}
	// 根据品质生成随机属性
	affix, _, err := svc.GenOneAffix(cfgList, 0, q)
	if err != nil {
		return nil, err
	}
	if affix == nil {
		return nil, errmsg.NewInternalErr("affix is nil")
	}

	info = dao.NewEnchant(ctx.RoleId, realId, req.SlotId, affix)
	if err := dao.SaveEnchant(ctx, info); err != nil {
		return nil, err
	}
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskEnchatNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskEnchatNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskEnchatNum: {
			Typ:     values.Integer(models.TaskType_TaskEnchatNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	svc.UpdateTargets(ctx, ctx.RoleId, tasks)

	return &servicepb.Enchant_EnchantGenResponse{
		Info: info.Affix,
	}, nil
}

func (svc *Service) EnchantReplace(ctx *ctx.Context, _ *servicepb.Enchant_EnchantReplaceRequest) (*servicepb.Enchant_EnchantReplaceResponse, *errmsg.ErrMsg) {
	info, err := dao.GetEnchant(ctx)
	if err != nil {
		return nil, err
	}
	if info.HeroId <= 0 || info.SlotId <= 0 || info.Affix == nil {
		return nil, errmsg.NewErrEnchantNotExist()
	}
	heroDao := dao.NewHero(ctx.RoleId)
	hero, ok, err := heroDao.GetOne(ctx, info.HeroId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getHeroEquippedEquipId(hero)...)
	if err != nil {
		return nil, err
	}
	if len(hero.EquipSlot) <= 0 {
		hero.EquipSlot = make(map[values.Integer]*models.HeroEquipSlot)
	}
	slotInfo, ok := hero.EquipSlot[info.SlotId]
	if !ok {
		slotInfo = &models.HeroEquipSlot{}
	}
	slotInfo.Enchant = info.Affix
	hero.EquipSlot[info.SlotId] = slotInfo

	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	svc.refreshHeroAttr(ctx, hero, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill, 0)

	heroModel := svc.dao2model(ctx, hero)
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: []*event.HeroAttrUpdateItem{{
			IsSkillChange: false,
			Hero:          heroModel,
		}},
	})

	info = dao.ResetEnchant(ctx.RoleId)
	if err := heroDao.Save(ctx, hero); err != nil {
		return nil, err
	}
	if err := dao.SaveEnchant(ctx, info); err != nil {
		return nil, err
	}

	return &servicepb.Enchant_EnchantReplaceResponse{
		Hero: heroModel,
	}, nil
}

func (svc *Service) EnchantDrop(ctx *ctx.Context, _ *servicepb.Enchant_EnchantDropRequest) (*servicepb.Enchant_EnchantDropResponse, *errmsg.ErrMsg) {
	info, err := dao.GetEnchant(ctx)
	if err != nil {
		return nil, err
	}
	info = dao.ResetEnchant(ctx.RoleId)
	if err := dao.SaveEnchant(ctx, info); err != nil {
		return nil, err
	}
	return &servicepb.Enchant_EnchantDropResponse{}, nil
}

func (svc *Service) handleProb(weightMap map[values.Quality]values.Integer) map[values.Quality]values.Integer {
	var totalWeight values.Float
	for _, weight := range weightMap {
		totalWeight += values.Float(weight)
	}
	list := make([]*WeightItem, 0, len(weightMap))
	for q, weight := range weightMap {
		v := values.Integer((values.Float(weight) / totalWeight) * 10000)
		list = append(list, &WeightItem{
			Quality: q,
			Weight:  v,
		})
	}
	// 按品质从高到低排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Quality > list[j].Quality
	})
	// 按品质从高到低，移除超过100%的品质，并把品质概率总和归为100%
	var totalProb int64
	max := values.Integer(10000)
	newList := make([]*WeightItem, 0)
	for _, item := range list {
		if item.Weight <= 0 {
			continue
		}
		if item.Weight >= max {
			totalProb = max
			item.Weight = max
			newList = append(newList, item)
			break
		} else if totalProb < max {
			totalProb += item.Weight
			if totalProb > max {
				item.Weight -= totalProb - max
				totalProb = max
			}
			newList = append(newList, item)
		}
	}
	// 概率不足100%，提高最低品质的概率直至100%
	if totalProb < max {
		newList[len(newList)-1].Weight += max - totalProb
	}
	// 概率从低到高排序
	sort.Slice(newList, func(i, j int) bool {
		if newList[i].Weight == newList[j].Weight {
			return newList[i].Quality < newList[j].Quality
		} else {
			return newList[i].Weight < newList[j].Weight
		}
	})
	retMap := make(map[values.Quality]values.Integer, len(newList))
	for _, item := range newList {
		retMap[item.Quality] = item.Weight
	}
	return retMap
}

func (svc *Service) isEnchantUnlock(ctx *ctx.Context, hero *pbdao.Hero, slotId values.Integer) bool {
	if hero.EquipSlot == nil {
		return false
	}
	slot, ok := hero.EquipSlot[slotId]
	if !ok {
		return false
	}
	return slot.MeltLevel >= rule.GetEquipEnchantingLimit(ctx)
}
