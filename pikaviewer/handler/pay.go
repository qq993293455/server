package handler

import (
	"time"

	"coin-server/common/dlock"
	"coin-server/common/logger"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
	"coin-server/pikaviewer/model"
	"coin-server/pikaviewer/utils"

	"go.uber.org/zap"
)

type Pay struct {
	SN         string `json:"sn" binding:"required"`
	PcId       int64  `json:"pc_id" binding:"required"`
	IggId      string `json:"igg_id" binding:"required"`
	PaidTime   int64  `json:"paid_time" binding:"required"`
	ExpireTime int64  `json:"expire_time"`
}

func (h *Pay) PaySuccess(req *Pay) error {
	logger.DefaultLogger.Info("pay callback",
		zap.String("sn", req.SN),
		zap.Int64("pc_id", req.PcId),
		zap.String("igg_id", req.IggId),
		zap.Int64("paid_time", req.PaidTime),
		zap.Int64("expire_time", req.ExpireTime),
	)
	locker := dlock.GetLocker()
	if err := locker.Lock(redisclient.GetLocker(), time.Second*5, req.SN); err != nil {
		logger.DefaultLogger.Error("lock err",
			zap.String("sn", req.SN),
			zap.Int64("pc_id", req.PcId),
			zap.String("igg_id", req.IggId),
			zap.Int64("paid_time", req.PaidTime),
			zap.Int64("expire_time", req.ExpireTime),
			zap.Error(err),
		)
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	defer func() {
		if err := locker.Unlock(); err != nil {
			logger.DefaultLogger.Error("unlock pay lock err",
				zap.String("sn", req.SN),
				zap.Int64("pc_id", req.PcId),
				zap.String("igg_id", req.IggId),
				zap.Int64("paid_time", req.PaidTime),
				zap.Int64("expire_time", req.ExpireTime),
				zap.Error(err),
			)
		}
		dlock.PutLocker(locker)
	}()
	user, ok, err := model.NewPlayer().GetUserData(req.IggId)
	if err != nil {
		logger.DefaultLogger.Error("get user data err",
			zap.String("sn", req.SN),
			zap.Int64("pc_id", req.PcId),
			zap.String("igg_id", req.IggId),
			zap.Int64("paid_time", req.PaidTime),
			zap.Int64("expire_time", req.ExpireTime),
			zap.Error(err),
		)
		return err
	}
	if !ok {
		logger.DefaultLogger.Error("PlayerNotExist",
			zap.String("sn", req.SN),
			zap.Int64("pc_id", req.PcId),
			zap.String("igg_id", req.IggId),
			zap.Int64("paid_time", req.PaidTime),
			zap.Int64("expire_time", req.ExpireTime),
			zap.Error(err),
		)
		return utils.PlayerNotExist
	}
	ok, err = model.NewPay().SnExist(user.RoleId, req.SN)
	if err != nil {
		logger.DefaultLogger.Error("SnExist err",
			zap.String("sn", req.SN),
			zap.Int64("pc_id", req.PcId),
			zap.String("igg_id", req.IggId),
			zap.Int64("paid_time", req.PaidTime),
			zap.Int64("expire_time", req.ExpireTime),
			zap.Error(err),
		)
		return err
	}
	if ok {
		logger.DefaultLogger.Error("DuplicateSN",
			zap.String("sn", req.SN),
			zap.Int64("pc_id", req.PcId),
			zap.String("igg_id", req.IggId),
			zap.Int64("paid_time", req.PaidTime),
			zap.Error(err),
		)
		return utils.DuplicateSN
	}
	// h.save2mysql(req, user.RoleId) // 已修改至pay-server里的saveRecord函数
	err = model.NewPay().Save2Queue(&pbdao.PayQueue{
		Sn:         req.SN,
		RoleId:     user.RoleId,
		ServerId:   user.ServerId,
		PcId:       req.PcId,
		PaidTime:   req.PaidTime,
		ExpireTime: req.ExpireTime,
	})
	if err != nil {
		logger.DefaultLogger.Error("Save2Queue err",
			zap.String("sn", req.SN),
			zap.Int64("pc_id", req.PcId),
			zap.String("igg_id", req.IggId),
			zap.Int64("paid_time", req.PaidTime),
			zap.Int64("expire_time", req.ExpireTime),
			zap.Error(err),
		)
	}
	return err
}

// 临时统计，方便查看
func (h *Pay) save2mysql(req *Pay, roleId values.RoleId) {
	pay := &model.Pay{
		RoleId:    roleId,
		SN:        req.SN,
		PcId:      req.PcId,
		IggId:     req.IggId,
		PaidTime:  req.PaidTime,
		CreatedAt: time.Now().Unix(),
	}
	if err := pay.Save(); err != nil {
		logger.DefaultLogger.Error("pay save err",
			zap.String("role_id", roleId),
			zap.String("sn", pay.SN),
			zap.Int64("pc_id", pay.PcId),
			zap.String("igg_id", pay.IggId),
			zap.Int64("paid_time", req.PaidTime),
			zap.Int64("expire_time", req.ExpireTime),
			zap.Error(err),
		)
	}
}
