package activity

import (
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	protosvc "coin-server/common/proto/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/service/activity/dao"
	"coin-server/rule"
	"go.uber.org/zap"
)

func (svc *Service) ZeroBuyInfo(c *ctx.Context, _ *protosvc.Activity_ZeroBuyInfoRequest) (*protosvc.Activity_ZeroBuyInfoResponse, *errmsg.ErrMsg) {
	data, err := dao.GetZeroBuy(c)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errmsg.NewErrActivityNotExist()
	}
	reader := rule.MustGetReader(c)
	if data.IsEnd {
		return nil, errmsg.NewErrEventHasEnd()
	}
	activityCfg, has := reader.Activity.GetActivityById(enum.ZeroPurchase)
	if !has {
		return nil, errmsg.NewErrActivityNotExist()
	}
	durationTime, er := strconv.ParseInt(activityCfg.DurationTime, 10, 64)
	if er != nil {
		return nil, errmsg.NewErrActivityNotExist()
	}
	cfgList := reader.Activity0yuanpurchase.List()
	cfg := make([]*models.ZeroBuyCfg, 0, len(cfgList))
	for _, originCfg := range cfgList {
		cfg = append(cfg, &models.ZeroBuyCfg{
			Id:                originCfg.Id,
			Day:               originCfg.Day,
			ActivityId:        originCfg.TypId,
			RewardImmediately: originCfg.RewardImmediately,
			RewardTomo:        originCfg.RewardTomo,
			CostItem:          originCfg.CostItem,
			PictureId:         originCfg.PictureId,
		})
	}
	now := timer.StartTime(c.StartTime).Unix()
	if now > data.StartAt+durationTime {
		svc.sendUnDrawMail(c, data)
	}
	return &protosvc.Activity_ZeroBuyInfoResponse{
		Cfg:     cfg,
		StartAt: data.StartAt,
		EndAt:   data.StartAt + durationTime,
		EachDay: data.EachDay,
	}, nil
}

func (svc *Service) DrawZeroBuy(c *ctx.Context, req *protosvc.Activity_DrawZeroBuyRequest) (*protosvc.Activity_DrawZeroBuyResponse, *errmsg.ErrMsg) {
	data, err := dao.GetZeroBuy(c)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errmsg.NewErrActivityNotExist()
	}
	reader := rule.MustGetReader(c)
	activityCfg, has := reader.Activity.GetActivityById(enum.ZeroPurchase)
	if !has {
		return nil, errmsg.NewErrActivityNotExist()
	}
	durationTime, er := strconv.ParseInt(activityCfg.DurationTime, 10, 64)
	if er != nil {
		return nil, errmsg.NewErrActivityNotExist()
	}
	if data.IsEnd {
		return nil, errmsg.NewErrEventHasEnd()
	}
	now := timer.StartTime(c.StartTime).Unix()
	if now > data.StartAt+durationTime {
		svc.sendUnDrawMail(c, data)
		return &protosvc.Activity_DrawZeroBuyResponse{
			EachDay: data.EachDay,
		}, nil
	}
	zeroCfg, has := reader.Activity0yuanpurchase.GetActivity0yuanpurchaseById(req.Day)
	if !has {
		return nil, errmsg.NewErrActivityDrawNotExist()
	}
	if int(req.Day) > len(data.EachDay) {
		return nil, errmsg.NewErrActivityDrawNotExist()
	}
	for day := values.Integer(0); day < req.Day-1; day++ {
		if data.EachDay[day].BuyAt[0] == -1 {
			return nil, errmsg.NewErrActivityPreGiftNotBuy()
		}
	}
	if req.Day > 1 && data.EachDay[req.Day-2].BuyAt[0] > svc.RefreshService.GetCurrDayFreshTime(c).Unix() {
		// 不是第一天购买，就要判断前一天购买是否过天了
		return nil, errmsg.NewErrActivityNotReadyForBuy()
	}
	if data.EachDay[req.Day-1].BuyAt[0] > 0 {
		return nil, errmsg.NewErrEventDoNotCollectAgain()
	}
	subItem := map[values.ItemId]values.Integer{}
	for i := 0; i < len(zeroCfg.CostItem); i += 2 {
		subItem[zeroCfg.CostItem[i]] = zeroCfg.CostItem[i+1]
	}
	if err = svc.SubManyItem(c, c.RoleId, subItem); err != nil {
		return nil, err
	}
	reward := map[values.ItemId]values.Integer{}
	for i := 0; i < len(zeroCfg.RewardImmediately); i += 2 {
		reward[zeroCfg.RewardImmediately[i]] = zeroCfg.RewardImmediately[i+1]
	}
	if _, err = svc.AddManyItem(c, c.RoleId, reward); err != nil {
		return nil, err
	}
	data.EachDay[req.Day-1].BuyAt[0] = now
	dao.SaveZeroBuy(c, data)
	c.PushMessage(&protosvc.Activity_BuyActivityNormalPush{
		Id:    enum.ZeroPurchase,
		BagId: req.Day,
		Items: reward,
	})
	return &protosvc.Activity_DrawZeroBuyResponse{
		EachDay: data.EachDay,
	}, nil
}

func (svc *Service) ZeroBuyDrawNextDay(c *ctx.Context, req *protosvc.Activity_ZeroBuyDrawNextDayRequest) (*protosvc.Activity_ZeroBuyDrawNextDayResponse, *errmsg.ErrMsg) {
	data, err := dao.GetZeroBuy(c)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errmsg.NewErrActivityNotExist()
	}
	reader := rule.MustGetReader(c)
	activityCfg, has := reader.Activity.GetActivityById(enum.ZeroPurchase)
	if !has {
		return nil, errmsg.NewErrActivityNotExist()
	}
	if data.IsEnd {
		return nil, errmsg.NewErrEventHasEnd()
	}
	durationTime, er := strconv.ParseInt(activityCfg.DurationTime, 10, 64)
	if er != nil {
		return nil, errmsg.NewErrActivityNotExist()
	}
	now := timer.StartTime(c.StartTime).Unix()
	if now > data.StartAt+durationTime {
		svc.sendUnDrawMail(c, data)
		return &protosvc.Activity_ZeroBuyDrawNextDayResponse{
			EachDay: data.EachDay,
		}, nil
	}
	zeroCfg, has := reader.Activity0yuanpurchase.GetActivity0yuanpurchaseById(req.Day)
	if !has {
		return nil, errmsg.NewErrActivityDrawNotExist()
	}
	if int(req.Day) > len(data.EachDay) {
		return nil, errmsg.NewErrActivityDrawNotExist()
	}
	for day := values.Integer(0); day < req.Day; day++ {
		if data.EachDay[day].BuyAt[0] == -1 {
			return nil, errmsg.NewErrActivityPreGiftNotBuy()
		}
	}
	todayBegin := svc.RefreshService.GetCurrDayFreshTime(c).Unix()
	if data.EachDay[req.Day-1].BuyAt[0] > todayBegin {
		return nil, errmsg.NewErrActivityNotReadyForCollection()
	}
	if data.EachDay[req.Day-1].BuyAt[1] > 0 {
		return nil, errmsg.NewErrEventDoNotCollectAgain()
	}
	reward := map[values.ItemId]values.Integer{}
	for i := 0; i < len(zeroCfg.RewardTomo); i += 2 {
		reward[zeroCfg.RewardTomo[i]] = zeroCfg.RewardTomo[i+1]
	}
	if _, err = svc.AddManyItem(c, c.RoleId, reward); err != nil {
		return nil, err
	}
	data.EachDay[req.Day-1].BuyAt[1] = now
	dao.SaveZeroBuy(c, data)
	c.PushMessage(&protosvc.Activity_BuyActivityNormalPush{
		Id:    enum.ZeroPurchase,
		BagId: req.Day,
		Items: reward,
	})
	return &protosvc.Activity_ZeroBuyDrawNextDayResponse{
		EachDay: data.EachDay,
	}, nil
}

func (svc *Service) handleZeroBuyUnlock(c *ctx.Context) *errmsg.ErrMsg {
	data, err := dao.GetZeroBuy(c)
	if err != nil {
		return err
	}
	if data != nil {
		return nil
	}
	now := timer.StartTime(c.StartTime).Unix()
	cfgLen := rule.MustGetReader(c).Activity0yuanpurchase.Len()
	newData := &pbdao.ZeroBuy{
		RoleId:  c.RoleId,
		StartAt: now,
		EachDay: make([]*models.ZeroBuy, cfgLen),
	}
	for idx := range newData.EachDay {
		newData.EachDay[idx] = &models.ZeroBuy{
			BuyAt: []values.Integer{-1, -1},
		}
	}
	dao.SaveZeroBuy(c, newData)
	c.PushMessage(&protosvc.Activity_ZeroBuyActivePush{})
	return nil
}

func (svc *Service) isZeroBuyActive(c *ctx.Context) (bool, values.Integer, values.Integer) {
	data, err := dao.GetZeroBuy(c)
	if err != nil {
		return false, 0, 0
	}
	if data == nil {
		return false, 0, 0
	}
	now := timer.StartTime(c.StartTime).Unix()
	activityCfg, has := rule.MustGetReader(c).Activity.GetActivityById(enum.ZeroPurchase)
	if !has {
		return false, 0, 0
	}
	if data.IsEnd {
		return false, 0, 0
	}
	durationTime, er := strconv.ParseInt(activityCfg.DurationTime, 10, 64)
	if er != nil {
		return false, 0, 0
	}
	if now < data.StartAt || now > data.StartAt+durationTime {
		svc.sendUnDrawMail(c, data)
		return false, 0, 0
	}
	return true, data.StartAt, data.StartAt + durationTime
}

func (svc *Service) sendUnDrawMail(c *ctx.Context, data *pbdao.ZeroBuy) {
	if data.IsEnd {
		return
	}
	reader := rule.MustGetReader(c)
	now := timer.StartTime(c.StartTime).Unix()
	for idx, each := range data.EachDay {
		if each.BuyAt[0] > 0 && each.BuyAt[1] == -1 {
			cfg, has := reader.Activity0yuanpurchase.GetActivity0yuanpurchaseById(int64(idx) + 1)
			if has {
				if len(cfg.RewardTomo) == 0 {
					continue
				}
				var rewardList []*models.Item
				for i := 0; i < len(cfg.RewardTomo); i += 2 {
					rewardList = append(rewardList, &models.Item{
						ItemId: cfg.RewardTomo[i],
						Count:  cfg.RewardTomo[i+1],
					})
				}
				if err := svc.MailService.Add(c, c.RoleId, &models.Mail{
					Type:       models.MailType_MailTypeSystem,
					TextId:     cfg.MailId,
					Attachment: rewardList,
					Args:       nil,
				}); err != nil {
					c.Error("send zero by un drew mail error", zap.Any("msg", err), zap.Any("reward", rewardList), zap.Any("roleid", c.RoleId))
					continue
				}
				each.BuyAt[1] = now
			}
		}
	}
	data.IsEnd = true
	dao.SaveZeroBuy(c, data)
}

func (svc *Service) CheatClearZeroBuy(c *ctx.Context, _ *protosvc.Activity_CheatClearZeroBuyRequest) (*protosvc.Activity_CheatClearZeroBuyResponse, *errmsg.ErrMsg) {
	data, err := dao.GetZeroBuy(c)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errmsg.NewErrActivityNotExist()
	}
	for _, each := range data.EachDay {
		each.BuyAt = []values.Integer{-1, -1}
	}
	dao.SaveZeroBuy(c, data)
	return &protosvc.Activity_CheatClearZeroBuyResponse{}, nil
}

func (svc *Service) CheatZeroBuyAheadOneDay(c *ctx.Context, _ *protosvc.Activity_CheatZeroBuyAheadOneDayRequest) (*protosvc.Activity_CheatZeroBuyAheadOneDayResponse, *errmsg.ErrMsg) {
	data, err := dao.GetZeroBuy(c)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errmsg.NewErrActivityNotExist()
	}
	data.StartAt -= 86400
	for _, each := range data.EachDay {
		if each.BuyAt[0] > 0 {
			each.BuyAt[0] -= 86400
		}
		if each.BuyAt[1] > 0 {
			each.BuyAt[1] -= 86400
		}
	}
	dao.SaveZeroBuy(c, data)
	return &protosvc.Activity_CheatZeroBuyAheadOneDayResponse{}, nil
}
