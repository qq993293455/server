package module

import (
	"coin-server/common/ctx"
	"coin-server/common/values"
)

type ArenaService interface {
	// GetArenaResetRemainSec 获取竞技场下次结算时间
	GetArenaResetRemainSec(c *ctx.Context) values.Integer
}
