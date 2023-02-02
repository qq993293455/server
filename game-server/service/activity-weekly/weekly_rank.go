package activity_weekly

import (
	"strconv"

	"coin-server/common/ActivityRankingRule"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/activity_ranking_service"
	daopb "coin-server/common/proto/dao"
	modelspb "coin-server/common/proto/models"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/service/activity-weekly/dao"
	heroDao "coin-server/game-server/service/hero/dao"
	"coin-server/game-server/util/trans"
	"coin-server/rule"

	"go.uber.org/zap"
)

const rankingMailId = 100017

func (svc *Service) GetRankingConfigs(c *ctx.Context, activityId values.Integer) []*modelspb.WeeklyRankCnf {
	var ret []*modelspb.WeeklyRankCnf
	cfgs := rule.MustGetReader(c).ActivityWeeklyRank.List()
	for _, cfg := range cfgs {
		if cfg.ActivityId != activityId {
			continue
		}
		ret = append(ret, &modelspb.WeeklyRankCnf{
			Id:             cfg.Id,
			ActivityId:     cfg.ActivityId,
			RankUpperLimit: cfg.RankUpperLimit,
			RankLowerLimit: cfg.RankLowerLimit,
			Reward:         cfg.Reward,
		})
	}
	return ret
}

func (svc *Service) startRanking(c *ctx.Context, aw *modelspb.ActivityWeekly) {
	rankingIdx, err := svc.joinRanking(c, aw)
	if err != nil {
		panic(err)
	}
	aw.RankingInfo.RankingIndex = rankingIdx
}

func (svc *Service) joinRanking(c *ctx.Context, aw *modelspb.ActivityWeekly) (string, *errmsg.ErrMsg) {
	version := aw.Version
	rankingEndTime := aw.EndTime

	out := &activity_ranking_service.ActivityRanking_JoinRankingResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(c, ActivityRankingRule.GetActivityRankingServer(aw.ActivityId), &activity_ranking_service.ActivityRanking_JoinRankingRequest{
		RoleId:       c.RoleId,
		Score:        0,
		RankingIndex: strconv.FormatInt(aw.ActivityId, 10),
		Version:      version,
		CreateInfo: &modelspb.ActivityRanking_Info{
			StartTime:    timer.Now().Unix(),
			Version:      version,
			EndTime:      rankingEndTime,
			DurationTime: rankingEndTime,
			RefreshTime:  GetActivityWeeklyRefreshTime(c),
		},
	}, out); err != nil {
		return "", err
	}

	c.Info("Join Activity Weekly Ranking", zap.String("rankingIndex", out.RankingIndex), zap.String("roleId", c.RoleId))
	return out.RankingIndex, nil
}

func (svc *Service) EndRanking(c *ctx.Context, aw *modelspb.ActivityWeekly) *errmsg.ErrMsg {
	if aw.RankingInfo == nil {
		c.Error("ActivityWeeklyData not have ranking data", zap.String("role_id", c.RoleId), zap.Int64("activityId", aw.ActivityId))
		return nil
	}

	return svc.procOverRanking(c, aw)
}

// procOverRanking 活动结束时处理排行榜发奖
func (svc *Service) procOverRanking(c *ctx.Context, aw *modelspb.ActivityWeekly) *errmsg.ErrMsg {
	version := aw.Version

	rankingIndex := aw.RankingInfo.RankingIndex
	has, rankData, err := dao.GetRankingData(c, rankingIndex, version, c.RoleId)
	if err != nil {
		c.Error("GetRankingData error", zap.Error(err), zap.String("rankingIndex", rankingIndex))
		return err
	}
	if !has {
		c.Error("GetRankingData not find data", zap.String("rankingIndex", rankingIndex))
		return nil
	}

	rank := rankData.Data.RankingId + 1
	cfgs := rule.MustGetReader(c).ActivityWeeklyRank.List()
	for _, cfg := range cfgs {
		if aw.ActivityId != cfg.ActivityId {
			continue
		}
		if rank < cfg.RankUpperLimit {
			continue
		}
		if rank > cfg.RankLowerLimit {
			continue
		}

		rewards := trans.ItemSliceToPb(cfg.Reward)
		if err := svc.MailService.Add(c, c.RoleId, &modelspb.Mail{
			Type:       modelspb.MailType_MailTypeSystem,
			TextId:     rankingMailId,
			Attachment: rewards,
		}); err != nil {
			c.Error("send mail error", zap.Error(err), zap.Any("reward", cfg.Reward), zap.Any("role_id", c.RoleId))
			panic(err)
		}

		break
	}

	dao.DelRankingData(c, rankingIndex, version, c.RoleId)
	return nil
}

func (svc *Service) updateRanking(ctx *ctx.Context, aw *modelspb.ActivityWeekly) (bool, *errmsg.ErrMsg) {
	out := &activity_ranking_service.ActivityRanking_UpdateRankingDataResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, ActivityRankingRule.GetActivityRankingServer(aw.ActivityId), &activity_ranking_service.ActivityRanking_UpdateRankingDataRequest{
		RoleId:       ctx.RoleId,
		Score:        aw.Score,
		RankingIndex: aw.RankingInfo.RankingIndex,
	}, out); err != nil {
		return false, err
	}

	return false, nil
}

func (svc *Service) GetRankingSelf(ctx *ctx.Context, aw *modelspb.ActivityWeekly) (int64, *errmsg.ErrMsg) {
	out := &activity_ranking_service.ActivityRanking_GetSelfRankingResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, ActivityRankingRule.GetActivityRankingServer(aw.ActivityId), &activity_ranking_service.ActivityRanking_GetSelfRankingRequest{
		RoleId:       ctx.RoleId,
		RankingIndex: aw.RankingInfo.RankingIndex,
	}, out); err != nil {
		return 0, err
	}
	return out.RankingId, nil
}

func (svc *Service) GetRankingList(ctx *ctx.Context, aw *modelspb.ActivityWeekly, startIndex int64, cnt int64) (int64, []*modelspb.ArenaInfo, int64, *errmsg.ErrMsg) {
	out := &activity_ranking_service.ActivityRanking_GetRankingListResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, ActivityRankingRule.GetActivityRankingServer(aw.ActivityId), &activity_ranking_service.ActivityRanking_GetRankingListRequest{
		RoleId:       ctx.RoleId,
		RankingIndex: aw.RankingInfo.RankingIndex,
		StartIndex:   startIndex,
		Count:        cnt,
	}, out); err != nil {
		return 0, nil, 0, err
	}

	roleInfos := make([]*modelspb.ArenaInfo, 0, len(out.RankingList))
	for _, rankInfo := range out.RankingList {
		roleInfo, err := svc.GetPlayerInfo(ctx, rankInfo)
		if err != nil {
			ctx.Error("GetRankingList GetPlayerInfo error", zap.Any("err msg", err))
			roleInfo = &modelspb.ArenaInfo{
				RoleId:    rankInfo.RoleId,
				RankingId: rankInfo.RankingId,
			}
		}
		roleInfos = append(roleInfos, roleInfo)
	}
	return out.SelfRankingId, roleInfos, out.NextRefreshTime, nil
}

func (svc *Service) GetPlayerInfo(c *ctx.Context, rankInfo *modelspb.ActivityRanking_RankInfo) (*modelspb.ArenaInfo, *errmsg.ErrMsg) {
	role, err := svc.Module.GetRoleModelByRoleId(c, rankInfo.RoleId)
	if err != nil {
		return nil, err
	}

	// fightHero, err := svc.Module.FormationService.GetDefaultHeroes(c, playerId)
	// if err != nil {
	//	return nil, err
	// }
	// fHero, err := svc.GetHero(c, playerId, fightHero)
	// if err != nil {
	//	return nil, err
	// }
	// power := int(fHero.GetHero_0Power()) + int(fHero.GetHero_1Power())

	guildName := ""
	guild, err := svc.GuildService.GetUserGuildInfo(c, rankInfo.RoleId)
	if err != nil {
		return nil, err
	}
	if guild != nil {
		guildName = guild.Name
	}

	ret := &modelspb.ArenaInfo{
		RoleId:      rankInfo.RoleId,
		Nickname:    role.Nickname,
		Level:       role.Level,
		AvatarId:    role.AvatarId,
		AvatarFrame: role.AvatarFrame,
		// Power:       int64(power),
		Title:     role.Title,
		RankingId: rankInfo.RankingId,
		GuildName: guildName,
		Score:     rankInfo.Score,
	}

	if rankInfo.RankingId < 3 {
		// 获取英雄信息
		heroesFormation, err2 := svc.Module.FormationService.GetDefaultHeroes(c, rankInfo.RoleId)
		if err2 != nil {
			return nil, err2
		}
		hero, _, err2 := svc.Module.GetHero(c, rankInfo.RoleId, heroesFormation.HeroOrigin_0)
		if err2 != nil {
			return nil, err2
		}
		ret.HeroId = hero.Id
		ret.HeroFashion = hero.Fashion.Dressed
	}

	return ret, nil
}

func (svc *Service) GetHero(ctx *ctx.Context, roleId string, assembleHero *modelspb.Assemble) (*modelspb.Assemble, *errmsg.ErrMsg) {
	if assembleHero.HeroOrigin_0 == 0 && assembleHero.HeroOrigin_1 == 0 {
		return nil, errmsg.NewErrArenaNotHeroes()
	}
	heros, err := heroDao.NewHero(roleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	var heroMap = make(map[int64]*daopb.Hero, len(heros))
	for _, hero := range heros {
		heroMap[hero.BuildId] = hero
	}
	if assembleHero.HeroOrigin_0 != 0 {
		isFind := false
		for _, heroInfo := range heroMap {
			if heroInfo.Id == assembleHero.HeroOrigin_0 {
				assembleHero.Hero_0 = heroInfo.BuildId
				assembleHero.HeroOrigin_0 = heroInfo.Id
				assembleHero.Hero_0Power = heroInfo.CombatValue.Total
				isFind = true
			}
		}
		if !isFind {
			return nil, errmsg.NewErrArenaHero()
		}
	}

	if assembleHero.HeroOrigin_1 != 0 {
		isFind := false
		for _, heroInfo := range heroMap {
			if heroInfo.Id == assembleHero.HeroOrigin_1 {
				assembleHero.Hero_1 = heroInfo.BuildId
				assembleHero.HeroOrigin_1 = heroInfo.Id
				assembleHero.Hero_1Power = heroInfo.CombatValue.Total
				isFind = true
			}
		}
		if !isFind {
			return nil, errmsg.NewErrArenaHero()
		}
	}
	return assembleHero, nil
}
