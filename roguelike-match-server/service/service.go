package service

import (
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/routine_limit_service"
	"coin-server/common/values"
	"go.uber.org/zap"

	"coin-server/roguelike-match-server/service/match"
)

type Service struct {
	svc        *routine_limit_service.RoutineLimitService
	log        *logger.Logger
	serverId   values.ServerId
	serverType models.ServerType
}

func NewService(
	urls []string,
	log *logger.Logger,
	serverId values.ServerId,
	serverType models.ServerType,
) *Service {
	svc := routine_limit_service.NewRoutineLimitService(urls, log, serverId, serverType, true, true, eventlocal.CreateEventLocal(true))
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
}

func (svc *Service) Stop() {
	svc.svc.Close()
}

func (svc *Service) Router() {
	match.NewMatchService(svc.serverId, svc.serverType, svc.svc, svc.log).Router()
}
