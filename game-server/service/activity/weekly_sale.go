package activity

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/activity/dao"
	"coin-server/rule"
)

const weekSec = 7 * 24 * 60 * 60

func (svc *Service) WeeklySaleInfo(ctx *ctx.Context, _ *servicepb.Activity_WeeklySaleInfoRequest) (*servicepb.Activity_WeeklySaleInfoResponse, *errmsg.ErrMsg) {
	ds, err := dao.GetWeeklySale(ctx)
	if err != nil {
		return nil, err
	}
	hasChange, err := svc.weeklySaleLazyUpdate(ctx, ds, weekSec)
	if err != nil {
		return nil, err
	}
	reader := rule.MustGetReader(ctx)
	cfg := make([]*models.WeeklySaleCfg, 0, len(ds.CanBuyIds))
	for _, id := range ds.CanBuyIds {
		r, ok := reader.ActivityDailygiftBuy.GetActivityDailygiftBuyById(id)
		if !ok {
			return nil, errmsg.NewErrActivityGiftNotExist()
		}
		cfg = append(cfg, &models.WeeklySaleCfg{
			Id:             r.Id,
			BuyType:        r.BuyType,
			PayId:          r.PayId,
			ActivityReward: r.ActivityReward,
			ItemPictureId:  r.ItemPictureId,
			PayNum:         r.PayNum,
			ActivityGridId: r.ActivityGridId,
		})
	}
	if hasChange {
		dao.SaveWeeklySale(ctx, ds)
	}
	return &servicepb.Activity_WeeklySaleInfoResponse{
		Cfg:          cfg,
		BuyTimes:     ds.BuyTimes,
		NxtRefreshAt: ds.LastFreshAt + weekSec,
	}, nil
}

func (svc *Service) weeklySalePayIdBuy(ctx *ctx.Context, payId values.Integer) *errmsg.ErrMsg {
	ds, err := dao.GetWeeklySale(ctx)
	if err != nil {
		return err
	}
	_, err = svc.weeklySaleLazyUpdate(ctx, ds, weekSec)
	if err != nil {
		return err
	}
	reader := rule.MustGetReader(ctx)
	for _, idx := range ds.CanBuyIds {
		cfg, ok := reader.ActivityDailygiftBuy.GetActivityDailygiftBuyById(idx)
		if !ok {
			continue
		}
		if cfg.BuyType == 2 && cfg.PayId[0] == payId {
			return svc.weeklySaleBuyIn(ctx, idx, ds)
		}
	}
	return nil
}

func (svc *Service) weeklySaleBuyIn(ctx *ctx.Context, idx values.Integer, ds *pbdao.WeeklySale) *errmsg.ErrMsg {
	cfg, ok := rule.MustGetReader(ctx).ActivityDailygiftBuy.GetActivityDailygiftBuyById(idx)
	if !ok {
		return errmsg.NewErrActivityGiftNotExist()
	}
	if ds.BuyTimes[cfg.ActivityGridId] >= cfg.PayNum {
		return errmsg.NewErrActivityGiftEmpty()
	}
	item := make(map[values.ItemId]values.Integer, 0)
	for i := 0; i < len(cfg.ActivityReward); i += 2 {
		item[cfg.ActivityReward[i]] += cfg.ActivityReward[i+1]
	}
	if len(item) > 0 {
		if _, err := svc.AddManyItem(ctx, ctx.RoleId, item); err != nil {
			return err
		}
		ctx.PushMessage(&servicepb.Activity_BuyActivityNormalPush{
			Id:    enum.WeeklySale,
			BagId: idx,
			Items: item,
		})
	}
	if cfg.BuyType == 1 {
		subItem := make(map[values.ItemId]values.Integer, 0)
		for i := 0; i < len(cfg.PayId); i += 2 {
			subItem[cfg.PayId[i]] += cfg.PayId[i+1]
		}
		if err := svc.SubManyItem(ctx, ctx.RoleId, subItem); err != nil {
			return err
		}
	}
	ds.BuyTimes[cfg.ActivityGridId] += 1
	reader := rule.MustGetReader(ctx)
	res := make([]*models.WeeklySaleCfg, 0, len(ds.CanBuyIds))
	for _, id := range ds.CanBuyIds {
		r, ok := reader.ActivityDailygiftBuy.GetActivityDailygiftBuyById(id)
		if !ok {
			return errmsg.NewErrActivityGiftNotExist()
		}
		res = append(res, &models.WeeklySaleCfg{
			Id:             r.Id,
			BuyType:        r.BuyType,
			PayId:          r.PayId,
			ActivityReward: r.ActivityReward,
			ItemPictureId:  r.ItemPictureId,
			PayNum:         r.PayNum,
			ActivityGridId: r.ActivityGridId,
		})
	}
	ctx.PushMessage(&servicepb.Activity_WeeklySaleUpdatePush{
		BuyTimes:     ds.BuyTimes,
		Cfg:          res,
		NxtRefreshAt: ds.LastFreshAt + weekSec,
	})
	dao.SaveWeeklySale(ctx, ds)
	return nil
}

func (svc *Service) WeeklySaleBuy(ctx *ctx.Context, req *servicepb.Activity_WeeklySaleBuyRequest) (*servicepb.Activity_WeeklySaleBuyResponse, *errmsg.ErrMsg) {
	ds, err := dao.GetWeeklySale(ctx)
	if err != nil {
		return nil, err
	}
	_, err = svc.weeklySaleLazyUpdate(ctx, ds, weekSec)
	if err != nil {
		return nil, err
	}
	cfg, ok := rule.MustGetReader(ctx).ActivityDailygiftBuy.GetActivityDailygiftBuyById(req.Idx)
	if !ok {
		return nil, errmsg.NewErrActivityGiftNotExist()
	}
	if cfg.BuyType != 1 {
		return nil, errmsg.NewErrActivityGiftNotExist()
	}
	if err = svc.weeklySaleBuyIn(ctx, req.Idx, ds); err != nil {
		return nil, err
	}
	return &servicepb.Activity_WeeklySaleBuyResponse{}, nil
}

func (svc *Service) CheatRefreshWeekly(c *ctx.Context, _ *servicepb.Activity_CheatRefreshWeeklyRequest) (*servicepb.Activity_CheatRefreshWeeklyResponse, *errmsg.ErrMsg) {
	ds, err := dao.GetWeeklySale(c)
	if err != nil {
		return nil, err
	}
	ds.LastFreshAt = svc.RefreshService.GetActivityCurrWeekFreshTime(c).Unix()
	ds.BuyTimes = map[int64]int64{}
	if err = svc.updateWeeklyCanBuyIds(c, ds); err != nil {
		return nil, err
	}
	reader := rule.MustGetReader(c)
	res := make([]*models.WeeklySaleCfg, 0, len(ds.CanBuyIds))
	for _, id := range ds.CanBuyIds {
		r, ok := reader.ActivityDailygiftBuy.GetActivityDailygiftBuyById(id)
		if !ok {
			return nil, errmsg.NewErrActivityGiftNotExist()
		}
		res = append(res, &models.WeeklySaleCfg{
			Id:             r.Id,
			BuyType:        r.BuyType,
			PayId:          r.PayId,
			ActivityReward: r.ActivityReward,
			ItemPictureId:  r.ItemPictureId,
			PayNum:         r.PayNum,
			ActivityGridId: r.ActivityGridId,
		})
	}
	c.PushMessage(&servicepb.Activity_WeeklySaleUpdatePush{
		BuyTimes:     ds.BuyTimes,
		Cfg:          res,
		NxtRefreshAt: ds.LastFreshAt + weekSec,
	})
	dao.SaveWeeklySale(c, ds)
	return &servicepb.Activity_CheatRefreshWeeklyResponse{}, nil
}

func (svc *Service) weeklySaleLazyUpdate(c *ctx.Context, sale *pbdao.WeeklySale, reset int64) (bool, *errmsg.ErrMsg) {
	now := timer.StartTime(c.StartTime).UTC()
	if now.Unix()-sale.LastFreshAt > reset {
		sale.LastFreshAt = svc.RefreshService.GetActivityCurrWeekFreshTime(c).Unix()
		sale.BuyTimes = map[int64]int64{}
		if err := svc.updateWeeklyCanBuyIds(c, sale); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (svc *Service) updateWeeklyCanBuyIds(c *ctx.Context, sale *pbdao.WeeklySale) *errmsg.ErrMsg {
	sale.CanBuyIds = make([]int64, 0, 16)
	list := rule.MustGetReader(c).ActivityDailygiftBuy.List()
	role, err := svc.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return err
	}
	rcg := role.Recharge / 100
	for _, v := range list {
		if v.TypId != enum.WeeklySale {
			continue
		}
		if role.Level >= v.GradeRange[0] && role.Level <= v.GradeRange[1] && rcg >= v.RechargeInterval[0] {
			sale.CanBuyIds = append(sale.CanBuyIds, v.Id)
		}
	}
	return nil
}

func (svc *Service) HandleWeeklySaleLevelUpEvent(c *ctx.Context, evt *event.UserLevelChange) *errmsg.ErrMsg {
	weeklySaleLevelLine, _ := rule.MustGetReader(c).CustomParse.GetWeeklySaleLine()
	for _, line := range weeklySaleLevelLine {
		if evt.Level == line {
			ds, err := dao.GetWeeklySale(c)
			if err != nil {
				return err
			}
			if err = svc.updateWeeklyCanBuyIds(c, ds); err != nil {
				return err
			}
			reader := rule.MustGetReader(c)
			cfg := make([]*models.WeeklySaleCfg, 0, len(ds.CanBuyIds))
			for _, id := range ds.CanBuyIds {
				r, ok := reader.ActivityDailygiftBuy.GetActivityDailygiftBuyById(id)
				if !ok {
					return errmsg.NewErrActivityGiftNotExist()
				}
				cfg = append(cfg, &models.WeeklySaleCfg{
					Id:             r.Id,
					BuyType:        r.BuyType,
					PayId:          r.PayId,
					ActivityReward: r.ActivityReward,
					ItemPictureId:  r.ItemPictureId,
					PayNum:         r.PayNum,
					ActivityGridId: r.ActivityGridId,
				})
			}
			c.PushMessage(&servicepb.Activity_WeeklySaleUpdatePush{
				BuyTimes:     ds.BuyTimes,
				Cfg:          cfg,
				NxtRefreshAt: ds.LastFreshAt + weekSec,
			})
			dao.SaveWeeklySale(c, ds)
		}
	}
	return nil
}

func (svc *Service) HandleWeeklySaleRechargeEvent(c *ctx.Context, evt *event.RechargeSuccEvt) *errmsg.ErrMsg {
	_, weeklySaleChargeLine := rule.MustGetReader(c).CustomParse.GetWeeklySaleLine()
	for _, line := range weeklySaleChargeLine {
		if evt.Old < line && evt.New >= line {
			ds, err := dao.GetWeeklySale(c)
			if err != nil {
				return err
			}
			if err = svc.updateWeeklyCanBuyIds(c, ds); err != nil {
				return err
			}
			reader := rule.MustGetReader(c)
			cfg := make([]*models.WeeklySaleCfg, 0, len(ds.CanBuyIds))
			for _, id := range ds.CanBuyIds {
				r, ok := reader.ActivityDailygiftBuy.GetActivityDailygiftBuyById(id)
				if !ok {
					return errmsg.NewErrActivityGiftNotExist()
				}
				cfg = append(cfg, &models.WeeklySaleCfg{
					Id:             r.Id,
					BuyType:        r.BuyType,
					PayId:          r.PayId,
					ActivityReward: r.ActivityReward,
					ItemPictureId:  r.ItemPictureId,
					PayNum:         r.PayNum,
					ActivityGridId: r.ActivityGridId,
				})
			}
			c.PushMessage(&servicepb.Activity_WeeklySaleUpdatePush{
				BuyTimes:     ds.BuyTimes,
				Cfg:          cfg,
				NxtRefreshAt: ds.LastFreshAt + weekSec,
			})
			dao.SaveWeeklySale(c, ds)
		}
	}
	return nil
}
