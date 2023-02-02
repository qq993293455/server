package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	rulemodel "coin-server/rule/rule-model"
)

type HeroService interface {
	AddHero(ctx *ctx.Context, id values.HeroId, register bool) (*models.Hero, *errmsg.ErrMsg)
	GetHero(ctx *ctx.Context, roleId values.RoleId, id values.HeroId) (*models.Hero, bool, *errmsg.ErrMsg)
	GetAllHero(ctx *ctx.Context, id values.RoleId) ([]*models.Hero, *errmsg.ErrMsg)
	GetHeroes(ctx *ctx.Context, roleId values.RoleId, ids []values.HeroId) ([]*models.Hero, *errmsg.ErrMsg)
	GetInitSkills(cfg *rulemodel.RowHero) []values.HeroSkillId
	GetAllHeroId(ctx *ctx.Context) ([]values.HeroId, *errmsg.ErrMsg) // 获取玩家已经拥有的所有英雄的originId
	GetHeroesEquippedEquipId(heroes []*models.Hero) []values.EquipId
	ActivateFashion(ctx *ctx.Context, id values.ItemId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg)
	FashionExpiredCheck(ctx *ctx.Context, role *dao.Role) *errmsg.ErrMsg
}
