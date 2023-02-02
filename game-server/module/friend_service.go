package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/values"
)

type FriendService interface {
	GetFriendIds(ctx *ctx.Context) ([]values.RoleId, *errmsg.ErrMsg)
}
