package activity

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	protosvc "coin-server/common/proto/service"
	"coin-server/common/values"
	"coin-server/game-server/service/activity/dao"
	rule2 "coin-server/game-server/service/activity/rule"
	"coin-server/rule"
)

func (svc *Service) StellargemShopData(c *ctx.Context, _ *protosvc.Activity_StellargemShopDataRequest) (*protosvc.Activity_StellargemShopDataResponse, *errmsg.ErrMsg) {
	data, err := dao.GetStellargemShop(c)
	if err != nil {
		return nil, err
	}
	list := rule.MustGetReader(c).ActivityStellargemShopmall.List()
	cfg := make([]*models.StellargemShopCfg, 0, len(list))
	for _, v := range list {
		cfg = append(cfg, &models.StellargemShopCfg{
			Id:                 v.Id,
			StellarGemNum:      v.StellarGemNum,
			SourceGemNum:       v.SourceGemNum,
			ExtraGemNum:        v.ExtraGemNum,
			GiftLanguage:       v.GiftLanguage,
			BuyGiftLanguage:    v.BuyGiftLanguage,
			StellarGemIconName: v.StellarGemIconName,
			PurchasePrice:      v.PurchasePrice,
			Tags:               v.Tags,
			GiftTags:           v.GiftTags,
		})
	}
	return &protosvc.Activity_StellargemShopDataResponse{
		IsDoubleOpen: true,
		BuyCnt:       data.BuyCnt,
		Cfg:          cfg,
	}, nil
}

// TODO 这个目前在PC上当作弊器用，后期上线需要去掉
func (svc *Service) Buy(c *ctx.Context, req *protosvc.Activity_StellargemShopBuyRequest) (*protosvc.Activity_StellargemShopBuyResponse, *errmsg.ErrMsg) {
	cfg, ok := rule.MustGetReader(c).ActivityStellargemShopmall.GetActivityStellargemShopmallById(req.Idx)
	if !ok {
		return nil, errmsg.NewInternalErr("activity_stellargem_shopmall config not found")
	}
	item, id, err := svc.StellargemShopBuy(c, cfg.PurchasePrice)
	if err != nil {
		return nil, err
	}
	c.PushMessage(&protosvc.Pay_NormalSuccessPush{
		PcId:         cfg.PurchasePrice,
		ShopNormalId: id,
		Items:        item,
	})
	return &protosvc.Activity_StellargemShopBuyResponse{}, nil
}

func (svc *Service) StellargemShopBuy(c *ctx.Context, id values.Integer) (map[values.ItemId]values.Integer, values.Integer, *errmsg.ErrMsg) {
	data, err := dao.GetStellargemShop(c)
	if err != nil {
		return nil, 0, err
	}
	cfg, ok := rule2.GetStellargemShopmallByPurchasePrice(c, id)
	if !ok {
		return nil, 0, errmsg.NewErrActivityGiftNotExist()
	}
	item := map[int64]int64{}
	for i := 0; i < len(cfg.StellarGemNum); i += 2 {
		item[cfg.StellarGemNum[i]] += cfg.StellarGemNum[i+1]
	}
	if data.BuyCnt[cfg.Id] == 0 {
		for i := 0; i < len(cfg.SourceGemNum); i += 2 {
			item[cfg.SourceGemNum[i]] += cfg.SourceGemNum[i+1]
		}
	} else {
		for i := 0; i < len(cfg.ExtraGemNum); i += 2 {
			item[cfg.ExtraGemNum[i]] += cfg.ExtraGemNum[i+1]
		}
	}
	if len(item) > 0 {
		if _, err := svc.AddManyItem(c, c.RoleId, item); err != nil {
			return nil, 0, err
		}
	}
	data.BuyCnt[cfg.Id]++
	dao.SaveStellargemShop(c, data)

	return item, cfg.Id, nil
}

func (svc *Service) CheatClearStellargemShop(c *ctx.Context, req *protosvc.Activity_CheatClearStellargemCntRequest) (*protosvc.Activity_CheatClearStellargemCntResponse, *errmsg.ErrMsg) {
	data, err := dao.GetStellargemShop(c)
	if err != nil {
		return nil, err
	}
	data.BuyCnt = map[int64]int64{}
	dao.SaveStellargemShop(c, data)
	return &protosvc.Activity_CheatClearStellargemCntResponse{}, nil
}
