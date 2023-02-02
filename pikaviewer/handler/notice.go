package handler

import (
	"time"

	"coin-server/common/proto/models"
	"coin-server/common/proto/notice_service"
	"coin-server/pikaviewer/utils"
	"github.com/rs/xid"
)

type GameNotice struct {
	Title         string `json:"title" binding:"required"`
	Content       string `json:"content" binding:"required"`
	RewardContent string `json:"reward_content"`
	IsCustom      bool   `json:"is_custom"`
	BeginAt       int64  `json:"begin_at" binding:"required"`
	ExpiredAt     int64  `json:"expired_at" binding:"required"`
}

func (h *GameNotice) Publish(req *GameNotice) error {
	if _, err := utils.NATS.RequestProto(0,
		&models.ServerHeader{ServerType: models.ServerType_GMServer},
		&notice_service.Notice_NoticeUpdateRequest{
			Data: &models.Notice{
				Id:        xid.New().String(),
				Title:     req.Title,
				ExpiredAt: req.ExpiredAt,
				CreatedAt: time.Now().Unix(),
			},
			Content: &models.NoticeContent{
				Content:       req.Content,
				RewardContent: req.RewardContent,
				IsCustom:      req.IsCustom,
				ExpiredAt:     req.ExpiredAt,
			},
			BeginAt: req.BeginAt,
		}); err != nil {
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}

func (h *GameNotice) Delete(id string) error {
	if _, err := utils.NATS.RequestProto(0, &models.ServerHeader{
		ServerType: models.ServerType_GMServer,
	}, &notice_service.Notice_NoticeDeleteRequest{
		NoticeId: id,
	}); err != nil {
		return utils.NewDefaultErrorWithMsg(err.Error())
	}
	return nil
}
