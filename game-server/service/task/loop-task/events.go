package loop_task

import (
	"errors"
	"fmt"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/task/loop-task/dao"
	"coin-server/game-server/util"
	"coin-server/rule"
	tasktarget "coin-server/rule/factory/task-target"
)

// 理论上一种任务类型最多在日常和周常各存在一个
// TODO: 完善任务类型

func (s *Service) HandleTargetUpdateEvent(ctx *ctx.Context, e *event.TargetUpdate) *errmsg.ErrMsg {
	if _, ok := rule.MustGetReader(ctx).Looptask.GetAllLoopTaskTypes()[e.Typ]; !ok {
		return nil
	}
	switch e.Typ {
	// 特殊处理的类型
	case models.TaskType_TaskLogin, models.TaskType_TaskCitySubmit, models.TaskType_TaskMapSubmit:
		return nil
	default:
		return s.checkAndUpdate(ctx, e.Id, e.Count, e.Typ, e.IsReplace)
	}
}

func (s *Service) HandleRoleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	if d.IsRegister {
		tasks, err := s.createTask(ctx)
		if err != nil {
			return err
		}
		dao.SaveLastLogin(ctx, &daopb.TaskLogin{
			RoleId:          ctx.RoleId,
			DailyNextReset:  util.DefaultNextRefreshTime().UnixMilli(),
			DayPass:         1,
			WeeklyNextReset: util.DefaultWeeklyRefreshTime().UnixMilli(),
		})

		r := rule.MustGetReader(ctx)
		for _, task := range tasks.Tasks {
			cfg, ok := r.Looptask.GetLooptaskById(task.TaskId)
			if !ok {
				continue
			}
			params := tasktarget.ParseParam(cfg.TaskTargetParam)
			if params == nil {
				return errmsg.NewInternalErr("loop_task_config_target_param_failed: " + strconv.Itoa(int(cfg.Id)))
			}
			if params.TaskType == models.TaskType_TaskLogin {
				s.updateProgress(task, params.TaskType, params.Count, 1, false)
			}
		}
		dao.SaveTask(ctx, tasks)
	} else {
		loginDao, err := dao.GetLastLogin(ctx, ctx.RoleId)
		if loginDao.DailyNextReset >= util.DefaultNextRefreshTime().UnixMilli() {
			return nil
		}
		err = s.checkAndRefresh(ctx, loginDao)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) HandleSubmitEvent(ctx *ctx.Context, d *event.Submit) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_Weekly && d.Kind != models.TaskKind_Daily {
		return nil
	}
	r := rule.MustGetReader(ctx)
	cfg, ok := r.Looptask.GetLooptaskById(d.TaskId)
	if !ok {
		return nil
	}
	params := tasktarget.ParseParam(cfg.TaskTargetParam)
	if params == nil {
		panic(errors.New("loop_task_config_target_param_failed: " + strconv.Itoa(int(cfg.Id))))
	}
	if params.TaskType != models.TaskType_TaskCitySubmit && params.TaskType != models.TaskType_TaskMapSubmit {
		return nil
	}
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	for _, task := range tasks.Tasks {
		if task.TaskId != d.TaskId {
			return nil
		}
		s.updateProgress(task, params.TaskType, params.Count, d.Count, false)

		err = s.SubItem(ctx, ctx.RoleId, params.Target, params.Count)
		if err != nil {
			return err
		}
	}
	s.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskCommitItem, d.ItemId, d.Count)
	dao.SaveTask(ctx, tasks)
	return nil
}

// HandleKillMonsterEvent 杀怪
func (s *Service) HandleKillMonsterEvent(ctx *ctx.Context, d *event.KillMonster) *errmsg.ErrMsg {
	if d.Kind != models.TaskKind_Weekly && d.Kind != models.TaskKind_Daily {
		return nil
	}
	return s.checkAndUpdate(ctx, d.MonsterId, d.Count, d.Typ, false)
}

func (s *Service) updateProgress(task *models.LoopTask, taskType models.TaskType, targetCount, add values.Integer, isReplace bool) {
	if task.Status != models.TaskStatus_Processing {
		return
	}
	if isReplace {
		task.Process = add
	} else {
		task.Process += add
	}
	cfg, ok := rule.MustGetReader(nil).TaskType.GetTaskTypeById(values.Integer(taskType))
	if !ok {
		panic(fmt.Sprintf("TaskType config not found. id: %d", taskType))
	}
	if cfg.IsReversed { // 是否反向判定 比如排名
		if task.Process <= targetCount {
			task.Status = models.TaskStatus_Completed
		}
	} else {
		if task.Process >= targetCount {
			task.Process = targetCount
			task.Status = models.TaskStatus_Completed
		}
	}
}

func (s *Service) checkAndUpdate(ctx *ctx.Context, target values.Integer, count values.Integer, typ models.TaskType, isReplace bool) *errmsg.ErrMsg {
	if count <= 0 {
		return nil
	}
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	r := rule.MustGetReader(ctx)

	isChange := false
	for _, task := range tasks.Tasks {
		if task.Status == models.TaskStatus_Processing {
			cfg, ok := r.Looptask.GetLooptaskById(task.TaskId)
			if !ok {
				continue
			}
			params := tasktarget.ParseParam(cfg.TaskTargetParam)
			if params == nil {
				panic(errors.New("loop_task_config_target_param_failed: " + strconv.Itoa(int(cfg.Id))))
			}
			if params.TaskType != typ || params.Target != target {
				continue
			}
			s.updateProgress(task, typ, params.Count, count, isReplace)
			isChange = true
		}
	}
	// 检查解锁任务
	for taskId, process := range tasks.Lock {
		cfg, ok := r.Looptask.GetLooptaskById(taskId)
		if !ok {
			continue
		}
		// 针对 以前有解锁条件，现在删掉了的情况 做容错
		if len(cfg.UnlockCondition) == 0 {
			tasks.Tasks = append(tasks.Tasks, &models.LoopTask{
				TaskId:  cfg.Id,
				Process: 0,
				Status:  models.TaskStatus_Processing,
				Kind:    cfg.Kind,
			})
			delete(tasks.Lock, taskId)
		}

		params := tasktarget.ParseParam(cfg.UnlockCondition)
		if params == nil {
			panic(errors.New("parse_loop_task_config_unlock_condition_failed: " + strconv.Itoa(int(cfg.Id))))
		}
		if params.TaskType != typ || params.Target != target {
			continue
		}
		process += count
		if process >= params.Count {
			tasks.Tasks = append(tasks.Tasks, &models.LoopTask{
				TaskId:  cfg.Id,
				Process: 0,
				Status:  models.TaskStatus_Processing,
				Kind:    cfg.Kind,
			})
			delete(tasks.Lock, taskId)
		} else {
			tasks.Lock[taskId] = process
		}
		isChange = true
	}
	if isChange {
		dao.SaveTask(ctx, tasks)
	}
	return nil
}
