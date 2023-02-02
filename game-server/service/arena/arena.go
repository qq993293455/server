package arena

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"coin-server/game-server/service/user/db"
	rulemodel "coin-server/rule/rule-model"

	"coin-server/common/ArenaRule"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/iggsdk"
	"coin-server/common/logger"
	"coin-server/common/proto/arena_service"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/redisclient"
	"coin-server/common/service"
	"coin-server/common/statistical"
	models2 "coin-server/common/statistical/models"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	arenaDao "coin-server/game-server/service/arena/dao"
	arenaRule "coin-server/game-server/service/arena/rule"
	guildDao "coin-server/game-server/service/guild/dao"
	heroDao "coin-server/game-server/service/hero/dao"
	heroRule "coin-server/game-server/service/hero/rule"
	valuesJourney "coin-server/game-server/service/journey/values"
	"coin-server/game-server/util"
	"coin-server/game-server/util/trans"

	"go.uber.org/zap"
)

const ArenaEvenKey = "ArenaDefault"

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	log *logger.Logger
	// challengeMap     map[int32]map[string]string
	// challengeLockMap map[int32]map[string]int64
	challengeMap     sync.Map
	challengeLockMap sync.Map
	fightLockMap     sync.Map
}

func NewArenaService(
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
		// challengeMap:     make(map[int32]map[string]string),
		// challengeLockMap: make(map[int32]map[string]int64),
	}
	module.ArenaService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("出战英雄设置", this_.SetArenaHeroSetRequest)
	this_.svc.RegisterFunc("排行查询", this_.GetArenaRankingRequest)
	this_.svc.RegisterFunc("自己排行查询", this_.GetArenaSelfRankingRequest)
	this_.svc.RegisterFunc("查询对方出战英雄", this_.GetArenaQueryFightHeroRequest)
	this_.svc.RegisterFunc("获取挑战队列", this_.GetArenaGetChallengeRequest)
	this_.svc.RegisterFunc("挑战", this_.ArenaChallenge)
	this_.svc.RegisterFunc("挑战结果", this_.ArenaChallengeResult)
	this_.svc.RegisterFunc("购买挑战票", this_.ArenaBuyChallengeTicket)
	this_.svc.RegisterFunc("获取挑战记录", this_.ArenaGetFightLog)
	this_.svc.RegisterFunc("检查奖励", this_.CheckReward)
	this_.svc.RegisterFunc("切搓", this_.VirtualCombat)
	this_.svc.RegisterFunc("获取发奖时间", this_.GetRewardTime)
	eventlocal.SubscribeEventLocal(this_.HandleRoleLogoutEvent)
	eventlocal.SubscribeEventLocal(this_.HandleRoleLoginEvent)
}

// GetArenaResetRemainSec 获取竞技场下次结算时间
func (this_ *Service) GetArenaResetRemainSec(c *ctx.Context) values.Integer {
	return util.DefaultNextRefreshTime().Unix() - timer.Now().Unix()
}

// msg proc ============================================================================================================================================================================================================
func (this_ *Service) SetArenaHeroSetRequest(ctx *ctx.Context, req *servicepb.Arena_ArenaHeroSetRequest) (*servicepb.Arena_ArenaHeroSetResponse, *errmsg.ErrMsg) {
	fightHero := &models.Assemble{
		HeroOrigin_0: req.Hero0_OriginId,
		HeroOrigin_1: req.Hero1_OriginId,
	}

	fightHero, err := this_.GetHero(ctx, ctx.RoleId, fightHero)
	if err != nil {
		return nil, err
	}

	err = this_.SetHero(ctx, req.Type, fightHero)
	if err != nil {
		return nil, err
	}

	return &servicepb.Arena_ArenaHeroSetResponse{
		Type:      req.Type,
		Assembles: fightHero,
	}, nil
}

func (this_ *Service) GetArenaRankingRequest(ctx *ctx.Context, req *servicepb.Arena_ArenaRankingRequest) (*servicepb.Arena_ArenaRankingResponse, *errmsg.ErrMsg) {
	_, data, err := this_.GetPlayerArenaData(ctx, ctx.RoleId, req.Type)
	if err != nil {
		return nil, err
	}

	arenaInfo, err := this_.SelectRankInfo(ctx, req.Type, data.RankingIndex, int64(req.StartIndex), int64(req.StartIndex+req.Count))
	if err != nil {
		return nil, err
	}

	return &servicepb.Arena_ArenaRankingResponse{
		Type:        req.Type,
		StartIndex:  req.StartIndex,
		RankingInfo: arenaInfo,
	}, nil
}

func (this_ *Service) GetArenaSelfRankingRequest(ctx *ctx.Context, req *servicepb.Arena_ArenaSelfRankingRequest) (*servicepb.Arena_ArenaSelfRankingResponse, *errmsg.ErrMsg) {
	arenaData, isNew := arenaDao.GetPlayerArenaData(ctx, ctx.RoleId)

	data, ok := arenaData.GetData().Data[int32(req.Type)]
	if !ok || data.RankingIndex == "" {
		rankingIndex, err := this_.JoinArenaRanking(ctx, req.Type)
		if err != nil {
			return nil, err
		}
		freeChallengeTime, err := this_.GetArenaFreeNum(req.Type)
		if err != nil {
			return nil, err
		}

		ticketConfigInfo, err := this_.GetArenaTicketCost(req.Type)
		if err != nil {
			return nil, err
		}

		data = &models.ArenaData{
			RankingIndex:       rankingIndex,
			LastRefreshTime:    timer.Now().Unix(),
			FreeChallengeTimes: freeChallengeTime,
			FightHero: &models.Assemble{
				HeroOrigin_0: 0,
				HeroOrigin_1: 0,
			},
			LeftTicketPurchasesNum: ticketConfigInfo[0],
		}

		arenaData.GetData().Data[int32(req.Type)] = data
		arenaData.Save()

		arenaKey := GetEventKey(req.Type)
		ctx.PublishEventLocal(&event.RedPointChange{
			RoleId: ctx.RoleId,
			Key:    arenaKey,
			Val:    data.FreeChallengeTimes,
		})
	}

	if isNew {
		ticketConfigInfo, err := this_.GetArenaTicketCost(req.Type)
		if err != nil {
			return nil, err
		}
		ok := arenaData.SetleftTicketPurchasesNumber(req.Type, ticketConfigInfo[0])
		if !ok {
			return nil, errmsg.NewErrArenaType()
		}
	}

	this_.CheckAndRefreshData(ctx, req.Type, arenaData, data)

	fightHero := data.FightHero
	var err *errmsg.ErrMsg
	if fightHero.Hero_0 == 0 && fightHero.Hero_1 == 0 {
		fightHero, err = this_.Module.FormationService.GetDefaultHeroes(ctx, ctx.RoleId)
		if err != nil {
			return nil, err
		}
	}

	fightHero, err = this_.GetHero(ctx, ctx.RoleId, fightHero)
	if err != nil {
		return nil, err
	}

	ret, err := this_.GetSelfRankingInfo(ctx, req.Type, data.RankingIndex)
	if err != nil {
		if err.ErrCode == errmsg.NewErrArenaRankingIndex().ErrCode {
			data.RankingIndex = ""
			arenaData.Save()
		}
		return nil, err
	}
	ret.LeftTicketPurchasesNum = data.LeftTicketPurchasesNum
	ret.FreeChallengeTimes = data.FreeChallengeTimes
	ret.Assembles = fightHero
	ret.Power = fightHero.Hero_0Power + fightHero.Hero_1Power

	return ret, nil
}

func (this_ *Service) GetArenaQueryFightHeroRequest(ctx *ctx.Context, req *servicepb.Arena_ArenaQueryFightHeroRequest) (*servicepb.Arena_ArenaQueryFightHeroResponse, *errmsg.ErrMsg) {
	_, data, err := this_.GetPlayerArenaData(ctx, ctx.RoleId, req.Type)
	if err != nil {
		return nil, err
	}

	isRobot, _, err := this_.IsRobot(ctx, req.Type, data.RankingIndex, req.RoleId)
	if err != nil {
		return nil, err
	}
	if isRobot {
		heros, err := this_.GetRobotHeroInfo(ctx, req.RoleId)
		if err != nil {
			return nil, err
		}

		return &servicepb.Arena_ArenaQueryFightHeroResponse{
			Assembles: heros,
			Type:      req.Type,
		}, nil
	}

	_, otherData, err := this_.GetPlayerArenaData(ctx, req.RoleId, req.Type)
	if err != nil {
		return nil, err
	}

	if otherData.FightHero != nil {
		fHero, err := this_.GetHero(ctx, req.RoleId, otherData.FightHero)
		if err != nil {
			return nil, err
		}
		otherData.FightHero.Hero_0Power = fHero.GetHero_0Power()
		otherData.FightHero.Hero_1Power = fHero.GetHero_1Power()
		return &servicepb.Arena_ArenaQueryFightHeroResponse{
			Assembles: otherData.FightHero,
			Type:      req.Type,
		}, nil
	}

	fightHero, err := this_.Module.FormationService.GetDefaultHeroes(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}

	return &servicepb.Arena_ArenaQueryFightHeroResponse{
		Assembles: fightHero,
		Type:      req.Type,
	}, nil
}

func (this_ *Service) GetArenaGetChallengeRequest(ctx *ctx.Context, req *servicepb.Arena_ArenaGetChallengeRequest) (*servicepb.Arena_ArenaGetChallengeResponse, *errmsg.ErrMsg) {
	_, data, err := this_.GetPlayerArenaData(ctx, ctx.RoleId, req.Type)
	if err != nil {
		return nil, err
	}

	datas, err := this_.GetChallengeRange(ctx, req.Type, data.RankingIndex, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	return &servicepb.Arena_ArenaGetChallengeResponse{
		Type:        req.Type,
		RankingInfo: datas,
	}, nil
}

func (this_ *Service) ArenaChallenge(ctx *ctx.Context, req *servicepb.Arena_ArenaChallengeRequest) (*servicepb.Arena_ArenaChallengeResponse, *errmsg.ErrMsg) {
	arenaData, data, err := this_.GetPlayerArenaData(ctx, ctx.RoleId, req.Type)
	if err != nil {
		return nil, err
	}

	if this_.HasFight(req.Type, ctx.RoleId) {
		return nil, errmsg.NewErrArenaChallengeStart()
	}

	challengeMapI, ok := this_.challengeMap.Load(int32(req.Type))
	if ok {
		challengeMap := challengeMapI.(*sync.Map)
		challengedRoleId, ok := challengeMap.Load(ctx.RoleId)
		if ok {
			// 挑战失败 结算一次
			this_.ChallengeFinish(ctx, req.Type, data.RankingIndex, ctx.RoleId, challengedRoleId.(string), false, true)
		}
	} else {
		this_.challengeMap.Store(int32(req.Type), new(sync.Map))
	}

	battleDuration, err := this_.GetBattleDuration(req.Type)
	if err != nil {
		return nil, err
	}
	challengeLockMapI, ok := this_.challengeLockMap.Load(int32(req.Type))
	if ok {
		challengeLockMap := challengeLockMapI.(*sync.Map)
		challengeTime, ok := challengeLockMap.Load(req.RoleId)
		if ok {
			if timer.Now().Unix()-challengeTime.(int64) < battleDuration {
				return nil, errmsg.NewErrArenaChallengeLock()
			}
		}
	} else {
		this_.challengeLockMap.Store(int32(req.Type), new(sync.Map))
	}

	out := &arena_service.ArenaRanking_LockRoleRankingResponse{}
	if err := this_.svc.GetNatsClient().RequestWithOut(ctx, ArenaRule.GetArenaServer(req.Type), &arena_service.ArenaRanking_LockRoleRankingRequest{
		Type:                req.Type,
		ChallengeRoleid:     ctx.RoleId,
		ChallengedRoleid:    req.RoleId,
		RankingIndex:        data.RankingIndex,
		ChallengeRankingId:  req.SelfRankingId,
		ChallengedRankingId: req.RankingId,
	}, out); err != nil {
		return nil, err
	}

	role, err := this_.GetRoleModelByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	isFree := false
	if data.FreeChallengeTimes > 0 {
		isFree = true
	}

	myFightHeros := data.FightHero
	if myFightHeros.HeroOrigin_0 == 0 && myFightHeros.HeroOrigin_1 == 0 {
		myFightHeros, err = this_.Module.FormationService.GetDefaultHeroes(ctx, ctx.RoleId)
		if err != nil {
			return nil, err
		}
	}

	myFightHeros, err = this_.GetHero(ctx, ctx.RoleId, myFightHeros)
	if err != nil {
		return nil, err
	}

	data.FightHero = myFightHeros
	myCppHero, err := this_.GetCppHero(ctx, ctx.RoleId, myFightHeros)
	if err != nil {
		return nil, err
	}

	if len(myCppHero) == 0 {
		panic("hero empty  FightHeros:0" + strconv.FormatInt(int64(myFightHeros.Hero_0), 10) + "  FightHeros:1" + strconv.FormatInt(int64(myFightHeros.Hero_1), 10))
	}

	isRobot, nickName, err := this_.IsRobot(ctx, req.Type, data.RankingIndex, req.RoleId)
	if err != nil {
		return nil, err
	}

	var challengeHeros *models.Assemble
	var challengeRole *models.Role
	var challengeCppHeros []*models.HeroForBattle
	var robotCfg *rulemodel.PvPRobot
	if isRobot {
		robotCfg, challengeCppHeros, err = this_.GetCppRobotHeroInfo(ctx, req.RoleId)
		if err != nil {
			return nil, err
		}

		challengeRole = &models.Role{
			RoleId:   req.RoleId,
			Nickname: nickName,
			Level:    robotCfg.Lv,
			Title:    robotCfg.RoleLvTitle,
		}
	} else {
		_, challengedata, err := this_.GetPlayerArenaData(ctx, req.RoleId, req.Type)
		if err != nil {
			return nil, err
		}
		if challengedata.FightHero != nil {
			challengeHeros = challengedata.FightHero
		} else {
			challengeHeros, err = this_.Module.FormationService.GetDefaultHeroes(ctx, req.RoleId)
			if err != nil {
				return nil, err
			}
		}
		challengeRole, err = this_.GetRoleModelByRoleId(ctx, req.RoleId)
		if err != nil {
			return nil, err
		}

		challengeHeros, err = this_.GetHero(ctx, req.RoleId, challengeHeros)
		if err != nil {
			return nil, err
		}

		challengeCppHeros, err = this_.GetCppHero(ctx, req.RoleId, challengeHeros)
		if err != nil {
			return nil, err
		}
	}

	if isFree {
		data.FreeChallengeTimes--
		arenaKey := GetEventKey(req.Type)
		ctx.PublishEventLocal(&event.RedPointChange{
			RoleId: ctx.RoleId,
			Key:    arenaKey,
			Val:    data.FreeChallengeTimes,
		})
	} else {
		ticketItemId, err := this_.GetArenaTicket(req.Type)
		if err != nil {
			return nil, err
		}
		var cost map[int64]int64 = make(map[int64]int64)
		cost[ticketItemId] = 1
		err = this_.BagService.SubManyItem(ctx, ctx.RoleId, cost)
		if err != nil {
			return nil, errmsg.NewErrArenaChallengeTimes()
		}
	}

	challengeMapI, ok = this_.challengeMap.Load(int32(req.Type))
	if !ok {
		challengeMapI = new(sync.Map)
		this_.challengeMap.Store(int32(req.Type), challengeMapI)
	}
	challengeMapI.(*sync.Map).Store(ctx.RoleId, req.RoleId)

	challengeLockMapI, ok = this_.challengeLockMap.Load(int32(req.Type))
	if ok {
		challengeLockMapI = new(sync.Map)
		this_.challengeLockMap.Store(int32(req.Type), challengeLockMapI)
	}
	challengeLockMapI.(*sync.Map).Store(req.RoleId, timer.Now().Unix())

	this_.SetFight(req.Type, ctx.RoleId)
	arenaData.Save()

	d, err00 := db.GetBattleSetting(ctx)
	if err00 != nil {
		return nil, err00
	}

	return &servicepb.Arena_ArenaChallengeResponse{
		Type:     req.Type,
		RoleId:   req.RoleId,
		BattleId: req.BattleId,
		Data: &models.SingleBattleParam{
			Role:      role,
			Heroes:    myCppHero,
			CountDown: battleDuration,
			HostilePlayers: []*models.SinglePlayerInfo{
				0: {
					Role:   challengeRole,
					Heroes: challengeCppHeros,
				},
			},
			AutoSoulSkill: d.Data.AutoSoulSkill,
		},
	}, nil
}

func (this_ *Service) ArenaBuyChallengeTicket(ctx *ctx.Context, req *servicepb.Arena_ArenaBuyChallengeTicketRequest) (*servicepb.Arena_ArenaBuyChallengeTicketResponse, *errmsg.ErrMsg) {
	ticketConfigInfo, err := this_.GetArenaTicketCost(req.Type)
	if err != nil {
		return nil, err
	}

	if len(ticketConfigInfo) < 3 {
		return nil, errmsg.NewErrArenaConfig()
	}

	ticketItemId, err := this_.GetArenaTicket(req.Type)
	if err != nil {
		return nil, err
	}

	arenaData, data, err := this_.GetPlayerArenaData(ctx, ctx.RoleId, req.Type)
	if err != nil {
		return nil, err
	}

	num := int64(req.Number)
	if num <= 0 {
		return nil, errmsg.NewErrArenaTicketPurchaseNumber()
	}

	if num > data.LeftTicketPurchasesNum {
		return nil, errmsg.NewErrArenaTicketPurchaseLimit()
	}

	ticketInfo := &models.Item{
		ItemId: ticketItemId,
		Count:  num,
	}

	ticketCostInfo := map[int64]int64{ticketConfigInfo[1]: ticketConfigInfo[2] * num}

	err = this_.BagService.SubManyItem(ctx, ctx.RoleId, ticketCostInfo)
	if err != nil {
		return nil, err
	}

	err = this_.BagService.AddManyItemPb(ctx, ctx.RoleId, ticketInfo)
	if err != nil {
		return nil, err
	}

	data.LeftTicketPurchasesNum -= num
	arenaData.Save()

	return &servicepb.Arena_ArenaBuyChallengeTicketResponse{
		Type:                   req.Type,
		Item:                   ticketInfo,
		LeftTicketPurchasesNum: data.LeftTicketPurchasesNum,
	}, nil
}

func (this_ *Service) ArenaChallengeResult(ctx *ctx.Context, req *servicepb.Arena_ArenaChallengeResultPrcRequest) (*servicepb.Arena_ArenaChallengeResultPrcResponse, *errmsg.ErrMsg) {
	this_.DelFight(req.Type, ctx.RoleId)
	_, data, err := this_.GetPlayerArenaData(ctx, ctx.RoleId, req.Type)
	if err != nil {
		return nil, err
	}

	ret, err := this_.ChallengeFinish(ctx, req.Type, data.RankingIndex, ctx.RoleId, req.PlayerId, req.IsWin, false)
	if err != nil {
		return nil, err
	}

	err = this_.Module.JourneyService.AddToken(ctx, ctx.RoleId, valuesJourney.JourneyArena, 1)
	return ret, nil
}

func (this_ *Service) GetRewardTime(ctx *ctx.Context, req *servicepb.Arena_ArenaGetRewardTimeRequest) (*servicepb.Arena_ArenaGetRewardTimeResponse, *errmsg.ErrMsg) {
	var ret *servicepb.Arena_ArenaGetRewardTimeResponse = new(servicepb.Arena_ArenaGetRewardTimeResponse)

	var serverIds map[int64]bool = make(map[int64]bool)
	for _, aType := range models.ArenaType_value {
		if aType == int32(models.ArenaType_ArenaType_None) {
			continue
		}
		sId := ArenaRule.GetArenaServer(models.ArenaType(aType))
		serverIds[sId] = true
	}

	for sId := range serverIds {
		out := &arena_service.ArenaRanking_ArenaGetRewardTimeResponse{}
		if err := this_.svc.GetNatsClient().RequestWithOut(ctx, sId, &arena_service.ArenaRanking_ArenaGetRewardTimeRequest{}, out); err != nil {
			continue
		}
		ret.Times = append(ret.Times, out.Times...)
	}
	return ret, nil
}

func (this_ *Service) ArenaGetFightLog(ctx *ctx.Context, req *servicepb.Arena_ArenaGetFightLogRequest) (*servicepb.Arena_ArenaGetFightLogResponse, *errmsg.ErrMsg) {
	fightLogs := arenaDao.GetAllFightLog(ctx, req.Type, ctx.RoleId)
	var delLogs []*dao.ArenaFightLogs
	var rets []*models.ArenaFightLog
	beginTime := timer.BeginOfDay(timer.Now()).Unix()

	for _, flogDatas := range fightLogs {
		if flogDatas.FightDayBegin < beginTime {
			delLogs = append(delLogs, flogDatas)
			continue
		}
		for _, flog := range flogDatas.FightLogs {
			if flog.FightTime > beginTime {
				rets = append(rets, flog)
			}
		}
	}

	if len(delLogs) > 0 {
		arenaDao.DelFightLog(ctx, req.Type, ctx.RoleId, delLogs)
	}

	for _, flog := range rets {
		var roleId string
		if flog.AttackerRoleId != ctx.RoleId {
			roleId = flog.AttackerRoleId
		} else {
			roleId = flog.DefenderRoleId
		}
		if flog.IsRobot {
			var err *errmsg.ErrMsg
			flog.PlayerInfo, err = this_.GetRobotInfo(ctx, roleId, 0, flog.NickName)
			if err != nil {
				return nil, err
			}
		} else {
			roleInfo, err := this_.GetPlayerInfo(ctx, req.Type, roleId, 0)
			if err != nil {
				ctx.Error("ArenaGetFightLog GetPlayerInfo erro", zap.Any("err msg", err))
				roleInfo = &models.ArenaInfo{
					RoleId: roleId,
				}
			}
			flog.PlayerInfo = roleInfo
		}
	}

	return &servicepb.Arena_ArenaGetFightLogResponse{
		Type:     req.Type,
		FightLog: rets,
	}, nil
}

func (this_ *Service) CheckReward(ctx *ctx.Context, req *servicepb.Arena_ArenaCheckRewardRequest) (*servicepb.Arena_ArenaCheckRewardResponse, *errmsg.ErrMsg) {
	this_.ReceiveReward(ctx, ctx.RoleId)
	return &servicepb.Arena_ArenaCheckRewardResponse{}, nil
}

func (this_ *Service) VirtualCombat(ctx *ctx.Context, req *servicepb.Virtual_CombatRequest) (*servicepb.Virtual_CombatResponse, *errmsg.ErrMsg) {
	coutDown, ok := arenaRule.GetVirtualCombatDuelDuration(ctx)
	if !ok {
		return nil, errmsg.NewErrVirtualCombatConfig()
	}
	role0, heros0, err := this_.GetPlayerFightInfo(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	role1, heros1, err := this_.GetPlayerFightInfo(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}

	d, err00 := db.GetBattleSetting(ctx)
	if err00 != nil {
		return nil, err00
	}

	return &servicepb.Virtual_CombatResponse{
		RoleId: req.RoleId,
		Data: &models.SingleBattleParam{
			CountDown: coutDown,
			Role:      role0,
			Heroes:    heros0,
			HostilePlayers: []*models.SinglePlayerInfo{
				0: {
					Role:   role1,
					Heroes: heros1,
				},
			},
			AutoSoulSkill: d.Data.AutoSoulSkill,
		},
	}, nil
}

func (this_ *Service) HandleRoleLogoutEvent(ctx *ctx.Context, d event.Logout) *errmsg.ErrMsg {
	return this_.CheckFight(ctx)
}

func (this_ *Service) HandleRoleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	err := this_.CheckFight(ctx)
	if err != nil {
		ctx.Error("HandleRoleLoginEvent CheckFight", zap.Any("err msg", err))
	}
	this_.ReceiveReward(ctx, ctx.RoleId)
	return nil
}

// func ============================================================================================================================================================================================================

func (this_ *Service) CheckFight(ctx *ctx.Context) *errmsg.ErrMsg {
	arenaData, isNew := arenaDao.GetPlayerArenaData(ctx, ctx.RoleId)
	if isNew {
		ticketConfigInfo, err := this_.GetArenaTicketCost(models.ArenaType_ArenaType_Default)
		if err != nil {
			return err
		}
		arenaData.SetleftTicketPurchasesNumber(models.ArenaType_ArenaType_Default, ticketConfigInfo[0])
		return nil
	}

	for _, aType := range models.ArenaType_value {
		if aType == int32(models.ArenaType_ArenaType_None) {
			continue
		}

		data, ok := arenaData.GetData().Data[aType]
		if !ok {
			continue
		}

		challengeMapI, ok := this_.challengeMap.Load(aType)
		if ok {
			challengedRoleId, ok := challengeMapI.(*sync.Map).Load(ctx.RoleId)
			if ok {
				// 挑战失败 结算一次
				this_.ChallengeFinish(ctx, models.ArenaType(aType), data.RankingIndex, ctx.RoleId, challengedRoleId.(string), false, true)
			}
		}
	}
	return nil
}

type stChangeInfo struct {
	index       int
	RewardIndex uint64
}

func GetRankingId(changeInfo []*dao.ArenaRoleRankingChangeInfo, rewardIndex uint64) int32 {
	lessId := stChangeInfo{
		index:       0,
		RewardIndex: 0,
	}

	equalId := stChangeInfo{
		index:       0,
		RewardIndex: 0,
	}

	for index, cInfo := range changeInfo {
		if cInfo.RewardIndex < rewardIndex && cInfo.RewardIndex > lessId.RewardIndex {
			lessId.index = index
			lessId.RewardIndex = cInfo.RewardIndex
		}

		if cInfo.RewardIndex == rewardIndex {
			equalId.index = index
			equalId.RewardIndex = cInfo.RewardIndex
			break
		}
	}

	if equalId.RewardIndex == rewardIndex {
		return changeInfo[equalId.index].RankingId
	}

	if lessId.RewardIndex > 0 {
		return changeInfo[equalId.index].RankingId
	}

	return 0
}

func (this_ *Service) ReceiveReward(ctx *ctx.Context, roleId string) {
	arenaData, isNew := arenaDao.GetPlayerArenaData(ctx, roleId)
	if isNew {
		return
	}

	haveChange := false
	for _, aType := range models.ArenaType_value {
		if aType == int32(models.ArenaType_ArenaType_None) {
			continue
		}

		data, ok := arenaData.GetData().Data[int32(aType)]
		if !ok {
			continue
		}

		dayRewardIndex := arenaDao.LoadArenaDayRewardIndex(ctx, models.ArenaType(aType))
		if dayRewardIndex != nil {
			if this_.ProcReward(ctx, arenaData, data, roleId, false, models.ArenaType(aType), dayRewardIndex) {
				haveChange = true
			}
		}

		seasonRewardIndex := arenaDao.LoadArenaSeasonRewardIndex(ctx, models.ArenaType(aType))
		if seasonRewardIndex != nil {
			if this_.ProcReward(ctx, arenaData, data, roleId, true, models.ArenaType(aType), seasonRewardIndex) {
				haveChange = true
			}
		}
	}

	if haveChange {
		arenaData.Save()
	}
}

func (this_ *Service) ProcReward(ctx *ctx.Context, arenaData arenaDao.ArenaI,
	data *models.ArenaData, roleId string, isSeasonOver bool, aType models.ArenaType, rewardIndex []*models.ArenaSendRewardTime) bool {
	changeInfo := arenaDao.GetRankingChangeInfo(ctx, roleId, isSeasonOver, int64(aType))
	if len(changeInfo) != 0 {
		sort.Slice(rewardIndex, func(i, j int) bool {
			return rewardIndex[i].RewardIndex < rewardIndex[j].RewardIndex
		})

		haveChange := false
		for i := 0; i < len(rewardIndex); i++ {
			if isSeasonOver {
				if rewardIndex[i].RewardIndex <= data.SeasonRewardIndex {
					continue
				}
			} else {
				if rewardIndex[i].RewardIndex <= data.DayRewardIndex {
					continue
				}
			}
			rankingId := GetRankingId(changeInfo, rewardIndex[i].RewardIndex)
			if rankingId == 0 {
				continue
			}

			ok := this_.ProcSendReward(ctx, roleId, isSeasonOver, int64(aType), int64(rankingId), rewardIndex[i].SendRewardTime, rewardIndex[i].RewardIndex)
			if !ok {
				break
			}

			if isSeasonOver {
				data.SeasonRewardIndex = rewardIndex[i].RewardIndex
			} else {
				data.DayRewardIndex = rewardIndex[i].RewardIndex

			}
			haveChange = true
		}

		var delChangeInfo []*dao.ArenaRoleRankingChangeInfo
		for _, changeInfo := range changeInfo {
			if isSeasonOver {
				if changeInfo.RewardIndex <= data.SeasonRewardIndex {
					delChangeInfo = append(delChangeInfo, changeInfo)
				}
			} else {
				if changeInfo.RewardIndex <= data.DayRewardIndex {
					delChangeInfo = append(delChangeInfo, changeInfo)
				}
			}
		}

		if len(delChangeInfo) > 1 {
			maxInfo := stChangeInfo{
				index:       0,
				RewardIndex: 0,
			}

			for index, cinfo := range delChangeInfo {
				if maxInfo.RewardIndex < cinfo.RewardIndex {
					maxInfo.index = index
					maxInfo.RewardIndex = cinfo.RewardIndex
				}
			}

			for index, changeInfo := range delChangeInfo {
				if index == maxInfo.index {
					continue
				}
				arenaDao.DelRankingChangeInfo(ctx, roleId, isSeasonOver, int64(aType), changeInfo.PK())
			}
		}
		return haveChange
	}
	return false
}

func (this_ *Service) ProcSendReward(ctx *ctx.Context, roleId string, isSeasonOver bool, aType int64, rankingId int64, sendRewardTime int64, rewardIndex uint64) bool {
	var items []*models.Item
	if models.ArenaType(aType) == models.ArenaType_ArenaType_Default {
		configs := arenaRule.GetAllPvpRankReward(ctx)
		for _, cnf := range configs {
			if rankingId < cnf.RankUpperLimit || rankingId > cnf.RankLowerLimit {
				continue
			}

			if isSeasonOver {
				for itemId, count := range cnf.SeasonReward {
					items = append(items, &models.Item{
						ItemId: itemId,
						Count:  count,
					})
				}
			} else {
				for itemId, count := range cnf.DailyReward {
					items = append(items, &models.Item{
						ItemId: itemId,
						Count:  count,
					})
				}
			}
			break
		}
	}

	if len(items) == 0 {
		return true
	}

	var mailId int64
	dayMailId, seasonMailId, err := this_.GetMailID(models.ArenaType(aType))
	if err != nil {
		ctx.Error("GetMailId error ", zap.Any("Type", aType))
		return false
	}

	if isSeasonOver {
		mailId = seasonMailId
	} else {
		mailId = dayMailId
	}

	_, err = this_.SendItem(ctx, items, aType, mailId, rankingId, sendRewardTime)
	if err != nil {
		ctx.Error("Send Mail fail", zap.Any("isOverSeason", isSeasonOver), zap.Any("aType", aType), zap.Any("time", sendRewardTime), zap.Any("rewardIndex", rewardIndex))
		return false
	}
	return true
}

func (this_ *Service) SendItem(ctx *ctx.Context, items []*models.Item, aType int64, mailId int64, rankingId int64, sendRewardTime int64) (bool, *errmsg.ErrMsg) {
	if err := this_.MailService.Add(ctx, ctx.RoleId, &models.Mail{
		Type:       models.MailType_MailTypeSystem,
		TextId:     mailId,
		Args:       []string{strconv.Itoa(int(rankingId)), strconv.Itoa(int(sendRewardTime)), strconv.Itoa(int(aType))},
		Attachment: items,
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (this_ *Service) ChallengeFinish(ctx *ctx.Context, aType models.ArenaType, rankingIndex string, challengerRoleId string, challengedRoleId string, isWin bool, isOverTime bool) (*servicepb.Arena_ArenaChallengeResultPrcResponse, *errmsg.ErrMsg) {
	challengeMapI, ok := this_.challengeMap.Load(int32(aType))
	if !ok {
		return nil, errmsg.NewErrArenaNotFoundChallenge()
	}

	cRoleId, ok := challengeMapI.(*sync.Map).Load(challengerRoleId)
	if !ok {
		return nil, errmsg.NewErrArenaNotFoundChallenge()
	}

	if cRoleId != challengedRoleId {
		return nil, errmsg.NewErrArenaRankingChallengePlayerId()
	}

	localTime := int64(0)
	challengedMapI, ok := this_.challengeLockMap.Load(int32(aType))
	if !ok {
		return nil, errmsg.NewErrArenaNotFoundChallenge()
	}
	lockTimeI, ok := challengedMapI.(*sync.Map).Load(challengedRoleId)
	if ok {
		localTime = lockTimeI.(int64)
	}

	challengeMapI.(*sync.Map).Delete(challengerRoleId)
	challengedMapI.(*sync.Map).Delete(challengedRoleId)

	fightLog := &models.ArenaFightLog{
		AttackerRoleId: challengerRoleId,
		DefenderRoleId: challengedRoleId,
		IsWin:          isWin,
		FightTime:      timer.Now().Unix(),
	}

	out := &arena_service.ArenaRanking_UnLockRoleRankingResponse{}
	if err := this_.svc.GetNatsClient().RequestWithOut(ctx, ArenaRule.GetArenaServer(aType), &arena_service.ArenaRanking_UnLockRoleRankingRequest{
		Type:         aType,
		RoleId:       challengerRoleId,
		RankingIndex: rankingIndex,
	}, out); err != nil {
		return nil, err
	}

	battleDuration, err := this_.GetBattleDuration(aType)
	if err != nil {
		return nil, err
	}

	this_.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskArenaBattleNum, 0, 1)

	iWin := int64(1)
	if !isWin {
		iWin = 0
	}

	localIsOverTime := false
	if localTime > 0 && localTime+battleDuration < timer.Now().Unix() {
		localIsOverTime = true
	}

	iOverTime := int64(0)
	if isOverTime || localIsOverTime {
		iOverTime = 1
	}

	// 埋点
	statistical.Save(ctx.NewLogServer(), &models2.Arena{
		IggId:       iggsdk.ConvertToIGGId(ctx.UserId),
		EventTime:   timer.Now(),
		GwId:        statistical.GwId(),
		RoleId:      ctx.RoleId,
		OtherRoleId: challengedRoleId,
		Type:        int64(aType),
		StartTime:   localTime,
		EndTime:     timer.Now().Unix(),
		IsOverTime:  iOverTime,
		IsWin:       iWin,
	})

	if isWin {

		if localIsOverTime {
			return nil, errmsg.NewErrArenaChallengeTimeOut()
		}

		swapInfo, isRobot, nickName, err := this_.SwapRanking(ctx, aType, rankingIndex, challengerRoleId, challengedRoleId)
		if err != nil {
			return nil, err
		}

		rewards, err := this_.GetArenaVictoryReward(aType)
		if err != nil {
			return nil, err
		}

		for _, item := range rewards {
			err = this_.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
			if err != nil {
				return nil, err
			}
		}

		fightLog.IsRobot = isRobot
		fightLog.RankingChangeInfo = swapInfo
		if isRobot {
			fightLog.NickName = nickName
		}

		lockKey := GetFightIndex(challengerRoleId, challengedRoleId)
		err = ctx.DRLock(redisclient.GetLocker(), lockKey)
		if err != nil {
			return nil, err
		}

		err = this_.AddChallengelog(ctx, aType, challengerRoleId, fightLog)
		if err != nil {
			return nil, err
		}

		if !isRobot {
			err = this_.AddChallengelog(ctx, aType, challengedRoleId, fightLog)
			if err != nil {
				return nil, err
			}
		}

		this_.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskKillPlayerNumAcc, 0, 1)
		this_.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskKillPlayerNum, 0, 1)
		this_.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskArenaRank, 0, values.Integer(swapInfo.Role_0NewRankingId), true)

		return &servicepb.Arena_ArenaChallengeResultPrcResponse{
			Type:              aType,
			IsWin:             isWin,
			Rewards:           rewards,
			SelfOldRankingId:  int64(swapInfo.Role_0OldRankingId),
			SelfNewRankingId:  int64(swapInfo.Role_0NewRankingId),
			OtherOldRankingId: int64(swapInfo.Role_1OldRankingId),
			OtherNewRankingId: int64(swapInfo.Role_1NewRankingId),
			PlayerId:          challengedRoleId,
		}, nil
	}

	err = this_.SetFirstChallenge(ctx, aType, rankingIndex, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	isRobot, nickName, err := this_.IsRobot(ctx, aType, rankingIndex, challengedRoleId)
	if err != nil {
		return nil, err
	}

	fightLog.IsRobot = isRobot
	if isRobot {
		fightLog.NickName = nickName
	}

	err = this_.AddChallengelog(ctx, aType, challengerRoleId, fightLog)
	if err != nil {
		return nil, err
	}
	if !isRobot {
		err = this_.AddChallengelog(ctx, aType, challengedRoleId, fightLog)
		if err != nil {
			return nil, err
		}
	}

	rewards, err := this_.GetArenaLoseReward(aType)
	if err != nil {
		return nil, err
	}

	for _, item := range rewards {
		err = this_.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
		if err != nil {
			return nil, err
		}
	}

	return &servicepb.Arena_ArenaChallengeResultPrcResponse{
		Type:     aType,
		IsWin:    isWin,
		Rewards:  rewards,
		PlayerId: challengedRoleId,
	}, nil
}

func (this_ *Service) GetCppHero(ctx *ctx.Context, roleId string, heros *models.Assemble) ([]*models.HeroForBattle, *errmsg.ErrMsg) {
	heroIds := make([]int64, 0, 2)
	if heros.Hero_0 > 0 && heros.HeroOrigin_0 > 0 {
		heroIds = append(heroIds, heros.HeroOrigin_0)
	}
	if heros.Hero_1 > 0 && heros.HeroOrigin_1 > 0 {
		heroIds = append(heroIds, heros.HeroOrigin_1)
	}
	heroes, err := this_.Module.GetHeroes(ctx, roleId, heroIds)
	if err != nil {
		return nil, err
	}
	equips, err := this_.GetManyEquipBagMap(ctx, roleId, this_.GetHeroesEquippedEquipId(heroes)...)
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
	return cppHeroes, nil
}

func (this_ *Service) GetRobotHeroInfo(ctx *ctx.Context, robotId string) (*models.Assemble, *errmsg.ErrMsg) {
	data, ok := arenaRule.GetRobot(ctx, robotId)
	if !ok {
		ctx.Error("robot info not find", zap.Any("robot id", robotId))
		return nil, errmsg.NewErrArenaNotFoundPlayer()
	}
	var haveHero bool = false
	var heros models.Assemble
	originId, ok := arenaRule.GetRobotHeroInfo(ctx, data.ConfigId1)
	if ok {
		haveHero = true
		heros.Hero_0 = data.ConfigId1
		heros.HeroOrigin_0 = originId
		heros.Hero_0Power = data.CombatEffectiveness / 2
	}

	originId, ok = arenaRule.GetRobotHeroInfo(ctx, data.ConfigId2)
	if ok {
		haveHero = true
		heros.Hero_1 = data.ConfigId2
		heros.HeroOrigin_1 = originId
		heros.Hero_1Power = data.CombatEffectiveness / 2
	}
	if haveHero {
		return &heros, nil
	}
	return nil, errmsg.NewErrArenaNotHeroes()
}

/*
message HeroForBattle{
  int64 config_id = 1; // hero表配置id
  map<int64, int64> attr = 2;  // hero属性(hp mp 等)
  map<int64, int64> equip = 3; // 各部位装备 itemId
  int64 hero_status = 4;
  repeated int64 skill_ids = 5; // 英雄技能
}
*/

func (this_ *Service) GetHeroDefaultSkill(ctx *ctx.Context, heroId int64) ([]*models.HeroSkillAndStone, *errmsg.ErrMsg) {
	config, ok := heroRule.GetHero(ctx, heroId)
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	skills := this_.GetInitSkills(config)
	res := make([]*models.HeroSkillAndStone, 0, len(skills))
	for _, skill := range skills {
		res = append(res, &models.HeroSkillAndStone{
			SkillId: skill,
		})
	}
	return res, nil
}

func (this_ *Service) GetCppRobotHeroInfo(ctx *ctx.Context, robotId string) (*rulemodel.PvPRobot, []*models.HeroForBattle, *errmsg.ErrMsg) {
	data, ok := arenaRule.GetRobot(ctx, robotId)
	if !ok {
		ctx.Error("robot info not find", zap.Any("robot id", robotId))
		return nil, nil, errmsg.NewErrArenaNotFoundPlayer()
	}
	var cppHero []*models.HeroForBattle

	_, ok = arenaRule.GetRobotHeroInfo(ctx, data.ConfigId1)
	if ok {
		skills, err := this_.GetHeroDefaultSkill(ctx, data.ConfigId1)
		if err != nil {
			return nil, nil, err
		}
		cppHero = append(cppHero, &models.HeroForBattle{
			ConfigId: data.ConfigId1,
			Attr:     data.Attr1,
			SkillIds: skills,
		})
	}

	_, ok = arenaRule.GetRobotHeroInfo(ctx, data.ConfigId2)
	if ok {
		skills, err := this_.GetHeroDefaultSkill(ctx, data.ConfigId2)
		if err != nil {
			return nil, nil, err
		}
		cppHero = append(cppHero, &models.HeroForBattle{
			ConfigId: data.ConfigId2,
			Attr:     data.Attr2,
			SkillIds: skills,
		})
	}

	if len(cppHero) == 0 {
		return nil, nil, errmsg.NewErrArenaNotHeroes()
	}

	return data, cppHero, nil
}

func (this_ *Service) GetSelfRankingInfo(ctx *ctx.Context, aType models.ArenaType, ranking_index string) (*servicepb.Arena_ArenaSelfRankingResponse, *errmsg.ErrMsg) {
	out := &arena_service.ArenaRanking_GetSelfDataResponse{}
	if err := this_.svc.GetNatsClient().RequestWithOut(ctx, ArenaRule.GetArenaServer(aType), &arena_service.ArenaRanking_GetSelfDataRequest{
		Type:         aType,
		RankingIndex: ranking_index,
		RoleId:       ctx.RoleId,
	}, out); err != nil {
		return nil, err
	}

	nextRefreshTime, err := this_.GetDayRefresh(ctx, aType)

	if err != nil {
		return nil, err
	}

	return &servicepb.Arena_ArenaSelfRankingResponse{
		Type:               aType,
		RankingId:          out.Element.RankingId,
		NextRefreshTime:    nextRefreshTime,
		NextSettlementTime: out.NextSettlementTime,
		SeasonOverTime:     out.SeasonOverTime,
		IsFirstChallenge:   out.Element.IsFirstChallenge,
	}, nil
}

func (this_ *Service) IsRobot(ctx *ctx.Context, aType models.ArenaType, ranking_index string, role_id string) (bool, string, *errmsg.ErrMsg) {
	out := &arena_service.ArenaRanking_GetPlayerIsRobotResponse{}
	if err := this_.svc.GetNatsClient().RequestWithOut(ctx, ArenaRule.GetArenaServer(aType), &arena_service.ArenaRanking_GetPlayerIsRobotRequest{
		Type:         aType,
		RankingIndex: ranking_index,
		RoleId:       role_id,
	}, out); err != nil {
		return true, "", err
	}
	return out.IsRobot, out.NickName, nil
}

func (this_ *Service) SelectRankInfo(ctx *ctx.Context, aType models.ArenaType, ranking_index string, start int64, end int64) ([]*models.ArenaInfo, *errmsg.ErrMsg) {
	out := &arena_service.ArenaRanking_GetDatasResponse{}
	if err := this_.svc.GetNatsClient().RequestWithOut(ctx, ArenaRule.GetArenaServer(aType), &arena_service.ArenaRanking_GetDatasRequest{
		Type:         aType,
		RankingIndex: ranking_index,
		StartIndex:   int32(start),
		Count:        int32(end - start),
	}, out); err != nil {
		return nil, err
	}
	guildIdList := make([]*models.ArenaInfo, 0)
	for _, player := range out.Datas {
		if player.IsRobot {
			robotInfo, err := this_.GetRobotInfo(ctx, player.RoleId, player.RankingId, player.NickName)
			if err != nil {
				continue
			}
			guildIdList = append(guildIdList, robotInfo)
		} else {
			roleInfo, err := this_.GetPlayerInfo(ctx, aType, player.RoleId, player.RankingId)
			if err != nil {
				ctx.Error("SelectRankInfo GetPlayerInfo error", zap.Any("err msg", err))
				roleInfo = &models.ArenaInfo{
					RoleId:    player.RoleId,
					RankingId: player.RankingId,
				}
			}
			guildIdList = append(guildIdList, roleInfo)
		}
	}
	return guildIdList, nil
}

func (this_ *Service) GetChallengeRange(ctx *ctx.Context, aType models.ArenaType, rankingIndex string, roleId string) ([]*models.ArenaInfo, *errmsg.ErrMsg) {
	out := &arena_service.ArenaRanking_GetChallengeRangeResponse{}
	if err := this_.svc.GetNatsClient().RequestWithOut(ctx, ArenaRule.GetArenaServer(aType), &arena_service.ArenaRanking_GetChallengeRangeRequest{
		Type:         aType,
		RankingIndex: rankingIndex,
		RoleId:       roleId,
	}, out); err != nil {
		return nil, err
	}
	guildIdList := make([]*models.ArenaInfo, 0)
	for _, player := range out.Datas {
		if player.IsRobot {
			robotInfo, err := this_.GetRobotInfo(ctx, player.RoleId, player.RankingId, player.NickName)
			if err != nil {
				continue
			}
			guildIdList = append(guildIdList, robotInfo)
		} else {
			roleInfo, err := this_.GetPlayerInfo(ctx, aType, player.RoleId, player.RankingId)
			if err != nil {
				ctx.Error("GetChallengeRange GetPlayerInfo erro", zap.Any("err msg", err))
				roleInfo = &models.ArenaInfo{
					RoleId: player.RoleId,
				}
			}
			guildIdList = append(guildIdList, roleInfo)
		}
	}
	return guildIdList, nil
}

func (this_ *Service) GetRobotInfo(ctx *ctx.Context, playerId string, rankingId int32, nickName string) (*models.ArenaInfo, *errmsg.ErrMsg) {
	data, ok := arenaRule.GetRobot(ctx, playerId)
	if !ok {
		ctx.Error("robot info not find", zap.Any("robot id", playerId))
		return nil, errmsg.NewErrArenaNotFoundPlayer()
	}
	return &models.ArenaInfo{
		RoleId:      playerId,
		Nickname:    nickName,
		Level:       data.Lv,
		AvatarId:    data.HeadSculpture,
		AvatarFrame: data.HeadSculptureFrame,
		Power:       data.CombatEffectiveness,
		Title:       data.RoleLvTitle,
		RankingId:   rankingId,
		IsRobot:     true,
	}, nil
}

func (this_ *Service) GetPlayerInfo(ctx *ctx.Context, aType models.ArenaType, playerId string, rankingId int32) (*models.ArenaInfo, *errmsg.ErrMsg) {
	role, err := this_.Module.GetRoleModelByRoleId(ctx, playerId)
	if err != nil {
		return nil, err
	}

	_, otherData, err := this_.GetPlayerArenaData(ctx, playerId, aType)
	if err != nil {
		return nil, err
	}

	power := 0
	if otherData.FightHero != nil {
		fHero, err := this_.GetHero(ctx, playerId, otherData.FightHero)
		if err != nil {
			return nil, err
		}
		power = int(fHero.GetHero_0Power()) + int(fHero.GetHero_1Power())
	} else {
		fightHero, err := this_.Module.FormationService.GetDefaultHeroes(ctx, playerId)
		if err != nil {
			return nil, err
		}
		fHero, err := this_.GetHero(ctx, playerId, fightHero)
		if err != nil {
			return nil, err
		}
		power = int(fHero.GetHero_0Power()) + int(fHero.GetHero_1Power())
	}

	guildName := ""
	user, err := guildDao.NewGuildUser(playerId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if user.GuildId != "" {
		guild, err := guildDao.NewGuild(user.GuildId).Get(ctx)
		if err != nil {
			return nil, err
		}
		guildName = guild.Name
	}

	return &models.ArenaInfo{
		RoleId:      playerId,
		Nickname:    role.Nickname,
		Level:       role.Level,
		AvatarId:    role.AvatarId,
		AvatarFrame: role.AvatarFrame,
		Power:       int64(power),
		Title:       role.Title,
		RankingId:   rankingId,
		GuildName:   guildName,
		IsRobot:     false,
	}, nil
}

func (this_ *Service) SetHero(ctx *ctx.Context, aType models.ArenaType, assembleHero *models.Assemble) *errmsg.ErrMsg {
	arenaData, data, err := this_.GetPlayerArenaData(ctx, ctx.RoleId, aType)
	if err != nil {
		return err
	}

	data.FightHero = assembleHero
	arenaData.Save()
	return nil
}

func (this_ *Service) GetHero(ctx *ctx.Context, roleId string, assembleHero *models.Assemble) (*models.Assemble, *errmsg.ErrMsg) {
	if assembleHero.HeroOrigin_0 == 0 && assembleHero.HeroOrigin_1 == 0 {
		return nil, errmsg.NewErrArenaNotHeroes()
	}
	heros, err := heroDao.NewHero(roleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	var heroMap map[int64]*dao.Hero = make(map[int64]*dao.Hero, len(heros))
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

func (this_ *Service) JoinArenaRanking(ctx *ctx.Context, aType models.ArenaType) (string, *errmsg.ErrMsg) {
	out := &arena_service.ArenaRanking_JoinRankingResponse{}
	if err := this_.svc.GetNatsClient().RequestWithOut(ctx, ArenaRule.GetArenaServer(aType), &arena_service.ArenaRanking_JoinRankingRequest{
		Type: aType,
		PlayerInfo: &models.ArenaRanking_Info{
			RoleId:  ctx.RoleId,
			IsRobot: false,
		},
	}, out); err != nil {
		ctx.Error("join arena ranking error", zap.Any("error", err))
		return "", err
	}
	return out.RankingIndex, nil
}

func (this_ *Service) SwapRanking(ctx *ctx.Context, aType models.ArenaType, rankingIndex string, challengerRoleId string, challengedRoleId string) (*models.RankingSwapInfo, bool, string, *errmsg.ErrMsg) {
	out := &arena_service.ArenaRanking_SwapRankingIdResponse{}
	if err := this_.svc.GetNatsClient().RequestWithOut(ctx, ArenaRule.GetArenaServer(aType), &arena_service.ArenaRanking_SwapRankingIdRequest{
		Type:         aType,
		RankingIndex: rankingIndex,
		RoleId_0:     challengerRoleId,
		RoleId_1:     challengedRoleId,
	}, out); err != nil {
		return nil, false, "", err
	}
	return out.SwapInfo, out.IsRobot, out.NickName, nil
}

func (this_ *Service) SetFirstChallenge(ctx *ctx.Context, aType models.ArenaType, rankingIndex string, roleId string) *errmsg.ErrMsg {
	out := &arena_service.ArenaRanking_SetFirstChallengeResponse{}
	if err := this_.svc.GetNatsClient().RequestWithOut(ctx, ArenaRule.GetArenaServer(aType), &arena_service.ArenaRanking_SetFirstChallengeRequest{
		Type:         aType,
		RankingIndex: rankingIndex,
		RoleId:       roleId,
	}, out); err != nil {
		return err
	}
	return nil
}

func (this_ *Service) GetArenaFreeNum(aType models.ArenaType) (int64, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx.GetContext(), this_.log)
		if config == nil {
			return 0, errmsg.NewErrArenaConfig()
		}
		return int64(config.ArenaFreeNum), nil
	}
	return 0, errmsg.NewErrArenaType()
}

func (this_ *Service) GetBattleDuration(aType models.ArenaType) (int64, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx.GetContext(), this_.log)
		if config == nil {
			return 0, errmsg.NewErrArenaConfig()
		}
		return int64(config.ArenaBattleDuration), nil
	}
	return 0, errmsg.NewErrArenaType()
}

func (this_ *Service) GetArenaTicket(aType models.ArenaType) (int64, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx.GetContext(), this_.log)
		if config == nil {
			return 0, errmsg.NewErrArenaConfig()
		}
		return int64(config.ArenaTicket), nil
	}
	return 0, errmsg.NewErrArenaType()
}

func (this_ *Service) GetArenaVictoryReward(aType models.ArenaType) ([]*models.Item, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx.GetContext(), this_.log)
		if config == nil {
			return nil, errmsg.NewErrArenaConfig()
		}
		return config.ArenaVictoryReward, nil
	}
	return nil, errmsg.NewErrArenaType()
}

func (this_ *Service) GetArenaLoseReward(aType models.ArenaType) ([]*models.Item, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx.GetContext(), this_.log)
		if config == nil {
			return nil, errmsg.NewErrArenaConfig()
		}
		return config.ArenaLoseReward, nil
	}
	return nil, errmsg.NewErrArenaType()
}

func (this_ *Service) GetArenaTicketCost(aType models.ArenaType) ([]int64, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx.GetContext(), this_.log)
		if config == nil {
			return nil, errmsg.NewErrArenaConfig()
		}
		return config.ArenaTicketCost, nil
	}
	return nil, errmsg.NewErrArenaType()
}

func (this_ *Service) GetMailID(aType models.ArenaType) (int64, int64, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx.GetContext(), this_.log)
		if config == nil {
			return 0, 0, errmsg.NewErrArenaConfig()
		}
		return config.ArenaDailyMail, config.ArenaSeasonMail, nil
	}
	return 0, 0, errmsg.NewErrArenaType()
}

func (this_ *Service) AddChallengelog(ctx *ctx.Context, aType models.ArenaType, roleId string, fightLog *models.ArenaFightLog) *errmsg.ErrMsg {
	fightLogs, err := arenaDao.GetFightLog(ctx, aType, roleId)
	if err != nil {
		return err
	}
	fightLogs.FightLogs = append(fightLogs.FightLogs, fightLog)
	arenaDao.SetFightHero(ctx, aType, roleId, fightLogs)
	return nil
}

func GetFightIndex(attackRoleId string, defenseRoleId string) string {
	if attackRoleId > defenseRoleId {
		return fmt.Sprintf("%s::%s", attackRoleId, defenseRoleId)
	} else {
		return fmt.Sprintf("%s::%s", defenseRoleId, attackRoleId)
	}
}

func GetHMS(timeUnix int64) (int, int, int) {
	local_time := time.Unix(timeUnix, 0).UTC()
	return local_time.Hour(), local_time.Minute(), local_time.Second()
}

func (this_ *Service) CheckAndRefreshData(ctx *ctx.Context, aType models.ArenaType, arenaData arenaDao.ArenaI, data *models.ArenaData) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx, this_.log)
		if config == nil {
			return
		}
		DefaultRefreshTime := int64(config.DefaultRefreshTime)
		nowTime := timer.BeginOfDay(timer.Now()).Unix()
		DefaultRefreshTime += nowTime
		h, m, _ := GetHMS(DefaultRefreshTime)
		isNewDay := timer.OverRefreshTime(data.LastRefreshTime, h, m)
		if isNewDay {
			data.LastRefreshTime = timer.Now().Unix()
			data.LeftTicketPurchasesNum = config.ArenaTicketCost[0]
			data.FreeChallengeTimes = int64(config.ArenaFreeNum)
			arenaKey := GetEventKey(aType)
			ctx.PublishEventLocal(&event.RedPointChange{
				RoleId: ctx.RoleId,
				Key:    arenaKey,
				Val:    data.FreeChallengeTimes,
			})
			arenaData.Save()
		}
	}
}

func (this_ *Service) GetPlayerFightInfo(ctx *ctx.Context, roleId string) (*models.Role, []*models.HeroForBattle, *errmsg.ErrMsg) {
	role, err := this_.GetRoleModelByRoleId(ctx, roleId)
	if err != nil {
		return nil, nil, err
	}
	heroesFormation, err := this_.Module.FormationService.GetDefaultHeroes(ctx, roleId)
	if err != nil {
		return nil, nil, err
	}

	heroIds := make([]int64, 0, 2)
	if heroesFormation.Hero_0 > 0 && heroesFormation.HeroOrigin_0 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_0)
	}
	if heroesFormation.Hero_1 > 0 && heroesFormation.HeroOrigin_1 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_1)
	}

	heroes, err := this_.Module.GetHeroes(ctx, roleId, heroIds)
	if err != nil {
		return nil, nil, err
	}
	equips, err := this_.GetManyEquipBagMap(ctx, roleId, this_.GetHeroesEquippedEquipId(heroes)...)
	if err != nil {
		return nil, nil, err
	}
	cppHeroes := trans.Heroes2CppHeroes(ctx, heroes, equips)
	if len(cppHeroes) == 0 {
		return nil, nil, errmsg.NewErrHeroNotFound()
	}

	for _, h := range cppHeroes {
		if len(h.SkillIds) == 0 {
			return nil, nil, errmsg.NewInternalErr("SkillIds empty")
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
	return role, cppHeroes, nil
}

func (this_ *Service) GetPlayerArenaData(ctx *ctx.Context, roleId string, aType models.ArenaType) (arenaDao.ArenaI, *models.ArenaData, *errmsg.ErrMsg) {
	arenaData, isNew := arenaDao.GetPlayerArenaData(ctx, roleId)
	data, ok := arenaData.GetData().Data[int32(aType)]
	if !ok {
		return nil, nil, errmsg.NewErrArenaType()
	}
	if isNew {
		ticketConfigInfo, err := this_.GetArenaTicketCost(aType)
		if err != nil {
			return nil, nil, err
		}

		if len(ticketConfigInfo) < 3 {
			return nil, nil, errmsg.NewErrArenaConfig()
		}

		arenaData.SetleftTicketPurchasesNumber(aType, ticketConfigInfo[0])
	}

	this_.CheckAndRefreshData(ctx, aType, arenaData, data)
	return arenaData, data, nil
}

func (this_ *Service) GetDayRefresh(ctx *ctx.Context, aType models.ArenaType) (int64, *errmsg.ErrMsg) {
	if aType == models.ArenaType_ArenaType_Default {
		config := arenaRule.GetArenaDefalutPublicInfo(ctx, this_.log)
		if config == nil {
			return 0, errmsg.NewErrArenaConfig()
		}
		defaultRefreshTime := int64(config.DefaultRefreshTime)
		nowTime := timer.BeginOfDay(timer.Now()).Unix()
		defaultRefreshTime += nowTime
		h, m, _ := GetHMS(defaultRefreshTime)
		defaultRefreshTime = timer.GetNextRefreshTime(h, m)
		return defaultRefreshTime, nil
	}
	return 0, errmsg.NewErrArenaType()
}

func (this_ *Service) SetFight(aType models.ArenaType, roleId string) {
	key := strconv.FormatInt(int64(aType), 10) + ":" + roleId
	this_.fightLockMap.Store(key, timer.Now().Unix())
}

func (this_ *Service) HasFight(aType models.ArenaType, roleId string) bool {
	key := strconv.FormatInt(int64(aType), 10) + ":" + roleId
	vI, ok := this_.fightLockMap.Load(key)
	if !ok {
		return false
	}

	v := vI.(int64)
	return timer.Now().Unix()-v <= 5
}

func (this_ *Service) DelFight(aType models.ArenaType, roleId string) int64 {
	key := strconv.FormatInt(int64(aType), 10) + ":" + roleId
	startTime := int64(0)
	vI, ok := this_.fightLockMap.Load(key)
	if ok {
		startTime = vI.(int64)
		this_.fightLockMap.Delete(key)
	}
	return startTime
}

func GetEventKey(aType models.ArenaType) string {
	switch aType {
	case models.ArenaType_ArenaType_Default:
		{
			return ArenaEvenKey
		}
	}
	return ""
}
