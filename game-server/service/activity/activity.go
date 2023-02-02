package activity

import (
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/module"
	rule2 "coin-server/game-server/service/activity/rule"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewActivityService(
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
	module.ActivityService = s
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取活动列表", svc.List)

	svc.svc.RegisterFunc("获取首充三日活动信息", svc.ThreeDayInfo)
	svc.svc.RegisterFunc("领取首充三日奖励", svc.ThreeDayGetReward)

	svc.svc.RegisterFunc("获取每日礼包活动信息", svc.DailySaleInfo)
	svc.svc.RegisterFunc("购买每日礼包", svc.DailySaleBuy)

	svc.svc.RegisterFunc("获取限时弹窗礼包列表", svc.GetLimitedTimePackages)

	svc.svc.RegisterFunc("获取等级成长基金数据", svc.LevelGrowthFundData)
	svc.svc.RegisterFunc("领取等级成长基金奖励", svc.LevelGrowthFundGetReward)

	svc.svc.RegisterFunc("获取累计充值列表", svc.AccRechargeData)
	svc.svc.RegisterFunc("领取累计充值奖励", svc.AccRechargeDraw)

	svc.svc.RegisterFunc("获取通行证列表", svc.PassesInfo)
	svc.svc.RegisterFunc("解锁通行证", svc.UnlockPasses)
	svc.svc.RegisterFunc("领取通行证奖励", svc.DrawPassesRewards)

	svc.svc.RegisterFunc("获取周礼包活动信息", svc.WeeklySaleInfo)
	svc.svc.RegisterFunc("购买周礼包", svc.WeeklySaleBuy)

	svc.svc.RegisterFunc("获取星钻商会活动信息", svc.StellargemShopData)
	svc.svc.RegisterFunc("购买星钻商会", svc.Buy)

	svc.svc.RegisterFunc("获取零元购活动信息", svc.ZeroBuyInfo)
	svc.svc.RegisterFunc("购买零元购", svc.DrawZeroBuy)
	svc.svc.RegisterFunc("零元购次日领取", svc.ZeroBuyDrawNextDay)

	eventlocal.SubscribeEventLocal(svc.HandleLoginEvent)
	eventlocal.SubscribeEventLocal(svc.HandleTargetUpdate)
	eventlocal.SubscribeEventLocal(svc.HandleDailySaleLevelUpEvent)
	eventlocal.SubscribeEventLocal(svc.HandleDailySaleRechargeEvent)
	eventlocal.SubscribeEventLocal(svc.HandleWeeklySaleLevelUpEvent)
	eventlocal.SubscribeEventLocal(svc.HandleWeeklySaleRechargeEvent)
	eventlocal.SubscribeEventLocal(svc.HandleRechargeAmountEvt)
	eventlocal.SubscribeEventLocal(svc.HandlePaySuccess)
	eventlocal.SubscribeEventLocal(svc.HandleDailyPaySuccess)
	eventlocal.SubscribeEventLocal(svc.HandlePassesPaySuccess)
	eventlocal.SubscribeEventLocal(svc.HandleLevelGrowthFundPaySuccess)
	eventlocal.SubscribeEventLocal(svc.HandleLimitedTimePackPaySuccess)
	eventlocal.SubscribeEventLocal(svc.HandleSystemUnlock)

	// 作弊器
	svc.svc.RegisterFunc("作弊器购买限时弹窗礼包", svc.CheatBuyLimitedTimePackage)
	svc.svc.RegisterFunc("购买付费内容", svc.CheatBuyLevelGrowthFund)
	svc.svc.RegisterFunc("作弊累计充值", svc.CheatRechargeRequest)
	svc.svc.RegisterFunc("作弊清楚星钻商会次数", svc.CheatClearStellargemShop)
	svc.svc.RegisterFunc("作弊器解锁高级通行证", svc.CheatUnlockAdvancePasses)
	svc.svc.RegisterFunc("作弊器加通行证经验", svc.CheatAddPassesExp)
	svc.svc.RegisterFunc("作弊器刷新每周礼包", svc.CheatRefreshWeekly)
	svc.svc.RegisterFunc("作弊清除零元购购买记录", svc.CheatClearZeroBuy)
	svc.svc.RegisterFunc("作弊零元购前移一天", svc.CheatZeroBuyAheadOneDay)
}

func (svc *Service) List(ctx *ctx.Context, _ *servicepb.Activity_ActivityListRequest) (*servicepb.Activity_ActivityListResponse, *errmsg.ErrMsg) {
	// out := &activity_service.ActivityService_ActivityListResponse{}
	// if err := svc.svc.GetNatsClient().RequestWithOut(ctx, 0, &activity_service.ActivityService_ActivityListRequest{}, out); err != nil {
	// 	return nil, err
	// }
	registerTime, err := svc.GetRegisterTime(ctx)
	if err != nil {
		return nil, err
	}
	activated, _ := rule2.GetAllAvailableActivity(ctx, registerTime)
	showThreeDay, err := svc.showThreeDayEvent(ctx)
	if err != nil {
		return nil, err
	}
	showLevelGrowthFund, err := svc.showLevelGrowthFundEvent(ctx)
	if err != nil {
		return nil, err
	}
	showAccRecharge := svc.showAccRecharge(ctx)
	list := make([]*models.Activity, 0)
	for _, item := range activated {
		if item.Id == enum.ThreeDay && !showThreeDay {
			continue
		}
		if item.Id == enum.SevenDay && svc.IsSevenDaysReceiveAll(ctx) {
			continue
		}
		if item.Id == enum.LevelGrowthFund && !showLevelGrowthFund {
			continue
		}
		if item.Id == enum.AccRecharge && !showAccRecharge {
			continue
		}
		if item.Id == enum.ZeroPurchase {
			isZeroActive, zeroStart, zeroEnd := svc.isZeroBuyActive(ctx)
			if !isZeroActive {
				continue
			} else {
				item.Begin = zeroStart
				item.End = zeroEnd
			}
		}
		list = append(list, item)
	}
	return &servicepb.Activity_ActivityListResponse{
		List: list,
	}, nil
}

func (svc *Service) GetRegisterTime(ctx *ctx.Context) (values.Integer, *errmsg.ErrMsg) {
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return 0, err
	}
	registerTime := time.Unix(role.CreateTime/1000, 0).UTC()
	refresh := svc.RefreshService.GetCurrDayFreshTimeWith(ctx, registerTime)

	return refresh.Unix(), nil
}
