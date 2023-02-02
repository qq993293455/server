package event

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/game-server/event"
)

const AllCount = -10000

type CondHandler func(ctx *ctx.Context, d *event.TargetUpdate, args any) *errmsg.ErrMsg
type Handler struct {
	Args any
	H    CondHandler
}
