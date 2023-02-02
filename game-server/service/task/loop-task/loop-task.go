package loop_task

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/utils/generic/maps"
	"coin-server/common/values"
	ItemId "coin-server/common/values/enum"
	"coin-server/game-server/module"
	"coin-server/game-server/service/task/loop-task/dao"
	"coin-server/game-server/util"
	"coin-server/rule"
	tasktarget "coin-server/rule/factory/task-target"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
}

func NewLoopTaskService(
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
	s.svc.RegisterFunc("获取所有循环任务", s.GetTaskRequest)
	s.svc.RegisterFunc("完成任务", s.ClearLoopTaskRequest)
	s.svc.RegisterFunc("领取阶段奖励", s.DrawStageReward)

	s.svc.RegisterFunc("作弊器完成所有任务", s.CheatFinishTaskRequest)
	s.svc.RegisterFunc("作弊器重置任务", s.CheatResetTaskRequest)

	eventlocal.SubscribeEventLocal(s.HandleRoleLoginEvent)
	eventlocal.SubscribeEventLocal(s.HandleSubmitEvent)
	eventlocal.SubscribeEventLocal(s.HandleKillMonsterEvent)
	eventlocal.SubscribeEventLocal(s.HandleTargetUpdateEvent)
}

func (s *Service) GetTaskRequest(ctx *ctx.Context, _ *servicepb.LoopTask_GetTaskRequest) (*servicepb.LoopTask_GetTaskResponse, *errmsg.ErrMsg) {
	lastLogin, err := dao.GetLastLogin(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	err = s.checkAndRefresh(ctx, lastLogin)
	if err != nil {
		return nil, err
	}
	task, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	sr, err := dao.GetTaskStageReward(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.LoopTask_GetTaskResponse{
		Tasks:  task.Tasks,
		Daily:  sr.Daily,
		Weekly: sr.Weekly,
	}, nil
}

func (s *Service) ClearLoopTaskRequest(ctx *ctx.Context, req *servicepb.LoopTask_ClearLoopTaskRequest) (*servicepb.LoopTask_ClearLoopTaskResponse, *errmsg.ErrMsg) {
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	var task *models.LoopTask
	for _, t := range tasks.Tasks {
		if t.TaskId == req.TaskId {
			task = t
			break
		}
	}

	if task == nil {
		return nil, errmsg.NewErrLoopTaskNotExist()
	}
	items := map[values.ItemId]values.Integer{}
	switch task.Status {
	case models.TaskStatus_Processing:
		return nil, errmsg.NewErrTaskNotCompleted()
	case models.TaskStatus_Completed:
		reader := rule.MustGetReader(ctx)
		for _, t := range tasks.Tasks {
			if t.Kind != task.Kind || t.Status != models.TaskStatus_Completed {
				continue
			}
			cfg, ok := reader.Looptask.GetLooptaskById(t.TaskId)
			if !ok {
				return nil, errmsg.NewErrLoopTaskInvalidType()
			}
			if models.RewardType(cfg.RewardType) == models.RewardType_TypeItem {
				t.Status = models.TaskStatus_Finished
				maps.Merge(items, cfg.Reward)
			}
		}
		_, err = s.AddManyItem(ctx, ctx.RoleId, items)
		if err != nil {
			return nil, err
		}
	case models.TaskStatus_Finished:
		return nil, errmsg.NewErrLoopTaskAlreadyGet()
	case models.TaskStatus_NotStarted:
		return nil, errmsg.NewErrLoopTaskLock()
	}

	dao.SaveTask(ctx, tasks)

	return &servicepb.LoopTask_ClearLoopTaskResponse{
		Reward: items,
	}, nil
}

func (s *Service) DrawStageReward(ctx *ctx.Context, req *servicepb.LoopTask_DrawStageRewardRequest) (*servicepb.LoopTask_DrawStageRewardResponse, *errmsg.ErrMsg) {
	cfg, ok := rule.MustGetReader(ctx).LoopTaskStageReward.GetLoopTaskStageRewardById(req.StageId)
	if !ok {
		return nil, errmsg.NewErrLoopTaskNotExist()
	}

	sr, err := dao.GetTaskStageReward(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	switch cfg.Typ {
	case 1: // 日常
		if sr.Daily[req.StageId] {
			return nil, errmsg.NewErrLoopTaskAlreadyGet()
		}

		active, err := s.BagService.GetItem(ctx, ctx.RoleId, ItemId.DailyTaskActive)
		if err != nil {
			return nil, err
		}
		if active < cfg.Points {
			return nil, errmsg.NewErrLoopTaskNotFinish()
		}

		sr.Daily[req.StageId] = true
	case 2: // 周常
		if sr.Weekly[req.StageId] {
			return nil, errmsg.NewErrLoopTaskAlreadyGet()
		}

		active, err := s.BagService.GetItem(ctx, ctx.RoleId, ItemId.WeeklyTaskActive)
		if err != nil {
			return nil, err
		}
		if active < cfg.Points {
			return nil, errmsg.NewErrLoopTaskNotFinish()
		}

		sr.Weekly[req.StageId] = true
	default:
		panic(fmt.Sprintf("LoopTaskStageReward config unknown Typ: %d, id: %d", cfg.Typ, req.StageId))
	}
	dao.SaveTaskStageReward(ctx, sr)

	_, err = s.BagService.AddManyItem(ctx, ctx.RoleId, cfg.Reward)
	if err != nil {
		return nil, err
	}
	return &servicepb.LoopTask_DrawStageRewardResponse{Rewards: cfg.Reward}, nil
}

// ------------------------------------------------------cheat---------------------------------------------------------//

func (s *Service) CheatFinishTaskRequest(ctx *ctx.Context, _ *servicepb.LoopTask_CheatFinishTaskRequest) (*servicepb.LoopTask_CheatFinishTaskResponse, *errmsg.ErrMsg) {
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks.Tasks {
		if task.Status != models.TaskStatus_Finished {
			task.Status = models.TaskStatus_Completed
		}
	}
	dao.SaveTask(ctx, tasks)
	return &servicepb.LoopTask_CheatFinishTaskResponse{}, nil
}

func (s *Service) CheatResetTaskRequest(ctx *ctx.Context, req *servicepb.LoopTask_CheatResetLoopTaskRequest) (*servicepb.LoopTask_CheatResetLoopTaskResponse, *errmsg.ErrMsg) {
	tasks, err := s.refreshTask(ctx, req.Kind)
	if err != nil {
		return nil, err
	}
	dao.SaveTask(ctx, tasks)
	return &servicepb.LoopTask_CheatResetLoopTaskResponse{}, nil
}

// ----------------------------------------------service---------------------------------------------------------------//

func (s *Service) createTask(ctx *ctx.Context) (*pbdao.LoopTask, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	task := &pbdao.LoopTask{RoleId: ctx.RoleId, Tasks: make([]*models.LoopTask, 0), Lock: map[int64]int64{}}

	for _, cfg := range r.Looptask.List() {
		// 活跃度任务只在选择后加入
		if models.TaskType(cfg.TaskTargetParam[0]) != models.TaskType_TaskDailyActivity &&
			models.TaskType(cfg.TaskTargetParam[0]) != models.TaskType_TaskWeeklyActivity {
			if len(cfg.UnlockCondition) > 2 {
				task.Lock[cfg.Id] = 0
			} else {
				task.Tasks = append(task.Tasks, &models.LoopTask{
					TaskId:  cfg.Id,
					Process: 0,
					Status:  models.TaskStatus_Processing,
					Kind:    cfg.Kind,
				})
			}
		}
	}

	m := map[values.ItemId]values.Integer{
		ItemId.DailyTaskActive:  0,
		ItemId.WeeklyTaskActive: 0,
	}
	_, err := s.AddManyItem(ctx, ctx.RoleId, m)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) refreshTask(ctx *ctx.Context, taskKind models.TaskKind) (*pbdao.LoopTask, *errmsg.ErrMsg) {
	tasks, err := dao.GetTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	// 邮件补发奖励逻辑
	//err = s.sendEmail(ctx, tasks.Tasks, taskKind)
	//if err != nil {
	//	return nil, err
	//}

	// --------------------刷新--------------------//
	r := rule.MustGetReader(ctx)
	// 删除指定类型的任务
	s.delTaskByKind(tasks, taskKind)
	// 根据配置添加新任务
	for _, cfg := range r.Looptask.List() {
		_, locked := tasks.Lock[cfg.Id]
		if models.TaskKind(cfg.Kind) == taskKind &&
			models.TaskType(cfg.TaskTargetParam[0]) != models.TaskType_TaskDailyActivity &&
			models.TaskType(cfg.TaskTargetParam[0]) != models.TaskType_TaskWeeklyActivity &&
			!locked {
			tasks.Tasks = append(tasks.Tasks, &models.LoopTask{
				TaskId:  cfg.Id,
				Process: 0,
				Status:  models.TaskStatus_Processing,
				Kind:    cfg.Kind,
			})
		}
	}

	sr, err := dao.GetTaskStageReward(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	// 重置活跃度道具数量 和 阶段奖励领取记录
	switch taskKind {
	case models.TaskKind_Daily:
		active, err1 := s.GetItem(ctx, ctx.RoleId, ItemId.DailyTaskActive)
		if err1 != nil {
			return nil, err
		}
		err1 = s.SubItem(ctx, ctx.RoleId, ItemId.DailyTaskActive, active)
		if err1 != nil {
			return nil, err
		}
		sr.Daily = map[int64]bool{}
	case models.TaskKind_Weekly:
		active, err1 := s.GetItem(ctx, ctx.RoleId, ItemId.WeeklyTaskActive)
		if err1 != nil {
			return nil, err
		}
		err1 = s.SubItem(ctx, ctx.RoleId, ItemId.WeeklyTaskActive, active)
		if err1 != nil {
			return nil, err
		}
		sr.Weekly = map[int64]bool{}
	}

	dao.SaveTaskStageReward(ctx, sr)
	return tasks, nil
}

func (s *Service) sendEmail(ctx *ctx.Context, tasks []*models.LoopTask, taskKind models.TaskKind) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	// 检查领取情况
	items := map[values.ItemId]values.Integer{}
	for _, task := range tasks {
		cfg, ok := r.Looptask.GetLooptaskById(task.TaskId)
		if !ok {
			continue
		}
		if models.TaskKind(cfg.Kind) == taskKind {
			if task.Status == models.TaskStatus_Completed {
				for k, v := range cfg.Reward {
					items[k] += v
				}
			}
		}
	}

	// 活跃度道具不补发
	delete(items, ItemId.DailyTaskActive)
	delete(items, ItemId.WeeklyTaskActive)
	// 未领取发邮件
	if len(items) != 0 {
		item := make([]*models.Item, 0)
		for id, cnt := range items {
			item = append(item, &models.Item{ItemId: id, Count: cnt})
		}
		mail := &models.Mail{
			Id:         "",
			TextId:     100009,
			ExpiredAt:  time.Now().AddDate(0, 1, 0).Unix() * 1000,
			Args:       nil,
			Attachment: item,
			CreatedAt:  0,
		}
		err := s.MailService.Add(ctx, ctx.RoleId, mail)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) checkAndRefresh(ctx *ctx.Context, loginDao *pbdao.TaskLogin) *errmsg.ErrMsg {
	refreshTime := util.DefaultNextRefreshTime().UnixMilli()
	// 没过一天不刷新
	if loginDao.DailyNextReset >= refreshTime {
		return nil
	}

	role, err := s.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil
	}
	day := timer.DayPass(time.UnixMilli(role.CreateTime), timer.StartTime(ctx.StartTime)) + 1

	loginDao.DayPass = day
	loginDao.DailyNextReset = refreshTime

	// 刷新每日任务
	tasks, err := s.refreshTask(ctx, models.TaskKind_Daily)
	if err != nil {
		return err
	}
	// 如果是每周第一天，刷新每周任务
	weeklyRefreshTime := util.DefaultWeeklyRefreshTime().UnixMilli()
	if loginDao.WeeklyNextReset < weeklyRefreshTime {
		tasks, err = s.refreshTask(ctx, models.TaskKind_Weekly)
		if err != nil {
			return err
		}
		loginDao.WeeklyNextReset = weeklyRefreshTime
	}

	// 刷新完后获取当天最新的任务，登录任务+1
	r := rule.MustGetReader(ctx)
	for _, t := range tasks.Tasks {
		cfg, ok := r.Looptask.GetLooptaskById(t.TaskId)
		if !ok {
			continue
		}
		params := tasktarget.ParseParam(cfg.TaskTargetParam)
		if params == nil {
			panic(errors.New("loop_task_config_target_param_failed: " + strconv.Itoa(int(cfg.Id))))
		}
		if params.TaskType == models.TaskType_TaskLogin {
			s.updateProgress(t, params.TaskType, params.Count, 1, false)
		}
	}
	dao.SaveTask(ctx, tasks)
	dao.SaveLastLogin(ctx, loginDao)
	return nil
}

func (s *Service) delTaskByKind(tasks *pbdao.LoopTask, kind models.TaskKind) {
	tempTasks := make([]*models.LoopTask, 0)
	for _, task := range tasks.Tasks {
		if models.TaskKind(task.Kind) != kind {
			tempTasks = append(tempTasks, task)
		}
	}
	tasks.Tasks = tempTasks
}
