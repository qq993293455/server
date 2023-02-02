package service

import (
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	nostatesvc "coin-server/role-state-server/no-role-service"
	"coin-server/role-state-server/service/state"
	"coin-server/role-state-server/service/state/raft"
	"go.uber.org/zap"
)

type Service struct {
	svc        *nostatesvc.NoRoleService
	log        *logger.Logger
	serverId   values.ServerId
	serverType models.ServerType
	opt        *raft.Options
}

func NewService(
	urls []string,
	log *logger.Logger,
	serverId values.ServerId,
	serverType models.ServerType,
	opt *raft.Options,
) *Service {
	svc := nostatesvc.NewNoRoleService(urls, log, serverId, serverType, true, false, eventlocal.CreateEventLocal(true))
	s := &Service{
		svc:        svc,
		log:        log,
		serverId:   serverId,
		serverType: serverType,
		opt:        opt,
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
	state.NewStateService(svc.serverId, svc.serverType, svc.svc, svc.log, svc.opt).Router()
}
