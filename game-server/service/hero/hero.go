package hero

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/common/values/enum/AttrId"
	"coin-server/common/values/enum/ItemType"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/hero/dao"
	"coin-server/game-server/service/hero/rule"
	rule2 "coin-server/rule"
	rulemodel "coin-server/rule/rule-model"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewHeroService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		log:        log,
		Module:     module,
	}
	module.HeroService = s
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取玩家拥有的英雄id列表", svc.List)
	svc.svc.RegisterFunc("装备槽升星", svc.UpgradeSlot)
	svc.svc.RegisterFunc("穿装备", svc.WearEquip)
	svc.svc.RegisterFunc("脱装备", svc.TakeDownEquip)
	// svc.svc.RegisterFunc("装备熔炼", svc.MeltEquip)
	svc.svc.RegisterFunc("装备熔炼升星", svc.UpgradeSlotMeltStar)
	// 附魔
	svc.svc.RegisterFunc("获取附魔信息", svc.GetEnchantInfo)
	svc.svc.RegisterFunc("附魔（生成词条）", svc.EnchantGen)
	svc.svc.RegisterFunc("替换附魔", svc.EnchantReplace)
	svc.svc.RegisterFunc("丢弃附魔", svc.EnchantDrop)

	// 传记
	svc.svc.RegisterFunc("获取单个英雄传记可领取的奖励信息", svc.BiographyRewardInfo)
	svc.svc.RegisterFunc("领取英雄传记奖励", svc.BiographyGetReward)

	// 魂契
	svc.svc.RegisterFunc("魂契升级", svc.UpgradeSoulContract)

	// 共鸣
	svc.svc.RegisterFunc("激活共鸣", svc.Activate)

	// 时装
	svc.svc.RegisterFunc("穿时装", svc.DressFashion)

	// 作弊器
	svc.svc.RegisterFunc("添加英雄", svc.CheatAddHero)
	svc.svc.RegisterFunc("满强化满精炼", svc.CheatSetSlotMax)
	svc.svc.RegisterFunc("魂契养成满级", svc.CheatSetSoulContractMax)

	// 注册Hero类型道具处理方法
	svc.BagService.RegisterUpdaterByType(ItemType.Hero, func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg {
		cfg, ok := rule2.MustGetReader(ctx).Item.GetItemById(itemId)
		if !ok {
			panic(errors.New("item_config_not_found: " + strconv.Itoa(int(itemId))))
		}
		_, err := svc.AddHero(ctx, cfg.TargetId, false)
		return err
	})

	svc.svc.RegisterFunc("时装过期", svc.FashionExpired)

	eventlocal.SubscribeEventLocal(svc.HandlerUserLevelUp)
	eventlocal.SubscribeEventLocal(svc.HandlerRoleAttrUpdate)
	eventlocal.SubscribeEventLocal(svc.HandleTalentUpdate)
	eventlocal.SubscribeEventLocal(svc.HandleSkillUpdate)
	eventlocal.SubscribeEventLocal(svc.HandleRoleSkillUpdateFinish)
	eventlocal.SubscribeEventLocal(svc.HandleTaskChange)
	eventlocal.SubscribeEventLocal(svc.HandleBlessActivated)
}

func (svc *Service) AddHero(ctx *ctx.Context, id values.HeroId, register bool) (*models.Hero, *errmsg.ErrMsg) {
	originId, err := svc.getHeroRealId(ctx, id)
	if err != nil {
		return nil, err
	}
	var level, levelIndex values.Level
	if register {
		level = enum.INIT_LEVEL
		levelIndex = enum.INIT_LEVEL
	} else {
		role, err := svc.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
		if err != nil {
			return nil, err
		}
		level = role.Level
		levelIndex = role.LevelIndex
	}
	cfg, ok := rule.GetHero(ctx, id)
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}

	// skillList := svc.GetInitSkills(cfg)
	// skillMap := make(map[values.HeroSkillId]*pbdao.SkillLevel)
	// for _, skillId := range skillList {
	// 	skillMap[skillId] = &pbdao.SkillLevel{}
	// }
	hero := &pbdao.Hero{
		Id: cfg.OriginId,
		// Skill:     skillMap,
		Attrs:     nil,
		EquipSlot: nil,
		BuildId:   id,
		SoulContract: &models.SoulContract{
			Rank:  1,
			Level: 1,
		},
		Fashion: &models.HeroFashion{
			Dressed: rule.GetDefaultFashion(ctx, originId),
			Data:    nil,
		},
	}
	dao.NilInit(hero)
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	svc.refreshHeroAttr(ctx, hero, level, levelIndex, true, nil, roleAttr, roleSkill, 0)

	if err := dao.NewHero(ctx.RoleId).Save(ctx, hero); err != nil {
		return nil, err
	}
	if err := svc.handleRedPointByGetNewHero(ctx, cfg.OriginId); err != nil {
		return nil, err
	}
	svc.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskRoleGetNum, cfg.OriginId, 1)

	heroModel := svc.dao2model(ctx, hero)
	if !register {
		ctx.PublishEventLocal(&event.HeroAttrUpdate{
			Data: []*event.HeroAttrUpdateItem{{
				IsSkillChange: false,
				Hero:          heroModel,
			}},
		})
	}
	ctx.PublishEventLocal(&event.GotHero{OriginId: hero.Id})
	ctx.PushMessage(&servicepb.Hero_GotHeroPush{
		Hero: []*models.Hero{heroModel},
	})
	return heroModel, nil
}

func (svc *Service) GetHero(ctx *ctx.Context, roleId values.RoleId, id values.HeroId) (*models.Hero, bool, *errmsg.ErrMsg) {
	hero, ok, err := dao.NewHero(roleId).GetOne(ctx, id)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	// formRoleSkills, err := svc.GetRoleSkill(ctx, ctx.RoleId)
	// if err != nil {
	// 	return nil, false, err
	// }
	return svc.dao2model(ctx, hero), true, nil
}

func (svc *Service) GetAllHero(ctx *ctx.Context, roleId values.RoleId) ([]*models.Hero, *errmsg.ErrMsg) {
	if roleId == "" {
		roleId = ctx.RoleId
	}
	heroes, err := dao.NewHero(roleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	// formRoleSkills, err := svc.GetRoleSkill(ctx, ctx.RoleId)
	// if err != nil {
	// 	return nil, err
	// }
	list := make([]*models.Hero, 0, len(heroes))
	for _, hero := range heroes {
		list = append(list, svc.dao2model(ctx, hero))
	}
	return list, nil
}

func (svc *Service) GetHeroes(ctx *ctx.Context, roleId values.RoleId, ids []values.HeroId) ([]*models.Hero, *errmsg.ErrMsg) {
	if roleId == "" {
		roleId = ctx.RoleId
	}
	if len(ids) == 0 {
		return nil, nil
	}
	heroes, err := dao.NewHero(roleId).GetSome(ctx, ids)
	if err != nil {
		return nil, err
	}
	// formRoleSkills, err := svc.GetRoleSkill(ctx, ctx.RoleId)
	// if err != nil {
	// 	return nil, err
	// }
	list := make([]*models.Hero, 0, len(heroes))
	for _, hero := range heroes {
		list = append(list, svc.dao2model(ctx, hero))
	}
	return list, nil
}

func (svc *Service) List(ctx *ctx.Context, _ *servicepb.Hero_HeroListRequest) (*servicepb.Hero_HeroListResponse, *errmsg.ErrMsg) {
	heroes, err := dao.NewHero(ctx.RoleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]values.Integer, 0, len(heroes))
	for _, hero := range heroes {
		list = append(list, hero.Id)
	}
	return &servicepb.Hero_HeroListResponse{
		HeroId: list,
	}, nil
}

func (svc *Service) UpgradeSlot(ctx *ctx.Context, req *servicepb.Hero_UpgradeSlotRequest) (*servicepb.Hero_UpgradeSlotResponse, *errmsg.ErrMsg) {
	heroId, err := svc.getHeroRealId(ctx, req.HeroId)
	if err != nil {
		return nil, err
	}
	hero, ok, err := dao.NewHero(ctx.RoleId).GetOne(ctx, heroId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	if hero.EquipSlot == nil {
		hero.EquipSlot = map[int64]*models.HeroEquipSlot{}
	}
	if req.SlotId == -1 {
		return svc.upgradeAllSlot(ctx, hero)
	}
	return svc.upgradeOneSlot(ctx, req.SlotId, hero)
}

func (svc *Service) upgradeAllSlot(ctx *ctx.Context, hero *pbdao.Hero) (*servicepb.Hero_UpgradeSlotResponse, *errmsg.ErrMsg) {
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	data := &UpgradeData{
		Cost:     make(map[values.ItemId]values.Integer),
		Upgraded: make(map[values.EquipSlot]*models.HeroEquipSlot),
		Count:    0,
	}
	allCostItemId := rule.GetUpgradeEquipAllCostItemId(ctx)
	items, err := svc.GetManyItem(ctx, ctx.RoleId, allCostItemId)
	if err != nil {
		return nil, err
	}
	disable := make(map[values.EquipSlot]struct{})
	err = svc.upgradeSlotOnce(ctx, hero, role.Level, items, disable, data)
	// 一次都没升级成功
	if err != nil {
		return nil, err
	}
	if data.Count <= 0 {
		return nil, errmsg.NewErrEquipNoUpgradeable()
	}

	if len(data.Cost) > 0 {
		if err := svc.SubManyItem(ctx, ctx.RoleId, data.Cost); err != nil {
			return nil, err
		}
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getHeroEquippedEquipId(hero)...)
	if err != nil {
		return nil, err
	}
	saveEquips := make([]*models.Equipment, 0)
	equipBonusTalent := make(map[values.TalentId]values.Level)
	for _, info := range data.Upgraded {
		equip, ok := equips[info.EquipId]
		if !ok {
			ctx.Warn("equip not found", zap.String("role_id", ctx.RoleId), zap.String("equip_id", info.EquipId))
			continue
		}
		last := svc.GetEquipTalentBonus(ctx, equip, true, false)
		if svc.UnlockEquipAffix(ctx, info.Star, equip) {
			current := svc.GetEquipTalentBonus(ctx, equip, true, false)
			saveEquips = append(saveEquips, equip)
			temp := svc.formatEquipBonusTalentData(current, last)
			for id, level := range temp {
				equipBonusTalent[id] += level
			}
		}
	}
	if len(equipBonusTalent) > 0 {
		ctx.PublishEventLocal(&event.EquipBonusTalent{
			HeroId: hero.Id,
			Data:   equipBonusTalent,
		})
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}

	svc.refreshEquipResonanceStatus(ctx, hero)

	svc.refreshHeroAttr(ctx, hero, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill, 0)
	if err := dao.NewHero(ctx.RoleId).Save(ctx, hero); err != nil {
		return nil, err
	}
	if len(saveEquips) > 0 {
		svc.SaveManyEquipment(ctx, ctx.RoleId, saveEquips)
	}

	if data.Count > 0 {
		tasks := map[models.TaskType]*models.TaskUpdate{
			models.TaskType_TaskEnhanceEquipLevel: {
				Typ:     values.Integer(models.TaskType_TaskEnhanceEquipLevel),
				Id:      0,
				Cnt:     data.Count,
				Replace: false,
			},
			models.TaskType_TaskEnhanceNum: {
				Typ:     values.Integer(models.TaskType_TaskEnhanceNum),
				Id:      0,
				Cnt:     data.Count,
				Replace: false,
			},
		}
		svc.UpdateTargets(ctx, ctx.RoleId, tasks)
	}
	heroModel := svc.dao2model(ctx, hero)
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: []*event.HeroAttrUpdateItem{{
			IsSkillChange: false,
			Hero:          heroModel,
		}},
	})

	return &servicepb.Hero_UpgradeSlotResponse{
		Hero: heroModel,
	}, nil
}

func (svc *Service) upgradeSlotOnce(ctx *ctx.Context, hero *pbdao.Hero, level values.Level, items map[values.ItemId]values.Integer, disable map[values.EquipSlot]struct{}, data *UpgradeData) *errmsg.ErrMsg {
	id, info := svc.getUpgradeSlot(hero, disable)
	if id <= 0 {
		return nil
	}
	cfg, ok := rule.GetUpgradeEquipStarCost(ctx, info.Star)
	if !ok {
		ctx.Warn("equip_star config not found", zap.Int64("star", info.Star))
		return errmsg.NewInternalErr("equip_star config not found")
	}
	if cfg.HeroLv > level {
		disable[id] = struct{}{}
		return svc.upgradeSlotOnce(ctx, hero, level, items, disable, data)
	}
	enough := true
	for costId, costCount := range cfg.Cost {
		if items[costId] < costCount {
			enough = false
			break
		}
	}
	if enough {
		for costId, costCount := range cfg.Cost {
			items[costId] -= costCount
			data.Cost[costId] += costCount
		}
		hero.EquipSlot[id].Star++
		data.Count++
		data.Upgraded[id] = hero.EquipSlot[id]
	} else {
		disable[id] = struct{}{}
	}

	return svc.upgradeSlotOnce(ctx, hero, level, items, disable, data)
}

func (svc *Service) getUpgradeSlot(hero *pbdao.Hero, disable map[values.EquipSlot]struct{}) (values.EquipSlot, *models.HeroEquipSlot) {
	var (
		info *models.HeroEquipSlot
		id   values.EquipSlot
	)

	for slotId, slotInfo := range hero.EquipSlot {
		if slotInfo == nil || slotInfo.EquipId == "" {
			continue
		}
		if _, ok := disable[slotId]; ok {
			continue
		}
		if info == nil || info.Star > slotInfo.Star {
			info = slotInfo
			id = slotId
		}
	}
	return id, info
}

func (svc *Service) upgradeOneSlot(ctx *ctx.Context, slotId values.EquipSlot, hero *pbdao.Hero) (*servicepb.Hero_UpgradeSlotResponse, *errmsg.ErrMsg) {
	if _, ok := rule.GetEquipSlot(ctx, slotId); !ok {
		return nil, errmsg.NewErrHeroInvalidSlot()
	}
	item, ok := hero.EquipSlot[slotId]
	if !ok || item.EquipId == "" {
		return nil, errmsg.NewErrHeroNeedEquip()
	}
	if item.Star >= rule.GetMaxEquipStar(ctx) {
		return nil, errmsg.NewErrHeroMaxEquipStar()
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getHeroEquippedEquipId(hero)...)
	if err != nil {
		return nil, err
	}
	equip, ok := equips[item.EquipId]
	if !ok {
		return nil, errmsg.NewErrHeroNeedEquip()
	}
	cfg, ok := rule.GetUpgradeEquipStarCost(ctx, item.Star)
	if !ok {
		return nil, errmsg.NewInternalErr("equip start config not found")
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if role.Level < cfg.HeroLv {
		return nil, errmsg.NewErrRoleLevelNotEnough()
	}
	if cfg.Cost != nil {
		if err := svc.SubManyItem(ctx, ctx.RoleId, cfg.Cost); err != nil {
			return nil, err
		}
	} else {
		svc.log.Warn("UpgradeSlot cost is nil", zap.Int64("star", item.Star))
	}
	item.Star++
	last := svc.GetEquipTalentBonus(ctx, equip, true, false)
	if svc.UnlockEquipAffix(ctx, item.Star, equip) {
		current := svc.GetEquipTalentBonus(ctx, equip, true, false)
		if err := svc.SaveEquipment(ctx, ctx.RoleId, equip); err != nil {
			return nil, err
		}
		// item.Equip = equip
		ctx.PublishEventLocal(&event.EquipBonusTalent{
			HeroId: hero.Id,
			Data:   svc.formatEquipBonusTalentData(current, last),
		})
	}
	hero.EquipSlot[slotId] = item
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	svc.refreshEquipResonanceStatus(ctx, hero)
	svc.refreshHeroAttr(ctx, hero, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill, 0)
	if err := dao.NewHero(ctx.RoleId).Save(ctx, hero); err != nil {
		return nil, err
	}

	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskEnhanceEquipLevel: {
			Typ:     values.Integer(models.TaskType_TaskEnhanceEquipLevel),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskEnhanceNum: {
			Typ:     values.Integer(models.TaskType_TaskEnhanceNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	svc.UpdateTargets(ctx, ctx.RoleId, tasks)
	heroModel := svc.dao2model(ctx, hero)
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: []*event.HeroAttrUpdateItem{{
			IsSkillChange: false,
			Hero:          heroModel,
		}},
	})
	return &servicepb.Hero_UpgradeSlotResponse{
		Hero: heroModel,
	}, nil
}

func (svc *Service) WearEquip(ctx *ctx.Context, req *servicepb.Hero_WearEquipRequest) (*servicepb.Hero_WearEquipResponse, *errmsg.ErrMsg) {
	heroId, err := svc.getHeroRealId(ctx, req.HeroId)
	if err != nil {
		return nil, err
	}
	hero, ok, err := dao.NewHero(ctx.RoleId).GetOne(ctx, heroId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	role, err := svc.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	equipIdList := svc.getHeroEquippedEquipId(hero)
	equipIdList = append(equipIdList, req.EquipId...)

	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, equipIdList...)
	if err != nil {
		return nil, err
	}
	updateEquips := make([]*models.Equipment, 0)
	itemId := make([]values.ItemId, 0)
	var wearCount values.Integer
	currentBonusData := make(map[values.TalentId]values.Level)
	lastBonusData := make(map[values.TalentId]values.Level)
	var capacity values.Integer
	for _, equipId := range req.EquipId {
		equip, ok := equips[equipId]
		if !ok {
			return nil, errmsg.NewErrEquipNotFound()
		}
		if equip.HeroId > 0 {
			return nil, errmsg.NewErrEquippedOtherHero()
		}
		itemId = append(itemId, equip.ItemId)
		updateEquips = append(updateEquips, equip)
		oldEquipId, _, firstWear, err := svc.wearOneEquip(ctx, hero, equip, role.Level)
		if err != nil {
			return nil, err
		}
		if firstWear {
			wearCount++
		}
		// refine, ok := rule.GetEquipRefine(ctx, slotId, meltLevel)
		// if !ok {
		// 	return nil, errmsg.NewInternalErr("equip refine config not found")
		// }
		// svc.refreshEquipAffixBonus(equip, refine)
		current := svc.GetEquipTalentBonus(ctx, equip, true, false)
		currentBonusData = svc.mergeEquipBonusTalentData(currentBonusData, current)
		if oldEquipId != "" {
			oldEquip, ok := equips[oldEquipId]
			if ok {
				delete(equips, oldEquipId)
				last := svc.GetEquipTalentBonus(ctx, oldEquip, true, false)
				lastBonusData = svc.mergeEquipBonusTalentData(lastBonusData, last)
				svc.clearEquipAffixBonus(oldEquip)
				svc.UnlockEquipAffix(ctx, enum.InitEquipStar, oldEquip)
				oldEquip.HeroId = 0
				updateEquips = append(updateEquips, oldEquip)
			}
		} else {
			capacity++
		}
	}
	if capacity > 0 {
		bagConfig, err := svc.GetBagConfig(ctx)
		if err != nil {
			return nil, err
		}
		bagConfig.Config.CapacityOccupied -= capacity
		if bagConfig.Config.CapacityOccupied < 0 {
			bagConfig.Config.CapacityOccupied = 0
		}
		svc.SaveBagConfig(ctx, bagConfig)
	}

	bonusData := svc.formatEquipBonusTalentData(currentBonusData, lastBonusData)
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	svc.refreshHeroAttr(ctx, hero, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill, 0)
	if err := dao.NewHero(ctx.RoleId).Save(ctx, hero); err != nil {
		return nil, err
	}
	svc.SaveManyEquipment(ctx, ctx.RoleId, updateEquips)
	// svc.SaveEquipmentBrief(ctx, ctx.RoleId, updateEquips...)

	heroModel := svc.dao2model(ctx, hero)
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: []*event.HeroAttrUpdateItem{{
			IsSkillChange: false,
			Hero:          heroModel,
		}},
	})
	if len(bonusData) > 0 {
		ctx.PublishEventLocal(&event.EquipBonusTalent{
			HeroId: hero.Id,
			Data:   bonusData,
		})
	}
	// 通知客户端，客户端把该装备从背包里隐藏
	ctx.PublishEventLocal(&event.EquipUpdate{
		RoleId: ctx.RoleId,
		Equips: updateEquips,
	})
	if wearCount > 0 {
		svc.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskWearEquipNumAcc, 0, wearCount)
	}
	return &servicepb.Hero_WearEquipResponse{
		Hero:   heroModel,
		ItemId: itemId,
	}, nil
}

func (svc *Service) wearOneEquip(ctx *ctx.Context, hero *pbdao.Hero, equip *models.Equipment, level values.Level) (values.EquipId, values.Level, bool, *errmsg.ErrMsg) {
	var firstWear bool
	levelLimit := rule.GetEquipLevelLimit(ctx, equip.Level)
	if level < levelLimit {
		return "", 0, firstWear, errmsg.NewErrRoleLevelNotEnough()
	}
	equipConfig, ok := rule.GetEquipByItemId(ctx, equip.ItemId)
	if !ok {
		return "", 0, firstWear, errmsg.NewErrEquipNotFound()
	}

	var find bool
	// 不限定职业
	if len(equipConfig.EquipJob) == 1 && equipConfig.EquipJob[0] == 0 {
		find = true
	} else {
		for _, v := range equipConfig.EquipJob {
			if v == hero.Id {
				find = true
				break
			}
		}
	}
	if !find {
		return "", 0, firstWear, errmsg.NewErrEquipHeroNotMatch()
	}
	// if hero.EquipSlot == nil {
	// 	hero.EquipSlot = make(map[int64]*models.HeroEquipSlot, 0)
	// }
	slotCfg, ok := rule.GetEquipSlot(ctx, equipConfig.EquipSlot)
	if !ok {
		return "", 0, firstWear, errmsg.NewErrEquipSlotNotUnlock()
	}
	sysCfg, ok := rule.GetSystemById(ctx, values.SystemId(slotCfg.SystemId))
	if !ok {
		return "", 0, firstWear, errmsg.NewErrEquipSlotNotUnlock()
	}
	if len(sysCfg.UnlockCondition) >= 3 {
		taskData, err := svc.GetCounterByType(ctx, models.TaskType(sysCfg.UnlockCondition[0]))
		if err != nil {
			return "", 0, firstWear, err
		}
		if taskData[0] < sysCfg.UnlockCondition[2] {
			return "", 0, firstWear, errmsg.NewErrEquipSlotNotUnlock()
		}
	}

	slotInfo, ok := hero.EquipSlot[equipConfig.EquipSlot]
	if !ok {
		slotInfo = &models.HeroEquipSlot{
			Star:      enum.InitEquipStar, // 首次装备 星级为0
			MeltLevel: 0,
		}
		firstWear = true
	}
	oldEquipId := slotInfo.EquipId
	// oldEquipInfo := slotInfo.Equip
	// slotInfo.Equip = equip
	slotInfo.EquipId = equip.EquipId
	slotInfo.EquipItemId = equip.ItemId
	equip.HeroId = hero.Id
	// 处理是否有词缀要激活
	svc.UnlockEquipAffix(ctx, slotInfo.Star, equip)
	hero.EquipSlot[equipConfig.EquipSlot] = slotInfo
	return oldEquipId, slotInfo.MeltLevel, firstWear, nil
}

func (svc *Service) TakeDownEquip(ctx *ctx.Context, req *servicepb.Hero_TakeDownEquipRequest) (*servicepb.Hero_TakeDownEquipResponse, *errmsg.ErrMsg) {
	heroId, err := svc.getHeroRealId(ctx, req.HeroId)
	if err != nil {
		return nil, err
	}
	hero, ok, err := dao.NewHero(ctx.RoleId).GetOne(ctx, heroId)
	if err != nil {
		return nil, err
	}
	role, err := svc.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	slotInfo, ok := hero.EquipSlot[req.SlotId]
	if !ok || slotInfo.EquipId == "" {
		return nil, errmsg.NewErrSlotHasNotEquip()
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getHeroEquippedEquipId(hero)...)
	if err != nil {
		return nil, err
	}
	equip, ok := equips[slotInfo.EquipId]
	if !ok {
		return nil, errmsg.NewErrEquipNotFound()
	}
	bagConfig, err := svc.GetBagConfig(ctx)
	if err != nil {
		return nil, err
	}
	if bagConfig.Config.Capacity <= bagConfig.Config.CapacityOccupied {
		return nil, errmsg.NewErrBagCapLimit()
	}
	bagConfig.Config.CapacityOccupied++
	svc.SaveBagConfig(ctx, bagConfig)

	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	svc.clearEquipAffixBonus(equip)
	// 将英雄对应slot的装备
	slotInfo.EquipId = ""
	slotInfo.EquipItemId = 0
	// 将装备的英雄id清空
	equip.HeroId = 0
	bonusData := svc.GetEquipTalentBonus(ctx, equip, true, true)
	// 将装备的词缀的解锁状态置为0星时的状态
	svc.UnlockEquipAffix(ctx, enum.InitEquipStar, equip)
	hero.EquipSlot[req.SlotId] = slotInfo
	delete(equips, equip.EquipId)
	svc.refreshHeroAttr(ctx, hero, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill, 0)
	if err := dao.NewHero(ctx.RoleId).Save(ctx, hero); err != nil {
		return nil, err
	}
	if err := svc.SaveEquipment(ctx, ctx.RoleId, equip); err != nil {
		return nil, err
	}
	// svc.SaveEquipmentBrief(ctx, ctx.RoleId, equip)

	heroModel := svc.dao2model(ctx, hero)
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: []*event.HeroAttrUpdateItem{{
			IsSkillChange: false,
			Hero:          heroModel,
		}},
	})

	if len(bonusData) > 0 {
		ctx.PublishEventLocal(&event.EquipBonusTalent{
			HeroId: hero.Id,
			Data:   bonusData,
		})
	}
	// 通知客户端，客户端把该装备从背包里显示出来
	ctx.PublishEventLocal(&event.EquipUpdate{
		RoleId: ctx.RoleId,
		Equips: []*models.Equipment{equip},
	})

	return &servicepb.Hero_TakeDownEquipResponse{
		Hero:   heroModel,
		ItemId: equip.ItemId,
	}, nil
}

func (svc *Service) MeltEquip(ctx *ctx.Context, req *servicepb.Hero_MeltEquipRequest) (*servicepb.Hero_MeltEquipResponse, *errmsg.ErrMsg) {
	if len(req.EquipId) <= 0 {
		return nil, errmsg.NewErrNeedSelectEquip()
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, req.EquipId...)
	if err != nil {
		return nil, err
	}
	var exp values.Integer
	equipId := make([]values.EquipId, 0)
	for _, id := range req.EquipId {
		equip, ok := equips[id]
		if !ok {
			return nil, errmsg.NewErrEquipNotFound()
		}
		if equip.HeroId > 0 {
			return nil, errmsg.NewErrCanNotMeltEquipped()
		}
		equipId = append(equipId, equip.EquipId)
		cfg, ok := rule.GetEquipByItemId(ctx, equip.ItemId)
		if !ok {
			svc.log.Warn("equip config not found", zap.Int64("id", equip.ItemId))
			continue
		}
		exp += cfg.MeltingValue
	}
	if err := svc.DelEquipment(ctx, equipId...); err != nil {
		return nil, err
	}
	if exp <= 0 {
		return &servicepb.Hero_MeltEquipResponse{}, nil
	}

	if err := svc.AddItem(ctx, ctx.RoleId, enum.MeltExp, exp); err != nil {
		return nil, err
	}

	return &servicepb.Hero_MeltEquipResponse{}, nil
}

func (svc *Service) UpgradeSlotMeltStar(ctx *ctx.Context, req *servicepb.Hero_UpgradeSlotMeltStarRequest) (*servicepb.Hero_UpgradeSlotMeltStarResponse, *errmsg.ErrMsg) {
	heroId, err := svc.getHeroRealId(ctx, req.HeroId)
	if err != nil {
		return nil, err
	}
	hero, ok, err := dao.NewHero(ctx.RoleId).GetOne(ctx, heroId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}

	// 后端拦截是否解锁
	if !svc.isRefineUnlock(ctx, hero, req.SlotId) {
		return nil, errmsg.NewErrEquipRefineNotUnlock()
	}

	taskData, err := svc.GetCounterByType(ctx, models.TaskType_TaskRefineEquipSlotNum)
	if err != nil {
		return nil, err
	}
	refineEquipSlotNum := taskData[0]

	slot, ok := hero.EquipSlot[req.SlotId]
	if !ok {
		return nil, errmsg.NewErrHeroNeedEquip()
	}
	if slot.EquipId == "" {
		return nil, errmsg.NewErrHeroNeedEquip()
	}
	if slot.MeltLevel >= rule.GetMaxEquipMeltLevel(ctx, req.SlotId) {
		return nil, errmsg.NewErrMaxMeltLevel()
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getHeroEquippedEquipId(hero)...)
	if err != nil {
		return nil, err
	}
	equip, ok := equips[slot.EquipId]
	if !ok {
		return nil, errmsg.NewErrEquipNotFound()
	}

	cfg, ok := rule.GetEquipRefine(ctx, req.SlotId, slot.MeltLevel)
	if !ok {
		return nil, errmsg.NewInternalErr("equip refine config not found")
	}
	// cfgNext, ok := rule.GetEquipRefine(ctx, req.SlotId, slot.MeltLevel+1)
	// if !ok {
	// 	return nil, errmsg.NewInternalErr("equip refine config not found")
	// }

	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	// 已去掉等级判断，改回通过强化等级判断（isRefineUnlock函数）
	// if role.Level < cfg.HeroLv {
	// 	return nil, errmsg.NewErrRoleLevelNotEnough()
	// }
	// attrFixed, attrPercent, err := svc.GetRoleAttr(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	cost := make(map[values.ItemId]values.Integer)
	for id, val := range cfg.Cost {
		cost[id] += val
	}
	// 扣除的经验也配在cfg.Cost字段里
	// if cfg.CostExp > 0 {
	// 	cost[enum.MeltExp] = cfg.CostExp
	// }
	if err := svc.SubManyItem(ctx, ctx.RoleId, cost); err != nil {
		return nil, err
	}
	// svc.refreshEquipAffixBonus(equip, cfgNext)

	slot.MeltLevel++
	// slot.Equip = equip
	hero.EquipSlot[req.SlotId] = slot
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	svc.refreshHeroAttr(ctx, hero, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill, 0)
	if err := dao.NewHero(ctx.RoleId).Save(ctx, hero); err != nil {
		return nil, err
	}
	if err := svc.SaveEquipment(ctx, ctx.RoleId, equip); err != nil {
		return nil, err
	}

	heroModel := svc.dao2model(ctx, hero)
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: []*event.HeroAttrUpdateItem{{
			IsSkillChange: false,
			Hero:          heroModel,
		}},
	})
	tasks := make(map[models.TaskType]*models.TaskUpdate, 0)
	if slot.MeltLevel > refineEquipSlotNum {
		tasks[models.TaskType_TaskRefineEquipSlotNum] = &models.TaskUpdate{
			Typ:     values.Integer(models.TaskType_TaskRefineEquipSlotNum),
			Id:      0,
			Cnt:     slot.MeltLevel,
			Replace: true,
		}
	}
	tasks[models.TaskType_TaskRefineEquipSlotNumAcc] = &models.TaskUpdate{
		Typ:     values.Integer(models.TaskType_TaskRefineEquipSlotNum),
		Id:      0,
		Cnt:     1,
		Replace: false,
	}
	tasks[models.TaskType_TaskMeltNStarsUpper] = &models.TaskUpdate{
		Typ:     values.Integer(models.TaskType_TaskMeltNStarsUpper),
		Id:      slot.MeltLevel,
		Cnt:     1,
		Replace: false,
	}
	svc.UpdateTargets(ctx, ctx.RoleId, tasks)

	return &servicepb.Hero_UpgradeSlotMeltStarResponse{
		Hero: heroModel,
	}, nil
}

func (svc *Service) combineAffixValue(affix *models.Affix) {
	if len(affix.Bonus) <= 0 && affix.AffixId > 0 {
		return
	}
	for _, v := range affix.Bonus {
		affix.AffixValue += v
	}
}

func (svc *Service) dao2model(ctx *ctx.Context, hero *pbdao.Hero) *models.Hero {
	// 将词缀的加成值合并后发给客户端
	// slots := make(map[values.Integer]*models.HeroEquipSlot)
	// for i, slot := range hero.EquipSlot {
	// 	temp := &models.HeroEquipSlot{
	// 		Star:        slot.Star,
	// 		EquipId:     slot.EquipId,
	// 		EquipItemId: slot.EquipItemId,
	// 		MeltLevel:   slot.MeltLevel,
	// 		Enchant:     slot.Enchant,
	// 	}
	// 	slots[i] = temp
	// }
	attrs := make(map[values.Integer]values.Integer)
	for ac, item := range hero.Attrs {
		if ac == primaryAttr || ac == secondaryAttr {
			for id, val := range item.Fixed {
				attrs[id] += val
			}
			for id, val := range item.Percent {
				attrs[id] += val
			}
		}

	}
	attrs[AttrId.CritDam] += rule.GetBaceCritDam(ctx)
	// skills := make([]values.HeroSkillId, 0)
	// for _, skill := range hero.Skill {
	// 	skills = append(skills, skill.Skill...)
	// }
	// for _, skill := range formRoleSkill {
	// 	skills = append(skills, skill)
	// }
	buff := make([]values.HeroBuffId, 0)
	buffMap := make(map[values.HeroBuffId]struct{})
	for _, item := range hero.Buff {
		for _, buffId := range item.Buff {
			if _, ok := buffMap[buffId]; !ok {
				buff = append(buff, buffId)
				buffMap[buffId] = struct{}{}
			}
		}
	}
	return &models.Hero{
		Id:           hero.BuildId,
		Attrs:        attrs,
		Skill:        svc.getSkills(ctx, hero),
		EquipSlot:    hero.EquipSlot,
		CombatValue:  hero.CombatValue,
		Buff:         buff,
		TalentBuff:   hero.TalentBuff,
		OriginId:     hero.Id,
		SoulContract: hero.SoulContract,
		Resonance:    hero.Resonance,
		Fashion:      hero.Fashion,
	}
}

func (svc *Service) oneEquipAttr(ctx *ctx.Context, hero *pbdao.Hero, slot *models.HeroEquipSlot, slotId values.Integer, equips map[values.EquipId]*models.Equipment) (map[values.AttrId]values.Integer, map[values.AttrId]values.Integer, map[values.HeroSkillId]struct{}) {
	id := slot.EquipItemId
	fixedMap := make(map[values.AttrId]values.Integer)
	equipCfg, ok := rule.GetEquipByItemId(ctx, id)
	if !ok {
		return nil, nil, nil
	}
	if len(equipCfg.AttrId) > 0 {
		star := slot.Star
		starToValueMap := make(map[values.AttrId]*equipStarToValue)
		for i, attrId := range equipCfg.AttrId {
			fixedMap[attrId] = equipCfg.AttrValue[i]
			starRange := make([]values.Integer, 0)
			value := make([]values.Integer, 0)
			if star > 0 {
				for j := 0; j < len(equipCfg.AttrStarRange[i]); j++ {
					if j == 0 && star <= equipCfg.AttrStarRange[i][j] {
						starRange = append(starRange, equipCfg.AttrStarRange[i][j])
						value = append(value, equipCfg.Attr[i][j])
						break
					}
					if star >= equipCfg.AttrStarRange[i][j] {
						starRange = append(starRange, equipCfg.AttrStarRange[i][j])
						value = append(value, equipCfg.Attr[i][j])
					}
					if star < equipCfg.AttrStarRange[i][j] {
						starRange = append(starRange, equipCfg.AttrStarRange[i][j])
						value = append(value, equipCfg.Attr[i][j])
						break
					}
				}
				starToValueMap[attrId] = &equipStarToValue{
					Range: starRange,
					Value: value,
				}
			}
		}
		for attrId, item := range starToValueMap {
			for i := 0; i < len(item.Range); i++ {
				v := item.Range[i]
				if i == 0 {
					if star >= v {
						fixedMap[attrId] += v * item.Value[i]
					} else if i == 0 {
						fixedMap[attrId] += star * item.Value[i]
					}
				} else if star > item.Range[i] {
					fixedMap[attrId] += (item.Range[i] - item.Range[i-1]) * item.Value[i]
				} else {
					fixedMap[attrId] += (star - item.Range[i-1]) * item.Value[i]
				}
			}
			if star > item.Range[len(item.Range)-1] {
				fixedMap[attrId] += (star - item.Range[len(item.Range)-1]) * item.Value[len(item.Range)-1]
			}
		}
	}
	// 精炼提升主要属性百分比
	if slot.MeltLevel > 0 {
		cfg, ok := rule.GetEquipRefine(ctx, slotId, slot.MeltLevel-1) // 因配置问题，这里属性取上一级的
		if !ok {
			ctx.Error("equip refine config not found", zap.Int64("slot_id", slotId), zap.Int64("refine_level", slot.MeltLevel))
		} else {
			for attrId, val := range fixedMap {
				fixedMap[attrId] += values.Integer(math.Ceil(values.Float(val) * (values.Float(cfg.AttrAdd) / 10000)))
			}
			// 精炼带来的属性值
			for id, val := range cfg.AttrId {
				fixedMap[id] += val
			}
		}
	}

	fixed := make(map[values.AttrId]values.Integer)
	percent := make(map[values.AttrId]values.Integer)
	skillMap := make(map[values.HeroSkillId]struct{})
	if equip, ok := equips[slot.EquipId]; ok {
		fixed, percent, skillMap = svc.oneEquipAffixAttr(ctx, equip, hero, slotId)
	}
	// 附魔属性
	svc.onAffixAttr(ctx, slot.Enchant, fixed, percent, skillMap, hero, slotId)

	for attrId, val := range fixed {
		fixedMap[attrId] += val
	}
	// 装备本身目前没有百分比属性
	return fixedMap, percent, skillMap
}

func (svc *Service) oneEquipAffixAttr(ctx *ctx.Context, equip *models.Equipment, hero *pbdao.Hero, slot values.Integer) (map[values.AttrId]values.Integer, map[values.AttrId]values.Integer, map[values.HeroSkillId]struct{}) {
	fixedMap := make(map[values.AttrId]values.Integer)
	percentMap := make(map[values.AttrId]values.Integer)
	skillMap := make(map[values.HeroSkillId]struct{})
	for _, affix := range equip.Detail.Affix {
		if !affix.Active {
			continue
		}
		svc.onAffixAttr(ctx, affix, fixedMap, percentMap, skillMap, hero, slot)
	}
	return fixedMap, percentMap, skillMap
}

func (svc *Service) onAffixAttr(
	ctx *ctx.Context,
	affix *models.Affix,
	fixedMap, percentMap map[values.AttrId]values.Integer,
	skillMap map[values.HeroSkillId]struct{},
	hero *pbdao.Hero,
	slotId values.Integer,
) {
	if affix == nil {
		return
	}
	if affix.AttrId > 0 { // 属性
		if affix.IsPercent {
			percentMap[affix.AttrId] += affix.AffixValue
			for _, val := range affix.Bonus {
				percentMap[affix.AttrId] += val
			}
		} else {
			fixedMap[affix.AttrId] += affix.AffixValue
			for _, val := range affix.Bonus {
				fixedMap[affix.AttrId] += val
			}
		}
	} else if affix.BuffId > 0 { // buff
		skillMap[affix.BuffId] = struct{}{}
	} else { // 对技能等级加成
		svc.affixUpgradeSkillLevel(ctx, affix, hero, slotId)
	}
}

func (svc *Service) affixUpgradeSkillLevel(ctx *ctx.Context, affix *models.Affix, hero *pbdao.Hero, slotId values.Integer) {
	cfg, ok := rule.GetEquipEntryById(ctx, affix.AffixId)
	if !ok {
		svc.log.Warn("EquipEntry config not found", zap.Int64("id", affix.AffixId))
		return
	}
	// if len(cfg.HeroId) > 0 { // 对单个英雄或某个流派
	// 	// 所有英雄所有流派都的所有技能生效
	// 	if len(cfg.HeroId) == 1 && cfg.HeroId[0] == 1 {
	// 		svc.affixUpgradeOneSkill(affix, hero, slotId, cfg)
	// 	} else {
	// 		for _, id := range cfg.HeroId {
	// 			if id == hero.Id {
	// 				svc.affixUpgradeOneSkill(affix, hero, slotId, cfg)
	// 				break
	// 			}
	// 		}
	// 	}
	// } else if
	if len(cfg.SkillId) <= 0 {
		return
	} // 对单个或多个技能
	for _, id := range cfg.SkillId {
		for skillId, level := range hero.Skill {
			if id == skillId {
				if level == nil {
					level = &pbdao.SkillLevel{}
				}
				if level.Equip == nil {
					level.Equip = map[int64]int64{}
				}
				level.Equip[slotId] = cfg.SkillLv[affix.Quality]
				hero.Skill[skillId] = level
				break
			}
		}
	}
}

func (svc *Service) affixUpgradeOneSkill(affix *models.Affix, hero *pbdao.Hero, slotId values.Integer, cfg *rulemodel.EquipEntry) {
	for skillId, level := range hero.Skill {
		if level == nil {
			level = &pbdao.SkillLevel{}
		}
		if level.Equip == nil {
			level.Equip = map[int64]int64{}
		}
		level.Equip[slotId] = cfg.SkillLv[affix.Quality]
		hero.Skill[skillId] = level
	}
}

// 属性转换
func (svc *Service) transformAttr(
	ctx *ctx.Context,
	hero *pbdao.Hero,
	base map[values.AttrId]values.Integer,
	detailFixed map[values.AttrId]values.Integer,
	detailPercent map[values.AttrId]values.Integer,
	detailOnHero map[values.AttrId]values.Integer,
) {
	// for ac, item := range hero.Attrs {
	//	list := make([]map[values.AttrId]values.Integer, 0)
	//	for attrId, val := range item.Attr {
	//		list = append(list, svc.transform1Attr(ctx, hero.Id, attrId, val))
	//	}
	//	attrs := make(map[values.Integer]values.Integer)
	//	for _, attrMap := range list {
	//		for id, v := range attrMap {
	//			attrs[id] += v
	//		}
	//	}
	//	hero.Attrs[ac] = &models.HeroAttrItem{
	//		Attr: attrs,
	//	}
	// }
	detail := make(map[values.AttrId]values.Integer)
	for id, val := range base {
		list := rule.GetAttrTransConfigById(ctx, id)
		for _, item := range list {
			if !svc.attrTransformVocationCheck(hero.Id, item.Limithero) {
				continue
			}
			fmt.Println(val)
			// if item.Transtype == ttPercentage {
			// 	// v = values.Integer(math.Ceil(values.Float(val) * values.Float(item.Transnum) / 10000))
			// 	if v, ok := detailPercent[item.TransattrId]; ok {
			// 		detailPercent[item.TransattrId] = v + val*item.Transnum
			// 	} else {
			// 		detailPercent[item.TransattrId] = val * item.Transnum
			// 	}
			// } else {
			// 	detail[item.TransattrId] = val * item.Transnum
			// }
		}
	}
	for id, val := range detailFixed {
		detail[id] += val
	}
	for id, val := range detailOnHero {
		detail[id] += val
	}
	for id, val := range detailPercent {
		v, ok := detail[id]
		if ok {
			detail[id] += values.Integer(math.Ceil(values.Float(v) * values.Float(val) / 10000.0))
		}
	}

	// hero.Attrs[acBase] = &models.HeroAttr{Fixed: base}
	// hero.Attrs[acDetail] = &models.HeroAttr{Fixed: detail}
}

// func (svc *Service) transform1Attr(ctx *ctx.Context, heroId values.HeroId, atrId values.AttrId, attrValue values.Integer) map[values.AttrId]values.Integer {
//	list := rule.GetAttrTransConfigById(ctx, atrId)
//	dataMap := map[values.AttrId]values.Integer{atrId: attrValue}
//	for _, item := range list {
//		if !svc.attrTransformVocationCheck(heroId, item.Limithero()) {
//			continue
//		}
//		var v values.Integer
//		if item.Transtype == ttPercentage {
//			v = values.Integer(math.Ceil(values.Float(attrValue) * values.Float(item.Transnum) / 10000))
//		} else {
//			v = attrValue * item.Transnum
//		}
//		dataMap[item.TransattrId] = v
//	}
//	return dataMap
// }

func (svc *Service) attrTransformVocationCheck(heroId values.HeroId, list []values.Integer) bool {
	// 没填或者填0就是对所有职业都生效
	if len(list) == 0 || (len(list) == 1 && list[0] == 0) {
		return true
	}
	for _, v := range list {
		if heroId == v {
			return true
		}
	}
	return false
}

// 已改为提高主要属性的百分比（直接在重新计算英雄属性的时候计算该值 oneEquipAttr 函数）
func (svc *Service) refreshEquipAffixBonus(equip *models.Equipment, cfg *rulemodel.EquipRefine) {
	for i := 0; i < len(equip.Detail.Affix); i++ {
		if equip.Detail.Affix[i].BuffId > 0 {
			continue
		}
		var bonus values.Integer
		if equip.Detail.Affix[i].IsPercent {
			// bonus = cfg.AttrAdd
			// 百分比也按照固定值属性加成计算
			bonus = values.Integer(math.Ceil(values.Float(equip.Detail.Affix[i].AffixValue) * (values.Float(cfg.AttrAdd) / 10000)))
		} else {
			bonus = values.Integer(math.Ceil(values.Float(equip.Detail.Affix[i].AffixValue) * (values.Float(cfg.AttrAdd) / 10000)))
		}
		if equip.Detail.Affix[i].Bonus == nil {
			equip.Detail.Affix[i].Bonus = map[int64]int64{}
		}
		equip.Detail.Affix[i].Bonus[values.Integer(models.AffixBonusType_MeltLevel)] = bonus
	}
}

func (svc *Service) clearEquipAffixBonus(equip *models.Equipment) {
	for i := 0; i < len(equip.Detail.Affix); i++ {
		if equip.Detail.Affix[i].BuffId > 0 {
			continue
		}
		if equip.Detail.Affix[i].Bonus == nil {
			equip.Detail.Affix[i].Bonus = map[int64]int64{}
		}
		equip.Detail.Affix[i].Bonus[values.Integer(models.AffixBonusType_MeltLevel)] = 0
	}
}

// TODO 后续优化：是否考虑将 roleAttrFixed，roleAttrPercent，formRoleSkills 持久化到每个英雄身上
// func (svc *Service) updateHeroAttrs(
// 	ctx *ctx.Context,
// 	hero *pbdao.Hero,
// 	level values.Level,
// 	calcLevel, calcEquip bool,
// 	roleAttrFixed, roleAttrPercent []*models.AttrBonus,
// ) {
// 	if calcLevel {
// 		svc.refreshAttrFromBase(ctx, hero, level)
// 	}
// 	if calcEquip {
// 		svc.refreshAttrFromEquip(ctx, hero)
// 	}
// 	baseFixed := make(map[values.AttrId]values.Integer)
// 	detailFixed := make(map[values.AttrId]values.Integer)
// 	for _, attr := range roleAttrFixed {
// 		if attr.AdvancedType == enum.BaseAttrType {
// 			for id, val := range attr.Attr {
// 				baseFixed[id] += val
// 			}
// 		} else {
// 			for id, val := range attr.Attr {
// 				detailFixed[id] += val
// 			}
// 		}
// 	}
// 	basePercent := make(map[values.AttrId]values.Integer)
// 	detailPercent := make(map[values.AttrId]values.Integer)
// 	for _, attr := range roleAttrPercent {
// 		if attr.AdvancedType == enum.BaseAttrType {
// 			for id, val := range attr.Attr {
// 				basePercent[id] += val
// 			}
// 		} else {
// 			for id, val := range attr.Attr {
// 				detailPercent[id] += val
// 			}
// 		}
// 	}
// 	detailOnHero := make(map[values.AttrId]values.Integer)
// 	// 先将基础（一级）属性算上来自角色的加成，再做属性转换
// 	for ac, item := range hero.Attrs {
// 		// if !isAttrConstitutePart(ac) {
// 		// 	continue
// 		// }
// 		fmt.Println(ac)
// 		// 固定值
// 		for id, val := range item.Fixed {
// 			cfg, ok := rule.GetAttrById(ctx, id)
// 			if !ok {
// 				svc.log.Warn("attr config not found", zap.Int64("id", id))
// 				continue
// 			}
// 			if cfg.AdvancedType == enum.BaseAttrType {
// 				baseFixed[id] += val
// 			} else {
// 				detailOnHero[id] += val
// 			}
// 		}
// 		// 百分比
// 		for id, val := range item.Percent {
// 			cfg, ok := rule.GetAttrById(ctx, id)
// 			if !ok {
// 				svc.log.Warn("attr config not found", zap.Int64("id", id))
// 				continue
// 			}
// 			if cfg.AdvancedType == enum.BaseAttrType {
// 				basePercent[id] += val
// 			} else {
// 				detailPercent[id] += val
// 			}
// 		}
// 	}
// 	for id, val := range basePercent {
// 		v, ok := baseFixed[id]
// 		if ok {
// 			baseFixed[id] += values.Integer(math.Ceil(values.Float(v) * values.Float(val) / 10000.0))
// 		}
// 	}
// 	svc.transformAttr(ctx, hero, baseFixed, detailFixed, detailPercent, detailOnHero)
// 	svc.updateHeroCombatValue(ctx, hero, level)
// }

// func (svc *Service) updateHeroCombatValue(ctx *ctx.Context, hero *pbdao.Hero, level values.Level) {
// 	// TODO 目前直接算一个总的战力即可，细节策划还未给出
// 	// 最终战力 =等级系数（万分比）* (属性1*系数1+属性2*系数2+++++++属性N*系数N)+等级技能系数（万分比）*(技能1系数+技能2系数+技能3系数+++++技能N系数)
// 	lvCfg, ok := rule.GetRoleLvConfigByLv(ctx, level)
// 	if !ok {
// 		svc.log.Warn("role_lv config not found", zap.Int64("level", level))
// 		return
// 	}
// 	var attrPart values.Integer
// 	// for ac, item := range hero.Attrs {
// 	// 	if isFinalAttrConstitute(ac) {
// 	// 		for id, val := range item.Fixed {
// 	// 			attrCfg, ok := rule.GetAttrById(ctx, id)
// 	// 			if !ok {
// 	// 				svc.log.Warn("attr config not found", zap.Int64("id", id))
// 	// 				continue
// 	// 			}
// 	// 			attrPart += val * attrCfg.PowerNum
// 	// 		}
// 	// 	}
// 	// }
// 	skillPart := values.Float(0)
// 	for _, item := range hero.Skill {
// 		for _, skillId := range item.Skill {
// 			skillCfg, ok := rule.GetSkillById(ctx, skillId)
// 			if !ok {
// 				svc.log.Warn("skill config not found", zap.Int64("id", skillCfg.Id))
// 				continue
// 			}
// 			skillPart += values.Float(skillCfg.PowerNum)
// 		}
// 	}
// 	skillPart = values.Float(lvCfg.SkillPowerNum) / 10000 * skillPart
// 	skillPartInteger := values.Integer(math.Ceil(skillPart))
// 	total := values.Integer(math.Ceil(values.Float(lvCfg.PowerNum*attrPart)/10000)) + skillPartInteger
// 	svc.log.Debug("update hero combat value",
// 		zap.String("role_id", ctx.RoleId),
// 		zap.Int64("hero_id", hero.Id),
// 		zap.Int64("attr_part", attrPart),
// 		zap.Int64("skill_part", skillPartInteger),
// 		zap.Int64("combat_value", total),
// 		zap.Int64("power_num", lvCfg.PowerNum),
// 		zap.Int64("skill_power_num", lvCfg.SkillPowerNum))
// 	if hero.CombatValue == nil {
// 		hero.CombatValue = &models.CombatValue{}
// 	}
// 	hero.CombatValue.Total = total
// }

// 获取英雄身上已装备的所有装备id
func (svc *Service) getHeroEquippedEquipId(hero *pbdao.Hero) []values.EquipId {
	var equipIds []values.EquipId
	for _, item := range hero.EquipSlot {
		if item.EquipId != "" {
			equipIds = append(equipIds, item.EquipId)
		}
	}
	return equipIds
}

func (svc *Service) GetHeroesEquippedEquipId(heroes []*models.Hero) []values.EquipId {
	var equipIds []values.EquipId
	for _, hero := range heroes {
		for _, slot := range hero.EquipSlot {
			if slot.EquipId != "" {
				equipIds = append(equipIds, slot.EquipId)
			}
		}
	}
	return equipIds
}

func (svc *Service) getAllHeroEquippedEquipId(heroes []*pbdao.Hero) []values.EquipId {
	list := make([]values.EquipId, 0)
	for _, hero := range heroes {
		list = append(list, svc.getHeroEquippedEquipId(hero)...)
	}
	return list
}

func (svc *Service) getEquipSetAttr(ctx *ctx.Context, equipItemId []values.ItemId) (map[values.AttrId]values.Integer, map[values.AttrId]values.Integer, []values.HeroBuffId) {
	cfg := rule2.MustGetReader(ctx).EquipComplete.List()
	fixedMap := make(map[values.AttrId]values.Integer)
	percentMap := make(map[values.AttrId]values.Integer)
	buffMap := make(map[values.HeroSkillId]struct{})
	for _, complete := range cfg {
		svc.getOnSetAttr(ctx, equipItemId, &complete, fixedMap, percentMap, buffMap)
	}
	buff := make([]values.HeroBuffId, 0, len(buffMap))
	for id := range buffMap {
		buff = append(buff, id)
	}
	return fixedMap, percentMap, buff
}

func (svc *Service) getOnSetAttr(
	ctx *ctx.Context, equipItemId []values.ItemId, cfg *rulemodel.EquipComplete,
	fixedMap, percentMap map[values.AttrId]values.Integer, buffMap map[values.HeroBuffId]struct{},
) {
	count := 0
	for _, id := range cfg.Complete {
		for _, itemId := range equipItemId {
			if itemId == id {
				count++
			}
		}
	}
	// 2件套
	if count >= 2 {
		svc.getOnSetItemResult(ctx, cfg.SkillTwo, fixedMap, percentMap, buffMap)
	}
	// 4件套
	if count >= 4 {
		svc.getOnSetItemResult(ctx, cfg.SkillFour, fixedMap, percentMap, buffMap)
	}
	// 6件套
	if count >= 6 {
		svc.getOnSetItemResult(ctx, cfg.SkillSix, fixedMap, percentMap, buffMap)
	}
}

func (svc *Service) getOnSetItemResult(ctx *ctx.Context, item [][]values.Integer, fixedMap, percentMap map[values.AttrId]values.Integer, buffMap map[values.HeroBuffId]struct{}) {
	attrOrSkill := item[0]
	fixedValue := item[1]
	percentValue := item[2]
	for i := 0; i < len(attrOrSkill); i++ {
		id := attrOrSkill[i]
		// 在属性表里能找到就认为是属性，否则认为是技能
		// 固定值或百分比只对属性有效
		_, ok := rule.GetAttrById(ctx, id)
		if ok {
			if fixedValue[i] > 0 {
				fixedMap[id] += fixedValue[i]
			}
			if percentValue[i] > 0 {
				percentMap[id] += percentValue[i]
			}
		} else {
			buffMap[id] = struct{}{}
		}
	}
}

func (svc *Service) GetInitSkills(cfg *rulemodel.RowHero) []values.HeroSkillId {
	skills := make([]values.HeroSkillId, 0)
	for _, skill := range cfg.AtKSkill {
		skills = append(skills, skill)
	}
	// for _, skill := range cfg.SkillId {
	// 	skills = append(skills, skill)
	// }
	return skills
}

func (svc *Service) GetAllHeroId(ctx *ctx.Context) ([]values.HeroId, *errmsg.ErrMsg) {
	heroes, err := dao.NewHero(ctx.RoleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]values.HeroId, 0, len(heroes))
	for _, hero := range heroes {
		list = append(list, hero.Id)
	}
	return list, nil
}

func (svc *Service) GetRoleAttr(ctx *ctx.Context) ([]*models.AttrBonus, []*models.AttrBonus, *errmsg.ErrMsg) {
	attrs, err := svc.UserService.GetRoleAttr(ctx, ctx.RoleId)
	if err != nil {
		return nil, nil, err
	}
	var attrFixed, attrPercent = make([]*models.AttrBonus, 0), make([]*models.AttrBonus, 0)
	for _, attr := range attrs {
		if len(attr.AttrFixed) > 0 {
			attrFixed = append(attrFixed, attr.AttrFixed...)
		}
		if len(attr.AttrPercent) > 0 {
			attrPercent = append(attrPercent, attr.AttrPercent...)
		}
	}
	return attrFixed, attrPercent, nil
}

func (svc *Service) getHeroRealId(ctx *ctx.Context, id values.Integer) (values.HeroId, *errmsg.ErrMsg) {
	hero, ok := rule.GetHero(ctx, id)
	if !ok {
		return 0, errmsg.NewErrHeroNotFound()
	}
	return hero.OriginId, nil
}

func (svc *Service) getSkills(ctx *ctx.Context, hero *pbdao.Hero) []*models.HeroSkillAndStone {
	cfg, ok := rule.GetHero(ctx, hero.BuildId)
	if ok {
		list := svc.GetInitSkills(cfg)
		for _, id := range list {
			_, ok = hero.Skill[id]
			if !ok {
				hero.Skill[id] = &pbdao.SkillLevel{}
			}
		}
	}
	skills := make([]*models.HeroSkillAndStone, 0, len(hero.Skill))
	for id, level := range hero.Skill {
		max := rule.GetMaxSkillId(ctx, id)
		skill := id
		if level != nil {
			skill += level.Talent
			for _, lv := range level.Equip {
				skill += lv
			}
			if skill > max {
				skill = max
			}
		}
		_, ok := rule.GetSkillById(ctx, skill)
		if !ok {
			svc.log.Warn("skill config not found", zap.Int64("id", skill))
			continue
		}
		var stones []values.Integer
		if level != nil {
			stones = level.Stones
		}
		skills = append(skills, &models.HeroSkillAndStone{
			SkillId: skill,
			Stones:  stones,
		})
	}
	return skills
}

func (svc *Service) isRefineUnlock(ctx *ctx.Context, hero *pbdao.Hero, slotId values.Integer) bool {
	if hero.EquipSlot == nil {
		return false
	}
	slot, ok := hero.EquipSlot[slotId]
	if !ok {
		return false
	}
	return slot.Star >= rule.GetEquipStarLimit(ctx)
}

func (svc *Service) formatEquipBonusTalentData(current, last map[values.HeroSkillId]values.Level) map[values.HeroSkillId]values.Level {
	ret := make(map[values.HeroSkillId]values.Level)
	for talentId, level := range current {
		lastLevel, ok := last[talentId]
		if ok {
			delete(last, talentId)
		}
		ret[talentId] = level - lastLevel
	}
	for talentId, level := range last {
		ret[talentId] -= level
	}
	return ret
}

func (svc *Service) mergeEquipBonusTalentData(data, target map[values.HeroSkillId]values.Level) map[values.HeroSkillId]values.Level {
	if len(data) <= 0 {
		return target
	}
	if len(target) <= 0 {
		return data
	}
	for id := range data {
		if lv, ok := target[id]; ok {
			data[id] += lv
			delete(target, id)
		}
	}
	for id, level := range target {
		data[id] += level
	}
	return data
}

func (svc *Service) GetRoleAttrAndSkill(ctx *ctx.Context) ([]*models.RoleAttr, []values.Integer, *errmsg.ErrMsg) {
	roleAttr, err := svc.UserService.GetRoleAttr(ctx, ctx.RoleId)
	if err != nil {
		return nil, nil, err
	}
	formRoleSkills, err := svc.GetRoleSkill(ctx, ctx.RoleId)
	if err != nil {
		return nil, nil, err
	}
	return roleAttr, formRoleSkills, nil
}

func (svc *Service) CheatAddHero(ctx *ctx.Context, req *servicepb.Hero_CheatAddHeroRequest) (*servicepb.Hero_CheatAddHeroResponse, *errmsg.ErrMsg) {
	hero, err := svc.AddHero(ctx, req.Id, false)
	if err != nil {
		return nil, err
	}
	return &servicepb.Hero_CheatAddHeroResponse{
		Hero: hero,
	}, nil
}

func (svc *Service) CheatSetSlotMax(ctx *ctx.Context, req *servicepb.Hero_CheatSetSlotMaxRequest) (*servicepb.Hero_CheatSetSlotMaxResponse, *errmsg.ErrMsg) {
	heroDao := dao.NewHero(ctx.RoleId)
	heroes, err := heroDao.GetSome(ctx, req.HeroId)
	if err != nil {
		return nil, err
	}
	if len(heroes) <= 0 {
		return nil, nil
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getAllHeroEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	maxStar := rule.GetMaxEquipStar(ctx)
	slots := make(map[int64]*models.HeroEquipSlot)
	for j := 1; j < 8; j++ {
		slots[values.Integer(j)] = &models.HeroEquipSlot{
			Star:      maxStar,
			MeltLevel: rule.GetMaxEquipMeltLevel(ctx, values.EquipSlot(j)),
		}
	}
	for i := 0; i < len(heroes); i++ {
		if heroes[i].EquipSlot == nil {
			heroes[i].EquipSlot = map[int64]*models.HeroEquipSlot{}
		}
		for j := 1; j < 8; j++ {
			if heroes[i].EquipSlot[values.Integer(j)] == nil {
				heroes[i].EquipSlot[values.Integer(j)] = &models.HeroEquipSlot{}
			}
			heroes[i].EquipSlot[values.Integer(j)].Star = maxStar
			heroes[i].EquipSlot[values.Integer(j)].MeltLevel = rule.GetMaxEquipMeltLevel(ctx, values.EquipSlot(j))
		}
		svc.refreshEquipResonanceStatus(ctx, heroes[i])
	}
	svc.refreshAllHeroAttr(ctx, heroes, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill)
	if err := heroDao.Save(ctx, heroes...); err != nil {
		return nil, err
	}
	ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: func() []*models.Hero {
			list := make([]*models.Hero, 0)
			for _, hero := range heroes {
				list = append(list, svc.dao2model(ctx, hero))
			}
			return list
		}(),
	})

	return &servicepb.Hero_CheatSetSlotMaxResponse{}, nil
}

func (svc *Service) CheatSetSoulContractMax(ctx *ctx.Context, req *servicepb.Hero_CheatSetSoulContractMaxRequest) (*servicepb.Hero_CheatSetSoulContractMaxResponse, *errmsg.ErrMsg) {
	heroDao := dao.NewHero(ctx.RoleId)
	heroes, err := heroDao.GetSome(ctx, req.HeroId)
	if err != nil {
		return nil, err
	}
	if len(heroes) <= 0 {
		return nil, nil
	}
	for i := 0; i < len(heroes); i++ {
		max, ok := rule.GetMaxSoulContract(ctx, heroes[i].Id)
		if !ok {
			return nil, errmsg.NewInternalErr("MaxSoulContract not exist")
		}
		heroes[i].SoulContract = &models.SoulContract{
			Rank:  max.Rank,
			Level: max.Level,
		}
	}
	if err := heroDao.Save(ctx, heroes...); err != nil {
		return nil, err
	}
	ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: func() []*models.Hero {
			list := make([]*models.Hero, 0)
			for _, hero := range heroes {
				list = append(list, svc.dao2model(ctx, hero))
			}
			return list
		}(),
	})
	return &servicepb.Hero_CheatSetSoulContractMaxResponse{}, nil
}
