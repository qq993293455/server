package match

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/proto/racingrank_service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
}

func NewMatchService(
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
	h.RegisterFunc("定时任务检查", svc.TaskCheck)
}

func (svc *Service) InitTask() {
	// if ownerCron == nil {
	// 	panic("ownerCron is nil")
	// }
	// count := 1000 // 一次查询1000条
	// start := 0
	// end := count
	// daoIns := dao.GetDao()
	// all := make([]*dao.EndTime, 0)
	// for {
	// 	list, err := daoIns.BatchGetEndTime(start, end)
	// 	if err != nil {
	// 		panic(fmt.Errorf("BatchGetEndTime error:%v", err))
	// 	}
	// 	if len(list) > 0 {
	// 		start = end
	// 		end = start + count
	// 		all = append(all, list...)
	// 		time.Sleep(time.Millisecond * 5)
	// 	} else {
	// 		break
	// 	}
	// }
	// cronExecutor := GetCronExecutor()
	// for _, item := range all {
	// 	ownerCron.AddCron(&Cron{
	// 		RoleId:     item.RoleId,
	// 		OrderKey:   0,
	// 		When:       item.EndTime,
	// 		RetryTimes: 0,
	// 		Exec:       cronExecutor,
	// 	})
	// }
}

func (svc *Service) TaskCheck(ctx *ctx.Context, req *racingrank_service.RacingRankService_CronTaskCheckRequest) (*racingrank_service.RacingRankService_CronTaskCheckResponse, *errmsg.ErrMsg) {
	if ownerCron == nil {
		return nil, nil
	}
	// 已经过了结算时间的不会请求过来，所以这里不用判断结算时间
	if req.EndTime > timer.StartTime(ctx.StartTime).UnixMilli() && !ownerCron.OwnerCronExist(ctx.RoleId) {
		ownerCron.AddCron(&Cron{
			RoleId:     ctx.RoleId,
			OrderKey:   0,
			When:       req.EndTime,
			RetryTimes: 0,
			Exec:       GetCronExecutor(),
		})
	}
	return nil, nil
}
