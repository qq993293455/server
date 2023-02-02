package refresh

import (
	"time"

	"coin-server/common/ctx"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/rule"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	module     *module.Module
	log        *logger.Logger
}

func NewRefreshService(
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
		module:     module,
		log:        log,
	}
	s.module.RefreshService = s
	return s
}

func (this_ *Service) GetCurrDayFreshTime(c *ctx.Context) time.Time {
	now := time.Unix(0, c.StartTime).UTC()
	return this_.GetCurrDayFreshTimeWith(c, now)
}

func (this_ *Service) GetCurrDayFreshTimeWith(c *ctx.Context, now time.Time) time.Time {

	r := rule.MustGetReader(c)
	defaultRefreshTime, ok := r.KeyValue.GetInt64("DefaultRefreshTime")
	if !ok {
		return timer.BeginOfDay(now)
	}
	begin := timer.BeginOfDay(now)
	offsetDuration := time.Second * time.Duration(defaultRefreshTime)
	currDayFreshTime := begin.Add(offsetDuration)

	if now.After(currDayFreshTime) {
		return currDayFreshTime
	}
	return timer.LastDay(begin).Add(offsetDuration)
}

func (this_ *Service) GetActivityCurrDayFreshTime(c *ctx.Context) time.Time {
	r := rule.MustGetReader(c)
	defaultRefreshTime, ok := r.KeyValue.GetInt64("ActivityEveryDayRefreshTime")
	offset := timer.Timer.GetOffset()
	now := time.Unix(0, c.StartTime).Add(offset).UTC()
	if !ok {
		return timer.BeginOfDay(now)
	}
	begin := timer.BeginOfDay(now)
	offsetDuration := time.Hour * time.Duration(defaultRefreshTime)
	currDayFreshTime := begin.Add(offsetDuration)

	if now.After(currDayFreshTime) {
		return currDayFreshTime
	}
	return timer.LastDay(begin).Add(offsetDuration)
}

func (this_ *Service) GetActivityCurrWeekFreshTime(c *ctx.Context) time.Time {
	r := rule.MustGetReader(c)
	defaultRefreshWeek, ok := r.KeyValue.GetInt64("DefaultRefreshWeekTime")
	offset := timer.Timer.GetOffset()
	now := time.Unix(0, c.StartTime).Add(offset).UTC()
	if !ok {
		return timer.BeginOfWeek(now)
	}
	defaultRefreshTime, ok := r.KeyValue.GetInt64("DefaultRefreshTime")
	begin := timer.BeginOfWeek(now)
	offsetDuration := 24*time.Hour*time.Duration(defaultRefreshWeek-1) + time.Second*time.Duration(defaultRefreshTime)
	currWeekFreshTime := begin.Add(offsetDuration)

	if now.After(currWeekFreshTime) {
		return currWeekFreshTime
	}
	return currWeekFreshTime.AddDate(0, 0, -7)
}
