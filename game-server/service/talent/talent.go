package talent

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/proto/models"
	protosvc "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/talent/dao"
	"coin-server/rule"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
}

func NewTalentService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
	}
	s.TalentService = s
	return s
}

// ---------------------------------------------------proto------------------------------------------------------------//

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取天赋列表", svc.Gets)
	svc.svc.RegisterFunc("天赋符文升级", svc.RuneLevelUp)
	svc.svc.RegisterFunc("天赋符文升级粉尘", svc.RuneLevelUpByDust)
	svc.svc.RegisterFunc("天赋盘解锁", svc.PlateUnlock)
	svc.svc.RegisterFunc("符文镶嵌", svc.RuneInlay)
	svc.svc.RegisterFunc("符文移动", svc.RuneMove)
	svc.svc.RegisterFunc("技能升级", svc.SkillLevelUp)
	svc.svc.RegisterFunc("技能选择", svc.SkillChoose)
	svc.svc.RegisterFunc("重置天赋", svc.Reset)

	svc.svc.RegisterFunc("插入宝石", svc.InlayStone)
	svc.svc.RegisterFunc("移除宝石", svc.RemoveStone)
	svc.svc.RegisterFunc("移除所有宝石", svc.RemoveAllStone)
	svc.svc.RegisterFunc("合成宝石", svc.ComposeStone)
	svc.svc.RegisterFunc("锁定宝石", svc.LockStone)

	eventlocal.SubscribeEventLocal(svc.HandleAddCommonPoints)
	eventlocal.SubscribeEventLocal(svc.HandleAddParticularPoints)
	eventlocal.SubscribeEventLocal(svc.HandleNewHero)
	eventlocal.SubscribeEventLocal(svc.HandleEquipBonusTalent)
	eventlocal.SubscribeEventLocal(svc.HandleLevelUp)

	svc.svc.RegisterFunc("作弊加通用天赋点", svc.CheatAddCommonPoints)
	svc.svc.RegisterFunc("作弊加职业天赋点", svc.CheatAddParticularPoints)
}

func (svc *Service) Gets(c *ctx.Context, _ *protosvc.Talent_GetsRequest) (*protosvc.Talent_GetsResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	return &protosvc.Talent_GetsResponse{
		CommonPoints: t.CommonPoints(),
		Talents:      t.Gets(),
		LockStoneIds: t.LockList(),
	}, nil
}

func (svc *Service) RuneLevelUp(c *ctx.Context, req *protosvc.Talent_RuneLevelUpRequest) (*protosvc.Talent_RuneLevelUpResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	r, err := svc.GetRuneById(c, c.RoleId, req.RuneId)
	if err != nil {
		return nil, err
	}
	oldLvl := r.Lvl
	obtainMap, err := svc.GetManyRunes(c, c.RoleId, req.UseRuneIds)
	if err != nil {
		return nil, err
	}
	obtain := make([]*models.TalentRune, 0, len(obtainMap))
	for _, ob := range obtainMap {
		obtain = append(obtain, ob)
	}
	configId, delIds, err := t.RuneLevelUp(c, r, obtain)
	if err != nil {
		return nil, err
	}
	evt, err := t.CalAttr(c, configId)
	if err != nil {
		return nil, err
	}
	if evt != nil {
		c.PublishEventLocal(&event.TalentChange{
			ConfigId: configId,
			Attr:     evt,
		})
	}
	delEvt, err := svc.DelManyRunes(c, c.RoleId, delIds)
	if err != nil {
		return nil, err
	}
	if delEvt != nil {
		c.PublishEventLocal(delEvt)
	}
	saveEvt, err := svc.SaveRune(c, c.RoleId, r)
	if err != nil {
		return nil, err
	}
	if saveEvt != nil {
		c.PublishEventLocal(saveEvt)
	}
	svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskRuneLvlUp, 0, r.Lvl-oldLvl)
	svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskRuneLvlUpAcc, 0, r.Lvl-oldLvl)
	dao.SaveTalent(c, t)
	return &protosvc.Talent_RuneLevelUpResponse{}, nil
}

func (svc *Service) RuneLevelUpByDust(c *ctx.Context, req *protosvc.Talent_RuneLevelUpUseDustRequest) (*protosvc.Talent_RuneLevelUpUseDustResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	r, err := svc.GetRuneById(c, c.RoleId, req.RuneId)
	if err != nil {
		return nil, err
	}
	oldLvl := r.Lvl
	configId, costCnt, err := t.RuneLevelUpByDust(c, r, req.DustCnt)
	if err != nil {
		return nil, err
	}
	if err = svc.SubManyItem(c, c.RoleId, map[values.ItemId]values.Integer{enum.RuneDust: costCnt}); err != nil {
		return nil, err
	}
	evt, err := t.CalAttr(c, configId)
	if err != nil {
		return nil, err
	}
	if evt != nil {
		c.PublishEventLocal(&event.TalentChange{
			ConfigId: configId,
			Attr:     evt,
		})
	}
	saveEvt, err := svc.SaveRune(c, c.RoleId, r)
	if err != nil {
		return nil, err
	}
	if saveEvt != nil {
		c.PublishEventLocal(saveEvt)
	}
	svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskRuneLvlUp, 0, r.Lvl-oldLvl)
	svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskRuneLvlUpAcc, 0, r.Lvl-oldLvl)
	dao.SaveTalent(c, t)
	return &protosvc.Talent_RuneLevelUpUseDustResponse{}, nil
}

func (svc *Service) PlateUnlock(c *ctx.Context, req *protosvc.Talent_PlateUnlockRequest) (*protosvc.Talent_PlateUnlockResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	heroP, err := t.UnlockPlate(c, req.ConfigId, req.PlateIdx, req.Loc)
	if err != nil {
		return nil, err
	}
	if heroP != nil {
		c.PushMessage(&protosvc.Talent_TalentSkillChangePush{
			ConfigId: req.ConfigId,
			Data:     heroP,
		})
	}
	dao.SaveTalent(c, t)
	return &protosvc.Talent_PlateUnlockResponse{}, nil
}

func (svc *Service) RuneInlay(c *ctx.Context, req *protosvc.Talent_RuneInlayRequest) (*protosvc.Talent_RuneInlayResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	r, err := svc.GetRuneById(c, c.RoleId, req.RuneId)
	if err != nil {
		return nil, err
	}
	tp, err := t.InlayRune(c, req.ConfigId, req.PlateIdx, req.Loc, r, req.IsInlay)
	if err != nil {
		return nil, err
	}
	saveEvt, err := svc.SaveRune(c, c.RoleId, r)
	if err != nil {
		return nil, err
	}
	if req.IsInlay {
		svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskInlayRune, 0, 1)
	} else {
		svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskInlayRune, 0, -1)
	}
	c.PublishEventLocal(saveEvt)
	evt, err := t.CalAttr(c, req.ConfigId)
	if err != nil {
		return nil, err
	}
	if evt != nil {
		c.PublishEventLocal(&event.TalentChange{
			ConfigId: req.ConfigId,
			Attr:     evt,
		})
	}
	c.PushMessage(&protosvc.Talent_PlateChangePush{
		ConfigId: req.ConfigId,
		PlateIdx: req.PlateIdx,
		Plate:    tp,
	})
	dao.SaveTalent(c, t)
	return &protosvc.Talent_RuneInlayResponse{}, nil
}

func (svc *Service) RuneMove(c *ctx.Context, req *protosvc.Talent_RuneMoveRequest) (*protosvc.Talent_RuneMoveResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	r, err := svc.GetRuneById(c, c.RoleId, req.RuneId)
	if err != nil {
		return nil, err
	}
	_, err = t.InlayRune(c, req.ConfigId, req.PlateIdx, req.OldLoc, r, false)
	if err != nil {
		return nil, err
	}
	tp, err := t.InlayRune(c, req.ConfigId, req.PlateIdx, req.NewLoc, r, true)
	if err != nil {
		return nil, err
	}
	saveEvt, err := svc.SaveRune(c, c.RoleId, r)
	if err != nil {
		return nil, err
	}
	c.PublishEventLocal(saveEvt)
	evt, err := t.CalAttr(c, req.ConfigId)
	if err != nil {
		return nil, err
	}
	if evt != nil {
		c.PublishEventLocal(&event.TalentChange{
			ConfigId: req.ConfigId,
			Attr:     evt,
		})
	}
	c.PushMessage(&protosvc.Talent_PlateChangePush{
		ConfigId: req.ConfigId,
		PlateIdx: req.PlateIdx,
		Plate:    tp,
	})
	dao.SaveTalent(c, t)
	return &protosvc.Talent_RuneMoveResponse{}, nil
}

func (svc *Service) SkillLevelUp(c *ctx.Context, req *protosvc.Talent_SkillLevelUpRequest) (*protosvc.Talent_SkillLevelUpResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	role, err := svc.Module.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	items, skill, err := t.SkillLevelUp(c, req.ConfigId, req.SkillId, role.Level)
	if err != nil {
		return nil, err
	}
	if err = svc.SubManyItem(c, c.RoleId, items); err != nil {
		return nil, err
	}
	skillUpCnt, heroT := t.SkillHandleHeroRoleLvlUp(c, role.Level, req.ConfigId)
	if skillUpCnt > 0 {
		c.PushMessage(&protosvc.Talent_TalentSkillChangePush{
			ConfigId: req.ConfigId,
			Data:     heroT,
		})
		svc.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskHaveLvlUpperSkill, 1, skillUpCnt)
	}
	evt, err := t.CalSkill(c, req.ConfigId)
	if err != nil {
		return nil, err
	}
	if evt != nil {
		c.PublishEventLocal(&event.SkillChange{
			ConfigId: req.ConfigId,
			Skills:   evt,
		})
	}
	dao.SaveTalent(c, t)
	svc.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskLvlUpSkillAcc, 0, 1)
	svc.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskLvlUpSkill, 0, 1)
	svc.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskTotalSkillLvl, 0, 1+skillUpCnt)
	svc.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskHaveLvlUpperSkill, skill.Lvl, 1)
	return &protosvc.Talent_SkillLevelUpResponse{
		ConfigId: req.ConfigId,
		Skill:    skill,
	}, nil
}

func (svc *Service) SkillChoose(c *ctx.Context, req *protosvc.Talent_SkillChooseRequest) (*protosvc.Talent_SkillChooseResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	role, err := svc.Module.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	changeCnt, heroT, err := t.SkillChoose(c, req.ConfigId, req.Idx, req.SkillId, role.Level)
	if err != nil {
		return nil, err
	}
	evt, err := t.CalSkill(c, req.ConfigId)
	if err != nil {
		return nil, err
	}
	if evt != nil {
		c.PublishEventLocal(&event.SkillChange{
			ConfigId: req.ConfigId,
			Skills:   evt,
		})
	}
	svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskChooseSkillAcc, 0, changeCnt)
	dao.SaveTalent(c, t)
	return &protosvc.Talent_SkillChooseResponse{
		Talent: heroT,
	}, nil
}

/*func (svc *Service) handleUpdateTalent(c *ctx.Context, t tval.TalentI, uts []*tval.UpdateTalent) {
	var tat, tas, tuta, tut, tmut int64 = 0, 0, 0, 0, 0
	for idx, updateTalent := range uts {
		if updateTalent != nil {
			if t.IsFirstUpdate(updateTalent.TalentId, updateTalent.Lvl) {
				t.FirstUpdate(updateTalent.TalentId, updateTalent.Lvl)
				if updateTalent.IsManualUp {
					tmut++
				}
				if updateTalent.IsActive {
					tat++
					if updateTalent.IsSkill {
						tas++
					}
				}
			}
			tuta++
			tut++
			tval.PutUpdateTalent(uts[idx])
			uts[idx] = nil
		}
	}
	if tat > 0 {
		svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskActiveTalent, 0, tat)
	}
	if tas > 0 {
		svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskActiveSkillCnt, 0, tas)
	}
	if tuta > 0 {
		svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskUpgradeTalentAcc, 0, tuta)
	}
	if tut > 0 {
		svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskUpgradeTalent, 0, tut)
	}
	if tmut > 0 {
		svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskManualUpTalent, 0, tut)
	}
}*/

func (svc *Service) Reset(c *ctx.Context, req *protosvc.Talent_ResetRequest) (*protosvc.Talent_ResetResponse, *errmsg.ErrMsg) {
	if err := svc.resetCost(c); err != nil {
		return nil, err
	}
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	tp, runeIds, err := t.Reset(c, req.ConfigId, req.PlateIdx)
	if err != nil {
		return nil, err
	}
	runes, err := svc.GetManyRunes(c, c.RoleId, runeIds)
	if err != nil {
		return nil, err
	}
	if len(runes) > 0 {
		rList := make([]*models.TalentRune, 0, len(runes))
		for _, r := range runes {
			r.Loc = 0
			r.PlateIdx = 0
			rList = append(rList, r)
		}
		sEvt, err := svc.SaveManyRunes(c, c.RoleId, rList)
		if err != nil {
			return nil, err
		}
		c.PublishEventLocal(sEvt)
	}
	evt, err := t.CalAttr(c, req.ConfigId)
	if err != nil {
		return nil, err
	}
	if evt != nil {
		c.PublishEventLocal(&event.TalentChange{
			ConfigId: req.ConfigId,
			Attr:     evt,
		})
	}
	c.PushMessage(&protosvc.Talent_PlateChangePush{
		ConfigId: req.ConfigId,
		PlateIdx: req.PlateIdx,
		Plate:    tp,
	})
	dao.SaveTalent(c, t)
	return &protosvc.Talent_ResetResponse{}, nil
}

func (svc *Service) resetCost(c *ctx.Context) *errmsg.ErrMsg {
	role, err := svc.Module.GetRoleModelByRoleId(c, c.RoleId)
	if err != nil {
		return err
	}
	freeLv, ok := rule.MustGetReader(c).KeyValue.GetInt64("SkillTalentResetFireLv")
	if !ok {
		return errmsg.NewInternalErr("kv not found")
	}
	if role.Level < freeLv {
		return nil
	}
	cost, ok := rule.MustGetReader(c).KeyValue.GetItem("SkillTalentResetConsume")
	if !ok {
		return errmsg.NewInternalErr("kv not found")
	}
	return svc.Module.SubItem(c, c.RoleId, cost.ItemId, cost.Count)
}

func (svc *Service) InlayStone(c *ctx.Context, req *protosvc.Talent_InlayStoneRequest) (*protosvc.Talent_InlayStoneResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	stoneRule, has := rule.MustGetReader(c).Skillstone.GetSkillstoneById(req.SkillStoneId)
	if !has {
		return nil, errmsg.NewErrSkillStoneNotExist()
	}
	skillRule, ok := rule.MustGetReader(c).RoleSkill.GetRoleSkillById(req.SkillId)
	if !ok {
		return nil, errmsg.NewErrTalentNotExist()
	}
	replacedId, detail, err := t.InlayStone(c, req.ConfigId, skillRule, stoneRule, int(req.HoleIdx))
	if err != nil {
		return nil, err
	}
	if replacedId != 0 {
		replacedRule, ok := rule.MustGetReader(c).Skillstone.GetSkillstoneById(replacedId)
		if !ok {
			return nil, errmsg.NewErrSkillStoneNotExist()
		}
		if replacedRule.ItemId != stoneRule.ItemId {
			if err = svc.Module.BagService.ExchangeManyItem(c, c.RoleId, map[values.ItemId]values.Integer{replacedRule.ItemId: 1}, map[values.ItemId]values.Integer{stoneRule.ItemId: 1}); err != nil {
				return nil, err
			}
		}
		for lvl := values.Integer(1); lvl <= replacedRule.SkillStoneLv; lvl++ {
			svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskInlayLvlUpperStone, lvl, -1)
		}
	} else {
		if err = svc.Module.BagService.SubItem(c, c.RoleId, stoneRule.ItemId, 1); err != nil {
			return nil, err
		}
	}
	for lvl := values.Integer(1); lvl <= stoneRule.SkillStoneLv; lvl++ {
		svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskInlayLvlUpperStone, lvl, 1)
	}
	evt, err := t.CalSkill(c, req.ConfigId)
	if err != nil {
		return nil, err
	}
	if evt != nil {
		c.PublishEventLocal(&event.SkillChange{
			ConfigId: req.ConfigId,
			Skills:   evt,
		})
	}
	dao.SaveTalent(c, t)
	return &protosvc.Talent_InlayStoneResponse{
		ConfigId: req.ConfigId,
		Skill:    detail,
	}, nil
}

func (svc *Service) RemoveStone(c *ctx.Context, req *protosvc.Talent_RemoveStoneRequest) (*protosvc.Talent_RemoveStoneResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	stoneId, detail, err := t.RemoveStone(c, req.ConfigId, req.SkillId, int(req.HoleIdx))
	if err != nil {
		return nil, err
	}
	stoneRule, has := rule.MustGetReader(c).Skillstone.GetSkillstoneById(stoneId)
	if !has {
		return nil, errmsg.NewErrSkillStoneNotExist()
	}
	if err = svc.Module.BagService.AddItem(c, c.RoleId, stoneRule.ItemId, 1); err != nil {
		return nil, err
	}
	for lvl := values.Integer(1); lvl <= stoneRule.SkillStoneLv; lvl++ {
		svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskInlayLvlUpperStone, lvl, -1)
	}
	evt, err := t.CalSkill(c, req.ConfigId)
	if err != nil {
		return nil, err
	}
	if evt != nil {
		c.PublishEventLocal(&event.SkillChange{
			ConfigId: req.ConfigId,
			Skills:   evt,
		})
	}
	dao.SaveTalent(c, t)
	return &protosvc.Talent_RemoveStoneResponse{
		ConfigId: req.ConfigId,
		Skill:    detail,
	}, nil
}

func (svc *Service) RemoveAllStone(c *ctx.Context, req *protosvc.Talent_RemoveAllStoneRequest) (*protosvc.Talent_RemoveAllStoneResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	stoneMap, detail, err := t.RemoveAllStone(c, req.ConfigId, req.SkillId)
	if err != nil {
		return nil, err
	}
	if _, err = svc.Module.BagService.AddManyItem(c, c.RoleId, stoneMap); err != nil {
		return nil, err
	}
	evt, err := t.CalSkill(c, req.ConfigId)
	if err != nil {
		return nil, err
	}
	if evt != nil {
		c.PublishEventLocal(&event.SkillChange{
			ConfigId: req.ConfigId,
			Skills:   evt,
		})
	}
	dao.SaveTalent(c, t)
	return &protosvc.Talent_RemoveAllStoneResponse{
		ConfigId: req.ConfigId,
		Skill:    detail,
	}, nil
}

func (svc *Service) ComposeStone(c *ctx.Context, req *protosvc.Talent_ComposeStoneRequest) (*protosvc.Talent_ComposeStoneResponse, *errmsg.ErrMsg) {
	mainRule, ok := rule.MustGetReader(c).Skillstone.GetSkillstoneById(req.MainId)
	if !ok {
		return nil, errmsg.NewErrSkillStoneNotExist()
	}
	if len(mainRule.UpgradeConsume) != 2 {
		return nil, errmsg.NewErrSkillStoneCantCompose()
	}
	tarId := req.MainId + 1
	tarRule, ok := rule.MustGetReader(c).Skillstone.GetSkillstoneById(tarId)
	if !ok {
		return nil, errmsg.NewErrSkillStoneCantCompose()
	}
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	for _, sa := range req.SacrificeIds {
		if t.IsLock(sa) {
			return nil, errmsg.NewErrSkillStoneIsLock()
		}
		if sa != req.MainId {
			return nil, errmsg.NewErrComposeSketchWrong()
		}
	}
	if len(req.SacrificeIds) != int(mainRule.UpgradeConsume[1]-1) {
		return nil, errmsg.NewErrComposeSketchWrong()
	}
	var subItem = map[values.ItemId]values.Integer{
		mainRule.ItemId: mainRule.UpgradeConsume[1],
	}
	if err := svc.Module.BagService.ExchangeManyItem(c, c.RoleId, map[values.ItemId]values.Integer{tarRule.ItemId: 1}, subItem); err != nil {
		return nil, err
	}
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskMergeStoneNum: {
			Typ:     values.Integer(models.TaskType_TaskMergeStoneNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskMergeStoneNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskMergeStoneNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	svc.UpdateTargets(c, c.RoleId, tasks)
	return &protosvc.Talent_ComposeStoneResponse{}, nil
}

func (svc *Service) LockStone(c *ctx.Context, req *protosvc.Talent_LockStoneRequest) (*protosvc.Talent_LockStoneResponse, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	if req.IsLock {
		t.Lock(req.ItemId)
	} else {
		t.Unlock(req.ItemId)
	}
	dao.SaveTalent(c, t)
	return &protosvc.Talent_LockStoneResponse{
		ItemId: req.ItemId,
		IsLock: req.IsLock,
	}, nil
}

// ---------------------------------------------------module------------------------------------------------------------//

func (svc *Service) GetTalentAttr(c *ctx.Context, configId values.Integer) (*models.TalentAdvance, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	return t.CalAttr(c, configId)
}

func (svc *Service) GetAllUnused(c *ctx.Context) (map[values.HeroId]map[values.Integer]values.Integer, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	return t.CalAllUnusedSkill(c)
}

func (svc *Service) GetUnused(c *ctx.Context, configId values.HeroId) (map[values.Integer]values.Integer, *errmsg.ErrMsg) {
	t, err := dao.GetTalent(c)
	if err != nil {
		return nil, err
	}
	return t.CalUnusedSkill(c, configId)
}

// ---------------------------------------------------event------------------------------------------------------------//

func (svc *Service) HandleAddCommonPoints(c *ctx.Context, d *event.GainTalentCommonPoint) *errmsg.ErrMsg {
	t, err := dao.GetTalent(c)
	if err != nil {
		return err
	}
	t.GainCommonP(d.Num)
	dao.SaveTalent(c, t)
	c.PushMessage(&protosvc.Talent_TalentPointChangePush{
		ConfigId: -1,
		Points:   d.Num,
	})
	return nil
}

func (svc *Service) HandleAddParticularPoints(c *ctx.Context, d *event.GainTalentParticularPoint) *errmsg.ErrMsg {
	t, err := dao.GetTalent(c)
	if err != nil {
		return err
	}
	t.GainParticularP(d.ConfigId, d.Num)
	dao.SaveTalent(c, t)
	c.PushMessage(&protosvc.Talent_TalentPointChangePush{
		ConfigId: d.ConfigId,
		Points:   d.Num,
	})
	return nil
}

func (svc *Service) HandleNewHero(c *ctx.Context, d *event.GotHero) *errmsg.ErrMsg {
	t, err := dao.GetTalent(c)
	if err != nil {
		return err
	}
	role, err := svc.Module.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return err
	}
	skillUpCnt, err := t.Init(c, d.OriginId, role.Level)
	if err == nil {
		dao.SaveTalent(c, t)
	}
	svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskTotalSkillLvl, 0, skillUpCnt)
	svc.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskHaveLvlUpperSkill, 1, skillUpCnt)
	heroIds := rule.MustGetReader(c).DeriveHeroMap(d.OriginId)
	if len(heroIds) == 0 {
		return nil
	}
	evt, err := t.CalSkill(c, heroIds[0])
	if err != nil {
		return err
	}
	if evt != nil {
		c.PublishEventLocal(&event.SkillChange{
			ConfigId: heroIds[0],
			Skills:   evt,
		})
	}
	evtAttr, err := t.CalAttr(c, heroIds[0])
	if err != nil {
		return err
	}
	if evtAttr != nil {
		c.PublishEventLocal(&event.TalentChange{
			ConfigId: heroIds[0],
			Attr:     evtAttr,
		})
	}
	return nil
}

func (svc *Service) HandleEquipBonusTalent(c *ctx.Context, d *event.EquipBonusTalent) *errmsg.ErrMsg {
	t, err := dao.GetTalent(c)
	if err != nil {
		return err
	}
	handled := t.UpdateExtraLvl(c, d.HeroId, d.Data)
	for configId := range handled {
		heroIds := rule.MustGetReader(c).DeriveHeroMap(configId)
		if len(heroIds) == 0 {
			continue
		}
		evt, err := t.CalSkill(c, heroIds[0])
		if err != nil {
			return err
		}
		if evt != nil {
			c.PublishEventLocal(&event.SkillChange{
				ConfigId: heroIds[0],
				Skills:   evt,
			})
		}
	}
	dao.SaveTalent(c, t)
	return nil
}

func (svc *Service) HandleLevelUp(c *ctx.Context, d *event.UserLevelChange) *errmsg.ErrMsg {
	t, err := dao.GetTalent(c)
	if err != nil {
		return err
	}
	if skillUpCnt, changeHeroes := t.SkillHandleRoleLvlUp(c, d.Level); skillUpCnt > 0 {
		svc.UpdateTarget(c, c.RoleId, models.TaskType_TaskTotalSkillLvl, 0, skillUpCnt)
		svc.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskHaveLvlUpperSkill, 1, skillUpCnt)
		evts, err := t.CalAllSkill(c)
		if err != nil {
			return err
		}
		if evts != nil {
			for heroId, evt := range evts {
				c.PublishEventLocal(&event.SkillChange{
					ConfigId: heroId,
					Skills:   evt,
				})
			}
		}
		for _, changeHero := range changeHeroes {
			c.PushMessage(&protosvc.Talent_TalentSkillChangePush{
				ConfigId: changeHero.ConfigId,
				Data:     changeHero,
			})
		}
		dao.SaveTalent(c, t)
	}
	return nil
}

// ---------------------------------------------------cheat------------------------------------------------------------//

func (svc *Service) CheatAddCommonPoints(ctx *ctx.Context, req *protosvc.Talent_CheatAddCommonRequest) (*protosvc.Talent_CheatAddCommonResponse, *errmsg.ErrMsg) {
	ctx.PublishEventLocal(&event.GainTalentCommonPoint{
		Num: req.Num,
	})
	return &protosvc.Talent_CheatAddCommonResponse{}, nil
}

func (svc *Service) CheatAddParticularPoints(ctx *ctx.Context, req *protosvc.Talent_CheatAddParticularRequest) (*protosvc.Talent_CheatAddParticularResponse, *errmsg.ErrMsg) {
	ctx.PublishEventLocal(&event.GainTalentParticularPoint{
		ConfigId: req.ConfigId,
		Num:      req.Num,
	})
	return &protosvc.Talent_CheatAddParticularResponse{}, nil
}
