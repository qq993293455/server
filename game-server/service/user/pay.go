package user

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/game-server/event"
	"coin-server/rule"
	"go.uber.org/zap"
)

func PayAuth(next handler.HandleFunc) handler.HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		if ctx.ServerType != models.ServerType_PayServer {
			ctx.Error("PayAuth err",
				zap.String("role_id", ctx.RoleId),
				zap.Int64("server_type", int64(ctx.ServerType)))
			return errmsg.NewProtocolErrorInfo("invalid server type request")
		}
		return next(ctx)
	}
}
func (this_ *Service) PaySuccess(ctx *ctx.Context, req *servicepb.Pay_Success) {
	ctx.Debug("PaySuccess", zap.String("role_id", ctx.RoleId),
		zap.Int64("pc_id", req.PcId),
		zap.Int64("paid_time", req.PaidTime),
		zap.Int64("expire_time", req.ExpireTime))
	ctx.PublishEventLocal(&event.PaySuccess{
		PcId:       req.PcId,
		PaidTime:   req.PaidTime,
		ExpireTime: req.ExpireTime,
	})
	cfg, ok := rule.MustGetReader(ctx).Charge.GetChargeById(req.PcId)
	if !ok {
		ctx.Error("charge config not found", zap.Int64("id", req.PcId), zap.String("role_id", ctx.RoleId))
		return
	}
	ctx.PublishEventLocal(&event.RechargeAmountEvt{
		Amount: cfg.ChargeNum / 100,
	})
}
