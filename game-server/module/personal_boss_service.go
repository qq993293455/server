package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/values"
)

type PersonalBossService interface {
	// GetPersonalBossResetRemainSec 获取个人BOSS距下次重置剩余多少秒
	GetPersonalBossResetRemainSec(c *ctx.Context) (values.Integer, *errmsg.ErrMsg)
}
