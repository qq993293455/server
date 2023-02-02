package activity_weekly

import (
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	daopb "coin-server/common/proto/dao"
	modelspb "coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/values"
	"coin-server/game-server/service/activity-weekly/dao"
	"coin-server/game-server/util/trans"
	"coin-server/rule"

	"go.uber.org/zap"
)

const (
	ChallengeSingleMailId = 100019
	ChallengeGuildMailId  = 100020
)

func (svc *Service) GetChallengeConfigs(c *ctx.Context, activityId values.Integer) []*modelspb.WeeklyChallengeCnf {
	var ret []*modelspb.WeeklyChallengeCnf
	cfgs := rule.MustGetReader(c).ActivityWeeklyChallenge.List()
	for _, cfg := range cfgs {
		if cfg.ActivityId != activityId {
			continue
		}
		ret = append(ret, &modelspb.WeeklyChallengeCnf{
			Id:         cfg.Id,
			ActivityId: cfg.ActivityId,
			Typ:        cfg.Typ,
			Times:      cfg.Times,
			Reward:     cfg.Reward,
		})
	}
	return ret
}

func (svc *Service) refreshChallengeInfo(c *ctx.Context, roleId values.RoleId, aw *modelspb.ActivityWeekly) *errmsg.ErrMsg {
	guild, err := svc.GuildService.GetUserGuildInfo(c, roleId)
	if err != nil {
		return err
	}
	guildId := ""
	if guild != nil {
		guildId = guild.Id
	}

	if aw.ChallengeInfo.GuildId != guildId { // 如果换公会了
		for k, v := range aw.ChallengeInfo.GuildRewards {
			if v == modelspb.RewardStatus_Received {
				continue
			}
			// 清除未领取的奖励，按新公会的进度重新计算
			delete(aw.ChallengeInfo.GuildRewards, k)
		}
		aw.ChallengeInfo.GuildId = guildId
	}

	return svc.refreshGuildChallengeRewards(c, aw)
}

func (svc *Service) refreshGuildChallengeRewards(c *ctx.Context, aw *modelspb.ActivityWeekly) *errmsg.ErrMsg {
	gci, err := dao.GetGuildChallengeInfo(c, aw.ActivityId, aw.Version, aw.ChallengeInfo.GuildId)
	if err != nil {
		return err
	}
	for _, cfg := range rule.MustGetReader(c).ActivityWeeklyChallenge.List() {
		if cfg.Typ != 2 {
			continue
		}
		if _, ok := aw.ChallengeInfo.GuildRewards[cfg.Id]; ok {
			continue
		}
		if gci.Score >= cfg.Times {
			aw.ChallengeInfo.GuildRewards[cfg.Id] = modelspb.RewardStatus_Unlocked
		}
	}
	return nil
}

func (svc *Service) updateChallenge(c *ctx.Context, aw *modelspb.ActivityWeekly, times int64) (bool, *errmsg.ErrMsg) {
	needSave := false

	// 更新公会（位面）挑战
	gci, err := svc.updateGuildChallenge(c, aw, times)
	if err != nil {
		return false, err
	}

	cfgs := rule.MustGetReader(c).ActivityWeeklyChallenge.List()
	for _, cfg := range cfgs {
		if cfg.ActivityId != aw.ActivityId {
			continue
		}

		switch cfg.Typ {
		case 1: // 个人挑战
			if _, ok := aw.ChallengeInfo.Rewards[cfg.Id]; !ok && aw.Score >= cfg.Times {
				aw.ChallengeInfo.Rewards[cfg.Id] = modelspb.RewardStatus_Unlocked
				needSave = true
			}
		case 2: // 公会挑战
			if gci == nil {
				continue
			}
			if _, ok := aw.ChallengeInfo.GuildRewards[cfg.Id]; !ok && gci.Score >= cfg.Times {
				aw.ChallengeInfo.GuildRewards[cfg.Id] = modelspb.RewardStatus_Unlocked
				needSave = true
			}
		}
	}

	return needSave, nil
}

func getGuildChallengeLockKey(activityId values.Integer, guildId string) string {
	return "aw_guild_challenge_lock:" + strconv.FormatInt(activityId, 10) + ":" + guildId
}

func (svc *Service) updateGuildChallenge(c *ctx.Context, aw *modelspb.ActivityWeekly, times int64) (*daopb.ActivityWeeklyGuildData, *errmsg.ErrMsg) {
	guildId, err := svc.GuildService.GetGuildIdByRole(c)
	if err != nil {
		c.Error("GetUserGuildInfo error", zap.Any("err", err))
		return nil, err
	}
	if guildId == "" {
		return nil, nil
	}

	err = c.DRLock(redisclient.GetLocker(), getGuildChallengeLockKey(aw.ActivityId, guildId))
	if err != nil {
		return nil, err
	}

	gci, _ := dao.GetGuildChallengeInfo(c, aw.ActivityId, aw.Version, guildId)
	gci.Score += times
	dao.SaveGuildChallengeInfo(c, gci)
	return gci, nil
}

//DrawChallengeReward 领取挑战奖励 一键领取所有
func (svc *Service) DrawChallengeReward(c *ctx.Context, aw *modelspb.ActivityWeekly, id int64) ([]*modelspb.Item, *errmsg.ErrMsg) {
	cfg, ok := rule.MustGetReader(c).ActivityWeeklyChallenge.GetActivityWeeklyChallengeById(id)
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyParam()
	}

	var rewards []*modelspb.Item
	if cfg.Typ == 1 {
		rewards = svc.drawAllChallengeReward(c, aw.ChallengeInfo.Rewards)
	} else {
		rewards = svc.drawAllChallengeReward(c, aw.ChallengeInfo.GuildRewards)
	}

	_, err := svc.BagService.AddManyItem(c, c.RoleId, trans.ItemProtoToMap(rewards))
	if err != nil {
		return nil, err
	}

	return rewards, nil
}

func (svc *Service) drawAllChallengeReward(c *ctx.Context, challengeRewards map[int64]modelspb.RewardStatus) []*modelspb.Item {
	rewards := make([]*modelspb.Item, 0)
	for cfgId, status := range challengeRewards {
		if status != modelspb.RewardStatus_Unlocked {
			continue
		}
		cfg, ok := rule.MustGetReader(c).ActivityWeeklyChallenge.GetActivityWeeklyChallengeById(cfgId)
		if !ok { // 容错
			delete(challengeRewards, cfgId)
			continue
		}
		items := trans.ItemSliceToPb(cfg.Reward)
		rewards = append(rewards, items...)
		challengeRewards[cfgId] = modelspb.RewardStatus_Received
	}
	return rewards
}

func (svc *Service) EndChallenge(c *ctx.Context, aw *modelspb.ActivityWeekly) *errmsg.ErrMsg {
	if aw.ChallengeInfo == nil {
		return nil
	}

	singleRewards := svc.drawAllChallengeReward(c, aw.ChallengeInfo.Rewards)
	if len(singleRewards) > 0 {
		if err := svc.MailService.Add(c, c.RoleId, &modelspb.Mail{
			Type:       modelspb.MailType_MailTypeSystem,
			TextId:     ChallengeSingleMailId,
			Attachment: singleRewards,
		}); err != nil {
			c.Error("send challenge singleRewards mail error", zap.Any("msg", err), zap.Any("rewards", singleRewards), zap.Any("role_id", c.RoleId))
			return err
		}
	}

	guildRewards := svc.drawAllChallengeReward(c, aw.ChallengeInfo.GuildRewards)
	if len(guildRewards) > 0 {
		if err := svc.MailService.Add(c, c.RoleId, &modelspb.Mail{
			Type:       modelspb.MailType_MailTypeSystem,
			TextId:     ChallengeGuildMailId,
			Attachment: guildRewards,
		}); err != nil {
			c.Error("send challenge guildRewards mail error", zap.Any("msg", err), zap.Any("rewards", guildRewards), zap.Any("role_id", c.RoleId))
			return err
		}
	}

	return nil
}
