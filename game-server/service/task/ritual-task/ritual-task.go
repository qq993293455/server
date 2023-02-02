package ritualtask

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
	models2 "coin-server/common/statistical/models"
	"coin-server/common/timer"
	"coin-server/common/utils/generic/maps"
	"coin-server/common/values"
	ItemId "coin-server/common/values/enum"
	"coin-server/game-server/module"
	"coin-server/game-server/service/task/ritual-task/dao"
	rule2 "coin-server/game-server/service/task/ritual-task/rule"
	"coin-server/rule"
	tasktarget "coin-server/rule/factory/task-target"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewRitualTaskService(
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
	svc.svc.RegisterFunc("获取仪式信息", svc.Info)
	svc.svc.RegisterFunc("提交仪式任务", svc.SubmitRitualTask)
	svc.svc.RegisterFunc("执行仪式", svc.PerformRitual)

	svc.svc.RegisterFunc("作弊完成任务", svc.CheatFinishTask)

	//eventlocal.SubscribeEventLocal(svc.HandleMainTaskFinishedEvent)
	eventlocal.SubscribeEventLocal(svc.HandleSubmitEvent)
	eventlocal.SubscribeEventLocal(svc.HandleTalkEvent)
	eventlocal.SubscribeEventLocal(svc.HandleKillMonsterEvent)
	eventlocal.SubscribeEventLocal(svc.HandleLevelUpEvent)
	eventlocal.SubscribeEventLocal(svc.HandleCollectTaskEvent)
	//eventlocal.SubscribeEventLocal(svc.HandleSysUnlockEvent)
	eventlocal.SubscribeEventLocal(svc.HandleTargetUpdate)
}

func (svc *Service) unlock(ctx *ctx.Context, roleId values.RoleId, targetId values.Integer, needPush bool) *daopb.ChaosRitual {
	targetCfg := rule2.MustGetTargetCfg(ctx, targetId)
	if len(targetCfg.TaskStage) == 0 {
		panic(fmt.Sprintf("MainTaskChapterTarget config TaskStage failed. id: %d", targetCfg.Id))
	}
	ritual := &daopb.ChaosRitual{
		RoleId: roleId,
		Ritual: &models.ChaosRitual{
			TargetId:      targetId,
			IsCompleted:   false,
			IsFinished:    false,
			TargetTaskId:  targetCfg.TaskStage[0],
			TargetTaskIdx: 0,
			Tasks:         map[int64]*models.RitualTask{},
			IsPlayedCg:    false,
		},
	}
	for _, stageCfg := range rule.MustGetReader(ctx).TargetTaskStage.ListByParent(ritual.Ritual.TargetTaskId) {
		if stageCfg.Front != 0 {
			continue
		}
		task := &models.RitualTask{
			TaskId: stageCfg.TargetTaskId,
		}

		svc.initTask(ctx, ritual.Ritual, task, stageCfg.Id)
		ritual.Ritual.Tasks[stageCfg.Id] = task
	}
	dao.Save(ctx, ritual)
	if needPush {
		ctx.PushMessageToRole(ctx.RoleId, &servicepb.RitualTask_RitualUnlockPush{
			Info: ritual.Ritual,
		})
	}
	return ritual
}

func (svc *Service) initTask(ctx *ctx.Context, ritual *models.ChaosRitual, task *models.RitualTask, nextId values.Integer) {
	delete(ritual.Tasks, task.SubTaskId)
	nextCfg := rule2.MustGetRitualTaskCfg(ctx, task.TaskId, nextId)
	params := tasktarget.ParseParam(nextCfg.TaskStageTargetParam)
	if params == nil {
		panic(fmt.Sprintf("TargetTaskStage 任务目标参数配置错误: parent_id: %d, id: %d", task.TaskId, nextId))
	}
	task.SubTaskId = nextId
	task.Type = params.TaskType
	sameTarget := task.Target == params.Target
	task.Target = params.Target
	task.ChoiceIdx = 0

	var count int64
	var err *errmsg.ErrMsg

	// 对以下类型初始化进度
	switch params.TaskType {
	case models.TaskType_TaskLevel:
		count, err = svc.GetLevel(ctx, ctx.RoleId)
		if err != nil {
			panic(err)
		}
	case models.TaskType_TaskCollect, models.TaskType_TaskMapSubmit, models.TaskType_TaskCitySubmit:
		if sameTarget {
			count = task.Progress
		} else {
			count, err = svc.GetItem(ctx, ctx.RoleId, params.Target)
			if err != nil {
				panic(err)
			}
		}
	default:
		cfg, ok := rule.MustGetReader(nil).TaskType.GetTaskTypeById(values.Integer(params.TaskType))
		if !ok {
			panic(errmsg.NewInternalErr("task_type not found: " + strconv.Itoa(int(params.TaskType))))
		}
		if cfg.IsAccumulate {
			counter, err := svc.TaskService.GetCounterByType(ctx, params.TaskType)
			if err != nil {
				panic(err)
			}
			count = counter[params.Target]
		}
	}

	task.Progress = count
	if count >= params.Count {
		// count = nextCfg.Count
		task.Status = models.TaskStatus_Completed
	} else {
		task.Status = models.TaskStatus_Processing
	}
	ritual.Tasks[task.SubTaskId] = task
}

func (svc *Service) getRitual(ctx *ctx.Context, roleId values.RoleId) (ret *daopb.ChaosRitual, err *errmsg.ErrMsg) {
	ret, err = dao.Get(ctx, roleId)
	if err != nil {
		return
	}
	if ret == nil {
		ret = svc.unlock(ctx, roleId, rule2.CureRitualTargetId, false)
	}
	return
}

func (svc *Service) getSuccessRate(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg) {
	up, err := svc.BagService.GetItem(ctx, roleId, ItemId.CureRitualSuccessRateUp)
	if err != nil {
		return 0, err
	}
	down, err := svc.BagService.GetItem(ctx, roleId, ItemId.CureRitualSuccessRateDown)
	if err != nil {
		return 0, err
	}
	return up - down, nil
}

// Info 获取仪式信息
func (svc *Service) Info(ctx *ctx.Context, _ *servicepb.RitualTask_RitualInfoRequest) (*servicepb.RitualTask_RitualInfoResponse, *errmsg.ErrMsg) {
	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.RitualTask_RitualInfoResponse{
		Info: ritual.Ritual,
	}, nil
}

// PlayCG 记录已播放CG动画
func (svc *Service) PlayCG(ctx *ctx.Context, _ *servicepb.RitualTask_PlayRitualCGRequest) (*servicepb.RitualTask_PlayRitualCGResponse, *errmsg.ErrMsg) {
	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	ritual.Ritual.IsPlayedCg = true
	dao.Save(ctx, ritual)
	return &servicepb.RitualTask_PlayRitualCGResponse{}, nil
}

// SubmitRitualTask 提交仪式任务
func (svc *Service) SubmitRitualTask(ctx *ctx.Context, req *servicepb.RitualTask_SubmitRitualTaskRequest) (*servicepb.RitualTask_SubmitRitualTaskResponse, *errmsg.ErrMsg) {
	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	task, ok := ritual.Ritual.Tasks[req.SubTaskId]
	if !ok {
		return nil, errmsg.NewErrTaskNotExist()
	}

	switch task.Status {
	case models.TaskStatus_Processing:
		return nil, errmsg.NewErrTaskNotCompleted()
	case models.TaskStatus_Completed:
		svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskRitualTaskFinish, task.TaskId, req.SubTaskId, true)
		reward, err := svc.complete(ctx, ritual, task)
		if err != nil {
			return nil, err
		}
		dao.Save(ctx, ritual)
		return &servicepb.RitualTask_SubmitRitualTaskResponse{
			Task:    task,
			Rewards: reward,
		}, nil
	case models.TaskStatus_Finished:
		return nil, errmsg.NewErrTaskAlreadyFinish()
	default:
		return &servicepb.RitualTask_SubmitRitualTaskResponse{
			Task: task,
		}, nil
	}
}

func (svc *Service) complete(ctx *ctx.Context, ritual *daopb.ChaosRitual, task *models.RitualTask) (reword map[int64]int64, err *errmsg.ErrMsg) {
	cfg := rule2.GetRitualTaskCfg(ctx, task.TaskId, task.SubTaskId)
	if cfg == nil {
		return nil, errmsg.NewErrTaskCfgNotExist()
	}
	params := tasktarget.ParseParam(cfg.TaskStageTargetParam)
	if params == nil {
		return nil, errmsg.NewErrTaskCfgNotExist()
	}

	switch params.TaskType {
	// 需要扣除物品的类型
	case models.TaskType_TaskCollect, models.TaskType_TaskMapSubmit, models.TaskType_TaskCitySubmit:
		if params.P1 == 0 {
			err = svc.SubItem(ctx, ctx.RoleId, task.Target, params.Count)
			if err != nil {
				return nil, err
			}
		}
	}

	// 发奖
	_, err = svc.AddManyItem(ctx, ctx.RoleId, cfg.Reward)
	if err != nil {
		return nil, err
	}

	// 初始化下一个任务
	next := cfg.Next
	switch len(next) {
	case 0:
		task.Status = models.TaskStatus_Finished
		isCompleted := !maps.InIf(ritual.Ritual.Tasks, func(_ int64, v *models.RitualTask) bool {
			return v.Status != models.TaskStatus_Finished
		})
		if isCompleted { // 当前阶段完成 切换下一阶段
			svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskRitualTargetFinish, task.TaskId, 1)
			svc.initNextTask(ctx, ritual)
		}
	case 1:
		svc.initTask(ctx, ritual.Ritual, task, next[0])
	default:
		svc.initTask(ctx, ritual.Ritual, task, next[int(task.ChoiceIdx)])
	}

	return cfg.Reward, nil
}

func (svc *Service) initNextTask(ctx *ctx.Context, ritual *daopb.ChaosRitual) {
	nextTaskId, ok := rule2.NextRitualTargetTaskId(ctx, ritual.Ritual.TargetId, ritual.Ritual.TargetTaskIdx)
	if !ok {
		ritual.Ritual.IsCompleted = true
	} else {
		ritual.Ritual.Tasks = map[int64]*models.RitualTask{}
		for _, stageCfg := range rule.MustGetReader(ctx).TargetTaskStage.ListByParent(nextTaskId) {
			if stageCfg.Front != 0 {
				continue
			}
			task := &models.RitualTask{
				TaskId: stageCfg.TargetTaskId,
			}

			svc.initTask(ctx, ritual.Ritual, task, stageCfg.Id)
			ritual.Ritual.Tasks[stageCfg.Id] = task
		}
		ritual.Ritual.TargetTaskId = nextTaskId
		ritual.Ritual.TargetTaskIdx += 1
	}
}

// 更新进度
func (svc *Service) updateProgress(ctx *ctx.Context, taskType models.TaskType, target, incr values.Integer, isReplace bool) {
	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return
	}
	isChange := false
	for _, task := range ritual.Ritual.Tasks {
		if svc.baseUpdateProgress(ctx, task, taskType, target, incr, isReplace) {
			isChange = true
		}
	}
	if isChange {
		dao.Save(ctx, ritual)
	}
}

// 更新进度 指定task
func (svc *Service) updateProgressByTaskId(ctx *ctx.Context, subTaskId, target, incr values.Integer, taskType ...models.TaskType) {
	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return
	}
	task, ok := ritual.Ritual.Tasks[subTaskId]
	if !ok {
		return
	}
	isChange := false
	for _, tt := range taskType {
		if svc.baseUpdateProgress(ctx, task, tt, target, incr, false) {
			isChange = true
		}
	}
	if isChange {
		dao.Save(ctx, ritual)
	}
}

// 批量更新进度 如物品更新事件
func (svc *Service) multiUpdateProgress(ctx *ctx.Context, taskType models.TaskType, targets, incr []values.Integer) {
	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return
	}
	isChange := false
	for _, task := range ritual.Ritual.Tasks {
		for i, tg := range targets {
			if svc.baseUpdateProgress(ctx, task, taskType, tg, incr[i], false) {
				isChange = true
			}
		}
	}
	if isChange {
		dao.Save(ctx, ritual)
	}
}

func (svc *Service) baseUpdateProgress(ctx *ctx.Context, task *models.RitualTask, tt models.TaskType, target, incr values.Integer, isReplace bool) bool {
	if task.Status == models.TaskStatus_Finished {
		return false
	}
	if task.Type != tt || task.Target != target {
		return false
	}
	taskCfg := rule2.MustGetRitualTaskCfg(ctx, task.TaskId, task.SubTaskId)
	params := tasktarget.ParseParam(taskCfg.TaskStageTargetParam)
	if params == nil {
		svc.log.Error("任务目标参数配置错误", zap.Int64s("params", taskCfg.TaskStageTargetParam))
		return false
	}
	if isReplace {
		task.Progress = incr
	} else {
		task.Progress += incr
	}

	if task.Status == models.TaskStatus_Processing {
		cfg, ok := rule.MustGetReader(nil).TaskType.GetTaskTypeById(values.Integer(task.Type))
		if !ok {
			panic(fmt.Sprintf("TaskType config not found. id: %d", task.Type))
		}
		if cfg.IsReversed { // 是否反向判定 比如排名
			if task.Progress <= params.Count {
				task.Status = models.TaskStatus_Completed
			}
		} else {
			if task.Progress >= params.Count {
				task.Progress = params.Count
				task.Status = models.TaskStatus_Completed
			}
		}
	}
	// 进度更新推送
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.RitualTask_RitualTaskUpdatePush{
		Task: task,
	})
	return true
}

// PerformRitual 执行仪式
func (svc *Service) PerformRitual(ctx *ctx.Context, _ *servicepb.RitualTask_PerformRitualRequest) (*servicepb.RitualTask_PerformRitualResponse, *errmsg.ErrMsg) {
	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if !ritual.Ritual.IsCompleted {
		return nil, errmsg.NewErrChaosRitualNotCompleted()
	}
	if ritual.Ritual.IsFinished {
		return nil, errmsg.NewErrTaskAlreadyFinish()
	}

	var heroId values.HeroId
	ritual.Ritual.IsFinished = true
	var finalReward map[int64]int64
	if ritual.Ritual.TargetId == rule2.CureRitualTargetId { // 救治仪式
		successRate, err := svc.getSuccessRate(ctx, ctx.RoleId)
		if err != nil {
			return nil, err
		}
		heroId = rule2.RandRitualUnlockHero(ctx, successRate)
		_, err = svc.AddHero(ctx, heroId, false)
		if err != nil {
			return nil, err
		}

		// 埋点
		statistical.Save(ctx.NewLogServer(), &models2.Ritual{
			IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
			EventTime: timer.Now(),
			GwId:      statistical.GwId(),
			RoleId:    ctx.RoleId,
			UserId:    ctx.UserId,
			HeroId:    heroId,
		})
		// 任务完成打点
		svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskRitualUnlockHero, heroId, 1)
	} else {
		cfg := rule2.MustGetTargetCfg(ctx, ritual.Ritual.TargetId)
		finalReward = cfg.FinalReward
		_, err = svc.BagService.AddManyItem(ctx, ctx.RoleId, cfg.FinalReward)
		if err != nil {
			return nil, err
		}
	}

	// 执行仪式打点
	svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskPerformRitual, ritual.Ritual.TargetId, 1)

	cfg, ok := rule2.NextRitualTargetCfg(ctx, ritual.Ritual.TargetId)
	if ok {
		ritual = svc.unlock(ctx, ctx.RoleId, cfg.Id, true)
	}

	dao.Save(ctx, ritual)
	return &servicepb.RitualTask_PerformRitualResponse{
		Info:    ritual.Ritual,
		Rewards: finalReward,
		HeroId:  heroId,
	}, nil
}

// CheatFinishTask 作弊完成任务
func (svc *Service) CheatFinishTask(ctx *ctx.Context, req *servicepb.RitualTask_CheatFinishTaskRequest) (*servicepb.RitualTask_CheatFinishTaskResponse, *errmsg.ErrMsg) {
	ritual, err := svc.getRitual(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	if req.SubTaskId == 0 { // 自动完成所有任务
		err = svc.cheatFinishAllTask(ctx, ritual)
		if err != nil {
			return nil, err
		}
		dao.Save(ctx, ritual)
		return &servicepb.RitualTask_CheatFinishTaskResponse{Info: ritual.Ritual}, nil
	}

	if req.SubTaskId == -1 {
		reward := make(map[values.Integer]values.Integer)
		for _, task := range ritual.Ritual.Tasks {
			if task.Status == models.TaskStatus_Finished {
				continue
			}

			cfg := rule2.GetRitualTaskCfg(ctx, task.TaskId, task.SubTaskId)
			for cfg != nil {
				// 发奖和加仪式概率
				maps.Merge(reward, cfg.Reward)

				if len(cfg.Next) == 0 {
					task.Status = models.TaskStatus_Finished
					break
				}
				svc.initTask(ctx, ritual.Ritual, task, cfg.Next[0])
				cfg = rule2.GetRitualTaskCfg(ctx, task.TaskId, task.SubTaskId)
			}
		}
		_, err2 := svc.AddManyItem(ctx, ctx.RoleId, reward)
		if err2 != nil {
			return nil, err2
		}
		svc.initNextTask(ctx, ritual)
		dao.Save(ctx, ritual)
		return &servicepb.RitualTask_CheatFinishTaskResponse{Info: ritual.Ritual}, nil
	}

	task, ok := ritual.Ritual.Tasks[req.SubTaskId]
	if !ok {
		return nil, errmsg.NewErrTaskNotExist()
	}
	if task.Status == models.TaskStatus_Finished {
		return &servicepb.RitualTask_CheatFinishTaskResponse{Info: ritual.Ritual}, nil
	}

	cfg := rule2.GetRitualTaskCfg(ctx, task.TaskId, task.SubTaskId)
	if len(cfg.Next) > 1 {
		return nil, errmsg.NewNormalErr("多选项任务不可自动完成", "多选项任务不可自动完成")
	}
	// 发奖和加仪式概率
	_, err = svc.AddManyItem(ctx, ctx.RoleId, cfg.Reward)
	if err != nil {
		return nil, err
	}

	if len(cfg.Next) == 0 {
		task.Status = models.TaskStatus_Finished
	} else {
		svc.initTask(ctx, ritual.Ritual, task, cfg.Next[0])
	}
	isCompleted := !maps.InIf(ritual.Ritual.Tasks, func(_ int64, v *models.RitualTask) bool {
		return v.Status != models.TaskStatus_Finished
	})
	if isCompleted {
		svc.initNextTask(ctx, ritual)
	}

	dao.Save(ctx, ritual)
	return &servicepb.RitualTask_CheatFinishTaskResponse{Info: ritual.Ritual}, nil
}

func (svc *Service) cheatFinishAllTask(ctx *ctx.Context, ritual *daopb.ChaosRitual) *errmsg.ErrMsg {
	reward := make(map[values.Integer]values.Integer)

	targetCfg := rule2.MustGetTargetCfg(ctx, ritual.Ritual.TargetId)
	for _, taskId := range targetCfg.TaskStage {
		if ritual.Ritual.TargetTaskId != taskId {
			continue
		}
		for _, stageCfg := range rule.MustGetReader(ctx).TargetTaskStage.ListByParent(taskId) {
			task, ok := ritual.Ritual.Tasks[stageCfg.Id]
			if !ok {
				continue
			}
			if task.Status == models.TaskStatus_Finished {
				continue
			}

			cfg := rule2.GetRitualTaskCfg(ctx, task.TaskId, task.SubTaskId)
			for cfg != nil {
				// 发奖和加仪式概率
				maps.Merge(reward, cfg.Reward)

				if len(cfg.Next) == 0 {
					task.Status = models.TaskStatus_Finished
					break
				}
				svc.initTask(ctx, ritual.Ritual, task, cfg.Next[0])
				cfg = rule2.GetRitualTaskCfg(ctx, task.TaskId, task.SubTaskId)
			}
		}
		// 切换下一阶段
		svc.initNextTask(ctx, ritual)
	}
	ritual.Ritual.IsCompleted = true

	_, err := svc.AddManyItem(ctx, ctx.RoleId, reward)
	return err
}
