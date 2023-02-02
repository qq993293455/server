package tower

import (
	"strconv"
	"sync"

	"coin-server/game-server/service/user/db"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/iggsdk"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/statistical"
	models2 "coin-server/common/statistical/models"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/module"
	valuesJourney "coin-server/game-server/service/journey/values"
	"coin-server/game-server/service/tower/dao"
	"coin-server/game-server/service/tower/rule"
	tower_rule "coin-server/game-server/service/tower/rule"
	"coin-server/game-server/util/trans"
	sys_rule "coin-server/rule"
	rulemodel "coin-server/rule/rule-model"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	log      *logger.Logger
	fightMap sync.Map
}

type RewardBase struct {
	item         []*models.Item
	stageItem    []*models.Item
	intervalTime int
}

type TowerRewardCalculateParam struct {
	ctx          *ctx.Context
	towerDaoData dao.TowerI
	towerData    *models.TowerData
	config       *tower_rule.TowerPublicConfig
	rewardBase   *RewardBase
}

type MeditationData struct {
	isFree bool
	cost   map[int64]int64
	reward []*models.Item
}

func NewTowerService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
		log:        log,
	}
	// s.fight_map = make(map[string]int64)
	module.TowerService = s
	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("获取爬塔数据", s.GetTowerRewardInfoRequest)
	s.svc.RegisterFunc("获取下一层数据", s.GetNextChallengeDataRequest)
	s.svc.RegisterFunc("查询 挂机奖励", s.TowerAccumulateInfoRequest)
	s.svc.RegisterFunc("收割 挂机奖励", s.TowerAccumulateHarvestRequest)
	s.svc.RegisterFunc("获取爬塔战斗数据(玩家自己的英雄等数据)", s.BattleStart)
	s.svc.RegisterFunc("结算", s.TowerChallengeSettlement)
	s.svc.RegisterFunc("快速计算战斗结果", s.TowerChallengeSoomCalculationRequest)
	s.svc.RegisterFunc("Gm设置层数", s.TowerGmSetLevel)
	s.svc.RegisterFunc("Gm重置冥想次数", s.TowerGmRefreshMeditationTimes)
	s.svc.RegisterFunc("单人战斗", s.SingleBattle)
}

// msg proc ============================================================================================================================================================================================================
func (s *Service) TowerGmSetLevel(ctx *ctx.Context, req *servicepb.Tower_CheatTowerGmSetlevelRequest) (*servicepb.Tower_CheatTowerGmSetlevelResponse, *errmsg.ErrMsg) {
	towerCalculateParam, err := s.GetTowerRewardCalculateParam(ctx, req.Type)
	if err != nil {
		return nil, err
	}

	maxLevel := int64(s.GetMaxLevel(ctx, req.Type))

	setLevel := int64(1)
	if req.CurrentLevel < 1 {
		req.CurrentLevel = 1
	}
	if req.CurrentLevel > maxLevel {
		req.CurrentLevel = maxLevel
	}
	setLevel = req.CurrentLevel

	towerCalculateParam.towerData.CurrentTowerLevel = setLevel
	towerCalculateParam.towerDaoData.SaveData(ctx)

	return &servicepb.Tower_CheatTowerGmSetlevelResponse{}, nil
}

func (s *Service) TowerGmRefreshMeditationTimes(ctx *ctx.Context, req *servicepb.Tower_CheatTowerGmRefreshMeditationTimesRequest) (*servicepb.Tower_CheatTowerGmRefreshMeditationTimesResponse, *errmsg.ErrMsg) {
	towerCalculateParam, err := s.GetTowerRewardCalculateParam(ctx, req.Type)
	if err != nil {
		return nil, err
	}
	towerCalculateParam.towerDaoData.RefreshMeditationData(ctx)
	return &servicepb.Tower_CheatTowerGmRefreshMeditationTimesResponse{}, nil
}

func (s *Service) TowerChallengeSoomCalculationRequest(ctx *ctx.Context, req *servicepb.Tower_TowerChallengeSoomCalculationRequest) (*servicepb.Tower_TowerChallengeSoomCalculationResponse, *errmsg.ErrMsg) {
	var towerCalculateParam *TowerRewardCalculateParam
	towerCalculateParam, err := s.GetTowerRewardCalculateParam(ctx, req.Type)
	if err != nil {
		return nil, err
	}

	recommendPower, err := s.GetRecommendNum(towerCalculateParam)
	if err != nil {
		return nil, err
	}

	monsterPower := float64(recommendPower) * float64(req.MonsterHpRatio)

	heroesFormation, err := s.Module.FormationService.GetDefaultHeroes(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	heroIds := make([]int64, 0, 2)
	if heroesFormation.Hero_0 > 0 && heroesFormation.HeroOrigin_0 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_0)
	}
	if heroesFormation.Hero_1 > 0 && heroesFormation.HeroOrigin_1 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_1)
	}
	heroes, err := s.Module.GetHeroes(ctx, ctx.RoleId, heroIds)
	if err != nil {
		return nil, err
	}

	var playerPower float64
	for _, v := range req.Heroes {
		for _, hero := range heroes {
			if v.HeroId == hero.Id {
				playerPower += float64(hero.CombatValue.Total*v.HeroHp) / float64(hero.Attrs[103])
			}
		}
	}

	if playerPower > monsterPower {
		err = s.Module.JourneyService.AddToken(ctx, ctx.RoleId, valuesJourney.JourneyTower, 1)
		if err != nil {
			return nil, err
		}
		s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskTowerChallengeSucc, 0, 1)
	}

	return &servicepb.Tower_TowerChallengeSoomCalculationResponse{
		Type:  req.Type,
		IsWin: playerPower > monsterPower,
	}, nil
}
func (s *Service) GetTowerRewardInfoRequest(ctx *ctx.Context, req *servicepb.Tower_GetTowerRewardInfoRequest) (*servicepb.Tower_GetTowerRewardInfoResponse, *errmsg.ErrMsg) {
	data, _, err := s.GetTowerInfo(ctx, req.Type)
	if err != nil {
		return nil, err
	}
	return &servicepb.Tower_GetTowerRewardInfoResponse{
		Type:    req.Type,
		DataMap: data.PassData,
	}, nil
}

func (s *Service) GetNextChallengeDataRequest(ctx *ctx.Context, req *servicepb.Tower_GetNextChallengeDataRequest) (*servicepb.Tower_GetNextChallengeDataResponse, *errmsg.ErrMsg) {
	data, err := s.GetLevelInfo(ctx, req.Type)
	if err != nil {
		return nil, err
	}
	return &servicepb.Tower_GetNextChallengeDataResponse{
		Type: req.Type,
		Data: data,
	}, nil
}

func (s *Service) TowerAccumulateInfoRequest(ctx *ctx.Context, req *servicepb.Tower_TowerAccumulateInfoRequest) (*servicepb.Tower_TowerAccumulateInfoResponse, *errmsg.ErrMsg) {
	var err *errmsg.ErrMsg
	var towerCalculateParam *TowerRewardCalculateParam
	towerCalculateParam, err = s.GetTowerRewardCalculateParam(ctx, req.Type)
	if err != nil {
		return nil, err
	}

	accumulateInfo, err := s.CalculateFreeCurrentAccumulateReward(ctx, towerCalculateParam)
	if err != nil {
		return nil, err
	}
	return &servicepb.Tower_TowerAccumulateInfoResponse{
		Type: req.Type,
		Info: accumulateInfo,
	}, nil
}

func (s *Service) TowerAccumulateHarvestRequest(ctx *ctx.Context, req *servicepb.Tower_TowerAccumulateHarvestRequest) (*servicepb.Tower_TowerAccumulateHarvestResponse, *errmsg.ErrMsg) {
	towerType := req.Type
	harvestType := req.HarvestType

	var err *errmsg.ErrMsg
	var towerCalculateParam *TowerRewardCalculateParam
	towerCalculateParam, err = s.GetTowerRewardCalculateParam(ctx, towerType)
	if err != nil {
		return nil, err
	}

	var accumulateInfo *models.TowerAccumulateHarvestInfo
	accumulateInfo, err = s.CalculateFreeCurrentAccumulateReward(ctx, towerCalculateParam)
	if err != nil {
		return nil, err
	}

	var rewards []*models.Item = accumulateInfo.ProfitsData
	var meditationData *MeditationData
	if harvestType == models.AccumulateHarvestType_AT_Meditation {
		meditationData, err = s.CalculateCostCurrentAccumulateReward(towerCalculateParam)
		if err != nil {
			return nil, err
		}
		err = s.ProcMeditationCost(ctx, meditationData)
		if err != nil {
			return nil, err
		}
		s.ProcMeditationCostTimes(ctx, towerCalculateParam, meditationData, accumulateInfo)
		rewards = meditationData.reward
	}

	for _, item := range rewards {
		err = s.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
		if err != nil {
			return nil, err
		}
	}

	if harvestType == models.AccumulateHarvestType_AT_Meditation {
		towerCalculateParam.towerDaoData.AddMeditationTimes(ctx, towerCalculateParam.towerData.Type, meditationData.isFree)
	} else {
		towerCalculateParam.towerDaoData.InitTowerRewardData(ctx, towerCalculateParam.towerData.Type, accumulateInfo.ModTime)
		accumulateInfo.ProfitsData = []*models.Item{}
		accumulateInfo.LastSettlementTime = timer.Now().Unix()
		accumulateInfo.NextSettlementTime = timer.Now().Unix() + int64(towerCalculateParam.config.Settlement_interval)
	}

	return &servicepb.Tower_TowerAccumulateHarvestResponse{
		Type:    req.Type,
		Info:    accumulateInfo,
		Rewards: rewards,
	}, nil
}

func (s *Service) SingleBattle(ctx *ctx.Context, req *servicepb.Tower_SingleBattleRequest) (*servicepb.Tower_SingleBattleResponse, *errmsg.ErrMsg) {
	role, err := s.GetRoleModelByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	cnf, ok := rule.GetTestSingleBattleById(ctx, req.Id)
	if !ok {
		return nil, errmsg.NewErrTowerConfig()
	}

	heroesFormation, err := s.Module.FormationService.GetDefaultHeroes(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	heroIds := make([]int64, 0, 2)
	if heroesFormation.Hero_0 > 0 && heroesFormation.HeroOrigin_0 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_0)
	}

	if heroesFormation.Hero_1 > 0 && heroesFormation.HeroOrigin_1 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_1)
	}

	heroes, err := s.Module.GetHeroes(ctx, ctx.RoleId, heroIds)
	if err != nil {
		return nil, err
	}
	equips, err := s.GetManyEquipBagMap(ctx, ctx.RoleId, s.GetHeroesEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	cppHeroes := trans.Heroes2CppHeroes(ctx, heroes, equips)
	if len(cppHeroes) == 0 {
		return nil, errmsg.NewErrHeroNotFound()
	}

	for _, h := range cppHeroes {
		if len(h.SkillIds) == 0 {
			return nil, errmsg.NewInternalErr("SkillIds empty")
		}
	}

	// cppHeroes := make([]*models.HeroForBattle, 0, len(heroes))
	// for _, h := range heroes {
	// 	hero := &models.HeroForBattle{}
	// 	equip := make(map[int64]int64)
	// 	for slot, item := range h.EquipSlot {
	// 		if item.Equip == nil {
	// 			equip[slot] = -1
	// 			continue
	// 		}
	// 		equip[slot] = item.Equip.ItemId
	// 	}
	// 	hero.SkillIds = h.Skill
	// 	hero.Equip = equip
	// 	hero.Attr = h.Attrs
	// 	hero.ConfigId = h.Id
	// 	cppHeroes = append(cppHeroes, hero)
	// }
	d, err00 := db.GetBattleSetting(ctx)
	if err00 != nil {
		return nil, err00
	}
	return &servicepb.Tower_SingleBattleResponse{
		Data: &models.SingleBattleParam{
			Role:             role,
			Heroes:           cppHeroes,
			CountDown:        cnf.Duration,
			MonsterGroupInfo: cnf.MonsterGroupInfo,
			AutoSoulSkill:    d.Data.AutoSoulSkill,
		},
		BattleId: req.BattleId,
	}, nil
}

func (s *Service) BattleStart(ctx *ctx.Context, req *servicepb.Tower_TowerBattleInfoRequest) (*servicepb.Tower_TowerBattleInfoResponse, *errmsg.ErrMsg) {
	if req.Type == models.TowerType_TT_None {
		return nil, errmsg.NewErrTowerNoFindType()
	}
	role, err := s.GetRoleModelByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	heroesFormation, err := s.Module.FormationService.GetDefaultHeroes(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	if s.HasFight(req.Type, ctx.RoleId) {
		return nil, errmsg.NewErrTowerBattleStart()
	}

	heroIds := make([]int64, 0, 2)
	if heroesFormation.Hero_0 > 0 && heroesFormation.HeroOrigin_0 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_0)
	}
	if heroesFormation.Hero_1 > 0 && heroesFormation.HeroOrigin_1 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_1)
	}
	heroes, err := s.Module.GetHeroes(ctx, ctx.RoleId, heroIds)
	if err != nil {
		return nil, err
	}
	equips, err := s.GetManyEquipBagMap(ctx, ctx.RoleId, s.GetHeroesEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	cppHeroes := trans.Heroes2CppHeroes(ctx, heroes, equips)
	if len(cppHeroes) == 0 {
		return nil, errmsg.NewErrHeroNotFound()
	}

	for _, h := range cppHeroes {
		if len(h.SkillIds) == 0 {
			return nil, errmsg.NewInternalErr("SkillIds empty")
		}
	}

	// cppHeroes := make([]*models.HeroForBattle, 0, len(heroes))
	// for _, h := range heroes {
	// 	hero := &models.HeroForBattle{}
	// 	equip := make(map[int64]int64)
	// 	for slot, item := range h.EquipSlot {
	// 		if item.Equip == nil {
	// 			equip[slot] = -1
	// 			continue
	// 		}
	// 		equip[slot] = item.Equip.ItemId
	// 	}
	// 	hero.SkillIds = h.Skill
	// 	hero.Equip = equip
	// 	hero.Attr = h.Attrs
	// 	hero.ConfigId = h.Id
	// 	cppHeroes = append(cppHeroes, hero)
	// }

	towerCalculateParam, err := s.GetTowerRewardCalculateParam(ctx, req.Type)
	if err != nil {
		return nil, err
	}

	if towerCalculateParam.towerData.CurrentTowerLevel != req.CurrentTowerLevel {
		return nil, errmsg.NewErrTowerBattleLevel()
	}

	// if int(tower_calculate_param.tower_data.ChallengeData.UseChallengeTimes) > tower_calculate_param.config.Challenge_limit {
	//	return nil, errmsg.NewErrTowerChallengeTimesNotEnough()
	// }

	monsterInfo, err := s.GetMonsterInfo(towerCalculateParam)
	if err != nil {
		return nil, err
	}

	countDown, err := s.GetCountDown(towerCalculateParam)
	if err != nil {
		return nil, err
	}

	s.SetFight(req.Type, ctx.RoleId)
	return &servicepb.Tower_TowerBattleInfoResponse{
		Data: &models.SingleBattleParam{
			Role:             role,
			Heroes:           cppHeroes,
			MonsterGroupInfo: *monsterInfo,
			CountDown:        countDown,
		},
		BattleId:          req.BattleId,
		CurrentTowerLevel: towerCalculateParam.towerData.CurrentTowerLevel,
	}, nil
}

func (s *Service) TowerChallengeSettlement(ctx *ctx.Context, req *servicepb.Tower_TowerChallengeResultPrcRequest) (*servicepb.Tower_TowerChallengeResultPrcResponse, *errmsg.ErrMsg) {
	towerCalculateParam, err := s.GetTowerRewardCalculateParam(ctx, req.Type)
	if err != nil {
		return nil, err
	}
	startTime := int64(0)
	if req.WinType == models.TowerWinType_TW_Pass {
		// 只对正常战斗进行埋点
		startTime = s.DelFight(req.Type, ctx.RoleId)
	}

	useTime := timer.Now().Unix() - startTime
	if startTime == 0 {
		useTime = 0
	}

	curLevel := towerCalculateParam.towerData.CurrentTowerLevel
	if req.IsWin {
		passLevel := req.PassLevel

		if passLevel < towerCalculateParam.towerData.CurrentTowerLevel {
			return nil, errmsg.NewErrTowerLevelPassed()
		}

		maxLevel := int64(s.GetMaxLevel(ctx, req.Type))
		if passLevel > maxLevel {
			passLevel = maxLevel
		}

		err = s.CheckChallengeSettlement(ctx, passLevel, towerCalculateParam)
		if err != nil {
			return nil, err
		}

		ret, err := s.CalculationReward(ctx, passLevel, towerCalculateParam)
		if err != nil {
			return nil, err
		}

		data, err := s.GetLevelInfo(ctx, req.Type)
		if err != nil {
			s.log.Error("TowerChallengeSettlement GetLevelInfo error", zap.Any("error code", err))
		}

		for i := curLevel; i <= passLevel; i++ {
			// 埋点
			statistical.Save(ctx.NewLogServer(), &models2.Tower{
				IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
				EventTime: timer.Now(),
				GwId:      statistical.GwId(),
				RoleId:    ctx.RoleId,
				Type:      int64(req.Type),
				Level:     req.PassLevel,
				UseTime:   useTime,
				IsWin:     1,
			})
		}
		addLevel := passLevel - curLevel + 1
		if addLevel > 0 {
			err = s.Module.JourneyService.AddToken(ctx, ctx.RoleId, valuesJourney.JourneyTower, addLevel)
			if err != nil {
				return nil, err
			}
			s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskTowerChallengeSucc, 0, addLevel)
		}

		ret.Data = data
		ret.WinType = req.WinType
		ret.Type = req.Type
		return ret, nil
	}
	// 埋点
	statistical.Save(ctx.NewLogServer(), &models2.Tower{
		IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
		EventTime: timer.Now(),
		GwId:      statistical.GwId(),
		RoleId:    ctx.RoleId,
		Type:      int64(req.Type),
		Level:     req.PassLevel,
		UseTime:   useTime,
		IsWin:     0,
	})
	return &servicepb.Tower_TowerChallengeResultPrcResponse{}, nil
}

// Module ================================================================================================================================================================
func (s *Service) GetTowerCnt(c *ctx.Context) ([2]values.Integer, *errmsg.ErrMsg) {
	var res [2]values.Integer
	towerCalculateParam, err := s.GetTowerRewardCalculateParam(c, models.TowerType_TT_Default)
	if err != nil {
		return res, err
	}
	CheckRefreshChallengeData(c, towerCalculateParam)

	challengeLimit, ok := sys_rule.MustGetReader(c).KeyValue.GetInt64("DefaultTowerBattleLimit")
	if !ok {
		return res, errmsg.NewErrTowerNoFindType()
	}
	data, _, err := s.GetTowerInfo(c, models.TowerType_TT_Default)
	if err != nil {
		return res, err
	}
	res[0] = data.ChallengeData.UseChallengeTimes
	res[1] = challengeLimit
	return res, nil
}

// assist func ============================================================================================================================================================================================================
func (s *Service) GetMonsterInfo(towerCalculateParam *TowerRewardCalculateParam) (*map[values.Integer]values.Integer, *errmsg.ErrMsg) {
	if towerCalculateParam.towerData.Type == models.TowerType_TT_Default {
		config, err := tower_rule.GetTowerDefalutById(towerCalculateParam.ctx, towerCalculateParam.towerData.CurrentTowerLevel)
		if !err {
			return nil, errmsg.NewErrTowerConfig()
		}
		return &config.MonsterGroupInfo, nil
	}
	return nil, errmsg.NewErrTowerConfig()
}

func (s *Service) GetCountDown(towerCalculateParam *TowerRewardCalculateParam) (int64, *errmsg.ErrMsg) {
	if towerCalculateParam.towerData.Type == models.TowerType_TT_Default {
		config, err := tower_rule.GetTowerDefalutById(towerCalculateParam.ctx, towerCalculateParam.towerData.CurrentTowerLevel)
		if !err {
			return 0, errmsg.NewErrTowerConfig()
		}
		return config.Duration, nil
	}
	return 0, errmsg.NewErrTowerConfig()
}

func (s *Service) GetLevelInfo(ctx *ctx.Context, cType models.TowerType) (*models.Tower, *errmsg.ErrMsg) {
	towerCalculateParam, err := s.GetTowerRewardCalculateParam(ctx, cType)
	if err != nil {
		return nil, err
	}
	nextFreshTime := timer.GetNextRefreshTime(towerCalculateParam.config.Tower_refresh_time, 0)
	localData, _, err := s.GetTowerInfo(ctx, cType)
	if err != nil {
		return nil, err
	}

	isAllPass, allLevel, err := s.IsAllPass(ctx, cType, int32(localData.CurrentTowerLevel))
	if err != nil {
		return nil, err
	}

	CheckRefreshChallengeData(ctx, towerCalculateParam)
	registerDay, err := s.GetRegisterDay(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	maxLevel := registerDay * int64(towerCalculateParam.config.Challenge_limit)
	if maxLevel > int64(allLevel) {
		maxLevel = int64(allLevel)
	}
	currLevel := localData.CurrentTowerLevel
	if currLevel > int64(allLevel) {
		currLevel = int64(allLevel)
	}
	return &models.Tower{
		CurrentTowerLevel: currLevel,
		IsAllPass:         isAllPass,
		UseChallengeTimes: towerCalculateParam.towerData.ChallengeData.UseChallengeTimes,
		MaxChallengeTimes: maxLevel,
		NextRefreshTime:   nextFreshTime,
	}, nil
}

func (s *Service) CheckChallengeSettlement(ctx *ctx.Context, passLevel int64, towerCalculateParam *TowerRewardCalculateParam) *errmsg.ErrMsg {
	CheckRefreshChallengeData(ctx, towerCalculateParam)
	// passNum := passLevel - (towerCalculateParam.tower_data.CurrentTowerLevel - 1)
	// if int(towerCalculateParam.tower_data.ChallengeData.UseChallengeTimes+passNum) > towerCalculateParam.config.Challenge_limit {
	//	return errmsg.NewErrTowerChallengeTimesNotEnough()
	// }
	isAllPass, _, err := s.IsAllPass(ctx, towerCalculateParam.towerData.Type, int32(towerCalculateParam.towerData.CurrentTowerLevel))
	if err != nil {
		return err
	}
	if isAllPass {
		return errmsg.NewErrTowerIsMaxLevel()
	}
	return nil
}

func (s *Service) LevelUp(ctx *ctx.Context, passLevel int64, towerCalculateParam *TowerRewardCalculateParam) *errmsg.ErrMsg {
	towerCalculateParam.towerData.ChallengeData.UseChallengeTimes += passLevel - (towerCalculateParam.towerData.CurrentTowerLevel - 1)
	towerCalculateParam.towerData.CurrentTowerLevel = passLevel + 1
	towerCalculateParam.towerDaoData.SaveData(ctx)
	s.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskTower, 0, passLevel, true)
	return nil
}

func (s *Service) CalculationReward(ctx *ctx.Context, passLevel int64, towerCalculateParam *TowerRewardCalculateParam) (*servicepb.Tower_TowerChallengeResultPrcResponse, *errmsg.ErrMsg) {
	err := s.CacheAccumulateHarvest(ctx, towerCalculateParam)
	if err != nil {
		return nil, err
	}

	rewardBase, err := s.GetLevelReward(passLevel, towerCalculateParam)
	if err != nil {
		return nil, err
	}

	rewards := rewardBase.stageItem

	for _, item := range rewards {
		err = s.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
		if err != nil {
			return nil, err
		}
	}

	err = s.LevelUp(ctx, passLevel, towerCalculateParam)
	if err != nil {
		return nil, err
	}

	return &servicepb.Tower_TowerChallengeResultPrcResponse{
		Rewards: rewards,
	}, nil
}

func (s *Service) CacheAccumulateHarvest(ctx *ctx.Context, towerCalculateParam *TowerRewardCalculateParam) *errmsg.ErrMsg {
	timeInterval := timer.Now().Unix() - towerCalculateParam.towerData.CacheInfo.LastSettlementTime - towerCalculateParam.towerData.CacheInfo.SettlementUseTime
	if towerCalculateParam.towerData.CacheInfo.SettlementUseTime+timeInterval > int64(towerCalculateParam.config.Accumulate_max) {
		timeInterval = int64(towerCalculateParam.config.Accumulate_max) - towerCalculateParam.towerData.CacheInfo.SettlementUseTime
	}

	number := int64(float64(timeInterval) / float64(towerCalculateParam.config.Settlement_interval))
	towerCalculateParam.towerData.CacheInfo.SettlementUseTime += number * int64(towerCalculateParam.config.Settlement_interval)

	if number == 0 {
		return nil
	}

	var tempMap = make(map[int64]*models.Item, len(towerCalculateParam.towerData.CacheInfo.ProfitsData))
	for _, item := range towerCalculateParam.towerData.CacheInfo.ProfitsData {
		_, ok := tempMap[item.ItemId]
		if ok {
			tempMap[item.ItemId].Count += item.Count
			continue
		}
		tempMap[item.ItemId] = item
	}

	for _, item := range towerCalculateParam.rewardBase.item {
		_, ok := tempMap[item.ItemId]
		if ok {
			tempMap[item.ItemId].Count += item.Count * number
			continue
		}
		tempMap[item.ItemId] = &models.Item{
			ItemId: item.ItemId,
			Count:  item.Count * number,
		}
	}

	towerCalculateParam.towerData.CacheInfo.ProfitsData = towerCalculateParam.towerData.CacheInfo.ProfitsData[:0]
	for _, item := range tempMap {
		towerCalculateParam.towerData.CacheInfo.ProfitsData = append(towerCalculateParam.towerData.CacheInfo.ProfitsData, item)
	}
	return nil
}

func (s *Service) ProcMeditationCostTimes(ctx *ctx.Context, towerCalculateParam *TowerRewardCalculateParam, meditationData *MeditationData, accumulateInfo *models.TowerAccumulateHarvestInfo) {
	if meditationData.isFree {
		// tower_calculate_param.tower_data.MeditationData.UseFreeMeditationTimes++
		accumulateInfo.LeftFreeMeditationTimes--
	} else {
		// tower_calculate_param.tower_data.MeditationData.UseCostMeditationTimes++
		accumulateInfo.UseCostMeditationTimes++
	}
	// tower_calculate_param.tower_dao_data.SaveData()
}

func (s *Service) ProcMeditationCost(ctx *ctx.Context, meditationData *MeditationData) *errmsg.ErrMsg {
	if meditationData.isFree {
		return nil
	}

	err := s.BagService.SubManyItem(ctx, ctx.RoleId, meditationData.cost)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) IsAllPass(ctx *ctx.Context, cType models.TowerType, currentLevel int32) (bool, int32, *errmsg.ErrMsg) {
	if cType == models.TowerType_TT_Default {
		return currentLevel > tower_rule.GetTowerDefalutMaxLevel(ctx), tower_rule.GetTowerDefalutMaxLevel(ctx), nil
	}
	s.log.Error("IsAllPass not find Tower Type", zap.Any("type", cType))
	return true, 0, nil
}

func (s *Service) GetMaxLevel(ctx *ctx.Context, cType models.TowerType) int32 {
	if cType == models.TowerType_TT_Default {
		return tower_rule.GetTowerDefalutMaxLevel(ctx)
	}
	s.log.Error("IsAllPass not find Tower Type", zap.Any("type", cType))
	return 0
}

func (s *Service) GetTowerConfig(ctx *ctx.Context, cType models.TowerType) (*tower_rule.TowerPublicConfig, *errmsg.ErrMsg) {
	if cType == models.TowerType_TT_Default {
		config := tower_rule.GetTowerDefalutPublicInfo(ctx)
		if config == nil {
			s.log.Error("GetTowerConfig Error")
			return nil, errmsg.NewErrTowerConfig()
		}
		return config, nil
	}
	s.log.Error("GetTowerAccumulateMax not find Tower Type", zap.Any("type", cType))
	return nil, errmsg.NewErrTowerConfig()
}

func (s *Service) GetLevelReward(level int64, towerCalculateParam *TowerRewardCalculateParam) (*RewardBase, *errmsg.ErrMsg) {
	var rewardBase *RewardBase = new(RewardBase)
	if towerCalculateParam.towerData.Type == models.TowerType_TT_Default {
		var ok bool
		var defaultConfig *rulemodel.TowerDefault = new(rulemodel.TowerDefault)
		var rewardMap map[int64]*models.Item = make(map[int64]*models.Item)

		CurrentTowerLevel := towerCalculateParam.towerData.CurrentTowerLevel
		isAllPass, maxLevel, err := s.IsAllPass(ctx.GetContext(), towerCalculateParam.towerData.Type, int32(towerCalculateParam.towerData.CurrentTowerLevel))
		if err != nil {
			return nil, err
		}
		if isAllPass {
			CurrentTowerLevel = int64(maxLevel)
			level = int64(maxLevel)
		}

		for startLevel := CurrentTowerLevel; startLevel <= level; startLevel++ {
			defaultConfig, ok = tower_rule.GetTowerDefalutById(towerCalculateParam.ctx, startLevel)
			if !ok {
				s.log.Error("GetLevelReward TowerDefalu not find level", zap.Any("level", towerCalculateParam.towerData.CurrentTowerLevel))
				return nil, errmsg.NewErrTowerConfig()
			}
			for itemId, itemNum := range defaultConfig.StageReward {
				data, ok := rewardMap[itemId]
				if ok {
					data.Count += itemNum
					continue
				}
				rewardMap[itemId] = &models.Item{
					ItemId: itemId,
					Count:  itemNum,
				}
			}
		}

		if len(rewardMap) > 0 {
			for _, item := range rewardMap {
				rewardBase.stageItem = append(rewardBase.stageItem, item)
			}
		}

		for itemId, itemNum := range defaultConfig.AccumulateReward {
			rewardBase.item = append(rewardBase.item, &models.Item{
				ItemId: itemId,
				Count:  itemNum,
			})
		}

		rewardBase.intervalTime = towerCalculateParam.config.Settlement_interval
		return rewardBase, nil
	}
	return nil, errmsg.NewErrTowerNoFindType()
}

func (s *Service) GetTowerAccumulateReward(towerCalculateParam *TowerRewardCalculateParam) (*RewardBase, *errmsg.ErrMsg) {
	return s.GetLevelReward(towerCalculateParam.towerData.CurrentTowerLevel, towerCalculateParam)
}

func (s *Service) GetTowerInfo(ctx *ctx.Context, cType models.TowerType) (*models.TowerData, dao.TowerI, *errmsg.ErrMsg) {
	daoTower, err := dao.GetTowerData(ctx, ctx.RoleId)
	if err != nil {
		return nil, nil, err
	}
	for _, towerData := range daoTower.TowerData().GetData() {
		if towerData.Type == cType {
			return towerData, daoTower, nil
		}
	}
	return nil, nil, errmsg.NewErrTowerNoData()
}

func (s *Service) ProcRewardItem(towerCalculateParam *TowerRewardCalculateParam) (*models.TowerAccumulateHarvestInfo, *errmsg.ErrMsg) {
	var accumulateInfo *models.TowerAccumulateHarvestInfo = &models.TowerAccumulateHarvestInfo{
		ProfitsData: towerCalculateParam.towerData.CacheInfo.ProfitsData,
	}
	maxTime := towerCalculateParam.config.Accumulate_max
	nowTime := timer.Unix()
	timeInterval := nowTime - towerCalculateParam.towerData.CacheInfo.LastSettlementTime - towerCalculateParam.towerData.CacheInfo.SettlementUseTime
	accumulateInfo.NowHandleupAllTime = (timeInterval + towerCalculateParam.towerData.CacheInfo.SettlementUseTime) * 1000
	accumulateInfo.NextSettlementTime = nowTime + (int64(towerCalculateParam.rewardBase.intervalTime) - timeInterval%int64(towerCalculateParam.rewardBase.intervalTime))
	accumulateInfo.SettlementInterval = int64(towerCalculateParam.rewardBase.intervalTime)
	accumulateInfo.ProfitsItem = towerCalculateParam.rewardBase.item

	accumulateInfo.LeftFreeMeditationTimes = int32(towerCalculateParam.config.Free_meditation_times) - towerCalculateParam.towerData.MeditationData.UseFreeMeditationTimes
	accumulateInfo.UseCostMeditationTimes = towerCalculateParam.towerData.MeditationData.UseCostMeditationTimes
	accumulateInfo.NextMeditationRefreshTime = timer.GetNextRefreshTime(towerCalculateParam.config.Tower_refresh_time, 0)

	if timeInterval+towerCalculateParam.towerData.CacheInfo.SettlementUseTime > int64(maxTime) {
		timeInterval = int64(maxTime) - towerCalculateParam.towerData.CacheInfo.SettlementUseTime
		accumulateInfo.IsMax = true
		accumulateInfo.NextSettlementTime = 0
	}

	settlementTimes := timeInterval / int64(towerCalculateParam.rewardBase.intervalTime)
	accumulateInfo.ModTime = timeInterval % int64(towerCalculateParam.rewardBase.intervalTime)

	var tempMap = make(map[int64]*models.Item, len(accumulateInfo.ProfitsData))
	for _, value := range accumulateInfo.ProfitsData {
		_, ok := tempMap[value.ItemId]
		if ok {
			tempMap[value.ItemId].Count += value.Count
			continue
		}
		tempMap[value.ItemId] = &models.Item{
			ItemId: value.ItemId,
			Count:  value.Count,
		}
	}

	if settlementTimes > 0 {
		for _, value := range towerCalculateParam.rewardBase.item {
			if data, ok := tempMap[value.ItemId]; ok {
				data.Count += value.Count * settlementTimes
				continue
			}
			tempMap[value.ItemId] = &models.Item{
				ItemId: value.ItemId,
				Count:  value.Count * settlementTimes,
			}
		}
	}

	accumulateInfo.ProfitsData = accumulateInfo.ProfitsData[:0]
	for _, item := range tempMap {
		accumulateInfo.ProfitsData = append(accumulateInfo.ProfitsData, item)
	}

	return accumulateInfo, nil
}

func CheckRefreshMeditation(ctx *ctx.Context, towerCalculateParam *TowerRewardCalculateParam) {
	isNewDay := timer.OverRefreshTime(towerCalculateParam.towerData.MeditationData.LastMeditationTime, towerCalculateParam.config.Tower_refresh_time, 0)
	if isNewDay {
		towerCalculateParam.towerDaoData.RefreshMeditationData(ctx)
	}
}

func CheckRefreshChallengeData(ctx *ctx.Context, towerCalculateParam *TowerRewardCalculateParam) {
	isNewDay := timer.OverRefreshTime(towerCalculateParam.towerData.ChallengeData.LastChallengeTime, towerCalculateParam.config.Tower_refresh_time, 0)
	if isNewDay {
		towerCalculateParam.towerDaoData.RefreshChallenge(ctx)
	}
}

func (s *Service) CalculateFreeCurrentAccumulateReward(ctx *ctx.Context, towerCalculateParam *TowerRewardCalculateParam) (*models.TowerAccumulateHarvestInfo, *errmsg.ErrMsg) {
	CheckRefreshMeditation(ctx, towerCalculateParam)

	var err *errmsg.ErrMsg
	var accumulateInfo *models.TowerAccumulateHarvestInfo
	accumulateInfo, err = s.ProcRewardItem(towerCalculateParam)

	if err != nil {
		return nil, err
	}
	return accumulateInfo, nil
}

func (s *Service) CalculateCostCurrentAccumulateReward(towerCalculateParam *TowerRewardCalculateParam) (*MeditationData, *errmsg.ErrMsg) {
	var meditationData = &MeditationData{
		isFree: false,
		cost:   make(map[int64]int64),
	}
	if towerCalculateParam.towerData.MeditationData.UseFreeMeditationTimes < int32(towerCalculateParam.config.Free_meditation_times) {
		meditationData.isFree = true
	}

	if !meditationData.isFree {
		for _, costData := range towerCalculateParam.config.Cost_meditation_times {
			if costData[0] == int64(towerCalculateParam.towerData.MeditationData.UseCostMeditationTimes+1) {
				meditationData.cost[towerCalculateParam.config.Cost_Item] = costData[1]
				break
			}
		}
	}

	if (meditationData.isFree && len(meditationData.cost) != 0) || (!meditationData.isFree && len(meditationData.cost) == 0) {
		return nil, errmsg.NewErrTowerMeditationTimesNotEnough()
	}

	for _, value := range towerCalculateParam.rewardBase.item {
		meditationData.reward = append(meditationData.reward, &models.Item{
			ItemId: value.ItemId,
			Count:  value.Count * int64(towerCalculateParam.config.Meditation_time/towerCalculateParam.rewardBase.intervalTime),
		})
	}

	return meditationData, nil
}

func (s *Service) GetTowerRewardCalculateParam(ctx *ctx.Context, cType models.TowerType) (*TowerRewardCalculateParam, *errmsg.ErrMsg) {
	var ret = &TowerRewardCalculateParam{
		ctx: ctx,
	}
	var err *errmsg.ErrMsg
	ret.towerData, ret.towerDaoData, err = s.GetTowerInfo(ctx, cType)

	if err != nil {
		return nil, err
	}

	ret.config, err = s.GetTowerConfig(ctx, cType)
	if err != nil {
		return nil, err
	}

	ret.rewardBase, err = s.GetTowerAccumulateReward(ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *Service) SetFight(cType models.TowerType, roleId string) {
	key := strconv.FormatInt(int64(cType), 10) + ":" + roleId
	// s.fight_map[key] = timer.Now().Unix()
	s.fightMap.Store(key, timer.Now().Unix())
}

func (s *Service) HasFight(cType models.TowerType, roleId string) bool {
	key := strconv.FormatInt(int64(cType), 10) + ":" + roleId
	vI, ok := s.fightMap.Load(key)
	if !ok {
		return false
	}
	v := vI.(int64)

	return timer.Now().Unix()-v <= 5
}

func (s *Service) DelFight(cType models.TowerType, roleId string) int64 {
	key := strconv.FormatInt(int64(cType), 10) + ":" + roleId
	startTime := int64(0)
	vI, ok := s.fightMap.Load(key)
	if ok {
		startTime = vI.(int64)
		s.fightMap.Delete(key)
	}
	return startTime
}

func (s *Service) GetRecommendNum(towerCalculateParam *TowerRewardCalculateParam) (int64, *errmsg.ErrMsg) {
	if towerCalculateParam.towerData.Type == models.TowerType_TT_Default {
		config, err := tower_rule.GetTowerDefalutById(towerCalculateParam.ctx, towerCalculateParam.towerData.CurrentTowerLevel)
		if !err {
			return 0, errmsg.NewErrTowerConfig()
		}
		return config.RecommendNum, nil
	}
	return 0, errmsg.NewErrTowerConfig()
}
