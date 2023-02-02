package maintask

import (
	"fmt"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/iggsdk"
	"coin-server/common/logger"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/statistical"
	stmodels "coin-server/common/statistical/models"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/task/main-task/dao"
	"coin-server/rule"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewMainTaskService(
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
		Module:     module,
		log:        log,
	}
	NewCantAccept(s)
	NewNotStart(s)
	NewProcessing(s)
	NewComplete(s)
	NewFinish(s)
	s.MainTaskService = s
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取主线任务", svc.GetMainTask)
	svc.svc.RegisterFunc("接主线任务", svc.AcceptMainTask)
	svc.svc.RegisterFunc("完成主线任务", svc.SubmitMainTask)
	svc.svc.RegisterFunc("领取章节奖励", svc.GetChapterReward)

	svc.svc.RegisterFunc("作弊设置当前主线任务为完成状态", svc.CheatFinishMainTask)
	svc.svc.RegisterFunc("作弊设置当前主线任务", svc.CheatSetMainTask)
	svc.svc.RegisterFunc("重置主线任务", svc.CheatResetMainTask)

	eventlocal.SubscribeEventLocal(svc.HandleTaskUpdateEvent)
	eventlocal.SubscribeEventLocal(svc.HandleCollectTaskEvent)
	eventlocal.SubscribeEventLocal(svc.HandleMapEventFinishedEvent)
	eventlocal.SubscribeEventLocal(svc.HandleSubmitEvent)
	eventlocal.SubscribeEventLocal(svc.HandleLevelUpEvent)
	eventlocal.SubscribeEventLocal(svc.HandleTalkEvent)
	eventlocal.SubscribeEventLocal(svc.HandleKillMonsterEvent)
	//eventlocal.SubscribeEventLocal(svc.HandleTeleportEvent)
	eventlocal.SubscribeEventLocal(svc.HandleGatherEvent)
	eventlocal.SubscribeEventLocal(svc.HandleMoveEvent)
	//eventlocal.SubscribeEventLocal(svc.HandlePlaneDungeonFinish)
	//eventlocal.SubscribeEventLocal(svc.HandleNpcDungeonFinish)
	eventlocal.SubscribeEventLocal(svc.HandleTargetUpdate)
}

func (svc *Service) GetMainTask(ctx *ctx.Context, _ *servicepb.MainTask_GetMainTaskRequest) (*servicepb.MainTask_GetMainTaskResponse, *errmsg.ErrMsg) {
	task, err := NewMainTask(ctx, svc)
	if err != nil {
		return nil, err
	}
	model, err := task.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &servicepb.MainTask_GetMainTaskResponse{
		Task:     model.Task,
		Finished: model.Finished,
	}, nil
}

func (svc *Service) AcceptMainTask(ctx *ctx.Context, _ *servicepb.MainTask_AcceptMainTaskRequest) (*servicepb.MainTask_AcceptMainTaskResponse, *errmsg.ErrMsg) {
	task, err := NewMainTask(ctx, svc)
	if err != nil {
		return nil, err
	}
	err = task.Accept(ctx)
	if err != nil {
		return nil, err
	}
	task.Save(ctx)
	return &servicepb.MainTask_AcceptMainTaskResponse{Task: task.Task}, nil
}

func (svc *Service) SubmitMainTask(ctx *ctx.Context, req *servicepb.MainTask_SubmitMainTaskRequest) (*servicepb.MainTask_SubmitMainTaskResponse, *errmsg.ErrMsg) {
	task, err := NewMainTask(ctx, svc)
	if err != nil {
		return nil, err
	}
	reward, err := task.Submit(ctx, req.Choose)
	if err != nil {
		return nil, err
	}
	task.Save(ctx)
	return &servicepb.MainTask_SubmitMainTaskResponse{
		Task:     task.Task,
		Finished: task.Finished,
		Rewards:  reward,
	}, nil
}

func (svc *Service) IsFinishMainTask(ctx *ctx.Context, taskId values.TaskId) (bool, *errmsg.ErrMsg) {
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return false, err
	}

	r := rule.MustGetReader(ctx)
	cfg, ok := r.MainTask.GetMainTaskById(taskId)
	if !ok {
		return false, nil
	}

	taskFinish := task.Finished
	chapter, ok := taskFinish.ChapterFinish[cfg.Chapter]
	if !ok {
		return false, nil
	}

	if chapter.Finish == nil {
		return false, nil
	}
	return chapter.Finish[cfg.Idx] > models.RewardStatus_Locked, nil
}

func (svc *Service) GetChapterReward(ctx *ctx.Context, req *servicepb.MainTask_GetChapterRewardRequest) (*servicepb.MainTask_GetChapterRewardResponse, *errmsg.ErrMsg) {
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	chapter, ok := task.Finished.ChapterFinish[req.Chapter]
	if !ok {
		return nil, errmsg.NewErrChapterRewardCantGet()
	}
	if chapter.Finish == nil {
		return nil, errmsg.NewErrChapterRewardCantGet()
	}

	status := chapter.Finish[req.Idx]
	if status != models.RewardStatus_Unlocked {
		return nil, errmsg.NewErrChapterRewardCantGet()
	}

	r := rule.MustGetReader(ctx)
	//chapterCfg, ok := r.MainTaskChapter.GetMainTaskChapterById(req.Chapter)
	//if !ok {
	//	panic(fmt.Sprintf("Chapter [%d] not exist!", req.Chapter))
	//}
	reward, ok := r.MainTaskChapterReward.GetMainTaskChapterRewardById(req.Chapter, req.Idx)
	if !ok {
		panic(fmt.Sprintf("Chapter [%d] Reward not exist!", req.Chapter))
	}

	_, err = svc.BagService.AddManyItem(ctx, ctx.RoleId, reward.Reward)
	if err != nil {
		return nil, err
	}

	SetMainTaskFinish(task.Finished, req.Chapter, req.Idx, models.RewardStatus_Received)
	dao.SaveMainTask(ctx, task)

	return &servicepb.MainTask_GetChapterRewardResponse{
		Finished: task.Finished,
		Rewards:  reward.Reward,
	}, nil
}

func (svc *Service) unlockMainTask(ctx *ctx.Context) (*daopb.MainTask, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	task := &daopb.MainTask{
		RoleId: ctx.RoleId,
		Task: &models.MainTask{
			TaskId: r.MainTask.List()[FirstTask].Id,
			Status: models.TaskStatus_Processing,
			Last:   0,
		},
	}
	sub := r.MainTask.List()[FirstTask].SubTask
	task.Task.Progress = map[values.Integer]values.Integer{}
	task.Task.Finish = map[values.Integer]bool{}
	for idx := range sub {
		task.Task.Progress[int64(idx)] = 0
		task.Task.Finish[int64(idx)] = false
	}
	task.Finished = &models.MainTaskFinish{ChapterFinish: map[int64]*models.MainTaskChapterFinish{}}

	dao.SaveMainTask(ctx, task)
	return task, nil
}

func (svc *Service) acceptTask(ctx *ctx.Context, task *MainTask) *errmsg.ErrMsg {
	task.Task.AcceptTime = timer.UnixMilli()
	err := svc.checkTaskTarget(ctx, task.Task)
	if err != nil {
		return err
	}
	complete := true
	for _, v := range task.Task.Finish {
		if v == false {
			complete = false
			break
		}
	}
	if complete {
		task.Task.Status = models.TaskStatus_Completed
		task.ITaskStatus = &CompleteStatus{svc: svc}
	} else {
		task.Task.Status = models.TaskStatus_Processing
		task.ITaskStatus = &ProcessingStatus{svc: svc}
	}
	cfg, ok := rule.MustGetReader(ctx).MainTask.GetMainTaskById(task.Task.TaskId)
	if !ok {
		return errmsg.NewInternalErr("main_task config not found: " + strconv.Itoa(int(task.Task.TaskId)))
	}
	statistical.Save(ctx.NewLogServer(), &stmodels.MainTask{
		IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
		EventTime: timer.Now(),
		GwId:      statistical.GwId(),
		RoleId:    ctx.RoleId,
		UserId:    ctx.UserId,
		TaskId:    task.Task.TaskId,
		Chapter:   cfg.Chapter,
		Index:     cfg.Idx,
		Action:    stmodels.TaskActionAccept,
	})
	return nil
}

func (svc *Service) checkAndUpdate(ctx *ctx.Context, task *daopb.MainTask, id values.Integer, count values.Integer, typ models.TaskType) (bool, *errmsg.ErrMsg) {
	if task.Task.Status != models.TaskStatus_Processing {
		return false, nil
	}

	r := rule.MustGetReader(ctx)
	cfg, ok := r.MainTask.GetMainTaskById(task.Task.TaskId)
	if !ok {
		return false, nil
	}
	update := false
	for idx, subTask := range cfg.SubTask {
		if models.TaskType(subTask[0]) == typ {
			targetId := subTask[1]
			targetCnt := subTask[2]
			if targetId == id {
				// 已完成任务，break
				if task.Task.Finish[int64(idx)] == true {
					continue
				}
				svc.updateProgress(ctx, task.Task, values.Integer(idx), targetCnt, count)
				update = true
			}
		}
	}
	return update, nil
}

func (svc *Service) updateProgress(ctx *ctx.Context, task *models.MainTask, subId, target, add values.Integer) {
	if add == 0 {
		return
	}
	progress, ok := task.Progress[subId]
	if !ok {
		return
	}

	value := progress + add
	if add < 0 {
		if value < target {
			if task.Finish[subId] == true {
				allDone := true
				for _, v := range task.Finish {
					if v == false {
						allDone = false
						break
					}
				}
				if allDone {
					task.Status = models.TaskStatus_Processing
				}
				task.Finish[subId] = false
			}
			if value < 0 {
				task.Progress[subId] = 0
			} else {
				task.Progress[subId] = value
			}
		}
	} else {
		if value >= target {
			task.Progress[subId] = target
			task.Finish[subId] = true
			allDone := true
			for _, v := range task.Finish {
				if v == false {
					allDone = false
					break
				}
			}
			if allDone {
				task.Status = models.TaskStatus_Completed
			}
		} else {
			task.Progress[subId] = value
		}
	}

	ctx.PublishEventLocal(&event.MainTaskUpdate{Task: task})
	return
}

// 新任务解锁前先检查任务目标完成情况
// todo: 丰富任务类型
func (svc *Service) checkTaskTarget(ctx *ctx.Context, task *models.MainTask) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	taskCfg, ok := r.MainTask.GetMainTaskById(task.TaskId)
	if !ok {
		return nil
	}
	for idx, subTask := range taskCfg.SubTask {
		switch taskType := models.TaskType(subTask[0]); taskType {
		case models.TaskType_TaskLevel:
			level, err := svc.UserService.GetLevel(ctx, ctx.RoleId)
			if err != nil {
				return err
			}
			if level >= subTask[2] {
				task.Finish[int64(idx)] = true
				task.Progress[int64(idx)] = subTask[2]
			} else {
				task.Progress[int64(idx)] = level
			}
		case models.TaskType_TaskCollect:
			count, err := svc.BagService.GetItem(ctx, ctx.RoleId, subTask[1])
			if err != nil {
				return err
			}
			if count >= subTask[2] {
				task.Finish[int64(idx)] = true
				task.Progress[int64(idx)] = subTask[2]
			} else {
				task.Progress[int64(idx)] = count
			}
		default:
			cfg, ok := r.TaskType.GetTaskTypeById(values.Integer(taskType))
			if !ok {
				return errmsg.NewInternalErr("task_type not found: " + strconv.Itoa(int(taskType)))
			}
			if !cfg.IsAccumulate {
				return nil
			}
			counter, err := svc.TaskService.GetCounterByType(ctx, taskType)
			if err != nil {
				return err
			}
			count := counter[subTask[1]]
			if count >= subTask[2] {
				task.Finish[int64(idx)] = true
				task.Progress[int64(idx)] = subTask[2]
			} else {
				task.Progress[int64(idx)] = count
			}
		}
	}
	return nil
}

func (svc *Service) findNextTask(ctx *ctx.Context, task *MainTask) (*errmsg.ErrMsg, bool) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.MainTask.GetMainTaskById(task.Task.TaskId)
	if !ok {
		return nil, false
	}
	// 只有单分支任务会被自动解锁
	if len(cfg.Next) != 1 {
		return nil, false
	}

	task.Task.Last = task.Task.TaskId
	task.Task.TaskId = cfg.Next[0]
	task.Task.Progress = map[values.Integer]values.Integer{}

	cfg, ok = r.MainTask.GetMainTaskById(task.Task.TaskId)
	if !ok {
		return nil, false
	}
	task.Task.Progress = map[values.Integer]values.Integer{}
	task.Task.Finish = map[values.Integer]bool{}
	for idx := range cfg.SubTask {
		task.Task.Progress[int64(idx)] = 0
		task.Task.Finish[int64(idx)] = false
	}
	level, err := svc.UserService.GetLevel(ctx, ctx.RoleId)
	if err != nil {
		return err, false
	}
	if level < cfg.MinLevel {
		task.Task.Status = models.TaskStatus_CantAccept
	} else {
		if cfg.Accept == 0 {
			err = svc.acceptTask(ctx, task)
			if err != nil {
				return err, false
			}
		} else {
			task.Task.Status = models.TaskStatus_NotStarted
		}
	}
	ctx.PublishEventLocal(&event.MainTaskUpdate{Task: task.Task})
	return nil, true
}

func (svc *Service) ChooseMainTask(ctx *ctx.Context, task *MainTask, taskId values.TaskId) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.MainTask.GetMainTaskById(task.Task.TaskId)
	if !ok {
		return nil
	}
	// 只有多分支任务可选
	if len(cfg.Next) == 1 {
		return nil
	}
	find := false
	for idx := range cfg.Next {
		if cfg.Next[idx] == taskId {
			find = true
		}
	}
	if !find {
		return errmsg.NewErrMainTaskCantChoose()
	}
	lastExp := int64(0)
	if task.Task.Last != 0 {
		last, has := r.MainTask.GetMainTaskById(task.Task.Last)
		if !has {
			return nil
		}
		lastExp = last.ExpProfit
	}
	expProfit := cfg.ExpProfit - lastExp
	illustration := cfg.StoryIllustration
	taskIdx := cfg.Idx

	task.Task.Last = task.Task.TaskId
	task.Task.TaskId = taskId
	task.Task.Progress = map[values.Integer]values.Integer{}
	task.Task.Finish = map[values.Integer]bool{}
	for idx := range cfg.SubTask {
		task.Task.Progress[int64(idx)] = 0
		task.Task.Finish[int64(idx)] = false
	}

	cfg, ok = r.MainTask.GetMainTaskById(task.Task.TaskId)
	if !ok {
		return nil
	}
	level, err := svc.UserService.GetLevel(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if level < cfg.MinLevel {
		task.Task.Status = models.TaskStatus_CantAccept
	} else {
		if cfg.Accept == 0 {
			err = svc.acceptTask(ctx, task)
			if err != nil {
				return err
			}
		} else {
			task.Task.Status = models.TaskStatus_NotStarted
		}
	}

	ctx.PublishEventLocal(&event.MainTaskUpdate{Task: task.Task})
	ctx.PublishEventLocal(&event.MainTaskFinished{
		TaskNo:       task.Task.Last,
		TaskIdx:      taskIdx,
		ExpProfit:    expProfit,
		Illustration: illustration,
	})
	// 主线任务完成打点
	svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskMainTask, task.Task.Last, 1)
	svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskMainTaskIndex, 0, taskIdx)
	return nil
}

//func (svc *Service) findNextChapterReward(ctx *ctx.Context, chapter *models.ChapterReward, taskFinish *daopb.MainTaskFinish) {
//	if chapter.Status == models.RewardStatus_Locked || chapter.Status == models.RewardStatus_Unlocked {
//		return
//	}
//	r := rule.MustGetReader(ctx)
//	chapterList := r.MainTaskChapter.List()
//
//	rewards, ok := r.MainTask.MainTaskChapter(chapterList[chapter.Chapter].Id)
//	if !ok {
//		return
//	}
//
//	chapter.Stage++
//	if int(chapter.Stage) >= len(rewards) {
//		if int(chapter.Chapter) < r.MainTaskChapter.Len()-1 {
//			chapter.Chapter++
//			chapter.Stage = 0
//			rewards, ok = r.MainTask.MainTaskChapter(chapterList[chapter.Chapter].Id)
//			if !ok {
//				return
//			}
//		} else {
//			chapter.Stage--
//			return
//		}
//	}
//
//	finish := false
//	reward := rewards[chapter.Stage]
//	if chapterFinish, ok := taskFinish.ChapterFinish[reward.MainTaskChapterId]; ok {
//		finish = chapterFinish.Finish != nil && chapterFinish.Finish[reward.Id]
//	}
//	if finish {
//		chapter.Status = models.RewardStatus_Unlocked
//	} else {
//		chapter.Status = models.RewardStatus_Locked
//	}
//
//	return
//}

// ************************************* cheat ****************************************//

// CheatFinishMainTask 作弊设置当前主线任务为完成状态
func (svc *Service) CheatFinishMainTask(ctx *ctx.Context, _ *servicepb.MainTask_CheatFinishMainTaskRequest) (*servicepb.MainTask_CheatFinishMainTaskResponse, *errmsg.ErrMsg) {
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	if task.Task.Status == models.TaskStatus_Completed || task.Task.Status == models.TaskStatus_Finished {
		return &servicepb.MainTask_CheatFinishMainTaskResponse{}, nil
	}

	task.Task.Status = models.TaskStatus_Completed
	for k := range task.Task.Finish {
		task.Task.Finish[k] = true
	}

	dao.SaveMainTask(ctx, task)
	return &servicepb.MainTask_CheatFinishMainTaskResponse{}, nil
}

// CheatSetMainTask 设置主线任务
func (svc *Service) CheatSetMainTask(ctx *ctx.Context, req *servicepb.MainTask_CheatSetMainTaskRequest) (*servicepb.MainTask_CheatSetMainTaskResponse, *errmsg.ErrMsg) {
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	r := rule.MustGetReader(ctx)
	cfg, ok := r.MainTask.GetMainTaskById(req.TaskId)
	if !ok {
		return nil, nil
	}
	task.Task.TaskId = req.TaskId
	task.Task.Status = models.TaskStatus_NotStarted
	task.Task.Finish = map[int64]bool{}
	task.Task.Progress = map[int64]int64{}
	for idx := range cfg.SubTask {
		task.Task.Finish[int64(idx)] = false
		task.Task.Progress[int64(idx)] = 0
	}

	finish := task.Finished
	if err != nil {
		return nil, err
	}
	for cfg.Front != 0 {
		cfg, ok = r.MainTask.GetMainTaskById(cfg.Front)
		if !ok {
			break
		}
		SetMainTaskFinish(finish, cfg.Chapter, cfg.Idx, models.RewardStatus_Unlocked)
	}

	dao.SaveMainTask(ctx, task)
	return &servicepb.MainTask_CheatSetMainTaskResponse{}, nil
}

// CheatResetMainTask 作弊重置主线任务
func (svc *Service) CheatResetMainTask(ctx *ctx.Context, _ *servicepb.MainTask_CheatResetMainTaskRequest) (*servicepb.MainTask_CheatResetMainTaskResponse, *errmsg.ErrMsg) {
	task, err := svc.unlockMainTask(ctx)
	if err != nil {
		return nil, err
	}
	r := rule.MustGetReader(ctx)
	first, ok := r.MainTask.GetMainTaskById(task.Task.TaskId)
	if !ok {
		panic(fmt.Sprintf("First SubTask not exist!"))
	}
	mainTask := &MainTask{
		MainTask:    task,
		ITaskStatus: nil,
	}
	if first.Accept == 0 {
		err = svc.acceptTask(ctx, mainTask)
		if err != nil {
			return nil, err
		}
	}
	dao.SaveMainTask(ctx, task)
	return &servicepb.MainTask_CheatResetMainTaskResponse{}, nil
}
