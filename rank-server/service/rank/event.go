package rank

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/rank-server/service/rank/values"
)

func (svc *Service) HandleMemRankValueChangeEvent(ctx *ctx.Context, d *values.MemRankValueChangeData) *errmsg.ErrMsg {
	return svc.MemRankValueChange(ctx, d)
}
