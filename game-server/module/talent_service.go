package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type TalentService interface {
	GetTalentAttr(c *ctx.Context, configId values.Integer) (*models.TalentAdvance, *errmsg.ErrMsg)
	GetAllUnused(c *ctx.Context) (map[values.HeroId]map[values.Integer]values.Integer, *errmsg.ErrMsg)
	GetUnused(c *ctx.Context, configId values.HeroId) (map[values.Integer]values.Integer, *errmsg.ErrMsg)
}
