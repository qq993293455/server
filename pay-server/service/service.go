package service

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/orm"
	"coin-server/common/pay"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/redisclient"
	"coin-server/common/service"
	"coin-server/common/statistical2"
	models3 "coin-server/common/statistical2/models"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/pay-server/service/dao"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type Service struct {
	svc        *service.Service
	log        *logger.Logger
	serverId   values.ServerId
	serverType models.ServerType
	stopped    int32
	done       chan struct{}
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
		done:       make(chan struct{}),
	}
	return s
}

func (svc *Service) Serve() {
	go svc.startConsumer()
	svc.svc.Start(func(event interface{}) {
		svc.log.Warn("unknown event", zap.Any("event", event))
	}, true)
}

func (svc *Service) Stop() {
	svc.svc.Close()
	svc.stopConsumer()
}

func (svc *Service) stopConsumer() {
	close(svc.done)
	for atomic.LoadInt32(&svc.stopped) == 0 {
		time.Sleep(time.Millisecond * 10)
	}
}

func (svc *Service) startConsumer() {
	defer func() {
		svc.log.Debug("consumer stopped")
	}()
	for {
		select {
		case <-svc.done:
			atomic.StoreInt32(&svc.stopped, 1)
			return
		default:
		}
		val, err := pay.GetPayQueueRedis().LPop(context.Background(), pay.PayQueueKey).Result()
		if err != nil {
			time.Sleep(time.Millisecond * 10)
			if errors.Is(err, redisclient.Nil) {
				time.Sleep(time.Second * 5)
			} else {
				svc.log.Error("LPop err", zap.Error(err))
			}
			continue
		}
		data := &pbdao.PayQueue{}

		if err := orm.ProtoUnmarshal(utils.StringToBytes(val), data); err != nil {
			svc.log.Error("proto.Unmarshal err",
				zap.String("data", val),
				zap.Error(err))
			time.Sleep(time.Millisecond * 10)
			continue
		}
		if err := dao.SavePay(data.RoleId, &pbdao.Pay{
			Sn:         data.Sn,
			PcId:       data.PcId,
			PaidTime:   data.PaidTime,
			ExpireTime: data.ExpireTime,
			CreatedAt:  time.Now().Unix(),
		}); err != nil {
			svc.log.Error("SavePay err",
				zap.String("role_id", data.RoleId),
				zap.String("sn", data.Sn),
				zap.Int64("pc_id", data.PcId),
				zap.Int64("paid_time", data.PaidTime),
				zap.Int64("expire_time", data.ExpireTime),
				zap.Error(err),
			)
			svc.save2queue(data)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		if err := svc.svc.GetNatsClient().Publish(data.ServerId, &models.ServerHeader{
			StartTime:  time.Now().UnixNano(),
			RoleId:     data.RoleId,
			ServerId:   data.ServerId,
			ServerType: models.ServerType_PayServer,
			InServerId: data.ServerId,
		}, &servicepb.Pay_Success{
			PcId:       data.PcId,
			PaidTime:   data.PaidTime,
			ExpireTime: data.ExpireTime,
		}); err != nil {
			svc.log.Error("nats publish err",
				zap.String("sn", data.Sn),
				zap.String("role_id", data.RoleId),
				zap.Int64("server_id", data.ServerId),
				zap.Int64("pc_id", data.PcId),
				zap.Int64("paid_time", data.PaidTime),
				zap.Int64("expire_time", data.ExpireTime),
				zap.Error(err))
			svc.save2queue(data)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		go svc.saveRecord(data)

		svc.log.Debug("nats publish success",
			zap.String("sn", data.Sn),
			zap.String("role_id", data.RoleId),
			zap.Int64("server_id", data.ServerId),
			zap.Int64("pc_id", data.PcId),
			zap.Int64("paid_time", data.PaidTime),
			zap.Int64("expire_time", data.ExpireTime))
	}
}

func (svc *Service) save2queue(data *pbdao.PayQueue) {
	if err := dao.Save2Queue(data); err != nil {
		svc.log.Error("save2queue err",
			zap.String("sn", data.Sn),
			zap.String("role_id", data.RoleId),
			zap.Int64("server_id", data.ServerId),
			zap.Int64("pc_id", data.PcId),
			zap.Int64("paid_time", data.PaidTime),
			zap.Int64("expire_time", data.ExpireTime),
			zap.Error(err))
	}
}

func (svc *Service) saveRecord(data *pbdao.PayQueue) {
	ch := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		ls := statistical2.GetLogServer(context.Background())
		statistical2.Save(ls, &models3.Pay{
			Time:       time.Now(),
			Xid:        xid.New().String(),
			Sn:         data.Sn,
			RoleId:     data.RoleId,
			ServerId:   data.ServerId,
			PcId:       data.PcId,
			PaidTime:   data.PaidTime,
			ExpireTime: data.ExpireTime,
		})
		if err := statistical2.Flush(ls); err != nil {
			svc.log.Error("saveRecord err",
				zap.String("sn", data.Sn),
				zap.String("role_id", data.RoleId),
				zap.Int64("server_id", data.ServerId),
				zap.Int64("pc_id", data.PcId),
				zap.Int64("paid_time", data.PaidTime),
				zap.Int64("expire_time", data.ExpireTime),
				zap.Error(err))
		}
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		svc.log.Debug("saveRecord success")
	case <-ctx.Done():
		svc.log.Error("saveRecord timeout", zap.Error(ctx.Err()))
	}
}
