package bag

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/cppbattle"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/bag/dao"
	"coin-server/rule"

	"go.uber.org/zap"
)

func (s *Service) HandleRoleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	info, err := dao.GetMedicineInfo(ctx, d.RoleId)
	if err != nil {
		return err
	}
	next := info.NextTake[1]
	ctx.PushMessage(&servicepb.GameBattle_MedicineCdPush{
		Cd: next,
	})
	return nil
}

func (s *Service) HandleMedicineChange(ctx *ctx.Context, d *event.ItemUpdate) *errmsg.ErrMsg {
	_, err := s.checkMedicine(ctx, 1)
	return err
}

func (s *Service) HandlerUserLevelUp(ctx *ctx.Context, d *event.UserLevelChange) *errmsg.ErrMsg {
	_, err := s.checkMedicine(ctx, 1)
	return err
}

func (s *Service) checkMedicine(ctx *ctx.Context, typ values.Integer) (*daopb.MedicineInfo, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	list, ok := r.Medicament.GetMedicineByType(typ)
	if !ok {
		return nil, errmsg.NewInternalErr("invalid medicine type")
	}
	info, err := dao.GetMedicineInfo(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	level, err1 := s.UserService.GetLevel(ctx, ctx.RoleId)
	if err1 != nil {
		return nil, err1
	}
	haveValidItems := values.Integer(0)
	for _, medicine := range list {
		if medicine.Level > level {
			continue
		}
		cnt, err2 := s.BagService.GetItem(ctx, ctx.RoleId, medicine.Id)
		if err2 != nil {
			return nil, err2
		}
		if cnt > 0 {
			haveValidItems = 1
			break
		}
	}
	old := info.Open[typ]
	if old != haveValidItems {
		info.Open[typ] = haveValidItems
		dao.SaveMedicineInfo(ctx, info)
		config := map[int64]*models.MedicineInfo{}
		for k, v := range info.AutoTake {
			inf := &models.MedicineInfo{
				CdTime:   0,
				Open:     info.Open[k],
				AutoTake: v,
			}
			if info.NextTake[k]-timer.StartTime(ctx.StartTime).UnixMilli() <= 0 {
				inf.CdTime = 0
			} else {
				inf.CdTime = info.NextTake[k] - timer.StartTime(ctx.StartTime).UnixMilli()
			}
			config[k] = inf
		}
		err3 := s.svc.GetNatsClient().PublishCtx(ctx, ctx.BattleServerId, &cppbattle.CPPBattle_MedicineItemsPush{
			Medicine: config,
		})
		s.log.Debug("checkMedicine1", zap.String("info", info.GoString()))
		return info, err3
	}
	s.log.Debug("checkMedicine2", zap.String("info", info.GoString()))
	return info, nil
}

func (s *Service) HandleItemUpdateEvent(ctx *ctx.Context, d *event.ItemUpdate) *errmsg.ErrMsg {
	negative := make(map[values.ItemId]values.Integer)
	var exp bool
	for i, item := range d.Items {
		if d.Incr[i] < 0 {
			negative[item.ItemId] += -d.Incr[i]
		}
		if item.ItemId == enum.RoleExp {
			exp = true
		}
	}
	// TODO 待任务那边优化
	for id, val := range negative {
		s.UpdateTarget(ctx, d.RoleId, models.TaskType_TaskConsumeItemAcc, id, val)
		s.UpdateTarget(ctx, d.RoleId, models.TaskType_TaskConsumeItem, id, val)
	}
	if exp {
		ctx.Debug("exp change",
			zap.String("ctx_role_id", ctx.RoleId),
			zap.String("role_id", d.RoleId), zap.Any("items", d.Items))
	}
	ctx.PushMessageToRole(d.RoleId, &servicepb.Bag_ItemUpdatePush{Items: d.Items})
	return nil
}

func (s *Service) HandleEquipUpdateEvent(ctx *ctx.Context, d *event.EquipUpdate) *errmsg.ErrMsg {
	equips := make([]*models.Equipment, 0, len(d.Equips))
	for i := 0; i < len(d.Equips); i++ {
		equip := d.Equips[i]
		equips = append(equips, equip)
	}
	if d.New {
		ctx.PushMessageToRole(d.RoleId, &servicepb.Bag_EquipGotPush{Equipments: equips})
	} else {
		ctx.PushMessageToRole(d.RoleId, &servicepb.Bag_EquipUpdatePush{Equipments: equips})
	}
	return nil
}

func (s *Service) HandleEquipDestroyedEvent(ctx *ctx.Context, d *event.EquipDestroyed) *errmsg.ErrMsg {
	ctx.PushMessageToRole(d.RoleId, &servicepb.Bag_EquipDestroyedPush{EquipId: d.EquipId})
	return nil
}

func (s *Service) HandleRelicsUpdate(ctx *ctx.Context, d *event.RelicsUpdate) *errmsg.ErrMsg {
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.Bag_RelicsUpdatePush{Relics: d.Relics})
	return nil
}

func (s *Service) HandleSkillStoneUpdate(ctx *ctx.Context, d *event.SkillStoneUpdate) *errmsg.ErrMsg {
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.Bag_SkillStoneUpdatePush{SkillStones: d.Stones})
	return nil
}

func (s *Service) HandleTalentRuneUpdate(ctx *ctx.Context, d *event.TalentRuneUpdate) *errmsg.ErrMsg {
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.Bag_RuneUpdatePush{Runes: d.Runes})
	return nil
}

func (s *Service) HandleTalentRuneDel(ctx *ctx.Context, d *event.TalentRuneDestroyed) *errmsg.ErrMsg {
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.Bag_RuneDestroyedPush{RuneIds: d.RuneIds})
	return nil
}

func (s *Service) HandleBattleSettingChange(ctx *ctx.Context, d *event.BattleSettingChange) *errmsg.ErrMsg {
	info, err := dao.GetMedicineInfo(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	info.AutoTake = map[int64]int64{
		1: d.Setting.Hp,
		2: d.Setting.Mp,
	}
	dao.SaveMedicineInfo(ctx, info)
	config := map[int64]*models.MedicineInfo{}
	for k, v := range info.AutoTake {
		inf := &models.MedicineInfo{
			CdTime:   0,
			Open:     info.Open[k],
			AutoTake: v,
		}
		if info.NextTake[k]-timer.StartTime(ctx.StartTime).UnixMilli() <= 0 {
			inf.CdTime = 0
		} else {
			inf.CdTime = info.NextTake[k] - timer.StartTime(ctx.StartTime).UnixMilli()
		}

		config[k] = inf
	}
	err = s.svc.GetNatsClient().PublishCtx(ctx, ctx.BattleServerId, &cppbattle.CPPBattle_MedicineItemsPush{
		Medicine: config,
	})
	return nil
}
