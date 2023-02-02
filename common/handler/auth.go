package handler

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
)

func GameServerAuth(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		if ctx.ServerType != models.ServerType_GameServer {
			return errmsg.NewProtocolErrorInfo("invalid server type request")
		}
		return next(ctx)
	}
}

func GVGServerAuth(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		if ctx.ServerType != models.ServerType_GVGGuildServer {
			return errmsg.NewProtocolErrorInfo("invalid server type request")
		}
		return next(ctx)
	}
}

func GVGOrGameServerAuth(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		if ctx.ServerType != models.ServerType_GVGGuildServer && ctx.ServerType != models.ServerType_GameServer {
			return errmsg.NewProtocolErrorInfo("invalid server type request")
		}
		return next(ctx)
	}
}

func GMAuth(next HandleFunc) HandleFunc {
	return func(ctx *ctx.Context) (err *errmsg.ErrMsg) {
		if ctx.ServerType != models.ServerType_GMServer {
			return errmsg.NewProtocolErrorInfo("invalid server type request")
		}
		return next(ctx)
	}
}
