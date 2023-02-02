package maintask

import (
	"fmt"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/game-server/service/task/main-task/dao"
	"coin-server/rule"
)

const (
	FirstTask = 0
)

type MainTask struct {
	*pbdao.MainTask
	ITaskStatus
}

func NewMainTask(ctx *ctx.Context, svc *Service) (*MainTask, *errmsg.ErrMsg) {
	task, err := dao.GetMainTask(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if task == nil {
		task, err = svc.unlockMainTask(ctx)
		if err != nil {
			return nil, err
		}
	}
	mainTask := &MainTask{
		MainTask:    task,
		ITaskStatus: nil,
	}

	r := rule.MustGetReader(ctx)
	if task.Task.TaskId == r.MainTask.List()[FirstTask].Id {
		first, ok := r.MainTask.GetMainTaskById(task.Task.TaskId)
		if !ok {
			panic(fmt.Sprintf("First SubTask not exist!"))
		}
		if first.Accept == 0 {
			err = svc.acceptTask(ctx, mainTask)
			if err != nil {
				return nil, err
			}
			dao.SaveMainTask(ctx, mainTask.MainTask)
		}
	}

	switch task.Task.Status {
	case models.TaskStatus_CantAccept:
		mainTask.ITaskStatus = CantAccept
	case models.TaskStatus_NotStarted:
		mainTask.ITaskStatus = NotStart
	case models.TaskStatus_Processing:
		mainTask.ITaskStatus = Processing
	case models.TaskStatus_Completed:
		mainTask.ITaskStatus = Complete
	case models.TaskStatus_Finished:
		mainTask.ITaskStatus = Finish
	}
	return mainTask, nil
}

func (t *MainTask) Get(ctx *ctx.Context) (*pbdao.MainTask, *errmsg.ErrMsg) {
	return t.GetMainTask(ctx, t)
}

func (t *MainTask) Accept(ctx *ctx.Context) *errmsg.ErrMsg {
	return t.AcceptMainTask(ctx, t)
}

func (t *MainTask) Submit(ctx *ctx.Context, choose values.TaskId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	return t.SubmitMainTask(ctx, t, choose)
}

func (t *MainTask) Save(ctx *ctx.Context) {
	dao.SaveMainTask(ctx, t.MainTask)
}
