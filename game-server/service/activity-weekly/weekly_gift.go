package activity_weekly

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	modelspb "coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/game-server/util/trans"
	"coin-server/rule"
)

func (svc *Service) GetGiftConfigs(c *ctx.Context, activityId values.Integer) []*modelspb.WeeklyGiftCnf {
	var ret []*modelspb.WeeklyGiftCnf
	cfgs := rule.MustGetReader(c).ActivityWeeklyGift.List()
	for _, cfg := range cfgs {
		if cfg.ActivityId != activityId {
			continue
		}
		ret = append(ret, &modelspb.WeeklyGiftCnf{
			Id:               cfg.Id,
			Language1:        cfg.Language1,
			ActivityId:       cfg.ActivityId,
			GridId:           cfg.GridId,
			GradeRange:       cfg.GradeRange,
			RechargeInterval: cfg.RechargeInterval,
			BuyType:          cfg.BuyType,
			PayId:            cfg.PayId,
			PayNum:           cfg.PayNum,
			IsRefresh:        cfg.IsRefresh,
			Reward:           cfg.Reward,
			ItemPictureID:    cfg.ItemPictureId,
		})
	}
	return ret
}

func (svc *Service) refreshGiftInfo(c *ctx.Context, aw *modelspb.ActivityWeekly) {
	// 处理每日刷新
	for _, cfg := range rule.MustGetReader(c).ActivityWeeklyGift.List() {
		if cfg.IsRefresh != 1 {
			continue
		}
		aw.GiftInfo.BuyTimes[cfg.Id] = 0
	}
}

//BuyGiftItems 道具购买
func (svc *Service) BuyGiftItems(c *ctx.Context, aw *modelspb.ActivityWeekly, id int64) *errmsg.ErrMsg {
	cfg, ok := rule.MustGetReader(c).ActivityWeeklyGift.GetActivityWeeklyGiftById(id)
	if !ok {
		return errmsg.NewErrActivityWeeklyParam()
	}
	if cfg.BuyType != 1 { // 只处理道具购买
		return errmsg.NewErrActivityWeeklyParam()
	}

	role, err := svc.UserService.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return err
	}
	if !utils.InRange(cfg.GradeRange, role.Level) { // 判断是否符合等级区间
		return errmsg.NewErrActivityWeeklyParam()
	}

	if cfg.PayNum > 0 && aw.GiftInfo.BuyTimes[id] >= cfg.PayNum { // 售罄
		return errmsg.NewErrActivityWeeklyNoGift()
	}

	if cfg.PayId[0] != 0 { // 判断是否是直接领取的
		cost := trans.ItemSliceToMap(cfg.PayId)
		err = svc.BagService.SubManyItem(c, c.RoleId, cost)
		if err != nil {
			return err
		}
	}

	rewards := trans.ItemSliceToPb(cfg.Reward)
	_, err = svc.BagService.AddManyItem(c, c.RoleId, trans.ItemProtoToMap(rewards))
	if err != nil {
		return err
	}

	aw.GiftInfo.BuyTimes[id]++
	c.PushMessage(&servicepb.ActivityWeekly_BuyGiftPush{Id: id, Items: rewards})
	return nil
}

func (svc *Service) buyGiftItemsByCash(c *ctx.Context, aw *modelspb.ActivityWeekly, id int64) *errmsg.ErrMsg {
	cfg, ok := rule.MustGetReader(c).ActivityWeeklyGift.GetActivityWeeklyGiftById(id)
	if !ok {
		return errmsg.NewErrActivityWeeklyParam()
	}

	if cfg.PayNum > 0 && aw.GiftInfo.BuyTimes[id] >= cfg.PayNum { // 售罄
		return errmsg.NewErrActivityWeeklyNoGift()
	}

	rewards := trans.ItemSliceToPb(cfg.Reward)
	_, err := svc.BagService.AddManyItem(c, c.RoleId, trans.ItemProtoToMap(rewards))
	if err != nil {
		return err
	}

	aw.GiftInfo.BuyTimes[id]++
	c.PushMessage(&servicepb.ActivityWeekly_BuyGiftPush{Id: id, Items: rewards})
	return nil
}
