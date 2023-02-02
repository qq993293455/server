package hero

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/hero/dao"
	"coin-server/game-server/service/hero/rule"
)

func (svc *Service) BiographyRewardInfo(ctx *ctx.Context, _ *servicepb.Hero_BiographyRewardInfoRequest) (*servicepb.Hero_BiographyRewardInfoResponse, *errmsg.ErrMsg) {
	list, err := svc.biographyCanGetList(ctx, -1)
	if err != nil {
		return nil, err
	}
	return &servicepb.Hero_BiographyRewardInfoResponse{
		List: list,
	}, nil
}

func (svc *Service) BiographyGetReward(ctx *ctx.Context, req *servicepb.Hero_BiographyGetRewardRequest) (*servicepb.Hero_BiographyGetRewardResponse, *errmsg.ErrMsg) {
	cfg, ok := rule.GetHeroBiographyById(ctx, req.RewardId)
	if !ok {
		return nil, errmsg.NewErrBiographyNotExist()
	}
	if len(cfg.Reward) <= 0 {
		return nil, errmsg.NewErrBiographyNoReward()
	}
	if len(cfg.UnlockCondition) == 3 {
		typ := models.TaskType(cfg.UnlockCondition[0])
		need := cfg.UnlockCondition[2]
		data, err := svc.GetCounterByType(ctx, typ)
		if err != nil {
			return nil, err
		}
		if data[0] < need {
			return nil, errmsg.NewErrBiographyNotUnlock()
		}
	}
	b, err := dao.GetBiography(ctx)
	if err != nil {
		return nil, err
	}
	if b.BiographyId == nil {
		b.BiographyId = map[int64]int64{}
	}
	if b.BiographyId[req.RewardId] > 0 {
		return nil, errmsg.NewErrBiographyDoNotRepeatGetReward()
	}
	if _, err := svc.AddManyItem(ctx, ctx.RoleId, cfg.Reward); err != nil {
		return nil, err
	}
	b.BiographyId[req.RewardId] = timer.StartTime(ctx.StartTime).Unix()
	dao.SaveBiography(ctx, b)

	ctx.PublishEventLocal(&event.RedPointAdd{
		RoleId: ctx.RoleId,
		Key:    enum.RedPointBiographyKey,
		Val:    -1,
	})

	return &servicepb.Hero_BiographyGetRewardResponse{
		Reward:   cfg.Reward,
		RewardId: req.RewardId,
	}, nil
}

func (svc *Service) biographyCanGetList(ctx *ctx.Context, notContainType models.TaskType) ([]values.Integer, *errmsg.ErrMsg) {
	heroId, err := svc.GetAllHeroId(ctx)
	if err != nil {
		return nil, err
	}
	ids, list, types := rule.GetHeroBiography(ctx, heroId, notContainType)
	b, err := dao.GetBiography(ctx)
	if err != nil {
		return nil, err
	}
	unlockCondition, err := svc.getTaskCount(ctx, types)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		if unlockCondition[item.TaskType] >= item.NeedCount {
			ids = append(ids, item.Id)
		}
	}
	// ids 为玩家当前英雄传记下可领取的所有奖励
	if len(b.BiographyId) <= 0 {
		return ids, nil
	}
	ret := make([]values.Integer, 0)
	for _, id := range ids {
		if _, ok := b.BiographyId[id]; !ok {
			ret = append(ret, id)
		}
	}
	return ret, nil
}

func (svc *Service) getTaskCount(ctx *ctx.Context, types []models.TaskType) (map[models.TaskType]values.Integer, *errmsg.ErrMsg) {
	ret := make(map[models.TaskType]values.Integer)
	taskData, err := svc.GetCounterByTypeList(ctx, types)
	if err != nil {
		return nil, err
	}
	for taskType, m := range taskData {
		ret[taskType] = m[0]
	}
	return ret, nil
}

func (svc *Service) handleRedPointByGetNewHero(ctx *ctx.Context, heroId values.HeroId) *errmsg.ErrMsg {
	ids, list, types := rule.GetHeroBiography(ctx, []values.HeroId{heroId}, -1)
	b, err := dao.GetBiography(ctx)
	if err != nil {
		return err
	}
	unlockCondition, err := svc.getTaskCount(ctx, types)
	if err != nil {
		return err
	}
	for _, item := range list {
		if unlockCondition[item.TaskType] >= item.NeedCount {
			ids = append(ids, item.Id)
		}
	}
	var count values.Integer
	// ids 为玩家当前英雄传记下可领取的所有奖励
	if len(b.BiographyId) <= 0 {
		count = values.Integer(len(ids))
	} else {
		for _, id := range ids {
			if _, ok := b.BiographyId[id]; !ok {
				count++
			}
		}
	}
	if count != 0 {
		ctx.PublishEventLocal(&event.RedPointAdd{
			RoleId: ctx.RoleId,
			Key:    enum.RedPointBiographyKey,
			Val:    count,
		})
	}
	return nil
}
