package npc_task

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/utils/generic/slices"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/task/npc-task/dao"
	"coin-server/rule"
)

// HandleTaskUpdateEvent 任务更新
func (s *Service) HandleTaskUpdateEvent(ctx *ctx.Context, d *event.NpcTaskUpdate) *errmsg.ErrMsg {
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.NpcTask_NpcTaskUpdatePush{Task: d.Task})
	return nil
}

// HandleLoginEvent 登录
func (s *Service) HandleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if d.IsRegister {
		err = s.unlockNpcTask(ctx, tasks.Tasks)
		if err != nil {
			return err
		}
		dao.SaveTask(ctx, tasks)
		return nil
	}

	// 登录时检测有没有新的任务配置
	r := rule.MustGetReader(ctx)
	for idx := range r.NpcTask.List() {
		if r.NpcTask.List()[idx].CanReset == 1 && len(r.NpcTask.List()[idx].AcceptTaskParam) == 0 {
			if _, ok := tasks.Tasks[r.NpcTask.List()[idx].Id]; !ok {
				t := s.newTask(ctx, r.NpcTask.List()[idx].Id)
				tasks.Tasks[r.NpcTask.List()[idx].Id] = t
			}
		}
	}

	// 检查刷新任务
	update := false
	for _, t := range tasks.Tasks {
		if s.checkTaskRefresh(ctx, t) {
			update = true
		}
	}
	if update {
		dao.SaveTask(ctx, tasks)
	}
	return nil

}

// HandleLevelUpEvent 升级事件
func (s *Service) HandleLevelUpEvent(ctx *ctx.Context, d *event.UserLevelChange) *errmsg.ErrMsg {
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	unlock, err := s.checkUnlock(ctx, 0, d.Level, models.TaskType_TaskLevel, tasks.Tasks)
	if err != nil {
		return err
	}
	update := s.checkAndUpdate(ctx, 0, d.Incr, models.TaskType_TaskLevel, tasks.Tasks, false)
	if unlock || update {
		dao.SaveTask(ctx, tasks)
	}
	return nil
}

// HandleTalkEvent 对话任务
func (s *Service) HandleTalkEvent(ctx *ctx.Context, d *event.Talk) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_NPC {
		return nil
	}
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	unlock, err := s.checkUnlock(ctx, d.DialogId, 1, d.Typ, tasks.Tasks)
	if err != nil {
		return err
	}
	var update bool
	task, ok := tasks.Tasks[d.TaskId]
	if ok {
		update = s.checkAndUpdate(ctx, d.DialogId, 1, d.Typ, map[values.TaskId]*models.NpcTask{d.TaskId: task}, false)
	}
	if unlock || update {
		dao.SaveTask(ctx, tasks)
	}
	return nil
}

// HandleKillMonsterEvent 杀怪
func (s *Service) HandleKillMonsterEvent(ctx *ctx.Context, d *event.KillMonster) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_NPC {
		return nil
	}
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	unlock, err := s.checkUnlock(ctx, d.MonsterId, d.Count, d.Typ, tasks.Tasks)
	if err != nil {
		return err
	}
	var update bool
	task, ok := tasks.Tasks[d.TaskId]
	if ok {
		update = s.checkAndUpdate(ctx, d.MonsterId, d.Count, d.Typ, map[values.TaskId]*models.NpcTask{d.TaskId: task}, false)
	}
	if unlock || update {
		dao.SaveTask(ctx, tasks)
	}
	return nil
}

// HandleSubmitEvent 提交任务
func (s *Service) HandleSubmitEvent(ctx *ctx.Context, d *event.Submit) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_NPC {
		return nil
	}
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	unlock, err := s.checkUnlock(ctx, d.ItemId, d.Count, d.Typ, tasks.Tasks)
	if err != nil {
		return err
	}
	var update bool
	task, ok := tasks.Tasks[d.TaskId]
	if ok {
		update = s.checkAndUpdate(ctx, d.ItemId, d.Count, d.Typ, map[values.TaskId]*models.NpcTask{d.TaskId: task}, false)
		if update {
			s.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskCommitItem, d.ItemId, d.Count)
			err = s.BagService.SubItem(ctx, ctx.RoleId, d.ItemId, d.Count)
			if err != nil {
				return err
			}
		}
	}

	if unlock || update {
		dao.SaveTask(ctx, tasks)
	}
	return nil
}

// HandleMainTaskUpdateEvent 任务更新
func (s *Service) HandleMainTaskUpdateEvent(ctx *ctx.Context, d *event.MainTaskFinished) *errmsg.ErrMsg {
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	unlock, err := s.checkUnlock(ctx, d.TaskNo, 1, models.TaskType_TaskMainTask, tasks.Tasks)
	if err != nil {
		return err
	}
	update := s.checkAndUpdate(ctx, d.TaskNo, 1, models.TaskType_TaskMainTask, tasks.Tasks, false)
	if unlock || update {
		dao.SaveTask(ctx, tasks)
	}
	return nil
}

func (s *Service) HandleTargetUpdate(ctx *ctx.Context, e *event.TargetUpdate) *errmsg.ErrMsg {
	if slices.In([]models.TaskType{models.TaskType_TaskLevel,
		models.TaskType_TaskMainTask, models.TaskType_TaskCollect},
		e.Typ,
	) {
		return nil
	}
	if _, ok := rule.MustGetReader(ctx).NpcTask.GetAllNpcTaskTypes()[e.Typ]; !ok {
		return nil
	}

	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	unlock, err := s.checkUnlock(ctx, e.Id, e.Incr, e.Typ, tasks.Tasks)
	if err != nil {
		return err
	}
	update := s.checkAndUpdate(ctx, e.Id, e.Incr, e.Typ, tasks.Tasks, e.IsReplace)
	if unlock || update {
		dao.SaveTask(ctx, tasks)
	}

	return nil
}

func (s *Service) checkAndUpdate(ctx *ctx.Context, id, count values.Integer, typ models.TaskType, tasks map[values.TaskId]*models.NpcTask, replace bool) bool {
	//检测任务完成
	update := false
	for _, task := range tasks {
		if task.Status == models.TaskStatus_Processing {
			r := rule.MustGetReader(ctx)
			taskCfg, ok := r.NpcTask.GetNpcTaskById(task.TaskId)
			if !ok {
				continue
			}
			for idx, sub := range taskCfg.SubTask {
				if !task.Finish[int64(idx)] {
					if models.TaskType(sub[0]) == typ {
						if sub[1] == id {
							s.updateProgress(ctx, task, typ, values.Integer(idx), sub[2], count, replace)
							update = true
						}
					}
				}
			}
		}
	}
	return update
}
