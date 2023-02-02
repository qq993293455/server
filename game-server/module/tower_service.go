package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type TowerService interface {
	GetTowerCnt(c *ctx.Context) ([2]values.Integer, *errmsg.ErrMsg)
	GetLevelInfo(ctx *ctx.Context, ctype models.TowerType) (*models.Tower, *errmsg.ErrMsg)
}
