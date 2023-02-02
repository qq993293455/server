package fashion

import (
	"fmt"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/proto/fashion_service"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/fashion-server/dao"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
}

func NewFashionService(
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
	h.RegisterFunc("更新任务", svc.Update)
	h.RegisterFunc("删除任务", svc.Remove)
}

func (svc *Service) InitTask() {
	if ownerCron == nil {
		panic("ownerCron is nil")
	}
	count := 1000 // 一次查询1000条
	start := 0
	end := count
	daoIns := dao.GetDao()
	all := make([]*dao.Fashion, 0)
	for {
		list, err := daoIns.BatchGet(start, end)
		if err != nil {
			panic(fmt.Errorf("BatchGet error:%v", err))
		}
		if len(list) > 0 {
			start = end
			end = start + count
			all = append(all, list...)
			time.Sleep(time.Millisecond * 5)
		} else {
			break
		}
	}
	for _, item := range all {
		ownerCron.AddCron(&Cron{
			RoleId:     item.RoleId,
			UserId:     item.UserId,
			ServerId:   item.ServerId,
			FashionId:  item.FashionId,
			HeroId:     item.HeroId,
			OrderKey:   0,
			When:       item.ExpiredAt * 1000,
			RetryTimes: 0,
			Exec:       svc.fashionExpired,
		})
	}
}

func (svc *Service) fashionExpired(cron *Cron) {
	svc.log.Debug("fashionExpired", cronData2ZapField(cron, nil)...)
	if cron.RoleId == "" || cron.HeroId == 0 || cron.FashionId == 0 {
		svc.log.Warn("[fashionExpired] data is empty", cronData2ZapField(cron, nil)...)
		return
	}
	// 1.删除mysql数据
	if err := dao.GetDao().Delete(cron.RoleId, cron.FashionId); err != nil {
		svc.log.Error("delete fashion task err", cronData2ZapField(cron, err)...)
		if cron.Retry() {
			cron.RetryTimes++
			ownerCron.AddCron(cron)
			return
		}
	}
	// 2.通知gameserver
	if _, err := svc.svc.GetNatsClient().RequestProto(cron.ServerId, &models.ServerHeader{
		StartTime:  timer.Now().UnixNano(),
		RoleId:     cron.RoleId,
		ServerId:   cron.ServerId,
		ServerType: models.ServerType_FashionServer,
		UserId:     cron.UserId,
		InServerId: cron.ServerId,
		TraceId:    xid.New().String(),
	}, &servicepb.Hero_FashionExpiredRequest{
		HeroId:    cron.HeroId,
		FashionId: cron.FashionId,
	}); err != nil {
		svc.log.Error("send to game server err", cronData2ZapField(cron, err)...)
		if cron.Retry() {
			cron.RetryTimes++
			ownerCron.AddCron(cron)
			return
		}
	}
}

func cronData2ZapField(cron *Cron, err error) []zap.Field {
	return []zap.Field{
		zap.String("role_id", cron.RoleId),
		zap.String("user_id", cron.UserId),
		zap.Int64("server_id", cron.ServerId),
		zap.Int64("fashion_id", cron.FashionId),
		zap.Int64("hero_id", cron.HeroId),
		zap.Int64("order_key", cron.OrderKey),
		zap.Int64("when", cron.When),
		zap.Int64("retry_times", cron.RetryTimes),
		zap.Error(err),
	}
}

func (svc *Service) Update(ctx *ctx.Context, req *fashion_service.Fashion_FashionTimerUpdateRequest) (*fashion_service.Fashion_FashionTimerUpdateResponse, *errmsg.ErrMsg) {
	daoIns := dao.GetDao()
	ok, err := daoIns.Exist(ctx.RoleId, req.FashionId)
	if err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	if !ok {
		if err := daoIns.Save(&dao.Fashion{
			RoleId:    ctx.RoleId,
			UserId:    ctx.UserId,
			ServerId:  ctx.InServerId,
			HeroId:    req.HeroId,
			FashionId: req.FashionId,
			ExpiredAt: req.ExpiredAt,
			CreatedAt: time.Now().Unix(),
		}); err != nil {
			return nil, errmsg.NewInternalErr(err.Error())
		}
	} else {
		if err := daoIns.UpdateExpire(ctx.RoleId, req.FashionId, req.ExpiredAt); err != nil {
			return nil, errmsg.NewInternalErr(err.Error())
		}
	}
	ownerCron.AddCron(&Cron{
		RoleId:     ctx.RoleId,
		UserId:     ctx.UserId,
		ServerId:   ctx.InServerId,
		FashionId:  req.FashionId,
		HeroId:     req.HeroId,
		OrderKey:   0,
		When:       req.ExpiredAt * 1000,
		RetryTimes: 0,
		Exec:       svc.fashionExpired,
	})
	return &fashion_service.Fashion_FashionTimerUpdateResponse{}, nil
}

func (svc *Service) Remove(ctx *ctx.Context, req *fashion_service.Fashion_FashionTimerRemoveRequest) (*fashion_service.Fashion_FashionTimerRemoveResponse, *errmsg.ErrMsg) {
	if err := dao.GetDao().Delete(ctx.RoleId, req.FashionId); err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	ownerCron.RemoveCron(&Cron{
		RoleId:    ctx.RoleId,
		FashionId: req.FashionId,
	})
	return &fashion_service.Fashion_FashionTimerRemoveResponse{}, nil
}
