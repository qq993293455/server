package system_unlock

import (
	"strings"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/handler"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/system-unlock/dao"
	"coin-server/rule"

	"github.com/gogo/protobuf/proto"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
}

func NewSysUnlockService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
	}
	s.SystemUnlockService = s
	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("获取系统解锁信息", s.GetSysUnlock)
	s.svc.RegisterFunc("点击解锁系统", s.ClickUnlock)
	s.svc.RegisterFunc("作弊器解锁所有系统", s.CheatUnlockSys)
	eventlocal.SubscribeEventLocal(s.HandleLoginEvent)
	eventlocal.SubscribeEventLocal(s.HandleSystemUnlock)

	s.RegisterCondition()
}

func (s *Service) GetSysUnlock(ctx *ctx.Context, _ *servicepb.SystemUnlock_GetSystemUnlockRequest) (*servicepb.SystemUnlock_GetSystemUnlockResponse, *errmsg.ErrMsg) {
	sysUnlock, err := dao.GetSysUnlock(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	temp := make([]values.SystemId, 0)
	r := rule.MustGetReader(ctx)
	for idx := range r.System.List() {
		sys := r.System.List()[idx]
		if !sysUnlock.Unlock[sys.Id] {
			if len(r.System.List()[idx].UnlockCondition) < 3 {
				sysUnlock.Unlock[sys.Id] = true
				// cache.SetUnlock(ctx, models.SystemType(sys.Id))
				temp = append(temp, models.SystemType(sys.Id))
				continue
			}
			typ, id, cnt := models.TaskType(sys.UnlockCondition[0]), sys.UnlockCondition[1], sys.UnlockCondition[2]
			_, unlock, err1 := s.TaskService.CheckCondition(ctx, typ, id, cnt)
			if err1 != nil {
				return nil, err
			}
			if unlock {
				sysUnlock.Unlock[sys.Id] = true
				// cache.SetUnlock(ctx, models.SystemType(sys.Id))
				temp = append(temp, models.SystemType(sys.Id))
			}
			sysUnlock.Unlock[sys.Id] = false
		}
	}
	if len(temp) != 0 {
		ctx.PublishEventLocal(&event.SystemUnlock{SystemId: temp})
		dao.SaveSysUnlock(ctx, sysUnlock)
	}
	sysId := make([]values.SystemId, 0)
	clicked := make([]values.SystemId, 0)
	for k, v := range sysUnlock.Unlock {
		if v {
			sysId = append(sysId, models.SystemType(k))
		}
	}
	for k, v := range sysUnlock.Click {
		if v {
			clicked = append(clicked, models.SystemType(k))
		}
	}
	return &servicepb.SystemUnlock_GetSystemUnlockResponse{SystemId: sysId, ClickId: clicked}, nil
}

func (s *Service) ClickUnlock(ctx *ctx.Context, req *servicepb.SystemUnlock_UnlockSystemRequest) (*servicepb.SystemUnlock_UnlockSystemResponse, *errmsg.ErrMsg) {
	unlock, err := dao.GetSysUnlock(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if _, ok := unlock.Click[int64(req.SystemId)]; !ok {
		unlock.Click[int64(req.SystemId)] = true
		dao.SaveSysUnlock(ctx, unlock)
	}

	return &servicepb.SystemUnlock_UnlockSystemResponse{}, nil
}

func (s *Service) CheatUnlockSys(ctx *ctx.Context, req *servicepb.SystemUnlock_CheatUnlockSystemRequest) (*servicepb.SystemUnlock_CheatUnlockSystemResponse, *errmsg.ErrMsg) {
	unlock, err := dao.GetSysUnlock(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	temp := make([]values.SystemId, 0)
	r := rule.MustGetReader(ctx)
	if req.SystemId == models.SystemType_SystemPadding {
		list := r.System.List()
		for idx := range list {
			un := unlock.Unlock[list[idx].Id]
			if !un {
				unlock.Unlock[list[idx].Id] = true
				// cache.SetUnlock(ctx, models.SystemType(list[idx].Id))
				temp = append(temp, values.SystemId(list[idx].Id))
			}
		}
		if len(temp) != 0 {
			ctx.PublishEventLocal(&event.SystemUnlock{SystemId: temp})
			dao.SaveSysUnlock(ctx, unlock)
		}
		// cache.init(ctx, unlock.Unlock)
	} else {
		cfg, ok := r.System.GetSystemById(values.Integer(req.SystemId))
		if !ok {
			return nil, nil
		}
		temp = append(temp, req.SystemId)
		unlock.Unlock[cfg.Id] = true
		// cache.SetUnlock(ctx, req.SystemId)
		ctx.PublishEventLocal(&event.SystemUnlock{SystemId: temp})
		dao.SaveSysUnlock(ctx, unlock)
	}
	return &servicepb.SystemUnlock_CheatUnlockSystemResponse{}, nil
}

func (s *Service) RegisterCondition() {
	r := rule.MustGetReader(&ctx.Context{})
	list := r.System.List()
	for idx := range list {
		if len(list[idx].UnlockCondition) < 3 {
			continue
		}
		typ, id, cnt := models.TaskType(list[idx].UnlockCondition[0]), list[idx].UnlockCondition[1], list[idx].UnlockCondition[2]
		s.TaskService.RegisterCondHandler(typ, id, cnt, s.HandleUnlock, list[idx].Id)
	}
}

func getDefaultUnlock(ctx *ctx.Context) map[values.Integer]bool {
	r := rule.MustGetReader(ctx)
	unlock := map[values.Integer]bool{}
	for idx := range r.System.List() {
		if len(r.System.List()[idx].UnlockCondition) == 0 {
			unlock[r.System.List()[idx].Id] = true
		} else {
			unlock[r.System.List()[idx].Id] = false
		}
	}
	return unlock
}

func (s *Service) GetMultiSysUnlock(ctx *ctx.Context, roleIds []values.RoleId) (map[values.RoleId]*pbdao.SystemUnlock, *errmsg.ErrMsg) {
	if len(roleIds) <= 0 {
		return nil, nil
	}
	return dao.GetMultiSysUnlock(ctx, roleIds)
}

func (s *Service) GetMultiSysUnlockBySys(ctx *ctx.Context, roleIds []values.RoleId, id values.SystemId) (map[values.RoleId]bool, *errmsg.ErrMsg) {
	if len(roleIds) <= 0 {
		return nil, nil
	}
	data, err := dao.GetMultiSysUnlock(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	ret := make(map[values.RoleId]bool)
	for roleId, unlock := range data {
		// 从db返回的 unlock.Unlock不会为nil，所以这里不用判断
		ret[roleId] = unlock.Unlock[values.Integer(id)]
	}
	return ret, nil
}

func SystemChecker(next handler.HandleFunc) handler.HandleFunc {
	return func(ctx *ctx.Context) *errmsg.ErrMsg {
		var reqMsgName string
		if ctx.ServerType == models.ServerType_GatewayStdTcp && ctx.Req != nil {
			reqMsgName = proto.MessageName(ctx.Req)
			str := strings.Split(reqMsgName, ".")
			if len(str) >= 3 {
				sys := str[1]
				r := rule.MustGetReader(ctx)
				typ, ok := r.System.GetSysTypeByName(sys)
				unlock, err := isUnlock(ctx, typ)
				if err != nil {
					return err
				}
				if ok && !unlock {
					return errmsg.NewErrSystemLock()
				}
			}
		}
		return next(ctx)
	}
}

func isUnlock(ctx *ctx.Context, typ models.SystemType) (bool, *errmsg.ErrMsg) {
	su, err := dao.GetSysUnlock(ctx, ctx.RoleId)
	if err != nil {
		return false, err
	}
	if su.Unlock == nil {
		return false, nil
	}
	return su.Unlock[values.Integer(typ)], nil
}
