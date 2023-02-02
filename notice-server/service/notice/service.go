package notice

import (
	"fmt"
	"sync"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/proto/notice_service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/notice-server/dao"
	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	Notice     *sync.Map
}

func NewNoticeService(
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
		Notice:     &sync.Map{},
	}
	return s
}

func (svc *Service) Router() {
	gs := svc.svc.Group(handler.GameServerAuth)
	gs.RegisterFunc("获取公告列表", svc.List)
	gs.RegisterFunc("获取公告详情", svc.Details)

	gm := svc.svc.Group(handler.GMAuth)
	gm.RegisterFunc("更新公告", svc.Update)
	gm.RegisterFunc("删除公告", svc.Delete)
}

func (svc *Service) InitTask() {
	if ownerCron == nil {
		panic("ownerCron is nil")
	}
	list, err := dao.GetDao().Find()
	if err != nil {
		panic(fmt.Errorf("find notice error:%v", err))
	}
	for _, item := range list {
		ownerCron.AddCron(&Cron{
			NoticeId:   item.Id,
			OrderKey:   0,
			When:       item.ExpiredAt * 1000,
			RetryTimes: 0,
			Exec:       svc.noticeExpired,
		})
		svc.Notice.Store(item.Id, item)
	}
}

func (svc *Service) noticeExpired(cron *Cron) {
	if cron.NoticeId == "" {
		svc.log.Warn("notice id is empty")
		return
	}
	if err := dao.GetDao().Delete(cron.NoticeId); err != nil {
		svc.log.Error("delete notice err", zap.String("id", cron.NoticeId), zap.Error(err))
		if cron.Retry() {
			cron.RetryTimes++
			ownerCron.AddCron(cron)
			return
		}
	}
	svc.Notice.Delete(cron.NoticeId)
	svc.log.Debug("notice expired", zap.Any("id", cron.NoticeId))
}

func (svc *Service) List(_ *ctx.Context, _ *notice_service.Notice_NoticeListRequest) (*notice_service.Notice_NoticeListResponse, *errmsg.ErrMsg) {
	now := time.Now().Unix()
	list := make([]*models.Notice, 0)
	svc.Notice.Range(func(key, value any) bool {
		notice, ok := value.(*dao.Notice)
		if !ok {
			svc.log.Warn("assert err", zap.Any("key", key), zap.Any("value", value))
		}
		if notice.ExpiredAt > now && notice.BeginAt <= now {
			list = append(list, &models.Notice{
				Id:        notice.Id,
				Title:     notice.Title,
				ExpiredAt: notice.ExpiredAt,
				CreatedAt: notice.CreatedAt,
				Read:      false,
			})
		}
		return true
	})

	return &notice_service.Notice_NoticeListResponse{
		List: list,
	}, nil
}

func (svc *Service) Details(_ *ctx.Context, req *notice_service.Notice_NoticeDetailsRequest) (*notice_service.Notice_NoticeDetailsResponse, *errmsg.ErrMsg) {
	var find *dao.Notice
	svc.Notice.Range(func(key, value any) bool {
		notice, ok := value.(*dao.Notice)
		if !ok {
			svc.log.Warn("assert err", zap.Any("key", key), zap.Any("value", value))
		}
		if notice.Id == req.Id {
			find = notice
			return false
		}
		return true
	})
	return &notice_service.Notice_NoticeDetailsResponse{
		Content: &models.NoticeContent{
			Content:       find.Content,
			RewardContent: find.RewardContent,
			IsCustom:      find.IsCustom,
			ExpiredAt:     find.ExpiredAt,
		},
	}, nil
}

func (svc *Service) Update(_ *ctx.Context, req *notice_service.Notice_NoticeUpdateRequest) (*notice_service.Notice_NoticeUpdateResponse, *errmsg.ErrMsg) {
	if req.Data == nil || req.Content == nil {
		return &notice_service.Notice_NoticeUpdateResponse{}, nil
	}
	notice := &dao.Notice{
		Id:            req.Data.Id,
		Title:         req.Data.Title,
		Content:       req.Content.Content,
		RewardContent: req.Content.RewardContent,
		IsCustom:      req.Content.IsCustom,
		BeginAt:       req.BeginAt,
		ExpiredAt:     req.Data.ExpiredAt,
		CreatedAt:     req.Data.CreatedAt,
	}
	if err := dao.GetDao().Save(notice); err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	ownerCron.AddCron(&Cron{
		NoticeId:   req.Data.Id,
		OrderKey:   0,
		When:       req.Data.ExpiredAt * 1000,
		RetryTimes: 0,
		Exec:       svc.noticeExpired,
	})
	svc.Notice.Store(notice.Id, notice)
	return &notice_service.Notice_NoticeUpdateResponse{}, nil
}

func (svc *Service) Delete(_ *ctx.Context, req *notice_service.Notice_NoticeDeleteRequest) (*notice_service.Notice_NoticeDeleteResponse, *errmsg.ErrMsg) {
	if req.NoticeId == "" {
		return nil, nil
	}
	if err := dao.GetDao().Delete(req.NoticeId); err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	svc.Notice.Delete(req.NoticeId)
	ownerCron.RemoveCron(&Cron{NoticeId: req.NoticeId})
	return &notice_service.Notice_NoticeDeleteResponse{}, nil
}
