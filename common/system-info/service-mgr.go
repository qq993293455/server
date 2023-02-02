package system_info

import (
	"time"

	"coin-server/common/logger"
	"coin-server/common/proto/broadcast"
	"coin-server/common/timer"
)

type ServiceStatusChange interface {
	OnServiceNew(s *broadcast.Stats_ServerStats)
	OnServiceLose(s *broadcast.Stats_ServerStats)
}

type ServiceStatusChangeBaseEvent struct {
	log *logger.Logger
}

func (this_ *ServiceStatusChangeBaseEvent) OnServiceNew(s *broadcast.Stats_ServerStats) {
	// this_.log.Info("OnServiceNew", zap.Int64("new_server_id", s.ServerId), zap.String("new_server_type", s.ServerType.String()))
}

func (this_ *ServiceStatusChangeBaseEvent) OnServiceLose(s *broadcast.Stats_ServerStats) {
	//this_.log.Warn("OnServiceLose", zap.Int64("lose_server_id", s.ServerId), zap.String("lose_server_type", s.ServerType.String()))
}

type ServiceMgr struct {
	servers          map[string]*broadcast.Stats_ServerStats
	log              *logger.Logger
	ssc              ServiceStatusChange
	uniqueId         string
}

func NewServiceMgr(uniqueId string, log *logger.Logger, ssc ServiceStatusChange) *ServiceMgr {
	s := &ServiceMgr{
		log:      log,
		uniqueId: uniqueId,
		servers:  make(map[string]*broadcast.Stats_ServerStats),
	}
	if ssc != nil {
		s.ssc = &ServiceStatusChangeBaseEvent{log: log}
	} else {
		s.ssc = ssc
	}
	return s
}

func (this_ *ServiceMgr) AddOrSet(s *broadcast.Stats_ServerStats) {
	if _, ok := this_.servers[s.UniqueId]; ok {
		this_.servers[s.UniqueId] = s
	} else {
		this_.servers[s.UniqueId] = s
		if this_.ssc != nil && s.UniqueId != this_.uniqueId {
			this_.ssc.OnServiceNew(s)
		}
	}
}

func (this_ *ServiceMgr) Remove(uniqueId string) {
	if v, ok := this_.servers[uniqueId]; ok {
		delete(this_.servers, uniqueId)
		if this_.ssc != nil && v.UniqueId != this_.uniqueId {
			this_.ssc.OnServiceLose(v)
		}
	}
}

func (this_ *ServiceMgr) CheckLose() {
	now := timer.Now()
	for uniqueId, s := range this_.servers {
		if now.Sub(time.Unix(0, s.Timestamp)) > time.Second*10 {
			this_.Remove(uniqueId)
		}
	}
}
