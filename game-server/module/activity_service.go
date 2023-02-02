package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/values"
)

type ActivityService interface {
	StellargemShopBuy(ctx *ctx.Context, id values.Integer) (map[values.ItemId]values.Integer, values.Integer, *errmsg.ErrMsg)
}
