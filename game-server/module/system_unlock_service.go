package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/values"
)

type SystemUnlockService interface {
	GetMultiSysUnlock(ctx *ctx.Context, roleIds []values.RoleId) (map[values.RoleId]*dao.SystemUnlock, *errmsg.ErrMsg)
	GetMultiSysUnlockBySys(ctx *ctx.Context, roleIds []values.RoleId, id values.SystemId) (map[values.RoleId]bool, *errmsg.ErrMsg)
}
