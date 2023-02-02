package maintask

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/iggsdk"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/statistical"
	stmodels "coin-server/common/statistical/models"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/task/main-task/dao"
	"coin-server/rule"
)

var (
	CantAccept *CantAcceptStatus
	NotStart   *NotStartStatus
	Processing *ProcessingStatus
	Complete   *CompleteStatus
	Finish     *FinishStatus
)

type ITaskStatus interface {
	GetMainTask(ctx *ctx.Context, task *MainTask) (*pbdao.MainTask, *errmsg.ErrMsg)
	AcceptMainTask(ctx *ctx.Context, task *MainTask) *errmsg.ErrMsg
	SubmitMainTask(ctx *ctx.Context, task *MainTask, choose values.TaskId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg)
}

type CantAcceptStatus struct {
	svc *Service
}

func NewCantAccept(svc *Service) {
	CantAccept = &CantAcceptStatus{svc: svc}
}

func (s *CantAcceptStatus) GetMainTask(ctx *ctx.Context, task *MainTask) (*pbdao.MainTask, *errmsg.ErrMsg) {
	return task.MainTask, nil
}

func (s *CantAcceptStatus) AcceptMainTask(ctx *ctx.Context, task *MainTask) *errmsg.ErrMsg {
	return errmsg.NewErrMainTaskStatusIsNotUnAccepted()
}

func (s *CantAcceptStatus) SubmitMainTask(ctx *ctx.Context, task *MainTask, choose values.TaskId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	return nil, errmsg.NewErrMainTaskUnfinished()
}

type NotStartStatus struct {
	svc *Service
}

func NewNotStart(svc *Service) {
	NotStart = &NotStartStatus{svc: svc}
}

func (s *NotStartStatus) GetMainTask(ctx *ctx.Context, task *MainTask) (*pbdao.MainTask, *errmsg.ErrMsg) {
	return task.MainTask, nil
}

func (s *NotStartStatus) AcceptMainTask(ctx *ctx.Context, task *MainTask) *errmsg.ErrMsg {
	if task.MainTask == nil {
		return errmsg.NewErrMainTaskNotUnlock()
	}

	err := s.svc.acceptTask(ctx, task)
	if err != nil {
		return err
	}
	return nil
}

func (s *NotStartStatus) SubmitMainTask(ctx *ctx.Context, task *MainTask, choose values.TaskId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	return nil, errmsg.NewErrMainTaskUnfinished()
}

type ProcessingStatus struct {
	svc *Service
}

func NewProcessing(svc *Service) {
	Processing = &ProcessingStatus{svc: svc}
}

func (s *ProcessingStatus) GetMainTask(ctx *ctx.Context, task *MainTask) (*pbdao.MainTask, *errmsg.ErrMsg) {
	return task.MainTask, nil
}

func (s *ProcessingStatus) AcceptMainTask(ctx *ctx.Context, task *MainTask) *errmsg.ErrMsg {
	return errmsg.NewErrMainTaskAccepted()
}

func (s *ProcessingStatus) SubmitMainTask(ctx *ctx.Context, task *MainTask, choose values.TaskId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	return nil, errmsg.NewErrMainTaskUnfinished()
}

type CompleteStatus struct {
	svc *Service
}

func NewComplete(svc *Service) {
	Complete = &CompleteStatus{svc: svc}
}

func (s *CompleteStatus) GetMainTask(ctx *ctx.Context, task *MainTask) (*pbdao.MainTask, *errmsg.ErrMsg) {
	return task.MainTask, nil
}

func (s *CompleteStatus) AcceptMainTask(ctx *ctx.Context, task *MainTask) *errmsg.ErrMsg {
	return errmsg.NewErrMainTaskAccepted()
}

func (s *CompleteStatus) SubmitMainTask(ctx *ctx.Context, task *MainTask, choose values.TaskId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.MainTask.GetMainTaskById(task.Task.TaskId)
	if !ok {
		return nil, errmsg.NewErrMainTaskNotExist()
	}

	// 收集任务在提交时扣道具
	for _, subTask := range cfg.SubTask {
		if models.TaskType(subTask[0]) == models.TaskType_TaskCollect {
			if subTask[3] == 1 {
				err := s.svc.BagService.SubItem(ctx, ctx.RoleId, subTask[0], subTask[1])
				if err != nil {
					return nil, err
				}
			}
		}
	}

	_, err := s.svc.BagService.AddManyItem(ctx, ctx.RoleId, cfg.Reward)
	if err != nil {
		return nil, err
	}
	task.Task.Status = models.TaskStatus_Finished
	taskFinish := task.Finished
	if err != nil {
		return nil, err
	}
	SetMainTaskFinish(taskFinish, cfg.Chapter, cfg.Idx, models.RewardStatus_Unlocked)

	// 埋点
	statistical.Save(ctx.NewLogServer(), &stmodels.MainTask{
		IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
		EventTime: timer.Now(),
		GwId:      statistical.GwId(),
		RoleId:    ctx.RoleId,
		UserId:    ctx.UserId,
		TaskId:    task.Task.TaskId,
		Chapter:   cfg.Chapter,
		Index:     cfg.Idx,
		Action:    stmodels.TaskActionFinish,
	})

	lastExp := int64(0)
	if task.Task.Last != 0 {
		last, ok := r.MainTask.GetMainTaskById(task.Task.Last)
		if !ok {
			return nil, errmsg.NewErrMainTaskNotExist()
		}
		lastExp = last.ExpProfit
	}
	expProfit := cfg.ExpProfit - lastExp
	illustration := cfg.StoryIllustration
	taskIdx := cfg.Idx

	ctx.PublishEventLocal(&event.MainTaskFinished{
		TaskNo:       task.Task.TaskId,
		TaskIdx:      taskIdx,
		ExpProfit:    expProfit,
		Illustration: illustration,
	})
	// 主线任务完成打点
	s.svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskMainTask, task.Task.TaskId, 1)
	s.svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskMainTaskIndex, 0, taskIdx)
	s.svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskFinishMainTaskCnt, 0, 1)

	// 传了choose先找选择的任务, 错误的选择报错
	if choose != 0 {
		err = s.svc.ChooseMainTask(ctx, task, choose)
		if err != nil {
			return nil, err
		}
		return cfg.Reward, nil
	}
	// 没传choose找下一个任务
	err, _ = s.svc.findNextTask(ctx, task)
	if err != nil {
		return nil, err
	}

	return cfg.Reward, nil
}

type FinishStatus struct {
	svc *Service
}

func NewFinish(svc *Service) {
	Finish = &FinishStatus{svc: svc}
}

func (s *FinishStatus) GetMainTask(ctx *ctx.Context, task *MainTask) (*pbdao.MainTask, *errmsg.ErrMsg) {
	err, find := s.svc.findNextTask(ctx, task)
	if err != nil {
		return nil, err
	}
	if find {
		dao.SaveMainTask(ctx, task.MainTask)
	}
	return task.MainTask, nil
}

func (s *FinishStatus) AcceptMainTask(ctx *ctx.Context, task *MainTask) *errmsg.ErrMsg {
	return errmsg.NewErrMainTaskAccepted()
}

func (s *FinishStatus) SubmitMainTask(ctx *ctx.Context, task *MainTask, choose values.TaskId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	return nil, errmsg.NewErrMainTaskUnfinished()
}
