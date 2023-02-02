package module

import "coin-server/common/ctx"

type SevenDaysService interface {
	IsSevenDaysReceiveAll(ctx *ctx.Context) bool
}
