package expedition

import (
	"math"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/expedition/dao"
	"coin-server/game-server/service/expedition/rule"
)

func (svc *Service) HandleTaskChange(ctx *ctx.Context, d *event.TargetUpdate) *errmsg.ErrMsg {
	data := rule.GetAllExpeditionQuantity(ctx)
	var find bool
	for taskType := range data {
		if taskType == d.Typ {
			find = true
			break
		}
	}
	if !find {
		return nil
	}
	e, ok, err := dao.Get(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	slotCount, err := svc.getSlotCount(ctx)
	if err != nil {
		return err
	}
	if e.NormalSlot != slotCount {
		e.NormalSlot = slotCount
		dao.Save(ctx, e)
	}
	return nil
}

func (svc *Service) HandlerVipChange(ctx *ctx.Context) {
	// TODO 待策划设计功能
	ctx.PushMessage(&servicepb.Expedition_ExecutionRecoverCountUpdatePush{
		ExecutionRecoverCount: 0,
		MaxExecution:          0,
	})
}

func (svc *Service) HandleExtraSkillTypTotalUpdate(ctx *ctx.Context, d *event.ExtraSkillTypTotal) *errmsg.ErrMsg {
	if d.TypId != models.EntrySkillType_ESTExpeditionExecutionLimitAdd && d.TypId != models.EntrySkillType_ESTExpeditionExecutionRecoverRateAdd {
		return nil
	}
	e, ok, err := dao.Get(ctx)
	if err != nil {
		return nil
	}
	if !ok {
		slotCount, err := svc.getSlotCount(ctx)
		if err != nil {
			return nil
		}
		e.NormalSlot = slotCount
	}
	svc.nilCheck(e)
	var update bool
	var recoverCount, maxExecution values.Integer
	if d.TypId == models.EntrySkillType_ESTExpeditionExecutionLimitAdd {
		value := d.TotalCnt
		max := rule.GetExpeditionCostLimit(ctx)
		// 百分比
		if d.ValueTyp == 2 {
			value = values.Integer(math.Ceil(values.Float(max) * values.Float(d.TotalCnt) / 10000.0))
		}
		if value > 0 {
			e.Execution.LimitBonus[relics] = value
			update = true
			maxExecution = svc.getMaxExecution(ctx, e.Execution)
		}
	} else if d.TypId == models.EntrySkillType_ESTExpeditionExecutionRecoverRateAdd {
		value := d.TotalCnt
		_, num := rule.GetExpeditionCostRecovery(ctx)
		// 百分比
		if d.ValueTyp == 2 {
			value = values.Integer(math.Ceil(values.Float(num) * values.Float(value) / 10000.0))
		}
		if value > 0 {
			e.Execution.ExtraRestoreCount[relics] = value
			update = true
			recoverCount = num
			for _, v := range e.Execution.ExtraRestoreCount {
				recoverCount += v
			}
		}
	}
	if update {
		dao.Save(ctx, e)
	}
	ctx.PushMessage(&servicepb.Expedition_ExecutionRecoverCountUpdatePush{
		ExecutionRecoverCount: recoverCount,
		MaxExecution:          maxExecution,
	})
	return nil
}

//
// func (svc *Service) HandleRelicsSkillUpdate(ctx *ctx.Context, d *event.RelicsSkillUpdate) *errmsg.ErrMsg {
// 	level := int(d.Level)
// 	cfg, ok := rule.GetRelicsSkill(ctx, d.SkillId)
// 	if !ok {
// 		return nil
// 	}
// 	var typ values.Integer
// 	if len(cfg.Typ()) > 0 {
// 		typ = cfg.Typ()[0]
// 	}
// 	// TODO 看typ是否要统一写在某个地方
// 	if typ != typ13 && typ != typ14 {
// 		return nil
// 	}
// 	e, ok, err := dao.Get(ctx)
// 	if err != nil {
// 		return nil
// 	}
// 	if !ok {
// 		slotCount, err := svc.getSlotCount(ctx)
// 		if err != nil {
// 			return nil
// 		}
// 		e.NormalSlot = slotCount
// 	}
// 	svc.nilCheck(e)
// 	var update bool
// 	var recoverCount, maxExecution values.Integer
// 	if typ == typ13 {
// 		if level > len(cfg.StarsValue()) {
// 			level = len(cfg.StarsValue())
// 		}
// 		value := cfg.StarsValue()[level-1]
// 		max := rule.GetExpeditionCostLimit(ctx)
// 		// 百分比
// 		if cfg.Value == 2 {
// 			value = values.Integer(math.Ceil(values.Float(max) * values.Float(value) / 10000.0))
// 		}
// 		if value > 0 {
// 			e.Execution.LimitBonus[relics] = value
// 			update = true
// 			maxExecution = svc.getMaxExecution(ctx, e.Execution)
// 		}
// 	} else if typ == typ14 {
// 		if level > len(cfg.StarsValue()) {
// 			level = len(cfg.StarsValue())
// 		}
// 		value := cfg.StarsValue()[level-1]
// 		_, num := rule.GetExpeditionCostRecovery(ctx)
// 		// 百分比
// 		if cfg.Value == 2 {
// 			value = values.Integer(math.Ceil(values.Float(num) * values.Float(value) / 10000.0))
// 		}
// 		if value > 0 {
// 			e.Execution.ExtraRestoreCount[relics] = value
// 			update = true
// 			recoverCount = num
// 			for _, v := range e.Execution.ExtraRestoreCount {
// 				recoverCount += v
// 			}
// 		}
// 	}
// 	if update {
// 		dao.Save(ctx, e)
// 	}
// 	ctx.PushMessage(&servicepb.Expedition_ExecutionRecoverCountUpdatePush{
// 		ExecutionRecoverCount: recoverCount,
// 		MaxExecution:          maxExecution,
// 	})
// 	return nil
// }
