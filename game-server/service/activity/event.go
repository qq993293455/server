package activity

import (
	"fmt"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	modelspb "coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/activity/dao"
	"coin-server/rule"
)

func (svc *Service) HandleLoginEvent(ctx *ctx.Context, _ *event.Login) *errmsg.ErrMsg {
	if err := svc.getThreeDayLock(ctx); err != nil {
		return err
	}
	td, err := dao.GetThreeDayData(ctx)
	if err != nil {
		return err
	}
	if td.Recharge && td.LoginDay < 3 {
		lastDayNum := td.LoginDay
		dft := svc.GetCurrDayFreshTime(ctx).Unix()
		// 这里要用time.Unix取上次时间，不能用timer.StartTime，因为可能存在偏移量，timer.StartTime会包含偏移量
		if td.LastLoginTime == 0 || td.LastLoginTime != dft {
			td.LoginDay++
			td.LastLoginTime = dft
		}
		if lastDayNum != td.LoginDay {
			dao.SaveThreeDayData(ctx, td)
		}
	}

	// 检查是否有新地限时弹窗礼包
	lock, err := dao.GetLimitedTimePackageLocks(ctx)
	if err != nil {
		return err
	}
	needSave := false
	for _, cfg := range rule.MustGetReader(ctx).ActivityLimitedtimePackage.List() {
		if _, ok := lock.Locks[cfg.Id]; ok {
			continue
		}
		lock.Locks[cfg.Id] = 0
		needSave = true
	}
	if needSave {
		dao.SaveLimitedTimePackageLocks(ctx, lock)
	}

	return nil
}

func (svc *Service) HandleTargetUpdate(ctx *ctx.Context, e *event.TargetUpdate) *errmsg.ErrMsg {
	if e.Typ == modelspb.TaskType_TaskGetItemTask && e.Id == enum.DailyTaskActive {
		err := svc.addPassesExp(ctx, e.Count)
		if err != nil {
			return err
		}
	}

	reader := rule.MustGetReader(ctx)
	targets, ok := reader.ActivityLimitedtimePackage.GetAllUnlockConditions()[e.Typ]
	if !ok {
		return nil
	}
	if _, ok = targets[e.Id]; !ok {
		return nil
	}

	count := e.Incr
	if e.IsAccumulate {
		count = e.Count
	}
	err := svc.checkUnlock(ctx, e.Typ, e.Id, count, e.IsAccumulate)
	if err != nil {
		return err
	}

	return nil
}

func (svc *Service) HandleRechargeAmountEvt(c *ctx.Context, d *event.RechargeAmountEvt) *errmsg.ErrMsg {
	role, err := svc.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return err
	}
	old := role.Recharge
	role.Recharge += d.Amount
	c.PublishEventLocal(&event.RechargeSuccEvt{
		Old: old,
		New: role.Recharge,
	})
	svc.SaveRole(c, role)
	return nil
}

func (svc *Service) HandlePaySuccess(ctx *ctx.Context, d *event.PaySuccess) *errmsg.ErrMsg {
	if err := svc.handlePayThreeDaySuccess(ctx); err != nil {
		return err
	}
	return nil
}

func (svc *Service) HandleSystemUnlock(ctx *ctx.Context, d *event.SystemUnlock) *errmsg.ErrMsg {
	for _, sys := range d.SystemId {
		if sys == modelspb.SystemType_SystemZeroBuy {
			return svc.handleZeroBuyUnlock(ctx)
		}
	}
	return nil
}

func (svc *Service) HandleDailyPaySuccess(ctx *ctx.Context, d *event.PaySuccess) *errmsg.ErrMsg {
	v, ok := rule.MustGetReader(ctx).Charge.GetChargeById(d.PcId)
	if !ok {
		return errmsg.NewErrActivityGiftEmpty()
	}
	if (v.FunctionType == 1 && v.TargetId == enum.DailySale) || (v.FunctionType == 2 && v.TargetId == values.Integer(modelspb.SystemType_SystemDailySale)) {
		return svc.dailySalePayIdBuy(ctx, d.PcId)
	} else if (v.FunctionType == 1 && v.TargetId == enum.WeeklySale) || (v.FunctionType == 2 && v.TargetId == values.Integer(modelspb.SystemType_SystemWeeklySale)) {
		return svc.weeklySalePayIdBuy(ctx, d.PcId)
	}
	return nil
}

// HandlePassesPaySuccess 购买高级通行证
func (svc *Service) HandlePassesPaySuccess(ctx *ctx.Context, d *event.PaySuccess) *errmsg.ErrMsg {
	v, ok := rule.MustGetReader(ctx).Charge.GetChargeById(d.PcId)
	if !ok {
		return errmsg.NewInternalErr(fmt.Sprintf("Change config not found. id: %d", d.PcId))
	}
	if v.FunctionType == 2 && v.TargetId == int64(modelspb.SystemType_SystemPasses) {
		return svc.BuyAdvancePasses(ctx)
	}
	return nil
}

func (svc *Service) HandleLevelGrowthFundPaySuccess(ctx *ctx.Context, d *event.PaySuccess) *errmsg.ErrMsg {
	v, ok := rule.MustGetReader(ctx).Charge.GetChargeById(d.PcId)
	if !ok {
		return errmsg.NewInternalErr(fmt.Sprintf("Change config not found. id: %d", d.PcId))
	}
	if v.FunctionType == 1 && v.TargetId == int64(modelspb.SystemType_SystemLevelGrowthFund) {
		data, err := dao.GetLevelGrowthFundData(ctx)
		if err != nil {
			return err
		}
		data.Buy = true
		if err := svc.rebateReward(ctx, data); err != nil {
			return err
		}
		ctx.PushMessage(&servicepb.Activity_LevelGrowthFundBuyPush{
			Buy: true,
		})
	}
	return nil
}

// HandleLimitedTimePackPaySuccess 购买限时弹窗礼包
func (svc *Service) HandleLimitedTimePackPaySuccess(ctx *ctx.Context, d *event.PaySuccess) *errmsg.ErrMsg {
	v, ok := rule.MustGetReader(ctx).Charge.GetChargeById(d.PcId)
	if !ok {
		return errmsg.NewInternalErr(fmt.Sprintf("Change config not found. id: %d", d.PcId))
	}
	if v.FunctionType == 2 && v.TargetId == int64(modelspb.SystemType_SystemLimitTimePack) {
		return svc.BuyLimitedTimePackage(ctx, d.PcId)
	}
	return nil
}
