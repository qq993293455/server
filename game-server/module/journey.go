package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
)

type JourneyService interface {
	AddToken(c *ctx.Context, roleId string, journeyListId int64, count int64) *errmsg.ErrMsg
}
