package guide

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/guide/dao"
	"coin-server/game-server/util/trans"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewGuideService(
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
	svc.svc.RegisterFunc("获取引导信息", svc.Info)
	svc.svc.RegisterFunc("记录引导进度", svc.Record)
}

func (svc *Service) getGuide(ctx *ctx.Context, roleId values.RoleId) (ret *daopb.Guide, err *errmsg.ErrMsg) {
	ret, err = dao.Get(ctx, roleId)
	if err != nil {
		return
	}
	if ret == nil {
		ret = &daopb.Guide{
			RoleId: roleId,
		}
		dao.Save(ctx, ret)
	}
	return
}

// Info 获取引导信息
func (svc *Service) Info(ctx *ctx.Context, _ *servicepb.Guide_GuideInfoRequest) (*servicepb.Guide_GuideInfoResponse, *errmsg.ErrMsg) {
	gui, err := svc.getGuide(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.Guide_GuideInfoResponse{
		Info: trans.GuideD2M(gui),
	}, nil
}

// Record 记录进度
func (svc *Service) Record(ctx *ctx.Context, req *servicepb.Guide_RecordGuideRequest) (*servicepb.Guide_RecordGuideResponse, *errmsg.ErrMsg) {
	gui, err := svc.getGuide(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if req.GuideId == 0 && gui.GuideId != 0 {
		gui.FinishedGuideIds = append(gui.FinishedGuideIds, gui.GuideId)
	}
	gui.GuideId = req.GuideId
	gui.StepId = req.StepId
	dao.Save(ctx, gui)
	return &servicepb.Guide_RecordGuideResponse{}, nil
}
