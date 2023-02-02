package divination

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/divination/dao"
	"coin-server/game-server/service/divination/rule"
	"coin-server/game-server/util"
	"coin-server/game-server/util/trans"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewDivinationService(
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
	svc.svc.RegisterFunc("获取占卜信息", svc.Info)
	svc.svc.RegisterFunc("占卜", svc.DivinationOnce)

	// 作弊器
	svc.svc.RegisterFunc("重置占卜次数", svc.CheatResetTimes)
	svc.svc.RegisterFunc("增加占卜次数上限", svc.CheatAddTotal)

	eventlocal.SubscribeEventLocal(svc.HandleRoleLvUpEvent)
	eventlocal.SubscribeEventLocal(svc.HandleExtraSkillTypTotal)
}

func (svc *Service) getDivination(ctx *ctx.Context, roleId values.RoleId) (ret *daopb.Divination, err *errmsg.ErrMsg) {
	ret, err = dao.Get(ctx, roleId)
	if err != nil {
		return
	}
	if ret == nil {
		ret = &daopb.Divination{
			RoleId:     roleId,
			TotalCount: rule.GetFreeTimes(ctx),
			ResetAt:    util.DefaultNextRefreshTime().UnixMilli(),
		}
		ret.AvailableCount = ret.TotalCount
		dao.Save(ctx, ret)
	}
	return
}

// Info 获取占卜信息
func (svc *Service) Info(ctx *ctx.Context, _ *servicepb.Divination_DivinationInfoRequest) (*servicepb.Divination_DivinationInfoResponse, *errmsg.ErrMsg) {
	div, err := svc.getDivination(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.Divination_DivinationInfoResponse{
		Info: trans.DivinationD2M(div),
	}, nil
}

// DivinationOnce 占卜
func (svc *Service) DivinationOnce(ctx *ctx.Context, _ *servicepb.Divination_DivinationOnceRequest) (*servicepb.Divination_DivinationOnceResponse, *errmsg.ErrMsg) {
	div, err := svc.getDivination(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if div.AvailableCount <= 0 {
		return nil, errmsg.NewErrTimesNotEnough()
	}
	level, err := svc.UserService.GetLevel(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	balls, exps := rule.DrawEnergy(ctx, level, int(div.AvailableCount))
	// 经验走自动升级流程不再直接往背包里加
	//err = svc.BagService.AddItem(ctx, ctx.RoleId, enum.RoleExp, exps)
	//if err != nil {
	//	return nil, err
	//}
	if err := svc.AddExp(ctx, ctx.RoleId, exps, false); err != nil {
		return nil, err
	}

	svc.TaskService.UpdateTargets(ctx, ctx.RoleId, map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskDivinationNum: {
			Typ:     values.Integer(models.TaskType_TaskDivinationNum),
			Id:      0,
			Cnt:     div.AvailableCount,
			Replace: false,
		},
		models.TaskType_TaskDivinationNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskDivinationNumAcc),
			Id:      0,
			Cnt:     div.AvailableCount,
			Replace: false,
		},
	})
	div.AvailableCount = 0
	dao.Save(ctx, div)

	return &servicepb.Divination_DivinationOnceResponse{
		Info:        trans.DivinationD2M(div),
		EnergyBalls: balls,
	}, nil
}

func (svc *Service) addTotal(c *ctx.Context, roleId values.RoleId, num values.Integer) (*daopb.Divination, *errmsg.ErrMsg) {
	div, err := svc.getDivination(c, roleId)
	if err != nil {
		return nil, err
	}
	div.TotalCount += num
	div.AvailableCount += num
	dao.Save(c, div)
	return div, nil
}

// CheatResetTimes 作弊器 重置占卜次数
func (svc *Service) CheatResetTimes(ctx *ctx.Context, _ *servicepb.Divination_CheatResetDivinationRequest) (*servicepb.Divination_CheatResetDivinationResponse, *errmsg.ErrMsg) {
	div, err := svc.getDivination(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	div.AvailableCount = div.TotalCount
	dao.Save(ctx, div)
	return &servicepb.Divination_CheatResetDivinationResponse{
		Info: trans.DivinationD2M(div),
	}, nil
}

// CheatAddTotal 作弊器 增加上限
func (svc *Service) CheatAddTotal(ctx *ctx.Context, req *servicepb.Divination_CheatAddDivinationTotalRequest) (*servicepb.Divination_CheatAddDivinationTotalResponse, *errmsg.ErrMsg) {
	if req.Num <= 0 {
		req.Num = 1
	}
	div, err := svc.addTotal(ctx, ctx.RoleId, req.Num)
	if err != nil {
		return nil, err
	}
	return &servicepb.Divination_CheatAddDivinationTotalResponse{
		Info: trans.DivinationD2M(div),
	}, nil
}
