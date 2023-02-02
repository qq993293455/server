package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
)

type GuildBossService interface {
	// IsGuildBossFighting 是否正在工会Boss战
	IsGuildBossFighting(c *ctx.Context, roleId string) (bool, *errmsg.ErrMsg)
}
