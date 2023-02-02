package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type UserService interface {
	GetRole(ctx *ctx.Context, roleIds []values.RoleId) (map[values.RoleId]*dao.Role, *errmsg.ErrMsg)
	GetRoleByRoleId(ctx *ctx.Context, roleId values.RoleId) (*dao.Role, *errmsg.ErrMsg)
	GetRoleModelByRoleId(ctx *ctx.Context, roleId values.RoleId) (*models.Role, *errmsg.ErrMsg)
	GetRoleByUserId(ctx *ctx.Context, userId string) (*dao.Role, *errmsg.ErrMsg)
	GetAvatar(c *ctx.Context) (values.Integer, *errmsg.ErrMsg)
	GetAvatarFrame(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg)
	GetTitle(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg)
	GetLevel(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg)
	GetPower(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg)
	GetMap(ctx *ctx.Context, userId string) (values.MapId, *errmsg.ErrMsg)
	GetRoleAttrByType(ctx *ctx.Context, roleId values.RoleId, typ models.AttrBonusType) (*models.RoleAttr, *errmsg.ErrMsg)
	GetRoleAttr(ctx *ctx.Context, roleId values.RoleId) ([]*models.RoleAttr, *errmsg.ErrMsg)
	GetUserById(ctx2 *ctx.Context, userId string) (*dao.User, *errmsg.ErrMsg)
	GetUserByRoleIds(ctx *ctx.Context, roleIds []values.RoleId) (map[values.RoleId]*dao.User, *errmsg.ErrMsg)
	GetRoleSkill(ctx *ctx.Context, roleId values.RoleId) ([]values.Integer, *errmsg.ErrMsg)
	GetRoguelikeCnt(ctx *ctx.Context) ([2]values.Integer, *errmsg.ErrMsg)
	GetRegisterDay(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg)

	SaveUser(c *ctx.Context, u *dao.User)
	SaveRole(ctx *ctx.Context, role *dao.Role) *errmsg.ErrMsg

	GetExtraSkillCnt(ctx *ctx.Context, typId, logicId values.Integer) (values.Integer, *errmsg.ErrMsg)

	AddExp(ctx *ctx.Context, roleId values.RoleId, count values.Integer, isHang bool) *errmsg.ErrMsg
}
