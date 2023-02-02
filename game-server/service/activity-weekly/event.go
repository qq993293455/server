package activity_weekly

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/activity-weekly/dao"
	"coin-server/rule"
)

func (svc *Service) HandleRoleLoginEvent(c *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	if d.IsRegister {
		return nil
	}
	_, err := svc.getActivityWeeklyData(c, c.RoleId)
	return err
}

func (svc *Service) HandlePaySuccess(c *ctx.Context, d *event.PaySuccess) *errmsg.ErrMsg {
	v, ok := rule.MustGetReader(c).Charge.GetChargeById(d.PcId)
	if !ok {
		return errmsg.NewErrActivityGiftEmpty()
	}
	if v.FunctionType == 1 && v.TargetId == enum.WeeklyCastSword {
		return svc.BuyGiftItemsByCash(c, d.PcId)
	}
	return nil
}

func (svc *Service) HandleGuildEvent(c *ctx.Context, e *event.GuildEvent) *errmsg.ErrMsg {
	data, err := svc.getActivityWeeklyData(c, e.RoleId)
	if err != nil {
		return err
	}
	for _, aw := range data.ActivityWeeklyData {
		err = svc.refreshChallengeInfo(c, e.RoleId, aw)
		if err != nil {
			return err
		}
		c.PushMessage(&servicepb.ActivityWeekly_WeeklyChallengeGuildRewardsPush{
			ActivityId:   aw.ActivityId,
			GuildId:      aw.ChallengeInfo.GuildId,
			GuildRewards: aw.ChallengeInfo.GuildRewards,
		})
	}
	dao.SaveActivityWeeklyData(c, data)
	return nil
}
