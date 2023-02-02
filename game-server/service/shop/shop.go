package shop

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	protosvc "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/module"
	"coin-server/game-server/service/shop/dao"
	rulemodel "coin-server/rule/rule-model"
)

const (
	RefFre      = "ShopRefFre"
	RefFreeNum  = "ShopRefFreeNum"
	RefCost     = "ShopRefCost"
	RefCostItem = "ShopRefCostItem"
	RefTime     = "ShopRefTime"

	RefFreAnecdotes     = "ShopRefFreAnecdotes"
	RefFreeNumAnecdotes = "ShopRefFreeNumAnecdotes"
	RefCostAnecdotes    = "ShopRefCostAnecdotes"
	RefTimeAnecdotes    = "ShopRefTimeAnecdotes"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
}

func NewShopService(
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
	return s
}

// ---------------------------------------------------proto------------------------------------------------------------//

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取商店", svc.GetShop)
	svc.svc.RegisterFunc("购买物品", svc.Buy)
	svc.svc.RegisterFunc("刷新商店", svc.Refresh)

	svc.svc.RegisterFunc("获取竞技场商店列表", svc.GetArenaShopList)
	svc.svc.RegisterFunc("购买竞技场商店", svc.BuyArena)

	svc.svc.RegisterFunc("获取公会商店列表", svc.GetGuildShopList)
	svc.svc.RegisterFunc("购买公会商店", svc.BuyGuild)

	svc.svc.RegisterFunc("获取势力商店列表", svc.GetCampShopList)
	svc.svc.RegisterFunc("购买势力商店", svc.BuyCamp)
	svc.svc.RegisterFunc("刷新势力商店", svc.RefreshCamp)

	svc.svc.RegisterFunc("作弊清除每日更新次数", svc.CheatClearRefreshCnt)
	svc.svc.RegisterFunc("作弊刷新", svc.CheatRefresh)
}

func (svc *Service) GetShop(ctx *ctx.Context, _ *protosvc.Shop_GetShopListRequest) (*protosvc.Shop_GetShopListResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetShop(ctx)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(ctx)
	if shopInfo.Data.CheckUpdate(ctx, tbg.Unix()) {
		lvl, err := svc.UserService.GetLevel(ctx, ctx.RoleId)
		if err != nil {
			return nil, err
		}
		err = shopInfo.Data.AutoUpdate(ctx, lvl, tbg.Unix())
		if err != nil {
			return nil, err
		}
		dao.SaveShop(ctx, shopInfo)
	}
	refTime, _ := rulemodel.GetReader().KeyValue.GetInt64(RefTime)
	gap := refTime * 60 * 60
	next := shopInfo.Data.LastRefreshAt() + gap

	return &protosvc.Shop_GetShopListResponse{
		TodayRefreshCnt: shopInfo.Data.TodayRefreshCnt(),
		NextRefreshAt:   next,
		Detail:          shopInfo.Data.GetDetail(),
	}, nil
}

func (svc *Service) Buy(ctx *ctx.Context, req *protosvc.Shop_BuyRequest) (*protosvc.Shop_BuyResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetShop(ctx)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(ctx)
	if shopInfo.Data.CheckUpdate(ctx, tbg.Unix()) {
		return nil, errmsg.NewErrShopAlreadyRefreshed()
	}
	detail, err := shopInfo.Data.Buy(req.DetailIdx, req.ShowTyp)
	if err != nil {
		return nil, err
	}
	err = svc.BagService.SubManyItem(ctx, ctx.RoleId, detail.Cost)
	if err != nil {
		return nil, err
	}
	err = svc.BagService.AddManyItemPb(ctx, ctx.RoleId, detail.Good)
	if err != nil {
		return nil, err
	}
	dao.SaveShop(ctx, shopInfo)
	// 购买道具打点
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskBuyShopItemNum: {
			Typ:     values.Integer(models.TaskType_TaskBuyShopItemNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskBuyShopItemNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskBuyShopItemNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	svc.TaskService.UpdateTargets(ctx, ctx.RoleId, tasks)
	refTime, _ := rulemodel.GetReader().KeyValue.GetInt64(RefTime)
	gap := refTime * 60 * 60
	next := shopInfo.Data.LastRefreshAt() + gap
	return &protosvc.Shop_BuyResponse{
		TodayRefreshCnt: shopInfo.Data.TodayRefreshCnt(),
		NextRefreshAt:   next,
		Detail:          shopInfo.Data.GetDetail(),
	}, nil
}

func (svc *Service) Refresh(ctx *ctx.Context, _ *protosvc.Shop_RefreshRequest) (*protosvc.Shop_RefreshResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetShop(ctx)
	if err != nil {
		return nil, err
	}
	lvl, err := svc.UserService.GetLevel(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(ctx)
	if shopInfo.Data.CheckUpdate(ctx, tbg.Unix()) {
		err = shopInfo.Data.AutoUpdate(ctx, lvl, tbg.Unix())
		if err != nil {
			return nil, err
		}
	}
	cnt := shopInfo.Data.TodayRefreshCnt()
	totalCnt, _ := rulemodel.GetReader().KeyValue.GetInt64(RefFre)
	freeCnt, _ := rulemodel.GetReader().KeyValue.GetInt64(RefFreeNum)
	totalCnt += freeCnt
	if cnt >= totalCnt {
		return nil, errmsg.NewErrHaveNoRefreshCnt()
	}
	if cnt >= freeCnt {
		cardItem, has := rulemodel.GetReader().KeyValue.GetItem(RefCostItem)
		if !has {
			return nil, errmsg.NewErrBagNoSuchItem()
		}
		if err = svc.Module.SubItem(ctx, ctx.RoleId, cardItem.ItemId, cardItem.Count); err != nil {
			// 重置卡不够
			cost, _ := rulemodel.GetReader().KeyValue.GetInt64(RefCost)
			if err = svc.Module.SubItem(ctx, ctx.RoleId, enum.BoundDiamond, cost); err != nil {
				return nil, err
			}
		}
	}
	if err = shopInfo.Data.Refresh(ctx, lvl); err != nil {
		return nil, err
	}
	dao.SaveShop(ctx, shopInfo)

	refTime, _ := rulemodel.GetReader().KeyValue.GetInt64(RefTime)
	gap := refTime * 60 * 60
	next := shopInfo.Data.LastRefreshAt() + gap

	return &protosvc.Shop_RefreshResponse{
		TodayRefreshCnt: shopInfo.Data.TodayRefreshCnt(),
		NextRefreshAt:   next,
		Detail:          shopInfo.Data.GetDetail(),
	}, nil
}

// ---------------------------------------------------arena------------------------------------------------------------//

func (svc *Service) GetArenaShopList(c *ctx.Context, _ *protosvc.Shop_GetAreaShopListRequest) (*protosvc.Shop_GetAreaShopListResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetArenaShop(c)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(c)
	if shopInfo.Data.CheckUpdate(c, tbg.Unix()) {
		lvl, err := svc.UserService.GetLevel(c, c.RoleId)
		if err != nil {
			return nil, err
		}
		err = shopInfo.Data.AutoUpdate(c, lvl, tbg.Unix())
		if err != nil {
			return nil, err
		}
		dao.SaveArenaShop(c, shopInfo)
	}
	refTime, _ := rulemodel.GetReader().KeyValue.GetInt64(RefTime)
	gap := refTime * 60 * 60
	next := shopInfo.Data.LastRefreshAt() + gap

	return &protosvc.Shop_GetAreaShopListResponse{
		NxtDailyRefreshAt: next,
		DailyList:         shopInfo.Data.GetDetail(),
	}, nil
}

func (svc *Service) BuyArena(c *ctx.Context, req *protosvc.Shop_BuyAreaRequest) (*protosvc.Shop_BuyAreaResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetArenaShop(c)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(c)
	if shopInfo.Data.CheckUpdate(c, tbg.Unix()) {
		return nil, errmsg.NewErrShopAlreadyRefreshed()
	}
	detail, err := shopInfo.Data.Buy(req.DetailIdx, req.ShowTyp)
	if err != nil {
		return nil, err
	}
	err = svc.BagService.SubManyItem(c, c.RoleId, detail.Cost)
	if err != nil {
		return nil, err
	}
	err = svc.BagService.AddManyItemPb(c, c.RoleId, detail.Good)
	if err != nil {
		return nil, err
	}
	dao.SaveArenaShop(c, shopInfo)
	// 购买道具打点
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskBuyShopItemNum: {
			Typ:     values.Integer(models.TaskType_TaskBuyShopItemNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskBuyShopItemNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskBuyShopItemNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	svc.TaskService.UpdateTargets(c, c.RoleId, tasks)
	refTime, _ := rulemodel.GetReader().KeyValue.GetInt64(RefTime)
	gap := refTime * 60 * 60
	next := timer.BeginOfDay(timer.StartTime(c.StartTime)).Unix() + gap
	return &protosvc.Shop_BuyAreaResponse{
		NxtDailyRefreshAt: next,
		DailyList:         shopInfo.Data.GetDetail(),
	}, nil
}

// ---------------------------------------------------guild------------------------------------------------------------//

func (svc *Service) GetGuildShopList(c *ctx.Context, _ *protosvc.Shop_GetGuildShopListRequest) (*protosvc.Shop_GetGuildShopListResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetGuildShop(c)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(c)
	if shopInfo.Data.CheckUpdate(c, tbg.Unix()) {
		lvl, err := svc.UserService.GetLevel(c, c.RoleId)
		if err != nil {
			return nil, err
		}
		err = shopInfo.Data.AutoUpdate(c, lvl, tbg.Unix())
		if err != nil {
			return nil, err
		}
		dao.SaveGuildShop(c, shopInfo)
	}
	refTime, _ := rulemodel.GetReader().KeyValue.GetInt64(RefTime)
	gap := refTime * 60 * 60
	next := shopInfo.Data.LastRefreshAt() + gap
	return &protosvc.Shop_GetGuildShopListResponse{
		NextRefreshAt: next,
		Detail:        shopInfo.Data.GetDetail(),
	}, nil
}

func (svc *Service) BuyGuild(c *ctx.Context, req *protosvc.Shop_BuyGuildRequest) (*protosvc.Shop_BuyGuildResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetGuildShop(c)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(c)
	if shopInfo.Data.CheckUpdate(c, tbg.Unix()) {
		return nil, errmsg.NewErrShopAlreadyRefreshed()
	}
	detail, err := shopInfo.Data.Buy(req.DetailIdx, req.ShowTyp)
	if err != nil {
		return nil, err
	}
	err = svc.BagService.SubManyItem(c, c.RoleId, detail.Cost)
	if err != nil {
		return nil, err
	}
	err = svc.BagService.AddManyItemPb(c, c.RoleId, detail.Good)
	if err != nil {
		return nil, err
	}
	dao.SaveGuildShop(c, shopInfo)
	// 购买道具打点
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskBuyShopItemNum: {
			Typ:     values.Integer(models.TaskType_TaskBuyShopItemNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskBuyShopItemNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskBuyShopItemNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	svc.TaskService.UpdateTargets(c, c.RoleId, tasks)
	refTime, _ := rulemodel.GetReader().KeyValue.GetInt64(RefTime)
	gap := refTime * 60 * 60
	next := shopInfo.Data.LastRefreshAt() + gap
	return &protosvc.Shop_BuyGuildResponse{
		NextRefreshAt: next,
		Detail:        shopInfo.Data.GetDetail(),
	}, nil
}

// ---------------------------------------------------camp------------------------------------------------------------//

func (svc *Service) GetCampShopList(c *ctx.Context, _ *protosvc.Shop_GetCampShopListRequest) (*protosvc.Shop_GetCampShopListResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetCampShop(c)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(c)
	if shopInfo.Data.CheckUpdate(c, tbg.Unix()) {
		lvl, err := svc.UserService.GetLevel(c, c.RoleId)
		if err != nil {
			return nil, err
		}
		err = shopInfo.Data.AutoUpdate(c, lvl, tbg.Unix())
		if err != nil {
			return nil, err
		}
		dao.SaveCampShop(c, shopInfo)
	}
	refTime, _ := rulemodel.GetReader().KeyValue.GetInt64(RefTimeAnecdotes)
	gap := refTime * 60 * 60
	next := shopInfo.Data.LastRefreshAt() + gap
	return &protosvc.Shop_GetCampShopListResponse{
		NextRefreshAt:   next,
		TodayRefreshCnt: shopInfo.Data.TodayRefreshCnt(),
		Detail:          shopInfo.Data.GetDetail(),
	}, nil
}

func (svc *Service) BuyCamp(c *ctx.Context, req *protosvc.Shop_BuyCampRequest) (*protosvc.Shop_BuyCampResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetCampShop(c)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(c)
	if shopInfo.Data.CheckUpdate(c, tbg.Unix()) {
		return nil, errmsg.NewErrShopAlreadyRefreshed()
	}
	detail, err := shopInfo.Data.Buy(req.DetailIdx, req.ShowTyp)
	if err != nil {
		return nil, err
	}
	err = svc.BagService.SubManyItem(c, c.RoleId, detail.Cost)
	if err != nil {
		return nil, err
	}
	err = svc.BagService.AddManyItemPb(c, c.RoleId, detail.Good)
	if err != nil {
		return nil, err
	}
	dao.SaveCampShop(c, shopInfo)
	// 购买道具打点
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskBuyShopItemNum: {
			Typ:     values.Integer(models.TaskType_TaskBuyShopItemNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskBuyShopItemNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskBuyShopItemNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	svc.TaskService.UpdateTargets(c, c.RoleId, tasks)
	refTime, _ := rulemodel.GetReader().KeyValue.GetInt64(RefTimeAnecdotes)
	gap := refTime * 60 * 60
	next := shopInfo.Data.LastRefreshAt() + gap
	return &protosvc.Shop_BuyCampResponse{
		NextRefreshAt: next,
		Detail:        shopInfo.Data.GetDetail(),
	}, nil
}

func (svc *Service) RefreshCamp(ctx *ctx.Context, _ *protosvc.Shop_RefreshCampRequest) (*protosvc.Shop_RefreshCampResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetCampShop(ctx)
	if err != nil {
		return nil, err
	}
	lvl, err := svc.UserService.GetLevel(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(ctx)
	if shopInfo.Data.CheckUpdate(ctx, tbg.Unix()) {
		err = shopInfo.Data.AutoUpdate(ctx, lvl, tbg.Unix())
		if err != nil {
			return nil, err
		}
	}
	cnt := shopInfo.Data.TodayRefreshCnt()
	totalCnt, _ := rulemodel.GetReader().KeyValue.GetInt64(RefFreAnecdotes)
	freeCnt, _ := rulemodel.GetReader().KeyValue.GetInt64(RefFreeNumAnecdotes)
	totalCnt += freeCnt
	if cnt >= totalCnt {
		return nil, errmsg.NewErrHaveNoRefreshCnt()
	}
	if cnt >= freeCnt {
		cost, _ := rulemodel.GetReader().KeyValue.GetItem(RefCostAnecdotes)
		if err = svc.Module.SubItem(ctx, ctx.RoleId, cost.ItemId, cost.Count); err != nil {
			return nil, err
		}
	}
	if err = shopInfo.Data.Refresh(ctx, lvl); err != nil {
		return nil, err
	}
	dao.SaveCampShop(ctx, shopInfo)

	refTime, _ := rulemodel.GetReader().KeyValue.GetInt64(RefTimeAnecdotes)
	gap := refTime * 60 * 60
	next := shopInfo.Data.LastRefreshAt() + gap

	return &protosvc.Shop_RefreshCampResponse{
		TodayRefreshCnt: shopInfo.Data.TodayRefreshCnt(),
		NextRefreshAt:   next,
		Detail:          shopInfo.Data.GetDetail(),
	}, nil
}

// ---------------------------------------------------cheat------------------------------------------------------------//

func (svc *Service) CheatClearRefreshCnt(ctx *ctx.Context, _ *protosvc.Shop_CheatClearRefreshCntRequest) (*protosvc.Shop_CheatClearRefreshCntResponse, *errmsg.ErrMsg) {
	shopInfo, err := dao.GetShop(ctx)
	if err != nil {
		return nil, err
	}
	tbg := svc.GetCurrDayFreshTime(ctx)
	shopInfo.Data.CheatClearRefreshCnt(ctx, tbg.Unix())
	dao.SaveShop(ctx, shopInfo)
	return &protosvc.Shop_CheatClearRefreshCntResponse{}, nil
}

func (svc *Service) CheatRefresh(ctx *ctx.Context, req *protosvc.Shop_CheatRefreshRequest) (*protosvc.Shop_CheatRefreshResponse, *errmsg.ErrMsg) {
	if req.Type == "" {
		shopInfo, err := dao.GetShop(ctx)
		if err != nil {
			return nil, err
		}
		lvl, err := svc.UserService.GetLevel(ctx, ctx.RoleId)
		if err != nil {
			return nil, err
		}
		err = shopInfo.Data.CheatRefresh(lvl)
		if err != nil {
			return nil, err
		}
		dao.SaveShop(ctx, shopInfo)
	} else if req.Type == "arena" {
		shopInfo, err := dao.GetArenaShop(ctx)
		if err != nil {
			return nil, err
		}
		lvl, err := svc.UserService.GetLevel(ctx, ctx.RoleId)
		if err != nil {
			return nil, err
		}
		err = shopInfo.Data.CheatRefresh(lvl)
		if err != nil {
			return nil, err
		}
		dao.SaveArenaShop(ctx, shopInfo)
	}
	return &protosvc.Shop_CheatRefreshResponse{}, nil
}
