package npc_task

import (
	"fmt"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/task/npc-task/dao"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
}

func NewNpcTaskService(
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
	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("获取所有npc任务", s.GetTaskRequest)
	s.svc.RegisterFunc("提交npc任务", s.SubmitTaskRequest)
	s.svc.RegisterFunc("快速完成任务", s.QuickFinishRequest)
	s.svc.RegisterFunc("作弊器接任务", s.CheatAddTaskRequest)

	eventlocal.SubscribeEventLocal(s.HandleLoginEvent)
	eventlocal.SubscribeEventLocal(s.HandleLevelUpEvent)
	eventlocal.SubscribeEventLocal(s.HandleTaskUpdateEvent)
	eventlocal.SubscribeEventLocal(s.HandleTalkEvent)
	eventlocal.SubscribeEventLocal(s.HandleKillMonsterEvent)
	eventlocal.SubscribeEventLocal(s.HandleSubmitEvent)
	eventlocal.SubscribeEventLocal(s.HandleMainTaskUpdateEvent)
	eventlocal.SubscribeEventLocal(s.HandleTargetUpdate)
}

func (s *Service) GetTaskRequest(ctx *ctx.Context, _ *servicepb.NpcTask_GetNpcTaskRequest) (*servicepb.NpcTask_GetNpcTaskResponse, *errmsg.ErrMsg) {
	task, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	// 检查刷新任务
	effect := make([]*models.NpcTask, 0)
	isRefresh := false
	for _, t := range task.Tasks {
		if s.checkTaskRefresh(ctx, t) {
			isRefresh = true
		}
		// 刷新任务未触发不发下去
		if t.Status != models.TaskStatus_Finished {
			effect = append(effect, t)
		}
	}
	if isRefresh {
		dao.SaveTask(ctx, task)
	}
	return &servicepb.NpcTask_GetNpcTaskResponse{Task: effect}, nil
}

func (s *Service) SubmitTaskRequest(ctx *ctx.Context, req *servicepb.NpcTask_SubmitNpcTaskRequest) (*servicepb.NpcTask_SubmitNpcTaskResponse, *errmsg.ErrMsg) {
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	task, ok := tasks.Tasks[req.TaskId]
	if !ok {
		return nil, errmsg.NewErrNpcTaskNotExist()
	}
	if task.Status != models.TaskStatus_Completed {
		return nil, errmsg.NewErrNpcTaskUnfinished()
	}

	r := rule.MustGetReader(ctx)
	cfg, ok := r.NpcTask.GetNpcTaskById(req.TaskId)
	if !ok {
		return nil, errmsg.NewErrNpcTaskNotExist()
	}
	err = s.submitTask(ctx, tasks.Tasks, req.TaskId, cfg)
	if err != nil {
		return nil, err
	}
	dao.SaveTask(ctx, tasks)

	// 任务完成打点
	s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskFinishNpc, req.TaskId, 1)
	return &servicepb.NpcTask_SubmitNpcTaskResponse{
		Task:    task,
		Rewards: cfg.Reward,
	}, nil
}

func (s *Service) QuickFinishRequest(ctx *ctx.Context, req *servicepb.NpcTask_QuickFinishRequest) (*servicepb.NpcTask_QuickFinishResponse, *errmsg.ErrMsg) {
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	task, ok := tasks.Tasks[req.TaskId]
	if !ok {
		return nil, errmsg.NewErrNpcTaskNotExist()
	}
	if task.Status != models.TaskStatus_Processing {
		return nil, errmsg.NewErrNpcTaskAlreadyFinish()
	}

	r := rule.MustGetReader(ctx)
	cfg, ok := r.NpcTask.GetNpcTaskById(req.TaskId)
	if !ok {
		return nil, nil
	}
	level, err := s.UserService.GetLevel(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	// todo: 暂时写死可快速完成的等级、消耗
	if level < cfg.MaxLevel {
		return nil, errmsg.NewErrNpcTaskLevelLow()
	}
	err = s.submitTask(ctx, tasks.Tasks, req.TaskId, cfg)
	if err != nil {
		return nil, err
	}
	err = s.BagService.SubItem(ctx, ctx.RoleId, enum.BoundDiamond, 20)
	if err != nil {
		return nil, err
	}
	dao.SaveTask(ctx, tasks)

	// 任务完成打点
	s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskFinishNpc, req.TaskId, 1)
	return &servicepb.NpcTask_QuickFinishResponse{
		Task:    task,
		Rewards: cfg.Reward,
	}, nil
}

func (s *Service) CheatAddTaskRequest(ctx *ctx.Context, req *servicepb.NpcTask_CheatAddNpcTaskRequest) (*servicepb.NpcTask_CheatAddNpcTaskResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	_, ok := r.NpcTask.GetNpcTaskById(req.TaskId)
	if !ok {
		return nil, nil
	}
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	t := s.newTask(ctx, req.TaskId)
	tasks.Tasks[req.TaskId] = t
	dao.SaveTask(ctx, tasks)
	ctx.PublishEventLocal(&event.NpcTaskUpdate{Task: t})
	return &servicepb.NpcTask_CheatAddNpcTaskResponse{}, nil
}

func (s *Service) updateProgress(ctx *ctx.Context, task *models.NpcTask, taskType models.TaskType, idx, target, add values.Integer, replace bool) {
	if task.Finish == nil {
		task.Finish = map[values.Integer]bool{}
	}
	if task.Progress == nil {
		task.Progress = map[values.Integer]values.Integer{}
	}
	if replace {
		task.Progress[idx] = add
	} else {
		task.Progress[idx] += add
	}

	cfg, ok := rule.MustGetReader(nil).TaskType.GetTaskTypeById(values.Integer(taskType))
	if !ok {
		panic(fmt.Sprintf("TaskType config not found. id: %d", taskType))
	}
	task.Finish[idx] = false
	if cfg.IsReversed { // 是否反向判定 比如排名
		if task.Progress[idx] <= target {
			task.Finish[idx] = true
		}
	} else {
		if task.Progress[idx] >= target {
			task.Finish[idx] = true
		}
	}

	allFinish := true
	for _, v := range task.Finish {
		if !v {
			allFinish = false
		}
	}
	if allFinish {
		task.Status = models.TaskStatus_Completed
	}
	ctx.PublishEventLocal(&event.NpcTaskUpdate{Task: task})
}

func (s *Service) unlockNpcTask(ctx *ctx.Context, tasks map[values.TaskId]*models.NpcTask) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	today := timer.BeginOfDay(timer.StartTime(ctx.StartTime)).Unix()
	now := timer.StartTime(ctx.StartTime).Unix()
	unlock, err := s.preCheckUnlock(ctx, tasks)
	if err != nil {
		return err
	}
	if !unlock {
		return nil
	}
	// 可重复刷新的任务
	for _, task := range tasks {
		cfg, _ := r.NpcTask.GetNpcTaskById(task.TaskId)
		if cfg.CanReset == 1 {
			for i := len(cfg.ResetTime) - 1; i >= 0; i-- {
				if now >= cfg.ResetTime[i]+today && task.LastRefresh != cfg.ResetTime[i]+today {
					task.LastRefresh = cfg.ResetTime[i] + today
					break
				}
			}
			// 没到第一次刷新时间
			if task.LastRefresh == 0 {
				continue
			}
		}
		tasks[task.TaskId] = task
	}
	return nil
}

func (s *Service) preCheckUnlock(ctx *ctx.Context, tasks map[values.TaskId]*models.NpcTask) (bool, *errmsg.ErrMsg) {
	// 检查等级条件
	level, err := s.UserService.GetLevel(ctx, ctx.RoleId)
	if err != nil {
		return false, nil
	}
	unlock, err := s.checkUnlock(ctx, 0, level, models.TaskType_TaskLevel, tasks)
	if err != nil {
		return false, err
	}
	return unlock, nil
}

func (s *Service) newTask(ctx *ctx.Context, taskId values.TaskId) *models.NpcTask {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.NpcTask.GetNpcTaskById(taskId)
	if !ok {
		return nil
	}
	progress := map[values.Integer]values.Integer{}
	Finish := map[values.Integer]bool{}
	for i := range cfg.SubTask {
		progress[values.Integer(i)] = 0
		Finish[values.Integer(i)] = false
	}
	t := &models.NpcTask{
		TaskId:      taskId,
		NpcId:       cfg.TaskCarrier[0],
		Progress:    progress,
		Finish:      Finish,
		Status:      models.TaskStatus_Processing,
		LastRefresh: 0,
		AcceptTime:  timer.UnixMilli(),
	}
	return t
}

func (s *Service) checkTaskRefresh(ctx *ctx.Context, task *models.NpcTask) bool {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.NpcTask.GetNpcTaskById(task.TaskId)
	if !ok {
		return false
	}
	// 不是每日刷新任务
	if cfg.CanReset != 1 {
		return false
	}
	// 重置次数小于等于1次，不用刷新
	if len(cfg.ResetTime) < 1 {
		return false
	}
	today := timer.BeginOfDay(timer.StartTime(ctx.StartTime).UTC()).Unix()
	// 当天最后一次刷新完成
	if task.LastRefresh == cfg.ResetTime[len(cfg.ResetTime)-1]+today {
		return false
	}
	for i := len(cfg.ResetTime) - 1; i >= 0; i-- {
		if timer.StartTime(ctx.StartTime).UTC().Unix() >= cfg.ResetTime[i]+today && task.LastRefresh < cfg.ResetTime[i]+today {
			progress := map[values.Integer]values.Integer{}
			Finish := map[values.Integer]bool{}
			for idx := range cfg.SubTask {
				progress[values.Integer(idx)] = 0
				Finish[values.Integer(idx)] = false
			}
			task.LastRefresh = cfg.ResetTime[i] + today
			task.Status = models.TaskStatus_Processing
			task.Progress = progress
			task.Finish = Finish
			return true
		}
	}
	return false
}

func (s *Service) refreshAll(ctx *ctx.Context, tasks map[values.TaskId]*models.NpcTask) {
	r := rule.MustGetReader(ctx)
	for _, task := range tasks {
		cfg, ok := r.NpcTask.GetNpcTaskById(task.TaskId)
		if !ok {
			continue
		}
		// 不是每日刷新任务
		if cfg.CanReset != 1 {
			continue
		}
		// 重置次数小于1次，不用刷新
		if len(cfg.ResetTime) < 1 {
			continue
		}
		today := timer.BeginOfDay(timer.StartTime(ctx.StartTime)).Unix()
		// 当天最后一次刷新完成
		if task.LastRefresh == cfg.ResetTime[len(cfg.ResetTime)-1]+today {
			continue
		}
		for i := len(cfg.ResetTime) - 1; i >= 0; i-- {
			if timer.StartTime(ctx.StartTime).Unix() >= cfg.ResetTime[i]+today && task.LastRefresh < cfg.ResetTime[i]+today {
				task.LastRefresh = cfg.ResetTime[i] + today
				task.Status = models.TaskStatus_Processing
				task.Progress = map[int64]int64{}
				task.Finish = map[int64]bool{}
				continue
			}
		}
	}
	return
}

func (s *Service) submitTask(ctx *ctx.Context, tasks map[values.TaskId]*models.NpcTask, taskId values.TaskId, cfg *rule_model.NpcTask) *errmsg.ErrMsg {
	task := tasks[taskId]
	// 是可重复出现的任务
	if cfg.CanReset == 1 {
		if !s.checkTaskRefresh(ctx, task) {
			// 未刷新, 置为完成
			task.Status = models.TaskStatus_Finished
		} else {
			// 有刷新，刷新所有任务
			s.refreshAll(ctx, tasks)
			return nil
		}
	} else {
		// 有后续
		if cfg.Next != 0 {
			nextTask := s.newTask(ctx, cfg.Next)
			tasks[nextTask.TaskId] = nextTask
			ctx.PublishEventLocal(&event.NpcTaskUpdate{Task: nextTask})
		}
		delete(tasks, taskId)
	}

	reward := cfg.Reward
	_, err := s.BagService.AddManyItem(ctx, ctx.RoleId, reward)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) checkUnlock(ctx *ctx.Context, id, count values.Integer, typ models.TaskType, tasks map[values.TaskId]*models.NpcTask) (bool, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cond, ok := r.NpcTask.NpcTaskUnlockCond(typ)
	if !ok {
		return false, nil
	}
	targetCond, ok := cond[id]
	if !ok {
		return false, nil
	}
	targetTasks, ok := targetCond[count]
	if !ok {
		return false, nil
	}

	update := false
	for _, taskId := range targetTasks {
		cfg, has := r.NpcTask.GetNpcTaskById(taskId)
		if !has {
			continue
		}
		needSave := false
		taskUnlock, err := dao.GetNpcTaskUnlock(ctx, ctx.RoleId, taskId)
		if err != nil {
			return false, err
		}
		if taskUnlock == nil {
			taskUnlock = &pbdao.NpcTaskUnlock{
				TaskId: taskId,
				Unlock: &models.NpcTaskUnlock{
					Achieve:  map[int64]bool{},
					IsUnlock: false,
				},
			}
			for idx := range cfg.AcceptTaskParam {
				taskUnlock.Unlock.Achieve[int64(idx)] = false
			}
		}
		if taskUnlock.Unlock.IsUnlock == true {
			continue
		}
		for idx, targetParam := range cfg.AcceptTaskParam {
			if typ == models.TaskType(targetParam[0]) && id == targetParam[1] && count >= targetParam[2] {
				taskUnlock.Unlock.Achieve[int64(idx)] = true
				needSave = true
				break
			}
		}
		if cfg.Front != 0 {
			continue
		}
		allFinish := true
		for _, finish := range taskUnlock.Unlock.Achieve {
			if !finish {
				allFinish = false
				break
			}
		}
		if allFinish {
			update = true
			taskUnlock.Unlock.IsUnlock = true
			newTask := s.newTask(ctx, taskId)
			tasks[taskId] = newTask
			ctx.PublishEventLocal(&event.NpcTaskUpdate{Task: newTask})
			needSave = true
		}
		if needSave {
			dao.SaveNpcTaskUnlock(ctx, ctx.RoleId, taskUnlock)
		}
	}
	return update, nil
}
