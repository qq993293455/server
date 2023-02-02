package ritualtask

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/common/utils/generic/slices"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/task/ritual-task/dao"
	rule2 "coin-server/game-server/service/task/ritual-task/rule"
	"coin-server/rule"
)

// todo: 丰富任务类型

//func (svc *Service) HandleMainTaskFinishedEvent(ctx *ctx.Context, event *event.MainTaskFinished) *errmsg.ErrMsg {
//	if event.TaskNo == rule2.GetRitualUnlockMainTaskId(ctx) {
//		svc.unlock(ctx, ctx.RoleId)
//	}
//	return nil
//}
//
//func (svc *Service) HandleSysUnlockEvent(ctx *ctx.Context, event *event.SystemUnlock) *errmsg.ErrMsg {
//	for _, id := range event.SystemId {
//		if id == models.SystemType_SystemChaosRitual {
//			svc.unlock(ctx, ctx.RoleId)
//			break
//		}
//	}
//	return nil
//}

// HandleSubmitEvent 提交道具任务
func (svc *Service) HandleSubmitEvent(ctx *ctx.Context, d *event.Submit) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_Ritual {
		return nil
	}
	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	task, ok := ritual.Ritual.Tasks[d.TaskId]
	if !ok {
		return nil
	}
	isChange := svc.baseUpdateProgress(ctx, task, d.Typ, d.ItemId, d.Count, false)
	if isChange {
		dao.Save(ctx, ritual)
	}
	return svc.BagService.SubItem(ctx, ctx.RoleId, d.ItemId, d.Count)
}

// HandleKillMonsterEvent 杀怪
func (svc *Service) HandleKillMonsterEvent(ctx *ctx.Context, d *event.KillMonster) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_Ritual {
		return nil
	}
	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	task, ok := ritual.Ritual.Tasks[d.TaskId]
	if !ok {
		return nil
	}
	isChange := svc.baseUpdateProgress(ctx, task, d.Typ, d.MonsterId, d.Count, false)
	if isChange {
		dao.Save(ctx, ritual)
	}
	return nil
}

// HandleTalkEvent 对话任务
func (svc *Service) HandleTalkEvent(ctx *ctx.Context, d *event.Talk) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_Ritual || !d.IsEnd {
		return nil
	}

	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	task, ok := ritual.Ritual.Tasks[d.TaskId]
	if !ok {
		return nil
	}

	cfg := rule2.GetRitualTaskCfg(ctx, task.TaskId, task.SubTaskId)
	if cfg == nil {
		return errmsg.NewErrTaskCfgNotExist()
	}

	if d.OptIdx < 0 {
		return errmsg.NewErrTaskChoiceNotExist()
	}

	isChange := svc.baseUpdateProgress(ctx, task, models.TaskType_TaskHalfBodyTalk, d.HeadDialogId, 1, false)
	if isChange {
		task.ChoiceIdx = d.OptIdx
		dao.Save(ctx, ritual)
	}

	return nil
}

// HandleLevelUpEvent 等级达到任务监测等级
func (svc *Service) HandleLevelUpEvent(ctx *ctx.Context, d *event.UserLevelChange) *errmsg.ErrMsg {
	svc.updateProgress(ctx, models.TaskType_TaskLevel, 0, d.Incr, false)
	return nil
}

// HandleCollectTaskEvent 收集任务道具监测道具更新
func (svc *Service) HandleCollectTaskEvent(ctx *ctx.Context, d *event.ItemUpdate) *errmsg.ErrMsg {
	targets := make([]values.Integer, 0, len(d.Items))
	for _, item := range d.Items {
		targets = append(targets, item.ItemId)
	}
	svc.multiUpdateProgress(ctx, models.TaskType_TaskCollect, targets, d.Incr)
	return nil
}

func (svc *Service) HandleTargetUpdate(ctx *ctx.Context, e *event.TargetUpdate) *errmsg.ErrMsg {
	if slices.In([]models.TaskType{models.TaskType_TaskHalfBodyTalk,
		models.TaskType_TaskLevel, models.TaskType_TaskCollect},
		e.Typ,
	) {
		return nil
	}
	if _, ok := rule.MustGetReader(ctx).TargetTask.GetAllTargetTaskTypes()[e.Typ]; !ok {
		return nil
	}
	svc.updateProgress(ctx, e.Typ, e.Id, e.Incr, e.IsReplace)
	return nil
}
