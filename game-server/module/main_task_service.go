package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/values"
)

type MainTaskService interface {
	IsFinishMainTask(ctx *ctx.Context, taskId values.TaskId) (bool, *errmsg.ErrMsg)
}
