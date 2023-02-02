package activity_weekly

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	modelspb "coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/game-server/util/trans"
	"coin-server/rule"
)

func (svc *Service) GetExchangeConfigs(c *ctx.Context, activityId values.Integer) []*modelspb.WeeklyExchangeCnf {
	var ret []*modelspb.WeeklyExchangeCnf
	cfgs := rule.MustGetReader(c).ActivityWeeklyExchange.List()
	for _, cfg := range cfgs {
		if cfg.ActivityId != activityId {
			continue
		}
		ret = append(ret, &modelspb.WeeklyExchangeCnf{
			Id:            cfg.Id,
			ActivityId:    cfg.ActivityId,
			RequiredProps: cfg.RequiredProps,
			Commodities:   cfg.Commodities,
			Maximum:       cfg.Maximum,
			IsRefresh:     cfg.IsRefresh,
		})
	}
	return ret
}

func (svc *Service) refreshExchangeInfo(c *ctx.Context, aw *modelspb.ActivityWeekly) {
	// 处理每日刷新
	for _, cfg := range rule.MustGetReader(c).ActivityWeeklyExchange.List() {
		if cfg.IsRefresh != 1 {
			continue
		}
		aw.ExchangeInfo.ExchangeTimes[cfg.Id] = 0
	}
}

//ExchangeItems 兑换道具
func (svc *Service) ExchangeItems(c *ctx.Context, aw *modelspb.ActivityWeekly, id int64) ([]*modelspb.Item, *errmsg.ErrMsg) {
	cfg, ok := rule.MustGetReader(c).ActivityWeeklyExchange.GetActivityWeeklyExchangeById(id)
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyParam()
	}

	times, ok := aw.ExchangeInfo.ExchangeTimes[id]
	if ok && cfg.Maximum > 0 && times >= cfg.Maximum {
		return nil, errmsg.NewErrActivityWeeklyNoExChange()
	}

	costs := trans.ItemSliceToPb(cfg.RequiredProps)
	err := svc.SubManyItemPb(c, c.RoleId, costs...)
	if err != nil {
		return nil, err
	}

	rewards := trans.ItemSliceToPb(cfg.Commodities)
	_, err = svc.BagService.AddManyItem(c, c.RoleId, trans.ItemProtoToMap(rewards))
	if err != nil {
		return nil, err
	}

	aw.ExchangeInfo.ExchangeTimes[id]++
	return rewards, nil
}
