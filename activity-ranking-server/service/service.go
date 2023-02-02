package service

import (
	"time"

	activity_ranking "coin-server/activity-ranking-server/service/ranking"
	"coin-server/common/ctx"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"

	"go.uber.org/zap"
)

type Service struct {
	svc                    *service.Service
	log                    *logger.Logger
	serverId               values.ServerId
	serverType             models.ServerType
	activityRankingService *activity_ranking.Service
}

func NewService(
	urls []string,
	log *logger.Logger,
	serverId values.ServerId,
	serverType models.ServerType,
) *Service {
	svc := service.NewService(urls, log, serverId, serverType, true, false, eventlocal.CreateEventLocal(true))
	s := &Service{
		svc:        svc,
		log:        log,
		serverId:   serverId,
		serverType: serverType,
	}
	return s
}

func (svc *Service) Serve() {
	svc.Router()
	svc.svc.Start(func(event interface{}) {
		svc.log.Warn("unknown event", zap.Any("event", event))
	}, true)
	timer.Ticker(100*time.Millisecond, svc.Tick)
}

func (svc *Service) Stop(ctx *ctx.Context) {
	svc.activityRankingService.Save(ctx)
	svc.svc.Close()
}

func (svc *Service) Router() {
	svc.activityRankingService = activity_ranking.NewActivityRankingService(svc.serverId, svc.serverType, svc.svc, svc.log)
	svc.activityRankingService.Router()
	ctxx := ctx.GetContext()
	defer ctxx.NewOrm().Do()
	svc.activityRankingService.Init(ctxx)
}

func (svc *Service) Tick() bool {
	now := timer.Now().Unix()
	svc.activityRankingService.Tick(now)
	return true
}
