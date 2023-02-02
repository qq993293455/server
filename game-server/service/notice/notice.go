package notice

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/proto/notice_service"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/module"
	"coin-server/game-server/service/notice/dao"
	json "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewNoticeService(
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
		log:        log,
		Module:     module,
	}
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取公告列表", svc.List)
	svc.svc.RegisterFunc("获取公告详情", svc.Details)
}

func (svc *Service) List(ctx *ctx.Context, _ *servicepb.Notice_NoticeListRequest) (*servicepb.Notice_NoticeListResponse, *errmsg.ErrMsg) {
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	readList, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	now := timer.StartTime(ctx.StartTime).Unix()
	readMap := make(map[string]struct{})
	delList := make([]*pbdao.NoticeRead, 0)
	for _, read := range readList {
		readMap[read.NoticeId] = struct{}{}
		if read.ExpiredAt <= now {
			delList = append(delList, read)
		}
	}
	out := &notice_service.Notice_NoticeListResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, 0, &notice_service.Notice_NoticeListRequest{}, out); err != nil {
		return nil, err
	}
	list := make([]*models.Notice, 0)
	for _, notice := range out.List {
		var read bool
		if _, ok := readMap[notice.Id]; ok {
			read = true
		}
		titleMap := make(map[values.Integer]string)
		if err := json.Unmarshal([]byte(notice.Title), &titleMap); err != nil {
			ctx.Error("unmarshal notice title err",
				zap.String("role_id", ctx.RoleId),
				zap.String("title", notice.Title),
				zap.Error(err),
				zap.Any("notice", notice),
			)
			continue
		}
		title, ok := titleMap[role.Language]
		if !ok {
			title = titleMap[enum.DefaultLanguage]
		}
		list = append(list, &models.Notice{
			Id:        notice.Id,
			Title:     title,
			ExpiredAt: notice.ExpiredAt,
			CreatedAt: notice.CreatedAt,
			Read:      read,
		})
	}
	if len(delList) > 0 {
		dao.Del(ctx, delList...)
	}
	return &servicepb.Notice_NoticeListResponse{
		List: list,
	}, nil
}

func (svc *Service) Details(ctx *ctx.Context, req *servicepb.Notice_NoticeDetailsRequest) (*servicepb.Notice_NoticeDetailsResponse, *errmsg.ErrMsg) {
	if req.NoticeId == "" {
		return &servicepb.Notice_NoticeDetailsResponse{}, nil
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	ok, err := dao.GetOne(ctx, req.NoticeId)
	if err != nil {
		return nil, err
	}
	out := &notice_service.Notice_NoticeDetailsResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, 0, &notice_service.Notice_NoticeDetailsRequest{
		Id: req.NoticeId,
	}, out); err != nil {
		return nil, err
	}
	if out.Content == nil {
		return nil, errmsg.NewErrNoticeExpired()
	}
	contentMap := make(map[values.Integer]string)
	if err := json.Unmarshal([]byte(out.Content.Content), &contentMap); err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	if !ok {
		dao.Save(ctx, &pbdao.NoticeRead{
			NoticeId:  req.NoticeId,
			ExpiredAt: out.Content.ExpiredAt,
		})
	}
	content, ok := contentMap[role.Language]
	if !ok {
		content = contentMap[enum.DefaultLanguage]
	}
	return &servicepb.Notice_NoticeDetailsResponse{
		Content: &models.NoticeContent{
			Content:       content,
			RewardContent: out.Content.RewardContent,
			IsCustom:      out.Content.IsCustom,
		},
	}, nil
}
