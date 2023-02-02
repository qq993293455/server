package handler

import (
	"context"
	"strconv"
	"time"

	"coin-server/common/logger"
	"coin-server/common/statistical"
	"coin-server/common/statistical/models"
	"coin-server/common/statistical2"
	models3 "coin-server/common/statistical2/models"
	"coin-server/common/timer"
	"coin-server/common/utils"
	utils2 "coin-server/pikaviewer/utils"

	"github.com/rs/xid"
	"go.uber.org/zap"
)

const md5Key = "a#OLShaHe-Gs@1BBibxpV5ymBCTX&^p&"

type Stat struct {
	IggId    string `form:"igg_id"`
	Device   string `form:"device" binding:"required"`
	Progress string `form:"progress" binding:"required"`
	UdId     string `form:"ud_id"`
	GameId   string `form:"game_id" binding:"required"`
	Time     int64  `form:"time" binding:"required"`
	Sign     string `form:"sign" binding:"required"`
}

func (h *Stat) Record(req *Stat) error {
	sign := utils.MD5String(req.Device + req.Progress + req.GameId + strconv.Itoa(int(req.Time)) + md5Key)
	if sign != req.Sign {
		logger.DefaultLogger.Error("Record invalid signature",
			zap.String("igg_id", req.IggId),
			zap.String("device", req.Device),
			zap.String("progress", req.Progress),
			zap.String("ud_id", req.UdId),
			zap.String("game_id", req.GameId),
			zap.Int64("time", req.Time),
			zap.String("sign", req.Sign))
		return utils2.NewDefaultErrorWithMsg("invalid signature")
	}
	ls := statistical.GetLogServer(context.Background())

	iggId, err := strconv.Atoi(req.IggId)
	if err != nil {
		logger.DefaultLogger.Warn("Record invalid igg_id",
			zap.String("igg_id", req.IggId),
			zap.String("device", req.Device),
			zap.String("progress", req.Progress),
			zap.String("ud_id", req.UdId),
			zap.String("game_id", req.GameId),
			zap.Int64("time", req.Time),
			zap.String("sign", req.Sign),
			zap.Error(err))
	}
	statistical.Save(ls, &models.Launch{
		IggId:     int64(iggId),
		EventTime: timer.Now(),
		GwId:      statistical.GwId(),
		UDId:      req.UdId,
		GameId:    req.GameId,
		Device:    req.Device,
		Progress:  req.Progress,
	})
	if err := statistical.Flush(ls); err != nil {
		logger.DefaultLogger.Error("Record Flush err",
			zap.String("igg_id", req.IggId),
			zap.String("device", req.Device),
			zap.String("progress", req.Progress),
			zap.String("ud_id", req.UdId),
			zap.String("game_id", req.GameId),
			zap.Int64("time", req.Time),
			zap.String("sign", req.Sign),
			zap.Error(err))
		return err
	}

	ls2 := statistical2.GetLogServer(context.Background())
	statistical2.Save(ls2, &models3.Launch{
		Time:     time.Now(),
		IggId:    req.IggId,
		Xid:      xid.New().String(),
		UDId:     req.UdId,
		GameId:   req.GameId,
		Device:   req.Device,
		Progress: req.Progress,
	})
	if err := statistical2.Flush(ls2); err != nil {
		logger.DefaultLogger.Error("Record Flush 2 err",
			zap.String("igg_id", req.IggId),
			zap.String("device", req.Device),
			zap.String("progress", req.Progress),
			zap.String("ud_id", req.UdId),
			zap.String("GameId", req.GameId),
			zap.Int64("time", req.Time),
			zap.String("sign", req.Sign),
			zap.Error(err))
		return err
	}
	return nil
}
