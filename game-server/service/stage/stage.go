package stage

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	protosvc "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/stage/dao"
	taskDao "coin-server/game-server/service/task/task/dao"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
}

func NewStageService(
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
	s.Module.StageService = s
	return s
}

// ---------------------------------------------------proto------------------------------------------------------------//

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取关卡列表", svc.GetStageList)
	svc.svc.RegisterFunc("领取关卡奖励", svc.GetStageReward)
	svc.svc.RegisterFunc("解锁所有关卡", svc.CheatUnlockAllStage)
	svc.svc.RegisterFunc("关卡进度达到满", svc.CheatExploreDegree)

	eventlocal.SubscribeEventLocal(svc.HandleEventFinished)
	eventlocal.SubscribeEventLocal(svc.HandleTargetUpdate)
}

func (svc *Service) CheatUnlockAllStage(ctx *ctx.Context, req *protosvc.Stage_CheatUnlockAllStageRequest) (*protosvc.Stage_CheatUnlockAllStageResponse, *errmsg.ErrMsg) {
	st, err := dao.GetStage(ctx)
	if err != nil {
		return nil, err
	}
	for {
		ms, ok := rule.MustGetReader(ctx).MapSelect.GetMapSelectById(st.CurrStage)
		if !ok {
			return nil, errmsg.NewErrStageIdNotExist()
		}
		next := ms.Next()
		if next == nil {
			break
		}

		st.CurrStage = next.Id
	}
	dao.SaveStage(ctx, st)
	return &protosvc.Stage_CheatUnlockAllStageResponse{}, nil
}

func (svc *Service) CheatExploreDegree(ctx *ctx.Context, req *protosvc.Stage_CheatSetExploreDegreeRequest) (*protosvc.Stage_CheatSetExploreDegreeResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.MapSelect.GetMapSelectById(req.StageId)
	if !ok {
		return nil, nil
	}
	st, err := dao.GetExploreById(ctx, ctx.RoleId, req.StageId)
	if err != nil {
		return nil, err
	}
	if st.Status != models.RewardStatus_Locked {
		return nil, nil
	}
	st.Explore = map[int64]int64{}
	st.Degree = 0
	for k, v := range cfg.ExploreDegree {
		st.Explore[int64(k+1)] = v
		st.Degree++
	}
	st.Status = models.RewardStatus_Unlocked
	dao.SaveExplore(ctx, ctx.RoleId, st)
	return &protosvc.Stage_CheatSetExploreDegreeResponse{}, nil
}

func (svc *Service) IsLock(ctx *ctx.Context, sceneId int64) (bool, *errmsg.ErrMsg) {
	id, ok := rule.MustGetReader(ctx).CustomParse.GetMapSelectIdBySceneId(sceneId)
	if !ok {
		return false, errmsg.NewErrStageIdNotExist()
	}

	st, err := dao.GetStage(ctx)
	if err != nil {
		return false, err
	}
	newStage, err := svc.tryLock(ctx, st)
	if err != nil {
		return false, err
	}
	if newStage != st.CurrStage {
		st.CurrStage = newStage
		st.TypeCount = map[int64]int64{}
		newExplore := &pbdao.StageExplore{
			StageId: newStage,
			Degree:  0,
			Explore: map[int64]int64{},
			Status:  models.RewardStatus_Locked,
		}
		dao.SaveExplore(ctx, ctx.RoleId, newExplore)
		dao.SaveStage(ctx, st)
	}
	return st.CurrStage >= id, nil

}

func (svc *Service) NextStage(c *ctx.Context, currStage int64) (*rule_model.MapSelect, *errmsg.ErrMsg) {
	if currStage == 0 {
		return dao.GetFirstStageID(c), nil
	}
	ms, ok := rule.MustGetReader(c).MapSelect.GetMapSelectById(currStage)
	if !ok {
		return nil, errmsg.NewErrStageIdNotExist()
	}
	return ms.Next(), nil
}

func (svc *Service) tryLock(ctx *ctx.Context, stage *pbdao.Stage) (int64, *errmsg.ErrMsg) {
	currStage := stage.CurrStage
	for {
		next, err := svc.NextStage(ctx, currStage)
		if err != nil {
			return 0, err
		}
		if next == nil {
			break
		}
		tt := rule.MustGetReader(ctx).TaskType
		for _, v := range next.UnlockCondition {
			t, ok := tt.GetTaskTypeById(v[0])
			if ok {
				if t.IsAccumulate {
					cc, err := taskDao.GetCondByType(ctx, ctx.RoleId, models.TaskType(t.Id))
					if err != nil {
						return 0, err
					}
					if cc.Count[v[1]] < v[2] {
						return currStage, nil
					}
				} else {
					if stage.TypeCount[v[1]] < v[2] {
						return currStage, nil
					}
				}
			}
		}

		currStage = next.Id
	}
	return currStage, nil
}

func (svc *Service) updateExplore(ctx *ctx.Context, eventId values.EventId, mapId values.MapId) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(eventId)
	if !ok {
		return nil
	}
	mapCfg, ok := r.MapScene.GetMapSceneById(mapId)
	if !ok {
		return nil
	}
	stageCfg, ok := r.MapSelect.GetMapSelectById(mapCfg.LevelMapId)
	if !ok {
		return nil
	}
	explore, err := dao.GetExploreById(ctx, ctx.RoleId, mapCfg.LevelMapId)
	if err != nil {
		return err
	}
	if explore.Status != models.RewardStatus_Locked {
		return nil
	}
	limit := stageCfg.ExploreDegree[cfg.Subtype-1]
	if explore.Explore[cfg.Subtype] >= limit {
		return nil
	}
	explore.Explore[cfg.Subtype]++
	explore.Degree++
	total := int64(0)
	for _, v := range stageCfg.ExploreDegree {
		total += v
	}
	if explore.Degree >= total {
		explore.Status = models.RewardStatus_Unlocked
	}
	dao.SaveExplore(ctx, ctx.RoleId, explore)
	return nil
}

func (svc *Service) GetStageList(ctx *ctx.Context, _ *protosvc.Stage_GetStageListRequest) (*protosvc.Stage_GetStageListResponse, *errmsg.ErrMsg) {
	st, err := dao.GetStage(ctx)
	if err != nil {
		return nil, err
	}
	newStage, err := svc.tryLock(ctx, st)
	if err != nil {
		return nil, err
	}
	if newStage != st.CurrStage {
		st.CurrStage = newStage
		st.TypeCount = map[int64]int64{}
		newExplore := &pbdao.StageExplore{
			StageId: newStage,
			Degree:  0,
			Explore: map[int64]int64{},
			Status:  models.RewardStatus_Locked,
		}
		dao.SaveExplore(ctx, ctx.RoleId, newExplore)
		dao.SaveStage(ctx, st)
	}
	explore, err := dao.GetExplore(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	return &protosvc.Stage_GetStageListResponse{
		CurrStageId: st.CurrStage,
		Explore:     ExploreDao2Models(explore),
	}, nil
}

func (svc *Service) GetStageReward(ctx *ctx.Context, req *protosvc.Stage_GetExploreRewardRequest) (*protosvc.Stage_GetExploreRewardResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.MapSelect.GetMapSelectById(req.StageId)
	if !ok {
		return nil, errmsg.NewErrStageExploreNotFinish()
	}
	explore, err := dao.GetExploreById(ctx, ctx.RoleId, req.StageId)
	if err != nil {
		return nil, err
	}
	if explore.Status != models.RewardStatus_Unlocked {
		return nil, errmsg.NewErrStageExploreNotFinish()
	}
	explore.Status = models.RewardStatus_Received
	_, err = svc.BagService.AddManyItem(ctx, ctx.RoleId, cfg.Reward)
	if err != nil {
		return nil, err
	}
	dao.SaveExplore(ctx, ctx.RoleId, explore)
	return &protosvc.Stage_GetExploreRewardResponse{Reward: cfg.Reward}, nil
}
