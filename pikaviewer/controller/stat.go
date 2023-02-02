package controller

import (
	"coin-server/common/logger"
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

var statHandler = new(handler.Stat)

type Stat struct {
}

func (c *Stat) Record(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.Stat{}
	if err := ctx.ShouldBind(req); err != nil {
		logger.DefaultLogger.Error("Record ShouldBindJSON err", zap.Error(err))
		resp.Send(utils.NewDefaultErrorWithMsg("invalid args"))
		return
	}
	logger.DefaultLogger.Info("stat",
		zap.String("igg_id", req.IggId),
		zap.String("device", req.Device),
		zap.String("progress", req.Progress),
		zap.String("ud_id", req.UdId),
		zap.String("game_id", req.GameId),
		zap.Int64("time", req.Time),
		zap.String("sign", req.Sign))
	if err := statHandler.Record(req); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}
