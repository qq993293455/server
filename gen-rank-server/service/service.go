package service

import (
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/values"
	"go.uber.org/zap"

	"coin-server/gen-rank-server/service/gen"
)

type Service struct {
	svc        *service.Service
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
}

func (svc *Service) Stop() {
	svc.svc.Close()
}

func (svc *Service) Router() {
	gen.NewGenRankService(svc.serverId, svc.serverType, svc.svc, svc.log).Router()
}
