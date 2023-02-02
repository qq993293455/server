package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/values"
)

type GuildService interface {
	GetGuildIdByRole(ctx *ctx.Context) (values.GuildId, *errmsg.ErrMsg)                                            // 如果玩家不在公会中，返回""
	GetUserGuildPositionByGuildId(ctx *ctx.Context, guildId values.GuildId) (values.GuildPosition, *errmsg.ErrMsg) // 如果没有公会，返回0
	GetUserGuildInfo(ctx *ctx.Context, roleId values.RoleId) (*dao.Guild, *errmsg.ErrMsg)                          // 如果没有公会，返回nil
	// GetMultiGuildByRoleId 如果对应的roleId没有公会返回nil
	GetMultiGuildByRoleId(ctx *ctx.Context, roleIds []values.RoleId) (map[values.RoleId]*dao.Guild, *errmsg.ErrMsg)
	// GetGuildMaxMemberCount 获取公会最大人数上限
	GetGuildMaxMemberCount(ctx *ctx.Context, guildId values.GuildId) (values.Integer, *errmsg.ErrMsg)
	IsUnlockGuildBoss(ctx *ctx.Context) *errmsg.ErrMsg
}
