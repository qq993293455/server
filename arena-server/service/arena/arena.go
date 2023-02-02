package arena

import (
	"coin-server/arena-server/service/arena/ranking"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/arena_service"
	"coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	ranking    *ranking.ArenaRankingManager
}

func NewArenaService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		log:        log,
	}
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("查询排行信息", this_.GetArenaRankingDatas)
	this_.svc.RegisterFunc("查询自己的排行信息", this_.GetArenaRankingSelfData)
	this_.svc.RegisterFunc("加入排行榜", this_.JoinRanking)
	this_.svc.RegisterFunc("获取挑战列表", this_.GetChallengeRange)
	this_.svc.RegisterFunc("查询玩家是否为机器人", this_.GetPlayerIsRobot)
	this_.svc.RegisterFunc("交换排行位置", this_.RankingSwap)
	this_.svc.RegisterFunc("设置是否参与挑战", this_.SetFirstChallenge)
	this_.svc.RegisterFunc("获取发奖时间", this_.GetRewardTime)
	this_.svc.RegisterFunc("锁定玩家", this_.LockRoleRanking)
	this_.svc.RegisterFunc("解锁玩家", this_.UnLockRoleRanking)
}

func (this_ *Service) Init() {
	this_.ranking = ranking.NewarenaRanking(this_.log, this_.svc)
	err := this_.ranking.Init(this_.serverId)
	if err != nil {
		panic(err)
	}
	this_.ranking.Tick(timer.Now().Unix())
}

func (this_ *Service) Save(ctx *ctx.Context) {
	this_.ranking.Save(ctx)
}

func (this_ *Service) Tick(now int64) {
	this_.ranking.Tick(now)
}

func (this_ *Service) GetArenaRankingDatas(ctx *ctx.Context, req *arena_service.ArenaRanking_GetDatasRequest) (*arena_service.ArenaRanking_GetDatasResponse, *errmsg.ErrMsg) {
	Data, err := this_.ranking.GetRanking(req.Type, req.RankingIndex, req.StartIndex, req.StartIndex+req.Count)
	if err != nil {
		return nil, err
	}
	return &arena_service.ArenaRanking_GetDatasResponse{
		Datas: Data,
	}, nil
}

func (this_ *Service) GetArenaRankingSelfData(ctx *ctx.Context, req *arena_service.ArenaRanking_GetSelfDataRequest) (*arena_service.ArenaRanking_GetSelfDataResponse, *errmsg.ErrMsg) {
	Data, err := this_.ranking.GetSelfRanking(req.Type, req.RankingIndex, req.RoleId)
	if err != nil {
		return nil, err
	}
	{
		this_.log.Debug("this selfData", zap.Any("rankingId", Data.Element.RankingId), zap.Any("rankingIndex", req.RankingIndex))
		local_time := time.Unix(Data.NextSettlementTime, 0).UTC()
		this_.log.Debug("next day settlement time", zap.Any("Time", fmt.Sprintf("%02d:%02d:%02d", local_time.Hour(), local_time.Minute(), local_time.Second())))
	}

	{
		this_.log.Debug("this selfData", zap.Any("rankingId", Data.Element.RankingId), zap.Any("rankingIndex", req.RankingIndex))
		local_time := time.Unix(Data.SeasonOverTime, 0).UTC()
		this_.log.Debug("next season settlement time", zap.Any("Time", fmt.Sprintf("%02d:%02d:%02d", local_time.Hour(), local_time.Minute(), local_time.Second())))
	}
	return Data, nil
}

func (this_ *Service) JoinRanking(ctx *ctx.Context, req *arena_service.ArenaRanking_JoinRankingRequest) (*arena_service.ArenaRanking_JoinRankingResponse, *errmsg.ErrMsg) {
	rankingIndex, err := this_.ranking.JoinRanking(ctx, req.Type, req.PlayerInfo.RoleId)
	if err != nil {
		return nil, err
	}
	return &arena_service.ArenaRanking_JoinRankingResponse{
		RankingIndex: rankingIndex,
	}, nil
}

func (this_ *Service) GetChallengeRange(ctx *ctx.Context, req *arena_service.ArenaRanking_GetChallengeRangeRequest) (*arena_service.ArenaRanking_GetChallengeRangeResponse, *errmsg.ErrMsg) {
	Datas, err := this_.ranking.GetChanllengeRange(req.Type, req.RankingIndex, req.RoleId)
	if err != nil {
		return nil, err
	}
	return &arena_service.ArenaRanking_GetChallengeRangeResponse{
		Datas: Datas,
	}, nil
}

func (this_ *Service) GetPlayerIsRobot(ctx *ctx.Context, req *arena_service.ArenaRanking_GetPlayerIsRobotRequest) (*arena_service.ArenaRanking_GetPlayerIsRobotResponse, *errmsg.ErrMsg) {
	IsRobot, NickName, err := this_.ranking.GetPlayerIsRobot(req.Type, req.RankingIndex, req.RoleId)
	if err != nil {
		return nil, err
	}
	return &arena_service.ArenaRanking_GetPlayerIsRobotResponse{
		IsRobot:  IsRobot,
		NickName: NickName,
	}, nil
}

func (this_ *Service) RankingSwap(ctx *ctx.Context, req *arena_service.ArenaRanking_SwapRankingIdRequest) (*arena_service.ArenaRanking_SwapRankingIdResponse, *errmsg.ErrMsg) {
	ret, isRobot, nickname, err := this_.ranking.ArenaRankingSwap(ctx, req.Type, req.RankingIndex, req.RoleId_0, req.RoleId_1)
	if err != nil {
		return nil, err
	}
	return &arena_service.ArenaRanking_SwapRankingIdResponse{
		SwapInfo: ret,
		IsRobot:  isRobot,
		NickName: nickname,
	}, nil
}

func (this_ *Service) SetFirstChallenge(ctx *ctx.Context, req *arena_service.ArenaRanking_SetFirstChallengeRequest) (*arena_service.ArenaRanking_SetFirstChallengeResponse, *errmsg.ErrMsg) {
	err := this_.ranking.ArenaRaningSetFirstChallenge(ctx, req.Type, req.RankingIndex, req.RoleId, req.Flg)
	if err != nil {
		return nil, err
	}
	return &arena_service.ArenaRanking_SetFirstChallengeResponse{}, nil
}

func (this_ *Service) GetRewardTime(ctx *ctx.Context, req *arena_service.ArenaRanking_ArenaGetRewardTimeRequest) (*arena_service.ArenaRanking_ArenaGetRewardTimeResponse, *errmsg.ErrMsg) {
	return this_.ranking.GetRewardTime(ctx), nil
}

func (this_ *Service) LockRoleRanking(ctx *ctx.Context, req *arena_service.ArenaRanking_LockRoleRankingRequest) (*arena_service.ArenaRanking_LockRoleRankingResponse, *errmsg.ErrMsg) {
	err := this_.ranking.ArenaLockRoleRanking(ctx, req.Type, req.RankingIndex, req.ChallengeRoleid, req.ChallengeRankingId, req.ChallengedRoleid, req.ChallengedRankingId)
	if err != nil {
		return nil, err
	}
	return &arena_service.ArenaRanking_LockRoleRankingResponse{}, nil
}

func (this_ *Service) UnLockRoleRanking(ctx *ctx.Context, req *arena_service.ArenaRanking_UnLockRoleRankingRequest) (*arena_service.ArenaRanking_UnLockRoleRankingResponse, *errmsg.ErrMsg) {
	err := this_.ranking.ArenaUnLockRoleRanking(ctx, req.Type, req.RankingIndex, req.RoleId)
	if err != nil {
		return nil, err
	}
	return &arena_service.ArenaRanking_UnLockRoleRankingResponse{}, nil
}
