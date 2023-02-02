package service

import (
	arena "coin-server/arena-server/service/arena"
	"coin-server/common/ctx"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	svc          *service.Service
	log          *logger.Logger
	serverId     values.ServerId
	serverType   models.ServerType
	arena_server *arena.Service
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
	timer.Ticker(time.Minute, svc.Tick)
}

func (svc *Service) Stop(ctx *ctx.Context) {
	svc.arena_server.Save(ctx)
	svc.svc.Close()
}

func (svc *Service) Router() {
	svc.arena_server = arena.NewArenaService(svc.serverId, svc.serverType, svc.svc, svc.log)
	svc.arena_server.Router()
	svc.arena_server.Init()
}

func (svc *Service) Tick() bool {
	now := timer.Now().Unix()
	svc.arena_server.Tick(now)
	return true
}
