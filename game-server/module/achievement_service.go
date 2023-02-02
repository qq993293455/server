package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/values"
)

type AchievementService interface {
	IsDone(ctx *ctx.Context, typ values.AchievementId, gear values.Integer) bool
	CurrGear(ctx *ctx.Context, typ values.AchievementId) (values.Integer, *errmsg.ErrMsg)
}
