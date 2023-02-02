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

const daySec = 24 * 60 * 60

func (svc *Service) DailySaleInfo(ctx *ctx.Context, _ *servicepb.Activity_DailySaleInfoRequest) (*servicepb.Activity_DailySaleInfoResponse, *errmsg.ErrMsg) {
	ds, err := dao.GetDailySale(ctx)
	if err != nil {
		return nil, err
	}
	hasChange, err := svc.dailySaleLazyUpdate(ctx, ds, daySec)
	if err != nil {
		return nil, err
	}
	reader := rule.MustGetReader(ctx)
	cfg := make([]*models.DailySaleCfg, 0, len(ds.CanBuyIds))
	for _, id := range ds.CanBuyIds {
		r, ok := reader.ActivityDailygiftBuy.GetActivityDailygiftBuyById(id)
		if !ok {
			return nil, errmsg.NewErrActivityGiftNotExist()
		}
		cfg = append(cfg, &models.DailySaleCfg{
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
		dao.SaveDailySale(ctx, ds)
	}
	return &servicepb.Activity_DailySaleInfoResponse{
		Cfg:      cfg,
		BuyTimes: ds.BuyTimes,
	}, nil
}

func (svc *Service) dailySalePayIdBuy(ctx *ctx.Context, payId values.Integer) *errmsg.ErrMsg {
	ds, err := dao.GetDailySale(ctx)
	if err != nil {
		return err
	}
	_, err = svc.dailySaleLazyUpdate(ctx, ds, daySec)
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
			return svc.dailySaleBuyIn(ctx, idx, ds)
		}
	}
	return nil
}

func (svc *Service) dailySaleBuyIn(ctx *ctx.Context, idx values.Integer, ds *pbdao.DailySale) *errmsg.ErrMsg {
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
			Id:    enum.DailySale,
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
	res := make([]*models.DailySaleCfg, 0, len(ds.CanBuyIds))
	for _, id := range ds.CanBuyIds {
		r, ok := reader.ActivityDailygiftBuy.GetActivityDailygiftBuyById(id)
		if !ok {
			return errmsg.NewErrActivityGiftNotExist()
		}
		res = append(res, &models.DailySaleCfg{
			Id:             r.Id,
			BuyType:        r.BuyType,
			PayId:          r.PayId,
			ActivityReward: r.ActivityReward,
			ItemPictureId:  r.ItemPictureId,
			PayNum:         r.PayNum,
			ActivityGridId: r.ActivityGridId,
		})
	}
	ctx.PushMessage(&servicepb.Activity_DailySaleUpdatePush{
		BuyTimes: ds.BuyTimes,
		Cfg:      res,
	})
	dao.SaveDailySale(ctx, ds)
	return nil
}

func (svc *Service) DailySaleBuy(ctx *ctx.Context, req *servicepb.Activity_DailySaleBuyRequest) (*servicepb.Activity_DailySaleBuyResponse, *errmsg.ErrMsg) {
	ds, err := dao.GetDailySale(ctx)
	if err != nil {
		return nil, err
	}
	_, err = svc.dailySaleLazyUpdate(ctx, ds, daySec)
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
	if err := svc.dailySaleBuyIn(ctx, req.Idx, ds); err != nil {
		return nil, err
	}
	return &servicepb.Activity_DailySaleBuyResponse{}, nil
}

func (svc *Service) dailySaleLazyUpdate(c *ctx.Context, sale *pbdao.DailySale, reset int64) (bool, *errmsg.ErrMsg) {
	now := timer.StartTime(c.StartTime).UTC()
	if now.Unix()-sale.LastFreshAt > reset {
		sale.LastFreshAt = svc.RefreshService.GetCurrDayFreshTime(c).Unix()
		sale.BuyTimes = map[int64]int64{}
		if err := svc.updateCanBuyIds(c, sale); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (svc *Service) updateCanBuyIds(c *ctx.Context, sale *pbdao.DailySale) *errmsg.ErrMsg {
	sale.CanBuyIds = make([]int64, 0, 16)
	list := rule.MustGetReader(c).ActivityDailygiftBuy.List()
	role, err := svc.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return err
	}
	rcg := role.Recharge / 100
	for _, v := range list {
		if v.TypId != enum.DailySale {
			continue
		}
		if role.Level >= v.GradeRange[0] && role.Level <= v.GradeRange[1] && rcg >= v.RechargeInterval[0] {
			sale.CanBuyIds = append(sale.CanBuyIds, v.Id)
		}
	}
	return nil
}

func (svc *Service) HandleDailySaleLevelUpEvent(c *ctx.Context, evt *event.UserLevelChange) *errmsg.ErrMsg {
	dailySaleLevelLine, _ := rule.MustGetReader(c).CustomParse.GetDailySaleLine()
	for _, line := range dailySaleLevelLine {
		if evt.Level == line {
			ds, err := dao.GetDailySale(c)
			if err != nil {
				return err
			}
			if err = svc.updateCanBuyIds(c, ds); err != nil {
				return err
			}
			reader := rule.MustGetReader(c)
			cfg := make([]*models.DailySaleCfg, 0, len(ds.CanBuyIds))
			for _, id := range ds.CanBuyIds {
				r, ok := reader.ActivityDailygiftBuy.GetActivityDailygiftBuyById(id)
				if !ok {
					return errmsg.NewErrActivityGiftNotExist()
				}
				cfg = append(cfg, &models.DailySaleCfg{
					Id:             r.Id,
					BuyType:        r.BuyType,
					PayId:          r.PayId,
					ActivityReward: r.ActivityReward,
					ItemPictureId:  r.ItemPictureId,
					PayNum:         r.PayNum,
					ActivityGridId: r.ActivityGridId,
				})
			}
			c.PushMessage(&servicepb.Activity_DailySaleUpdatePush{
				BuyTimes: ds.BuyTimes,
				Cfg:      cfg,
			})
			dao.SaveDailySale(c, ds)
		}
	}
	return nil
}

func (svc *Service) HandleDailySaleRechargeEvent(c *ctx.Context, evt *event.RechargeSuccEvt) *errmsg.ErrMsg {
	_, dailySaleChargeLine := rule.MustGetReader(c).CustomParse.GetDailySaleLine()
	for _, line := range dailySaleChargeLine {
		if evt.Old < line && evt.New >= line {
			ds, err := dao.GetDailySale(c)
			if err != nil {
				return err
			}
			if err = svc.updateCanBuyIds(c, ds); err != nil {
				return err
			}
			reader := rule.MustGetReader(c)
			cfg := make([]*models.DailySaleCfg, 0, len(ds.CanBuyIds))
			for _, id := range ds.CanBuyIds {
				r, ok := reader.ActivityDailygiftBuy.GetActivityDailygiftBuyById(id)
				if !ok {
					return errmsg.NewErrActivityGiftNotExist()
				}
				cfg = append(cfg, &models.DailySaleCfg{
					Id:             r.Id,
					BuyType:        r.BuyType,
					PayId:          r.PayId,
					ActivityReward: r.ActivityReward,
					ItemPictureId:  r.ItemPictureId,
					PayNum:         r.PayNum,
					ActivityGridId: r.ActivityGridId,
				})
			}
			c.PushMessage(&servicepb.Activity_DailySaleUpdatePush{
				BuyTimes: ds.BuyTimes,
				Cfg:      cfg,
			})
			dao.SaveDailySale(c, ds)
		}
	}
	return nil
}
