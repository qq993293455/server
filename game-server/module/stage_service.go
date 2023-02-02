package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
)

type StageService interface {
	IsLock(ctx *ctx.Context, sceneId int64) (bool, *errmsg.ErrMsg)
}
