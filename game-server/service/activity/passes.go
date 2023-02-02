package activity

import (
	"math"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	daopb "coin-server/common/proto/dao"
	modelspb "coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/timer"
	"coin-server/common/utils/generic/maps"
	"coin-server/common/values"
	"coin-server/game-server/service/activity/dao"
	"coin-server/game-server/util"
	"coin-server/rule"
)

func (svc *Service) genPassesRewardCfgs(c *ctx.Context) []*modelspb.PassesReward {
	cfgs := rule.MustGetReader(c).ActivityPassesAwards.List()
	list := make([]*modelspb.PassesReward, 0, len(cfgs))
	for _, cfg := range cfgs {
		if len(cfg.GradeRange) < 2 {
			panic("ActivityPassesAwards config GradeRange len < 2: " + strconv.Itoa(int(cfg.Id)))
		}
		list = append(list, &modelspb.PassesReward{
			Id:          cfg.Id,
			Lv:          cfg.PassesLevel,
			RewardType:  cfg.RewardType,
			FreeReward:  cfg.ActivityReward,
			PaidReward:  cfg.ActivityPayReward,
			RequiredExp: cfg.PassesLevelExp,
			LvRange:     cfg.GradeRange,
		})
	}
	return list
}

func (svc *Service) getPassesReset(c *ctx.Context) (*daopb.PassesReset, *errmsg.ErrMsg) {
	reset, err := dao.GetPassesReset(c)
	if err != nil {
		return nil, err
	}

	if reset == nil {
		reset = &daopb.PassesReset{
			Key:        dao.GlobalPassesResetKey,
			ResetAt:    util.DefaultCurWeekRefreshTime().AddDate(0, 0, 14).UnixMilli(),
			RewardCfgs: svc.genPassesRewardCfgs(c),
		}
		dao.SavePassesReset(c, reset)
	}

	if timer.Now().UnixMilli() >= reset.ResetAt {
		reset.ResetAt = util.DefaultCurWeekRefreshTime().AddDate(0, 0, 14).UnixMilli()
		reset.RewardCfgs = svc.genPassesRewardCfgs(c)
		dao.SavePassesReset(c, reset)
	}

	return reset, nil
}

func (svc *Service) getPassesRewardCfgsByLv(c *ctx.Context, reset *daopb.PassesReset, roleLv values.Level) []*modelspb.PassesReward {
	ret := make([]*modelspb.PassesReward, 0, 36)
	for _, cfg := range reset.RewardCfgs {
		if cfg.LvRange[1] == -1 {
			cfg.LvRange[1] = math.MaxInt64
		}
		if roleLv >= cfg.LvRange[0] && roleLv <= cfg.LvRange[1] {
			ret = append(ret, cfg)
		}
	}
	return ret
}

func (svc *Service) getPasses(c *ctx.Context) (*daopb.Passes, *daopb.PassesReset, *errmsg.ErrMsg) {
	reset, err := svc.getPassesReset(c)
	if err != nil {
		return nil, nil, err
	}

	passes, err := dao.GetPasses(c)
	if err != nil {
		return nil, nil, err
	}

	if passes == nil {
		role, err2 := svc.UserService.GetRoleByRoleId(c, c.RoleId)
		if err2 != nil {
			return nil, nil, err2
		}
		passes = &daopb.Passes{
			RoleId: c.RoleId,
			Data: &modelspb.PassesData{
				Lv:         1,
				Exp:        0,
				Status:     0,
				FreeUnlock: map[int64]modelspb.RewardStatus{1: modelspb.RewardStatus_Unlocked},
				PaidUnlock: map[int64]modelspb.RewardStatus{1: modelspb.RewardStatus_Unlocked},
				ResetAt:    reset.ResetAt,
			},
			RoleLv: role.Level,
		}
		dao.SavePasses(c, passes)
	}

	if passes.Data.ResetAt < reset.ResetAt {
		// 未领取的发邮件
		err = svc.sendPassesMail(c, passes, reset)
		if err != nil {
			return nil, nil, err
		}

		role, err := svc.UserService.GetRoleByRoleId(c, c.RoleId)
		if err != nil {
			return nil, nil, err
		}
		passes.Data.Lv = 1
		passes.Data.Exp = 0
		passes.Data.Status = 0
		passes.Data.FreeUnlock = map[int64]modelspb.RewardStatus{1: modelspb.RewardStatus_Unlocked}
		passes.Data.PaidUnlock = map[int64]modelspb.RewardStatus{1: modelspb.RewardStatus_Unlocked}
		passes.Data.ResetAt = reset.ResetAt
		passes.RoleLv = role.Level
		dao.SavePasses(c, passes)
	}

	return passes, reset, nil
}

func (svc *Service) sendPassesMail(c *ctx.Context, passes *daopb.Passes, reset *daopb.PassesReset) *errmsg.ErrMsg {
	rewards := svc.drawPassesAllReward(c, passes, reset)
	items := make([]*modelspb.Item, 0, len(rewards))

	for id, cnt := range rewards {
		items = append(items, &modelspb.Item{ItemId: id, Count: cnt})
	}

	mail := &modelspb.Mail{
		Type:       modelspb.MailType_MailTypeSystem,
		TextId:     100021,
		Attachment: items,
	}
	err := svc.MailService.Add(c, c.RoleId, mail)
	if err != nil {
		return err
	}

	return nil
}

func (svc *Service) nextPassesLvCfg(c *ctx.Context, passes *daopb.Passes, reset *daopb.PassesReset) *modelspb.PassesReward {
	cfgs := svc.getPassesRewardCfgsByLv(c, reset, passes.RoleLv)
	max := cfgs[len(cfgs)-1]
	if passes.Data.Lv == max.Lv {
		return nil
	}
	for _, cfg := range cfgs {
		if cfg.Lv > passes.Data.Lv {
			return cfg
		}
	}
	return nil
}

func (svc *Service) addPassesExp(c *ctx.Context, incr values.Integer) *errmsg.ErrMsg {
	passes, reset, err := svc.getPasses(c)
	if err != nil {
		return err
	}
	old := passes.Data.Exp
	passes.Data.Exp += incr
	for {
		nextLv := svc.nextPassesLvCfg(c, passes, reset)
		if nextLv == nil {
			break
		}
		if old < nextLv.RequiredExp && passes.Data.Exp >= nextLv.RequiredExp {
			passes.Data.Lv = nextLv.Lv
			passes.Data.FreeUnlock[nextLv.Lv] = modelspb.RewardStatus_Unlocked
			passes.Data.PaidUnlock[nextLv.Lv] = modelspb.RewardStatus_Unlocked
		} else {
			break
		}
	}
	dao.SavePasses(c, passes)
	c.PushMessage(&servicepb.Activity_AddPassesExpPush{
		Lv:  passes.Data.Lv,
		Exp: passes.Data.Exp,
	})
	return nil
}

func (svc *Service) PassesInfo(c *ctx.Context, _ *servicepb.Activity_PassesInfoRequest) (*servicepb.Activity_PassesInfoResponse, *errmsg.ErrMsg) {
	role, err := svc.UserService.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	passes, reset, err := svc.getPasses(c)
	if err != nil {
		return nil, err
	}

	cfgs := svc.getPassesRewardCfgsByLv(c, reset, role.Level)
	price, ok := rule.MustGetReader(c).KeyValue.GetInt64("PremiumPasssPrice")
	if !ok {
		panic("KeyValue PremiumPasssPrice notfound")
	}

	return &servicepb.Activity_PassesInfoResponse{
		Passes: &modelspb.Passes{
			Data:         passes.Data,
			RewardCfgs:   cfgs,
			AdvancePrice: price,
		},
	}, nil
}

func (svc *Service) UnlockPasses(c *ctx.Context, _ *servicepb.Activity_UnlockPassesRequest) (*servicepb.Activity_UnlockPassesResponse, *errmsg.ErrMsg) {
	passes, _, err := svc.getPasses(c)
	if err != nil {
		return nil, err
	}
	if passes.Data.Status == modelspb.PassesStatus_PassesStatusLocked {
		passes.Data.Status = modelspb.PassesStatus_PassesStatusGeneral
		dao.SavePasses(c, passes)
	}
	return &servicepb.Activity_UnlockPassesResponse{Status: passes.Data.Status}, nil
}

func (svc *Service) drawPassesAllReward(c *ctx.Context, passes *daopb.Passes, reset *daopb.PassesReset) map[values.Integer]values.Integer {
	rewards := make(map[values.Integer]values.Integer)

	cfgs := svc.getPassesRewardCfgsByLv(c, reset, passes.RoleLv)
	for _, cfg := range cfgs {
		lv := cfg.Lv
		status := passes.Data.FreeUnlock[lv]
		// 普通奖励
		if status == modelspb.RewardStatus_Unlocked {
			maps.Merge(rewards, cfg.FreeReward)
			passes.Data.FreeUnlock[lv] = modelspb.RewardStatus_Received
		}
		// 高级奖励
		if passes.Data.Status == modelspb.PassesStatus_PassesStatusAdvanced {
			status := passes.Data.PaidUnlock[lv]
			if status == modelspb.RewardStatus_Unlocked {
				maps.Merge(rewards, cfg.PaidReward)
				passes.Data.PaidUnlock[lv] = modelspb.RewardStatus_Received
			}
		}
	}
	return rewards
}

func (svc *Service) DrawPassesRewards(c *ctx.Context, req *servicepb.Activity_DrawPassesRewardRequest) (*servicepb.Activity_DrawPassesRewardResponse, *errmsg.ErrMsg) {
	passes, reset, err := svc.getPasses(c)
	if err != nil {
		return nil, err
	}
	if passes.Data.Status == modelspb.PassesStatus_PassesStatusLocked {
		return nil, errmsg.NewErrPassesNotUnlock()
	}

	rewards := map[values.Integer]values.Integer{}
	if req.IsAll { // 一键领取
		rewards = svc.drawPassesAllReward(c, passes, reset)
	} else {
		cfgs := svc.getPassesRewardCfgsByLv(c, reset, passes.RoleLv)
		for _, cfg := range cfgs {
			if cfg.Lv != req.Lv {
				continue
			}
			if status := passes.Data.FreeUnlock[req.Lv]; status == modelspb.RewardStatus_Unlocked {
				// 普通奖励
				maps.Merge(rewards, cfg.FreeReward)
				passes.Data.FreeUnlock[req.Lv] = modelspb.RewardStatus_Received
			}

			// 高级奖励
			if passes.Data.Status == modelspb.PassesStatus_PassesStatusAdvanced {
				if status := passes.Data.PaidUnlock[req.Lv]; status == modelspb.RewardStatus_Unlocked {
					// 普通奖励
					maps.Merge(rewards, cfg.PaidReward)
					passes.Data.PaidUnlock[req.Lv] = modelspb.RewardStatus_Received
				}
			}
			break
		}
	}

	if len(rewards) == 0 {
		return nil, nil
	}

	_, err = svc.BagService.AddManyItem(c, c.RoleId, rewards)
	if err != nil {
		return nil, err
	}
	dao.SavePasses(c, passes)
	return &servicepb.Activity_DrawPassesRewardResponse{Rewards: rewards}, nil
}

// BuyAdvancePasses 购买解锁高级通行证
func (svc *Service) BuyAdvancePasses(c *ctx.Context) *errmsg.ErrMsg {
	passes, _, err := svc.getPasses(c)
	if err != nil {
		return err
	}
	passes.Data.Status = modelspb.PassesStatus_PassesStatusAdvanced
	dao.SavePasses(c, passes)
	c.PushMessage(&servicepb.Activity_UnlockAdvancePassesPush{})
	return nil
}

// CheatUnlockAdvancePasses 作弊器解锁高级通行证
func (svc *Service) CheatUnlockAdvancePasses(c *ctx.Context, _ *servicepb.Activity_CheatUnlockAdvancePassesRequest) (*servicepb.Activity_CheatUnlockAdvancePassesResponse, *errmsg.ErrMsg) {
	err := svc.BuyAdvancePasses(c)
	if err != nil {
		return nil, err
	}
	return &servicepb.Activity_CheatUnlockAdvancePassesResponse{}, nil
}

// CheatAddPassesExp 作弊器加通行证经验
func (svc *Service) CheatAddPassesExp(c *ctx.Context, req *servicepb.Activity_CheatAddPassesExpRequest) (*servicepb.Activity_CheatAddPassesExpResponse, *errmsg.ErrMsg) {
	err := svc.addPassesExp(c, req.Exp)
	if err != nil {
		return nil, err
	}
	return &servicepb.Activity_CheatAddPassesExpResponse{}, nil
}
