package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/values"
)

type MapEventService interface {
	// 添加事件
	AddEvent(ctx *ctx.Context, roleId values.RoleId, eventId values.EventId, mapId values.MapId) *errmsg.ErrMsg
}
