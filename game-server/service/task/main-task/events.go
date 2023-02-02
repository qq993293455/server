package maintask

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/utils/generic/slices"
	"coin-server/game-server/event"
	"coin-server/game-server/service/task/main-task/dao"
	"coin-server/rule"

	"go.uber.org/zap"
)

// todo: 丰富任务类型

// HandleTaskUpdateEvent 任务更新
func (svc *Service) HandleTaskUpdateEvent(ctx *ctx.Context, d *event.MainTaskUpdate) *errmsg.ErrMsg {
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.MainTask_MainTaskUpdatePush{Task: d.Task})
	return nil
}

// HandleGatherEvent 完成采集事件任务
func (svc *Service) HandleGatherEvent(ctx *ctx.Context, d *event.Gather) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_Main {
		return nil
	}
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}
	if task.Task.TaskId != d.TaskId {
		return nil
	}
	ok, err := svc.checkAndUpdate(ctx, task, 0, 1, models.TaskType_TaskGather)
	if err != nil {
		return err
	}
	if ok {
		dao.SaveMainTask(ctx, task)
	}
	return nil
}

// HandleMoveEvent 完成移动任务
func (svc *Service) HandleMoveEvent(ctx *ctx.Context, d *event.Move) *errmsg.ErrMsg {
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}
	ok, err := svc.checkAndUpdate(ctx, task, d.Id, d.Count, models.TaskType_TaskMove)
	if err != nil {
		return err
	}
	if ok {
		dao.SaveMainTask(ctx, task)
	}
	return nil
}

// HandleMapEventFinishedEvent 完成地图事件任务
func (svc *Service) HandleMapEventFinishedEvent(ctx *ctx.Context, d *event.MapEventFinished) *errmsg.ErrMsg {
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}
	ok, err := svc.checkAndUpdate(ctx, task, d.EventId, 1, models.TaskType_TaskMapEvent)
	if err != nil {
		return err
	}
	if ok {
		dao.SaveMainTask(ctx, task)
	}
	return nil
}

// HandleSubmitEvent 提交道具任务
func (svc *Service) HandleSubmitEvent(ctx *ctx.Context, d *event.Submit) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_Main {
		return nil
	}
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}
	if task.Task.TaskId != d.TaskId {
		return nil
	}
	ok, err := svc.checkAndUpdate(ctx, task, d.ItemId, d.Count, d.Typ)
	if err != nil {
		return err
	}
	if ok {
		dao.SaveMainTask(ctx, task)
	}
	//svc.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskCommitItem, d.ItemId, d.Count)
	return svc.BagService.SubItem(ctx, ctx.RoleId, d.ItemId, d.Count)
}

// HandleKillMonsterEvent 杀怪
func (svc *Service) HandleKillMonsterEvent(ctx *ctx.Context, d *event.KillMonster) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_Main {
		return nil
	}
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}
	if task.Task.TaskId != d.TaskId {
		return nil
	}
	ok, err := svc.checkAndUpdate(ctx, task, d.MonsterId, d.Count, d.Typ)
	if err != nil {
		return err
	}
	if ok {
		dao.SaveMainTask(ctx, task)
	}
	return nil
}

// HandleTalkEvent 对话任务
func (svc *Service) HandleTalkEvent(ctx *ctx.Context, d *event.Talk) *errmsg.ErrMsg {
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}
	if task.Task.TaskId != d.TaskId {
		return nil
	}
	ok, err := svc.checkAndUpdate(ctx, task, d.DialogId, 1, d.Typ)
	if err != nil {
		return err
	}
	if ok {
		dao.SaveMainTask(ctx, task)
	}
	return nil
}

// HandleLevelUpEvent 等级达到任务监测等级
func (svc *Service) HandleLevelUpEvent(ctx *ctx.Context, d *event.UserLevelChange) *errmsg.ErrMsg {
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}
	r := rule.MustGetReader(ctx)
	if task.Task.Status == models.TaskStatus_CantAccept {
		cfg, ok := r.MainTask.GetMainTaskById(task.Task.TaskId)
		if !ok {
			return nil
		}
		if d.Level >= cfg.MinLevel {
			task.Task.Status = models.TaskStatus_NotStarted
			dao.SaveMainTask(ctx, task)
			ctx.PublishEventLocal(&event.MainTaskUpdate{Task: task.Task})
			return nil
		}
	}

	ok, err := svc.checkAndUpdate(ctx, task, 0, d.Level, models.TaskType_TaskLevel)
	if err != nil {
		return err
	}
	if ok {
		dao.SaveMainTask(ctx, task)
	}
	return nil
}

// HandleCollectTaskEvent 收集任务道具监测道具更新
func (svc *Service) HandleCollectTaskEvent(ctx *ctx.Context, d *event.ItemUpdate) *errmsg.ErrMsg {
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}
	ok := false
	for i := range d.Items {
		ok1, err1 := svc.checkAndUpdate(ctx, task, d.Items[i].ItemId, d.Incr[i], models.TaskType_TaskCollect)
		if err1 != nil {
			return err1
		}
		if ok1 {
			ok = true
		}
	}
	if ok {
		dao.SaveMainTask(ctx, task)
	}
	return nil
}

//// HandlePlaneDungeonFinish 通关位面副本
//func (svc *Service) HandlePlaneDungeonFinish(ctx *ctx.Context, d *event.PlaneDungeonFinish) *errmsg.ErrMsg {
//	if !d.IsSuccess {
//		return nil
//	}
//	task, err := dao.GetMainTask(ctx, ctx.RoleId)
//	if err != nil {
//		return err
//	}
//	if task == nil {
//		return nil
//	}
//	ok, err := svc.checkAndUpdate(ctx, task, d.PlaneId, 1, d.TaskType)
//	if err != nil {
//		return err
//	}
//	if ok {
//		dao.SaveMainTask(ctx, task)
//	}
//	return nil
//}
//
//// HandleNpcDungeonFinish 通关Npc挑战副本
//func (svc *Service) HandleNpcDungeonFinish(ctx *ctx.Context, d *event.NpcDungeonFinish) *errmsg.ErrMsg {
//	if !d.IsSuccess {
//		return nil
//	}
//	task, err := dao.GetMainTask(ctx, ctx.RoleId)
//	if err != nil {
//		return err
//	}
//	if task == nil {
//		return nil
//	}
//	ok, err := svc.checkAndUpdate(ctx, task, d.DungeonId, 1, d.TaskType)
//	if err != nil {
//		return err
//	}
//	if ok {
//		dao.SaveMainTask(ctx, task)
//	}
//	return nil
//}

func (svc *Service) HandleTargetUpdate(ctx *ctx.Context, e *event.TargetUpdate) *errmsg.ErrMsg {
	ctx.TraceLogger.Debug("HandleTargetUpdate start", zap.Any("event.TargetUpdate", e))
	if slices.In([]models.TaskType{models.TaskType_TaskGather, models.TaskType_TaskMove,
		models.TaskType_TaskMapEvent, models.TaskType_TaskLevel, models.TaskType_TaskCollect},
		e.Typ,
	) {
		return nil
	}
	ctx.TraceLogger.Debug("HandleTargetUpdate 1", zap.Any("event.TargetUpdate", e))
	if _, ok := rule.MustGetReader(ctx).MainTask.GetAllMainTaskTypes()[e.Typ]; !ok {
		return nil
	}
	ctx.TraceLogger.Debug("HandleTargetUpdate 2", zap.Any("event.TargetUpdate", e))

	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		ctx.TraceLogger.Debug("HandleTargetUpdate GetMainTask error", zap.Any("event.TargetUpdate", e), zap.Error(err))
		return err
	}
	if task == nil {
		ctx.TraceLogger.Warn("HandleTargetUpdate MainTask not found", zap.Any("event.TargetUpdate", e))
		return nil
	}
	ctx.TraceLogger.Debug("HandleTargetUpdate 3", zap.Any("event.TargetUpdate", e), zap.Any("task", task))
	ok, err := svc.checkAndUpdate(ctx, task, e.Id, e.Incr, e.Typ)
	ctx.TraceLogger.Debug("HandleTargetUpdate 4", zap.Any("event.TargetUpdate", e), zap.Bool("ok", ok), zap.Error(err))
	if err != nil {
		return err
	}
	if ok {
		dao.SaveMainTask(ctx, task)
		ctx.TraceLogger.Debug("HandleTargetUpdate 5", zap.Any("event.TargetUpdate", e))
	}
	return nil
}
