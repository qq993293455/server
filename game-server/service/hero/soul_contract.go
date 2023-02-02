package hero

import (
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

func (svc *Service) UpgradeSoulContract(ctx *ctx.Context, req *servicepb.Hero_UpgradeSoulContractRequest) (*servicepb.Hero_UpgradeSoulContractResponse, *errmsg.ErrMsg) {
	heroDao := dao.NewHero(ctx.RoleId)
	heroes, err := heroDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	var soulLv values.Integer
	var hero *pbdao.Hero
	for _, d := range heroes {
		if d.Id == req.HeroOriginId {
			hero = d
		}
		if d.SoulContract != nil && soulLv < d.SoulContract.Rank {
			soulLv = d.SoulContract.Rank
		}
	}
	// hero, ok, err := heroDao.GetOne(ctx, req.HeroOriginId)
	// if err != nil {
	// 	return nil, err
	// }
	if hero == nil {
		return nil, errmsg.NewErrHeroNotFound()
	}
	max, ok := rule.GetMaxSoulContract(ctx, req.HeroOriginId)
	if !ok {
		return nil, errmsg.NewInternalErr("MaxSoulContract not exist")
	}
	if hero.SoulContract.Rank >= max.Rank && hero.SoulContract.Level >= max.Level {
		return nil, errmsg.NewErrSoulContractMaxLevel()
	}
	cfg, ok := rule.GetSoulContract(ctx, req.HeroOriginId, hero.SoulContract.Rank, hero.SoulContract.Level)
	if !ok {
		return nil, errmsg.NewInternalErr("SoulContract not exist")
	}
	next, ok := rule.GetNextSoulContract(ctx, req.HeroOriginId, hero.SoulContract.Rank, hero.SoulContract.Level)
	if !ok {
		return nil, errmsg.NewInternalErr("NextSoulContract not exist")
	}
	// 消耗
	if len(cfg.RankCost) > 0 {
		if err := svc.SubManyItem(ctx, ctx.RoleId, cfg.RankCost); err != nil {
			return nil, err
		}
	}
	// 物品奖励
	if len(cfg.RankItemReward) > 0 {
		if _, err := svc.AddManyItem(ctx, ctx.RoleId, cfg.RankItemReward); err != nil {
			return nil, err
		}
	}
	// 判断限定条件满不满足
	if len(cfg.Condition) > 0 && len(cfg.Condition) == 3 {
		typ := models.TaskType(cfg.Condition[0])
		v := cfg.Condition[1]
		need := cfg.Condition[2]
		data, err := svc.GetCounterByType(ctx, typ)
		if err != nil {
			return nil, err
		}
		if data[v] < need {
			return nil, errmsg.NewErrSoulContractConditionNotEnough()
		}
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	hero.SoulContract.Rank = next.Rank
	hero.SoulContract.Level = next.Level
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getHeroEquippedEquipId(hero)...)
	if err != nil {
		return nil, err
	}
	svc.refreshHeroAttr(ctx, hero, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill, 0)
	if err := heroDao.Save(ctx, hero); err != nil {
		return nil, err
	}

	heroModel := svc.dao2model(ctx, hero)
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: []*event.HeroAttrUpdateItem{{
			IsSkillChange: false,
			Hero:          heroModel,
		}},
	})
	// 天赋点奖励
	if cfg.GiftPoint > 0 {
		ctx.PublishEventLocal(&event.GainTalentParticularPoint{
			ConfigId: hero.Id,
			Num:      cfg.GiftPoint,
		})
	}
	if next.Rank > soulLv {
		soulLv = next.Rank
	}
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskSoulContractTotalLevel: {
			Typ:     values.Integer(models.TaskType_TaskSoulContractTotalLevel),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskSoulContractHeroLevel: {
			Typ:     values.Integer(models.TaskType_TaskSoulContractHeroLevel),
			Id:      hero.Id,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskSoulContractUpdateNum: {
			Typ:     values.Integer(models.TaskType_TaskSoulContractUpdateNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskSoulLvlUpCnt: {
			Typ:     values.Integer(models.TaskType_TaskSoulLvlUpCnt),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskSoulLvlReach: {
			Typ:     values.Integer(models.TaskType_TaskSoulLvlReach),
			Id:      0,
			Cnt:     soulLv,
			Replace: true,
		},
	}
	svc.UpdateTargets(ctx, ctx.RoleId, tasks)

	return &servicepb.Hero_UpgradeSoulContractResponse{
		Hero:   heroModel,
		Reward: cfg.RankItemReward,
	}, nil
}
