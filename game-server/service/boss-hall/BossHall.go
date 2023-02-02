package bossHall

import (
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/im"
	"coin-server/common/logger"
	"coin-server/common/proto/cppbattle"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	newcenterpb "coin-server/common/proto/newcenter"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum/Notice"
	"coin-server/common/values/env"
	"coin-server/game-server/module"
	bossHallDao "coin-server/game-server/service/boss-hall/dao"
	valuesJourney "coin-server/game-server/service/journey/values"
	"coin-server/rule"

	"go.uber.org/zap"
)

const (
	maxCount = 5
)

var mustOpen bool

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	log *logger.Logger
}

func BossHallService(
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
	module.BossHallService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("获取boss大厅的战斗服信息", this_.GetBattleInfo)
	this_.svc.RegisterFunc("查询Boss可领奖励", this_.GetBossHalInfo)
	this_.svc.RegisterFunc("是否强制开启恶魔秘境", this_.CheatForceOpen)
	this_.svc.RegisterEvent("killBoss", this_.KillBoss)
	this_.svc.RegisterEvent("JoinkillBoss", this_.JoinKillBoss)
}

func CanEnterBossHall(ctx *ctx.Context, nowNS int64) bool {
	if mustOpen {
		return true
	}
	ut := time.Unix(0, nowNS).UTC()
	bt := timer.BeginOfDay(ut)
	s := int64(ut.Sub(bt).Seconds())
	rangeTime := rule.MustGetReader(ctx).GetBossHallOpenTime()
	for _, v := range rangeTime {
		if s >= v[0] && s <= v[1] {
			return true
		}
		if s < v[0] {
			return false
		}
	}
	return false
}

func (this_ *Service) CheatForceOpen(ctx *ctx.Context, req *servicepb.BossHall_MustOpenRequest) (*servicepb.BossHall_MustOpenResponse, *errmsg.ErrMsg) {
	mustOpen = req.Open
	return &servicepb.BossHall_MustOpenResponse{}, nil
}

// msg proc ============================================================================================================================================================================================================
func (this_ *Service) GetBattleInfo(ctx *ctx.Context, req *servicepb.BossHall_GetBattleInfoRequest) (*servicepb.BossHall_GetBattleInfoResponse, *errmsg.ErrMsg) {
	if !CanEnterBossHall(ctx, ctx.StartTime) {
		return nil, errmsg.NewErrBossHallNotOpen()
	}
	cnf, ok := rule.MustGetReader(ctx).BossHall.GetBossHallById(req.BossId)
	if !ok {
		return nil, errmsg.NewErrBossId()
	}

	r, err := this_.Module.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	if r.Level < cnf.RoleLvUnlock {
		return nil, errmsg.NewErrBossHallNotUnlock()
	}

	centerId := env.GetCenterServerId()
	out := &newcenterpb.NewCenter_CurBossHallResponse{}
	err = this_.svc.GetNatsClient().RequestWithOut(ctx, centerId, &newcenterpb.NewCenter_CurBossHallRequest{
		BossId: req.BossId,
	}, out)
	if err != nil {
		return nil, err
	}
	return &servicepb.BossHall_GetBattleInfoResponse{
		BattleId:   out.BattleServerId,
		MapSceneId: out.MapScenceId,
	}, nil
}

func (this_ *Service) GetBossHalInfo(ctx *ctx.Context, req *servicepb.BossHall_BossHallGetInfoRequest) (*servicepb.BossHall_BossHallGetInfoResponse, *errmsg.ErrMsg) {
	bossHallInfo, err := bossHallDao.GetBossHallData(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	err = this_.CheckRefresh(ctx, bossHallInfo)
	if err != nil {
		return nil, err
	}

	killLimit, joinLimit, err := GetKillJoinLimit(ctx, this_.log)
	if err != nil {
		return nil, err
	}

	bossKillJoinInfo, err := bossHallDao.GetBossKillJoinData(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	ret := &servicepb.BossHall_BossHallGetInfoResponse{
		LeftKillTimes: killLimit - bossHallInfo.Info.KillTimes,
		LeftJoinTimes: joinLimit - bossHallInfo.Info.JoinTimes,
		KillInfo:      bossKillJoinInfo.KillInfo,
		JoinInfo:      bossKillJoinInfo.JoinInfo,
		CanEnter:      CanEnterBossHall(ctx, ctx.StartTime),
	}

	return ret, nil
}

func (this_ *Service) KillBoss(ctx *ctx.Context, req *servicepb.BattleEvent_MonsterBossKill) {
	bossHallInfo, err := bossHallDao.GetBossHallData(ctx, ctx.RoleId)
	if err != nil {
		this_.log.Error("KillBoss error", zap.Error(err), zap.Any("req", req))
		return
	}
	err = this_.CheckRefresh(ctx, bossHallInfo)
	if err != nil {
		this_.log.Error("CheckRefresh error", zap.Error(err), zap.Any("req", req))
		return
	}
	killLimit, joinLimit, err := GetKillJoinLimit(ctx, this_.log)
	if err != nil {
		this_.log.Error("kill boss GetKillJoinLimit Error ", zap.Any("Err Msg", err))
		return
	}

	role, err := this_.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		this_.log.Error("kill boss GetRoleByRoleId Error ", zap.Any("Err Msg", err))
		return
	}

	err = this_.ImService.SendNotice(ctx, im.ParseTypeNoticeBossHall, Notice.BossHall, role.Nickname, req.MonsterConfigId)
	if err != nil {
		this_.log.Error("kill boss SendNotice Error ", zap.Any("Err Msg", err))
		return
	}

	if bossHallInfo.Info.KillTimes >= killLimit {
		this_.log.Info("killLimit", zap.Any("roleid", ctx.RoleId), zap.Any("kill Times", bossHallInfo.Info.KillTimes))
		return
	}

	curRes, err1 := this_.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		this_.log.Info("GetCurrBattleInfo error", zap.Any("roleid", ctx.RoleId), zap.Any("err msg", err1))
		return
	}

	bossHallInfo.Info.KillTimes++
	killData := &models.BossKillJoinInfo{
		Time: timer.Now().Unix(),
		BossInfo: &models.BossCollectInfo{
			BossId:   req.MonsterConfigId,
			BattleId: curRes.BattleId,
			MapId:    curRes.SceneId,
		},
	}

	for itemId, cnt := range req.Items {
		item := &models.Item{
			ItemId: itemId,
			Count:  cnt,
		}

		killData.Rewards = append(killData.Rewards, item)
	}

	if len(killData.Rewards) > 0 {
		err = this_.BagService.AddManyItemPb(ctx, ctx.RoleId, killData.Rewards...)
		if err != nil {
			this_.log.Error("sevenday ReceiveReward AddManyItemPb err",
				zap.Error(err), zap.String("role", ctx.RoleId), zap.Any("items", killData.Rewards))

		}
	}

	bossKillJoinInfo, err := bossHallDao.GetBossKillJoinData(ctx, ctx.RoleId)
	if err != nil {
		this_.log.Error("sGetBossKillJoinData err", zap.Error(err), zap.Any("req", req))
	}
	bossKillJoinInfo.KillInfo = append(bossKillJoinInfo.KillInfo, killData)
	bossHallDao.SaveBossKillJoinData(ctx, bossKillJoinInfo)
	bossHallDao.SaveBossHallData(ctx, bossHallInfo)
	killItems := make([]*models.Item, 0, len(req.Items))
	for k, v := range req.Items {
		killItems = append(killItems, &models.Item{
			ItemId: k,
			Count:  v,
		})
	}
	err = this_.svc.GetNatsClient().PublishCtx(ctx, curRes.BattleId, &cppbattle.CPPBattle_MonsterBossKillPush{
		BossId:          req.BossId,
		BossPos:         req.BossPos,
		MonsterConfigId: req.MonsterConfigId,
		KillerId:        req.KillerId,
		Items:           killItems,
	})
	if err != nil {
		this_.log.Error("push MonsterBossKillPush error", zap.Error(err))
	}

	ctx.PushMessage(&servicepb.BossHall_BossHallGetInfoResponse{
		LeftKillTimes: killLimit - bossHallInfo.Info.KillTimes,
		LeftJoinTimes: joinLimit - bossHallInfo.Info.JoinTimes,
		KillInfo:      bossKillJoinInfo.KillInfo,
		JoinInfo:      bossKillJoinInfo.JoinInfo,
	})
	this_.Module.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskKillBossHallKillBoos, 0, 1)
	this_.Module.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskBossHallKillRewardAcc, 0, 1)
	this_.Module.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskBossHallKillReward, 0, 1)
}

func (this_ *Service) JoinKillBoss(ctx *ctx.Context, req *servicepb.BattleEvent_MonsterBossJoinKill) {
	bossHallInfo, err := bossHallDao.GetBossHallData(ctx, ctx.RoleId)
	if err != nil {
		this_.log.Error("GetBossHallData error", zap.Error(err), zap.Any("req", req))
		return
	}
	err = this_.CheckRefresh(ctx, bossHallInfo)
	if err != nil {
		this_.log.Error("CheckRefresh error", zap.Error(err), zap.Any("req", req))
		return
	}
	killLimit, joinLimit, err := GetKillJoinLimit(ctx, this_.log)
	if err != nil {
		this_.log.Error("GetKillJoinLimit error", zap.Error(err), zap.Any("req", req))
		return
	}
	curRes, err1 := this_.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		this_.log.Error("GetCurrBattleInfo error", zap.Any("roleid", ctx.RoleId), zap.Error(err1))
		return
	}
	if bossHallInfo.Info.JoinTimes >= joinLimit {

		err = this_.svc.GetNatsClient().PublishCtx(ctx, curRes.BattleId, &cppbattle.CPPBattle_MonsterJoinKillPush{
			HasKill: bossHallInfo.Info.KillTimes < killLimit,
			HasJoin: false,
		})

		this_.log.Info("JoinkillLimit", zap.Any("roleid", ctx.RoleId), zap.Any("Join kill Times", bossHallInfo.Info.JoinTimes))
		return
	}

	bossHallCnf, ok := rule.MustGetReader(ctx).BossHall.GetBossHallById(req.MonsterConfigId)
	if !ok {
		this_.log.Error("Not Find Boss In BossHall", zap.Int64("BossId", req.MonsterConfigId))
		return
	}

	bossHallInfo.Info.JoinTimes++

	joinData := &models.BossKillJoinInfo{
		Time: timer.Now().Unix(),
		BossInfo: &models.BossCollectInfo{
			BossId:   req.MonsterConfigId,
			BattleId: curRes.BattleId,
			MapId:    curRes.SceneId,
		},
	}

	for itemId, cnt := range bossHallCnf.ParticipationAward {
		item := &models.Item{
			ItemId: itemId,
			Count:  cnt,
		}
		joinData.Rewards = append(joinData.Rewards, item)
	}

	if len(joinData.Rewards) > 0 {
		err = this_.BagService.AddManyItemPb(ctx, ctx.RoleId, joinData.Rewards...)
		if err != nil {
			this_.log.Error("sevenday ReceiveReward AddManyItemPb err",
				zap.Error(err), zap.String("role", ctx.RoleId), zap.Any("items", joinData.Rewards))
		}
	}

	bossKillJoinInfo, err := bossHallDao.GetBossKillJoinData(ctx, ctx.RoleId)
	if err != nil {
		this_.log.Error("GetBossKillJoinData BossHall", zap.Any("req", req), zap.Error(err))
		return
	}
	bossKillJoinInfo.JoinInfo = append(bossKillJoinInfo.JoinInfo, joinData)
	bossHallDao.SaveBossKillJoinData(ctx, bossKillJoinInfo)
	bossHallDao.SaveBossHallData(ctx, bossHallInfo)

	err = this_.svc.GetNatsClient().PublishCtx(ctx, curRes.BattleId, &cppbattle.CPPBattle_MonsterJoinKillPush{
		DropPos: &models.Vec2{
			X: req.BossPos.X,
			Y: req.BossPos.Y,
		},
		JoinItems: joinData.Rewards,
		HasJoin:   true,
		HasKill:   bossHallInfo.Info.KillTimes < killLimit,
	})
	if err != nil {
		this_.log.Error("push MonsterBossKillPush error", zap.Any("Error Msg", err))
	}
	err = this_.Module.JourneyService.AddToken(ctx, ctx.RoleId, valuesJourney.JourneyBossHall, 1)
	if err != nil {
		this_.log.Error("JourneyService.AddToken JourneyBossHall error", zap.Error(err))
	}
	ctx.PushMessage(&servicepb.BossHall_BossHallGetInfoResponse{
		LeftKillTimes: killLimit - bossHallInfo.Info.KillTimes,
		LeftJoinTimes: joinLimit - bossHallInfo.Info.JoinTimes,
		KillInfo:      bossKillJoinInfo.KillInfo,
		JoinInfo:      bossKillJoinInfo.JoinInfo,
	})
	this_.Module.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskBossHallJoinRewardAcc, 0, 1)
	this_.Module.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskBossHallJoinReward, 0, 1)
}

func (this_ *Service) CheckRefresh(ctx *ctx.Context, bossHallInfo *dao.BossHallData) *errmsg.ErrMsg {
	if timer.Now().Unix() < bossHallInfo.Info.NextRefreshTime {
		return nil
	}

	refreshTime, err := GetRefreshTime(ctx, this_.log)
	if err != nil {
		return err
	}

	beginTime := timer.BeginOfDay(timer.Now()).Unix()
	refreshTime += beginTime
	if timer.Now().Unix() > refreshTime {
		refreshTime += 86400
	}
	bossHallInfo.Info.NextRefreshTime = refreshTime
	bossHallInfo.Info.JoinTimes = 0
	bossHallInfo.Info.KillTimes = 0
	bossHallDao.SaveBossHallData(ctx, bossHallInfo)

	bossHallDao.SaveBossKillJoinData(ctx, &dao.BossKillJoinData{
		RoleId:          ctx.RoleId,
		NextRefreshTime: refreshTime,
	})
	return nil
}

func GetKillJoinLimit(ctx *ctx.Context, log *logger.Logger) (int64, int64, *errmsg.ErrMsg) {
	joinLimit, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("DevilSecretParticipationAwardNum")
	if !ok {
		log.Error("DevilSecretParticipationAwardNum error")
		return 0, 0, errmsg.NewErrBossHallConfig()
	}

	killLimit, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("DevilSecretKillingAwardNum")
	if !ok {
		log.Error("DevilSecretKillingAwardNum error")
		return 0, 0, errmsg.NewErrBossHallConfig()
	}
	return killLimit, joinLimit, nil
}

func GetRefreshTime(ctx *ctx.Context, log *logger.Logger) (int64, *errmsg.ErrMsg) {
	ActivityEveryDayRefreshTime, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("DefaultRefreshTime")
	if !ok {
		log.Error("DefaultRefreshTime error")
		return 0, errmsg.NewErrBossHallConfig()
	}
	return ActivityEveryDayRefreshTime, nil
}
