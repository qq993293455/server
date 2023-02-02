package task

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/utils/generic/slices"
	"coin-server/common/values"
	event2 "coin-server/common/values/event"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/task/task/dao"
	"coin-server/rule"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	register *Register
	checker  *ConditionChecker
}

func NewTaskService(
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
	s.register = NewRegister()
	s.checker = NewConditionChecker(s)
	module.TaskService = s
	return s
}

func (s *Service) Router() {
	// 任务通用协议（请求发出事件）
	s.svc.RegisterFunc("提交任务道具", s.SubmitRequest)
	s.svc.RegisterFunc("采集道具", s.GatherRequest)
	s.svc.RegisterFunc("移动到某个点", s.MoveRequest)
	s.svc.RegisterFunc("杀怪计数", s.KillMonsterRequest)

	// 任务通用事件（战斗服事件）
	//s.svc.RegisterEvent("杀怪事件", s.KillMonsterEvent)
	s.svc.RegisterEvent("传送事件", s.TeleportEvent)
	s.svc.RegisterEvent("完成肉鸽副本", s.FinishRLEvent)

	s.svc.RegisterEvent("远程统计事件", s.RemoteUpdateTarget)

	// 接收各系统发出的目标打点条件
	eventlocal.SubscribeEventLocal(s.HandleUpdateTarget)
	// 发出目标更新事件
	eventlocal.SubscribeEventLocal(s.HandleTargetUpdate)
}

func (s *Service) UpdateTarget(ctx *ctx.Context, roleId values.RoleId, typ models.TaskType, id, cnt values.Integer, replace ...bool) {
	var isReplace bool
	if len(replace) > 0 {
		isReplace = replace[0]
	}
	ctx.PublishEventLocal(&event.UpdateTarget{
		RoleId:  roleId,
		Typ:     typ,
		Id:      id,
		Count:   cnt,
		Replace: isReplace,
	})
}

func (s *Service) UpdateTargets(ctx *ctx.Context, roleId values.RoleId, tasks map[models.TaskType]*models.TaskUpdate) {
	for typ, task := range tasks {
		ctx.PublishEventLocal(&event.UpdateTarget{
			RoleId:  roleId,
			Typ:     typ,
			Id:      task.Id,
			Count:   task.Cnt,
			Replace: task.Replace,
		})
	}
}

// 所有用户公用的条件，注册回调
func (s *Service) RegisterCondHandler(taskType models.TaskType, targetId, targetCnt values.Integer, handler event2.CondHandler, args any) {
	s.register.RegisterHandler(taskType, targetId, targetCnt, handler, args)
}

func (s *Service) GetCounter(ctx *ctx.Context) (map[string]*pbdao.CondCounter, *errmsg.ErrMsg) {
	counter, err := dao.GetCond(ctx, ctx.RoleId)
	return counter, err
}

func (s *Service) GetCounterByType(ctx *ctx.Context, taskType models.TaskType) (map[values.Integer]values.Integer, *errmsg.ErrMsg) {
	counter, err := dao.GetCondByType(ctx, ctx.RoleId, taskType)
	return counter.Count, err
}

func (s *Service) GetCounterByTypeList(ctx *ctx.Context, taskTypes []models.TaskType) (map[models.TaskType]map[values.Integer]values.Integer, *errmsg.ErrMsg) {
	res := make(map[models.TaskType]map[values.Integer]values.Integer)
	for _, taskType := range taskTypes {
		counter, err := dao.GetCondByType(ctx, ctx.RoleId, taskType)
		if err != nil {
			return nil, err
		}
		res[taskType] = counter.Count
	}

	return res, nil
}

func (s *Service) CheckCondition(ctx *ctx.Context, taskType models.TaskType, targetId, targetCnt values.Integer) (values.Integer, bool, *errmsg.ErrMsg) {
	return s.checker.CheckCondition(ctx, taskType, targetId, targetCnt)
}

// -----------------------------------------------------service---------------------------------------------------------//

func (s *Service) SubmitRequest(ctx *ctx.Context, req *servicepb.Task_SubmitTaskRequest) (*servicepb.Task_SubmitTaskResponse, *errmsg.ErrMsg) {
	ctx.PublishEventLocal(&event.Submit{
		TaskId: req.TaskId,
		ItemId: req.ItemId,
		Count:  req.Count,
		Kind:   req.Kind,
		Typ:    req.Typ,
	})
	return &servicepb.Task_SubmitTaskResponse{}, nil
}

func (s *Service) GatherRequest(ctx *ctx.Context, req *servicepb.Task_GatherTaskRequest) (*servicepb.Task_GatherTaskResponse, *errmsg.ErrMsg) {
	ctx.PublishEventLocal(&event.Gather{
		TaskId: req.TaskId,
		Kind:   req.Kind,
	})
	return &servicepb.Task_GatherTaskResponse{}, nil
}

func (s *Service) MoveRequest(ctx *ctx.Context, req *servicepb.Task_MoveTaskRequest) (*servicepb.Task_MoveTaskResponse, *errmsg.ErrMsg) {
	ctx.PublishEventLocal(&event.Move{
		TaskId: req.TaskId,
		Id:     0,
		Count:  1,
		Kind:   req.Kind,
	})
	return &servicepb.Task_MoveTaskResponse{}, nil
}

func (s *Service) KillMonsterRequest(ctx *ctx.Context, req *servicepb.Task_KillMonsterRequest) (*servicepb.Task_KillMonsterResponse, *errmsg.ErrMsg) {
	list := []models.TaskType{
		models.TaskType_TaskKillMonster,
		models.TaskType_TaskKillAllocateMonsterNum,
		models.TaskType_TaskKillAnyMonsterNum,
	}
	if slices.In(list, req.Typ) {
		ctx.PublishEventLocal(&event.KillMonster{
			TaskId:    req.TaskId,
			MonsterId: req.MonsterId,
			Count:     req.Count,
			Kind:      req.Kind,
			Typ:       req.Typ,
		})
	}
	return &servicepb.Task_KillMonsterResponse{}, nil
}

// -----------------------------------------------------events---------------------------------------------------------//

//func (s *Service) KillMonsterEvent(ctx *ctx.Context, req *servicepb.BattleEvent_KillMonsterPush) {
//	s.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskKillMonster, req.ConfigId, 1)
//	ctx.PublishEventLocal(&event.KillMonster{
//		RoleId:    ctx.RoleId,
//		MonsterId: req.ConfigId,
//		Num:       1,
//	})
//}

func (s *Service) TeleportEvent(ctx *ctx.Context, req *servicepb.Task_TeleportEvent) {
	ctx.PublishEventLocal(&event.Teleport{
		RoleId: req.RoleId,
		MapId:  req.MapId,
		Pos:    event.Vector2{X: req.Pos.X, Y: req.Pos.Y},
	})
}

func (s *Service) FinishRLEvent(c *ctx.Context, req *servicepb.Task_FinishRLEvent) {
	s.UpdateTarget(c, c.RoleId, models.TaskType_TaskPassDungeonNumAcc, 0, 1)
	s.UpdateTarget(c, c.RoleId, models.TaskType_TaskPassDungeonNum, 0, 1)
	rlCfg, ok := rule.MustGetReader(c).RoguelikeDungeon.GetRoguelikeDungeonById(req.RoguelikeId)
	if ok {
		s.UpdateTarget(c, c.RoleId, models.TaskType_TaskFinishRLDungeon, rlCfg.HeroDifficulty[1], 1)
	}
}

func (s *Service) RemoteUpdateTarget(c *ctx.Context, req *less_service.User_UpdateTargetPush) {
	c.PublishEventLocal(&event.UpdateTarget{
		RoleId:  c.RoleId,
		Typ:     models.TaskType(req.Typ),
		Id:      req.Id,
		Count:   req.Count,
		Replace: req.Replace,
	})
}
