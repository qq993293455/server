package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type FormationService interface {
	HeroChange(c *ctx.Context, heroId values.HeroId) *errmsg.ErrMsg
	GetDefaultHeroes(c *ctx.Context, roleId values.RoleId) (*models.Assemble, *errmsg.ErrMsg)
}
