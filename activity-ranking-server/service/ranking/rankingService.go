package activity_ranking

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/activity_ranking_service"
	"coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	ranking    *RankingManager
}

func NewActivityRankingService(
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
	this_.svc.RegisterFunc("加入排行榜", this_.JoinRanking)
	this_.svc.RegisterFunc("更新排行", this_.UpdateRankingData)
	this_.svc.RegisterFunc("获取自己排行数据", this_.GetSelfRanking)
	this_.svc.RegisterFunc("获取排行数据", this_.GetRankingList)
}

func (this_ *Service) Init(ctx *ctx.Context) {
	this_.ranking = NewActivityRanking(this_.log, this_.svc)
	err := this_.ranking.Init(ctx, this_.serverId)
	if err != nil {
		panic(err)
	}
	this_.ranking.Tick(timer.Now().Unix())
}

func (this_ *Service) Save(ctx *ctx.Context) {
	this_.ranking.Save(ctx, nil)
}

func (this_ *Service) Tick(now int64) {
	this_.ranking.Tick(now)
}

func (this_ *Service) JoinRanking(ctx *ctx.Context, req *activity_ranking_service.ActivityRanking_JoinRankingRequest) (*activity_ranking_service.ActivityRanking_JoinRankingResponse, *errmsg.ErrMsg) {
	rankingIndex, err := this_.ranking.Join(ctx, req)
	if err != nil {
		return nil, err
	}
	return &activity_ranking_service.ActivityRanking_JoinRankingResponse{
		RankingIndex: rankingIndex,
	}, nil
}

func (this_ *Service) UpdateRankingData(ctx *ctx.Context, req *activity_ranking_service.ActivityRanking_UpdateRankingDataRequest) (*activity_ranking_service.ActivityRanking_UpdateRankingDataRequest, *errmsg.ErrMsg) {
	err := this_.ranking.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return &activity_ranking_service.ActivityRanking_UpdateRankingDataRequest{}, nil
}

func (this_ *Service) GetSelfRanking(ctx *ctx.Context, req *activity_ranking_service.ActivityRanking_GetSelfRankingRequest) (*activity_ranking_service.ActivityRanking_GetSelfRankingResponse, *errmsg.ErrMsg) {
	msg, err := this_.ranking.GetSelfRankingInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (this_ *Service) GetRankingList(ctx *ctx.Context, req *activity_ranking_service.ActivityRanking_GetRankingListRequest) (*activity_ranking_service.ActivityRanking_GetRankingListResponse, *errmsg.ErrMsg) {
	msg, err := this_.ranking.GetRankingList(ctx, req)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
