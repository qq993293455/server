package activity

import (
	"time"

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

	"github.com/rs/xid"

	"go.uber.org/zap"
)

func (svc *Service) LevelGrowthFundData(ctx *ctx.Context, _ *servicepb.Activity_LevelGrowthFundDataRequest) (*servicepb.Activity_LevelGrowthFundDataResponse, *errmsg.ErrMsg) {
	data, err := dao.GetLevelGrowthFundData(ctx)
	if err != nil {
		return nil, err
	}
	list := rule.GetLevelGrowthFund(ctx)
	cfgList := make([]*models.LevelGrowthFundCfg, 0, len(list))
	for _, item := range list {
		cfgList = append(cfgList, &models.LevelGrowthFundCfg{
			Id:     item.Id,
			TypeId: item.TypId,
			Level:  item.Level,
			ActivityReward: func() []values.Integer {
				temp := make([]values.Integer, 0)
				for _, reward := range item.ActivityReward {
					temp = append(temp, reward)
				}
				return temp
			}(),
			ActivityPayReward: func() []values.Integer {
				temp := make([]values.Integer, 0)
				for _, reward := range item.ActivityPayReward {
					temp = append(temp, reward)
				}
				return temp
			}(),
		})
	}
	return &servicepb.Activity_LevelGrowthFundDataResponse{
		Buy:  data.Buy,
		Info: data.Info,
		Cfg:  cfgList,
	}, nil
}

func (svc *Service) LevelGrowthFundGetReward(ctx *ctx.Context, req *servicepb.Activity_LevelGrowthFundGetRewardRequest) (*servicepb.Activity_LevelGrowthFundGetRewardResponse, *errmsg.ErrMsg) {
	data, err := dao.GetLevelGrowthFundData(ctx)
	if err != nil {
		return nil, err
	}

	cfg, ok := rule.GetLevelGrowthFundById(ctx, req.Id)
	if !ok {
		return nil, errmsg.NewErrActivityGiftNotExist()
	}
	rewards := make(map[values.ItemId]values.Integer, 0)
	if _, ok := data.Info[req.Id]; !ok {
		data.Info[req.Id] = &models.LevelGrowthFundItem{}
	}
	// 免费奖励
	if !data.Info[req.Id].Free {
		data.Info[req.Id].Free = true
		for i := 0; i < len(cfg.ActivityReward); i += 2 {
			rewards[cfg.ActivityReward[i]] = cfg.ActivityReward[i+1]
		}
	}
	// 付费奖励
	if data.Buy && !data.Info[req.Id].Paid {
		data.Info[req.Id].Paid = true
		for i := 0; i < len(cfg.ActivityPayReward); i += 2 {
			rewards[cfg.ActivityPayReward[i]] = cfg.ActivityPayReward[i+1]
		}
	}
	if len(rewards) <= 0 {
		// 重复领取
		return nil, errmsg.NewErrEventDoNotCollectAgain()
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if role.Level < cfg.Level {
		return nil, errmsg.NewErrRoleLevelNotEnough()
	}

	if _, err := svc.AddManyItem(ctx, ctx.RoleId, rewards); err != nil {
		return nil, err
	}
	dao.SaveLevelGrowthFundData(ctx, data)
	return &servicepb.Activity_LevelGrowthFundGetRewardResponse{
		Info: data.Info,
	}, nil
}

func (svc *Service) showLevelGrowthFundEvent(ctx *ctx.Context) (bool, *errmsg.ErrMsg) {
	data, err := dao.GetLevelGrowthFundData(ctx)
	if err != nil {
		return false, err
	}
	var count int
	for _, item := range data.Info {
		if item.Free && item.Paid {
			count++
		}
	}
	total := rule.GetLevelGrowthFundLen(ctx)
	return count < total, nil
}

func (svc *Service) rebateReward(ctx *ctx.Context, data *pbdao.LevelGrowthFund) *errmsg.ErrMsg {
	if !data.Buy {
		return nil
	}
	reward, ok := rule.GetRebateReward(ctx)
	if !ok {
		ctx.Error("ActivityGrowthfundRebateReward config not found", zap.String("role_id", ctx.RoleId))
	}
	attachment := make([]*models.Item, 0)
	for id, count := range reward {
		attachment = append(attachment, &models.Item{
			ItemId: id,
			Count:  count,
		})
	}
	id := values.Integer(enum.GrowthFundRebateRewardId)
	var expiredAt values.Integer
	cfg, ok := rule.GetMailConfigTextId(ctx, id)
	if ok {
		expiredAt = timer.StartTime(ctx.StartTime).Add(time.Second * time.Duration(cfg.Overdue)).UnixMilli()
	}
	if err := svc.Add(ctx, ctx.RoleId, &models.Mail{
		Id:         xid.New().String(),
		Type:       models.MailType_MailTypeSystem,
		TextId:     id,
		ExpiredAt:  expiredAt,
		Attachment: attachment,
	}); err != nil {
		return err
	}
	dao.SaveLevelGrowthFundData(ctx, data)
	return nil
}

func (svc *Service) CheatBuyLevelGrowthFund(ctx *ctx.Context, _ *servicepb.Activity_CheatBuyLevelGrowthFundRequest) (*servicepb.Activity_CheatBuyLevelGrowthFundResponse, *errmsg.ErrMsg) {
	data, err := dao.GetLevelGrowthFundData(ctx)
	if err != nil {
		return nil, err
	}
	data.Buy = true
	if err := svc.rebateReward(ctx, data); err != nil {
		return nil, err
	}

	ctx.PushMessage(&servicepb.Activity_LevelGrowthFundBuyPush{
		Buy: true,
	})

	return &servicepb.Activity_CheatBuyLevelGrowthFundResponse{}, nil
}
