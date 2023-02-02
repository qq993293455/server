package hero

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/hero/dao"
	"coin-server/game-server/service/hero/rule"
	"go.uber.org/zap"
)

// HandlerUserLevelUp 角色升级事件，更新英雄属性
func (svc *Service) HandlerUserLevelUp(ctx *ctx.Context, d *event.UserLevelChange) *errmsg.ErrMsg {
	if d.Level <= 0 {
		return nil
	}
	heroDao := dao.NewHero(ctx.RoleId)
	heroes, err := heroDao.Get(ctx)
	if err != nil {
		return err
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getAllHeroEquippedEquipId(heroes)...)
	if err != nil {
		return err
	}
	// temp := make([]string, 0)
	// for id := range equips {
	// 	temp = append(temp, id)
	// }
	svc.refreshAllHeroAttr(ctx, heroes, d.Level, d.LevelIndex, true, equips, roleAttr, roleSkill)
	if err := heroDao.Save(ctx, heroes...); err != nil {
		return err
	}
	list := make([]*models.Hero, 0, len(heroes))
	heroAttrUpdateList := make([]*event.HeroAttrUpdateItem, 0, len(heroes))
	for _, hero := range heroes {
		heroModel := svc.dao2model(ctx, hero)
		list = append(list, heroModel)
		heroAttrUpdateList = append(heroAttrUpdateList, &event.HeroAttrUpdateItem{
			IsSkillChange: false,
			Hero:          heroModel,
		})
	}
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: heroAttrUpdateList,
	})

	ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: list,
	})
	return nil
}

// HandlerRoleAttrUpdate 全局属性加成更新事件
func (svc *Service) HandlerRoleAttrUpdate(ctx *ctx.Context, d *event.RoleAttrUpdate) *errmsg.ErrMsg {
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	heroDao := dao.NewHero(ctx.RoleId)
	heroes, err := heroDao.Get(ctx)
	if err != nil {
		return err
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getAllHeroEquippedEquipId(heroes)...)
	if err != nil {
		return err
	}
	svc.refreshAllHeroAttr(ctx, heroes, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill)
	if err := heroDao.Save(ctx, heroes...); err != nil {
		return err
	}
	list := make([]*models.Hero, 0, len(heroes))
	heroAttrUpdateList := make([]*event.HeroAttrUpdateItem, 0, len(heroes))
	for _, hero := range heroes {
		heroModel := svc.dao2model(ctx, hero)
		list = append(list, heroModel)
		heroAttrUpdateList = append(heroAttrUpdateList, &event.HeroAttrUpdateItem{
			IsSkillChange: false,
			Hero:          heroModel,
		})
	}
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: heroAttrUpdateList,
	})
	ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: list,
	})
	return nil
}

func (svc *Service) HandleTalentUpdate(ctx *ctx.Context, d *event.TalentChange) *errmsg.ErrMsg {
	if d.Attr == nil {
		return nil
	}
	cfg, ok := rule.GetHero(ctx, d.ConfigId)
	if !ok {
		return errmsg.NewErrHeroNotFound()
	}
	heroDao := dao.NewHero(ctx.RoleId)
	hero, ok, err := heroDao.GetOne(ctx, cfg.OriginId)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	/*if rule.IsInitTalent(ctx, d.ConfigId) {
		skills := svc.GetInitSkills(cfg)
		skillMap := make(map[values.HeroSkillId]*pbdao.SkillLevel)
		for _, skillId := range skills {
			skillMap[skillId] = &pbdao.SkillLevel{}
		}
		hero.Skill = skillMap
	} else {
		newSkills := make(map[values.HeroSkillId]*pbdao.SkillLevel)
		for _, skillId := range d.Attr.SkillIds {
			skillCfg, ok := rule.GetSkillById(ctx, skillId)
			if !ok {
				svc.log.Warn("skill config not found", zap.Int64("id", skillId))
				continue
			}
			if hero.Skill == nil {
				hero.Skill = map[int64]*pbdao.SkillLevel{}
			}
			levelInfo := &pbdao.SkillLevel{
				Talent: 0,
				Equip:  map[int64]int64{},
				Role:   0,
			}
			info, ok := hero.Skill[skillCfg.SkillBaseId]
			if ok {
				levelInfo = info
			}
			levelInfo.Talent = skillId - skillCfg.SkillBaseId
			if levelInfo.Talent < 0 {
				levelInfo.Talent = 0
			}
			newSkills[skillCfg.SkillBaseId] = levelInfo
		}
		hero.Skill = newSkills
	}*/

	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if len(hero.Attrs) == 0 {
		hero.Attrs = make(map[values.Integer]*models.HeroAttr)
	}
	if d.Attr.Attr == nil {
		d.Attr.Attr = &models.TalentAttr{}
	}
	hero.Attrs[talentAttr] = &models.HeroAttr{
		Fixed:   d.Attr.Attr.AttrFixed,
		Percent: d.Attr.Attr.AttrPercent,
	}
	hero.TalentBuff = d.Attr.TalentBuff
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getHeroEquippedEquipId(hero)...)
	if err != nil {
		return err
	}
	svc.refreshHeroAttr(ctx, hero, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill, 0)
	hero.BuildId = d.ConfigId
	if err := heroDao.Save(ctx, hero); err != nil {
		return err
	}

	heroModel := svc.dao2model(ctx, hero)
	dataItem := &event.HeroAttrUpdateItem{
		// IsSkillChange: true,
		Hero: heroModel,
	}
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: []*event.HeroAttrUpdateItem{dataItem},
	})
	ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: []*models.Hero{heroModel},
	})
	return nil
}

func (svc *Service) HandleSkillUpdate(c *ctx.Context, d *event.SkillChange) *errmsg.ErrMsg {
	cfg, ok := rule.GetHero(c, d.ConfigId)
	if !ok {
		return errmsg.NewErrHeroNotFound()
	}
	heroDao := dao.NewHero(c.RoleId)
	hero, ok, err := heroDao.GetOne(c, cfg.OriginId)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	newSkills := make(map[values.HeroSkillId]*pbdao.SkillLevel)
	// 保留初始技能的存档（event.SkillChange不会改变初始技能）
	for _, id := range cfg.AtKSkill {
		sl, ok := hero.Skill[id]
		if ok {
			newSkills[id] = sl
		}
	}
	for skillId, skillAdv := range d.Skills.Skills {
		skillCfg, ok := rule.GetSkillById(c, skillId)
		if !ok {
			svc.log.Warn("skill config not found", zap.Int64("id", skillId))
			continue
		}
		if hero.Skill == nil {
			hero.Skill = map[int64]*pbdao.SkillLevel{}
		}
		levelInfo := &pbdao.SkillLevel{
			Talent: 0,
			Equip:  map[int64]int64{},
			Role:   0,
		}
		info, ok := hero.Skill[skillCfg.SkillBaseId]
		if ok {
			levelInfo = info
		}
		levelInfo.Stones = skillAdv.Stones
		levelInfo.Talent = skillId - skillCfg.SkillBaseId
		if levelInfo.Talent < 0 {
			levelInfo.Talent = 0
		}
		newSkills[skillCfg.SkillBaseId] = levelInfo
	}
	hero.Skill = newSkills
	if d.Skills != nil && d.Skills.StoneAttr != nil {
		hero.Attrs[talentSkillAttr] = &models.HeroAttr{
			Fixed:   d.Skills.StoneAttr.AttrFixed,
			Percent: d.Skills.StoneAttr.AttrPercent,
		}
	}
	role, err := svc.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return err
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(c)
	if err != nil {
		return err
	}
	equips, err := svc.GetManyEquipBagMap(c, c.RoleId, svc.getHeroEquippedEquipId(hero)...)
	if err != nil {
		return err
	}
	var specialStoneAtk values.Integer
	if d.Skills != nil {
		specialStoneAtk = d.Skills.SpecialStoneAtk
	}
	svc.refreshHeroAttr(c, hero, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill, specialStoneAtk)
	if err := heroDao.Save(c, hero); err != nil {
		return err
	}

	heroModel := svc.dao2model(c, hero)
	dataItem := &event.HeroAttrUpdateItem{
		IsSkillChange: true,
		Hero:          heroModel,
	}
	c.PublishEventLocal(&event.HeroAttrUpdate{
		Data: []*event.HeroAttrUpdateItem{dataItem},
	})
	c.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: []*models.Hero{heroModel},
	})
	return nil
}

func (svc *Service) HandleRoleSkillUpdateFinish(ctx *ctx.Context, d *event.RoleSkillUpdateFinish) *errmsg.ErrMsg {
	if d.Skill == nil {
		d.Skill = []values.HeroSkillId{}
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	heroDao := dao.NewHero(ctx.RoleId)
	heroes, err := heroDao.Get(ctx)
	if err != nil {
		return err
	}
	roleAttr, err := svc.UserService.GetRoleAttr(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getAllHeroEquippedEquipId(heroes)...)
	if err != nil {
		return err
	}
	svc.refreshAllHeroAttr(ctx, heroes, role.Level, role.LevelIndex, false, equips, roleAttr, d.Skill)
	if err := heroDao.Save(ctx, heroes...); err != nil {
		return err
	}
	list := make([]*models.Hero, 0, len(heroes))
	heroAttrUpdateList := make([]*event.HeroAttrUpdateItem, 0, len(heroes))
	for _, hero := range heroes {
		heroModel := svc.dao2model(ctx, hero)
		list = append(list, heroModel)
		heroAttrUpdateList = append(heroAttrUpdateList, &event.HeroAttrUpdateItem{
			IsSkillChange: false,
			Hero:          heroModel,
		})
	}
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: heroAttrUpdateList,
	})
	ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: list,
	})
	return nil
}

func (svc *Service) HandleTaskChange(ctx *ctx.Context, d *event.TargetUpdate) *errmsg.ErrMsg {
	list := rule.GetBiographyByTaskType(ctx, d.Typ)
	if len(list) <= 0 {
		return nil
	}
	ids := make([]values.Integer, 0)
	for _, biography := range list {
		if biography.NeedCount <= d.Count {
			ids = append(ids, biography.Id)
		}
	}
	if len(ids) <= 0 {
		return nil
	}
	idList, err := svc.biographyCanGetList(ctx, d.Typ)
	if err != nil {
		return err
	}
	b, err := dao.GetBiography(ctx)
	if err != nil {
		return err
	}
	if b.BiographyId == nil {
		b.BiographyId = map[int64]int64{}
	}
	for _, id := range ids {
		if _, ok := b.BiographyId[id]; !ok {
			idList = append(idList, id)
		}
	}
	ctx.PushMessage(&servicepb.Hero_BiographyCanGetRewardPush{
		List: idList,
	})

	ctx.PublishEventLocal(&event.RedPointChange{
		RoleId: ctx.RoleId,
		Key:    enum.RedPointBiographyKey,
		Val:    values.Integer(len(idList)),
	})
	return nil
}

func (svc *Service) HandleBlessActivated(ctx *ctx.Context, d *event.BlessActivated) *errmsg.ErrMsg {
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	heroDao := dao.NewHero(ctx.RoleId)
	heroes, err := heroDao.Get(ctx)
	if err != nil {
		return err
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getAllHeroEquippedEquipId(heroes)...)
	if err != nil {
		return err
	}

	svc.refreshAllHeroAttrByGuildBless(ctx, heroes, role.Level, role.LevelIndex, true, equips, roleAttr, roleSkill, d)
	if err := heroDao.Save(ctx, heroes...); err != nil {
		return err
	}
	list := make([]*models.Hero, 0, len(heroes))
	heroAttrUpdateList := make([]*event.HeroAttrUpdateItem, 0, len(heroes))
	for _, hero := range heroes {
		heroModel := svc.dao2model(ctx, hero)
		list = append(list, heroModel)
		heroAttrUpdateList = append(heroAttrUpdateList, &event.HeroAttrUpdateItem{
			IsSkillChange: false,
			Hero:          heroModel,
		})
	}
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: heroAttrUpdateList,
	})

	ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: list,
	})

	return nil
}
