package activity

import (
	"fmt"
	"time"

	rule2 "coin-server/activity-server/service/activity/rule"
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/proto/activity_service"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/rule"

	"go.uber.org/zap"

	"github.com/gogo/protobuf/proto"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
}

func NewActivityService(
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

func (svc *Service) Router() {
	h := svc.svc.Group(handler.GameServerAuth)
	h.RegisterFunc("获取所有正在进行的活动", svc.List)
}

func (svc *Service) InitTask() {
	if endOC == nil || beginOC == nil {
		panic("ownerCron is nil")
	}
	now := time.Now().UnixMilli()
	list := rule.MustGetReader(ctx.GetContext()).Activity.List()
	for _, item := range list {
		if item.TimeType != enum.AbsoluteTime {
			continue
		}
		begin, err := time.ParseInLocation("2006-01-02 15:04:05", item.ActivityOpenTime, time.UTC)
		if err != nil {
			panic(err)
		}
		end, err := time.ParseInLocation("2006-01-02 15:04:05", item.DurationTime, time.UTC)
		if err != nil {
			panic(err)
		}
		if begin.UnixMilli() >= end.UnixMilli() {
			panic(fmt.Errorf("活动开始时间必须小于结束时间：%d", item.Id))
		}
		if begin.UnixMilli() > now {
			beginOC.AddCron(&Cron{
				Id:         item.Id,
				ActivityId: item.Id,
				OrderKey:   0,
				When:       begin.UnixMilli(),
				RetryTimes: 0,
				Exec:       svc.activityStartedPush,
			})
		}
		if end.UnixMilli() > now {
			endOC.AddCron(&Cron{
				Id:         item.Id,
				ActivityId: item.Id,
				OrderKey:   0,
				When:       end.UnixMilli(),
				RetryTimes: 0,
				Exec:       svc.activityEndedPush,
			})
		}
	}

	list2 := rule.MustGetReader(ctx.GetContext()).ActivityCircular.List()
	for _, item := range list2 {
		if item.TimeType != enum.AbsoluteTime {
			continue
		}
		begin, err := time.ParseInLocation("2006-01-02 15:04:05", item.ActivityOpenTime, time.UTC)
		if err != nil {
			panic(err)
		}
		end, err := time.ParseInLocation("2006-01-02 15:04:05", item.DurationTime, time.UTC)
		if err != nil {
			panic(err)
		}
		if begin.UnixMilli() >= end.UnixMilli() {
			panic(fmt.Errorf("活动开始时间必须小于结束时间：%d", item.Id))
		}
		if begin.UnixMilli() > now {
			beginOC.AddCron(&Cron{
				Id:         item.Id,
				ActivityId: item.ActivityId,
				OrderKey:   0,
				When:       begin.UnixMilli(),
				RetryTimes: 0,
				Exec:       svc.activityStartedPush,
			})
		}
		if end.UnixMilli() > now {
			endOC.AddCron(&Cron{
				Id:         item.Id,
				ActivityId: item.ActivityId,
				OrderKey:   0,
				When:       end.UnixMilli(),
				RetryTimes: 0,
				Exec:       svc.activityEndedPush,
			})
		}
	}
}

func (svc *Service) List(ctx *ctx.Context, _ *activity_service.ActivityService_ActivityListRequest) (*activity_service.ActivityService_ActivityListResponse, *errmsg.ErrMsg) {
	// 在gameserver读配置，不需要走这里
	return nil, nil
	// cfg, _ := rule2.GetAllAvailableActivity(ctx)
	// list := make([]*models.Activity, 0, len(cfg))
	// for _, item := range cfg {
	// 	model, err := NewActivityModel(item)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	list = append(list, model)
	// }
	// return &activity_service.ActivityService_ActivityListResponse{
	// 	List: list,
	// }, nil
}

func (svc *Service) activityStartedPush(activityId values.Integer) {
	cfg, ok := rule2.GetEventById(ctx.GetContext(), activityId)
	if !ok {
		return
	}
	activity, err := NewActivityModel(cfg)
	if err != nil {
		svc.log.Error("activity config to proto model error", zap.Int64("id", cfg.Id), zap.Error(err))
		return
	}
	svc.svc.PushToAllOnlineClient([]proto.Message{&servicepb.Activity_ActivityStartedPush{
		Activity: activity,
	}})
}

func (svc *Service) activityEndedPush(activityId values.Integer) {
	cfg, ok := rule2.GetEventById(ctx.GetContext(), activityId)
	if !ok {
		return
	}
	activity, err := NewActivityModel(cfg)
	if err != nil {
		svc.log.Error("activity config to proto model error", zap.Int64("id", cfg.Id), zap.Error(err))
		return
	}
	svc.svc.PushToAllOnlineClient([]proto.Message{&servicepb.Activity_ActivityEndedPush{
		Activity: activity,
	}})
}
