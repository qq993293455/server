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
	"coin-server/game-server/service/activity/dao"
	"coin-server/game-server/service/activity/rule"
)

const threeDayLocker = "lock:three_day:"

func (svc *Service) handlePayThreeDaySuccess(ctx *ctx.Context) *errmsg.ErrMsg {
	if err := svc.getThreeDayLock(ctx); err != nil {
		return err
	}
	td, err := dao.GetThreeDayData(ctx)
	if err != nil {
		return err
	}
	if td.Recharge {
		return nil
	}
	td.Recharge = true
	// 充值当天算第一天
	td.LoginDay = 1
	td.LastLoginTime = svc.GetCurrDayFreshTime(ctx).Unix()
	dao.SaveThreeDayData(ctx, td)
	ctx.PushMessage(&servicepb.Activity_BuyActivityNormalPush{
		Id: enum.ThreeDay,
	})
	return nil
}

func (svc *Service) ThreeDayInfo(ctx *ctx.Context, _ *servicepb.Activity_ActivityThreeDayInfoRequest) (*servicepb.Activity_ActivityThreeDayInfoResponse, *errmsg.ErrMsg) {
	if err := svc.getThreeDayLock(ctx); err != nil {
		return nil, err
	}
	td, err := dao.GetThreeDayData(ctx)
	if err != nil {
		return nil, err
	}
	update := svc.handleAcrossDay(ctx, td)

	list := rule.GetFirstPay(ctx)
	cfgList := make([]*models.FirstPayCfg, 0, len(list))
	for _, item := range list {
		cfgList = append(cfgList, &models.FirstPayCfg{
			Id:             item.Id,
			TypeId:         item.TypId,
			ItemPictureId:  item.ItemPictureId,
			SpecialFieldId: item.SpecialFieldId,
			Day:            item.Day,
			ActivityReward: func() []values.Integer {
				temp := make([]values.Integer, 0)
				for _, reward := range item.ActivityReward {
					temp = append(temp, reward)
				}
				return temp
			}(),
		})
	}
	if update {
		dao.SaveThreeDayData(ctx, td)
	}
	return &servicepb.Activity_ActivityThreeDayInfoResponse{
		LoginDay: td.LoginDay,
		Receive:  td.Receive,
		Cfg:      cfgList,
		Recharge: td.Recharge,
	}, nil
}

func (svc *Service) ThreeDayGetReward(ctx *ctx.Context, req *servicepb.Activity_ActivityThreeDayGetRewardRequest) (*servicepb.Activity_ActivityThreeDayGetRewardResponse, *errmsg.ErrMsg) {
	if err := svc.getThreeDayLock(ctx); err != nil {
		return nil, err
	}
	td, err := dao.GetThreeDayData(ctx)
	if err != nil {
		return nil, err
	}
	if !td.Recharge {
		return nil, errmsg.NewErrNeedRecharge()
	}
	if _, ok := td.Receive[req.Day]; ok {
		return nil, errmsg.NewErrEventDoNotCollectAgain()
	}
	cfg, ok := rule.GetFirstPayById(ctx, req.Day)
	if !ok {
		return nil, errmsg.NewErrEventHasEnd()
	}
	svc.handleAcrossDay(ctx, td)

	if req.Day > td.LoginDay {
		return nil, errmsg.NewErrActivityNotReadyForCollection()
	}

	item := make(map[values.ItemId]values.Integer)
	for i := 0; i < len(cfg.ActivityReward); i += 2 {
		item[cfg.ActivityReward[i]] = cfg.ActivityReward[i+1]
	}
	if len(item) > 0 {
		if _, err := svc.AddManyItem(ctx, ctx.RoleId, item); err != nil {
			return nil, err
		}
	}
	td.Receive[req.Day] = timer.StartTime(ctx.StartTime).UnixMilli()
	td.ReceiveTime = timer.BeginOfDay(timer.StartTime(ctx.StartTime)).UnixMilli()
	dao.SaveThreeDayData(ctx, td)
	return &servicepb.Activity_ActivityThreeDayGetRewardResponse{
		Reward: item,
	}, nil
}

func (svc *Service) showThreeDayEvent(ctx *ctx.Context) (bool, *errmsg.ErrMsg) {
	td, err := dao.GetThreeDayData(ctx)
	if err != nil {
		return false, err
	}
	return len(td.Receive) < 3, nil
}

func (svc *Service) handleAcrossDay(ctx *ctx.Context, td *pbdao.ThreeDay) bool {
	if !td.Recharge {
		return false
	}
	dft := svc.GetCurrDayFreshTime(ctx).Unix()
	if td.LastLoginTime != dft {
		td.LoginDay++
		td.LastLoginTime = dft
		return true
	}
	return false
}

func (svc *Service) getThreeDayLock(ctx *ctx.Context) *errmsg.ErrMsg {
	// 不需要锁，都是在玩家自己的队列里执行的
	return nil
	// return ctx.DRLock(redisclient.GetLocker(), threeDayLocker+ctx.RoleId)
}
