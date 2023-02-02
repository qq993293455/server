package handler

import (
	"coin-server/common/ctx"
	"coin-server/common/proto/models"
	newcenterpb "coin-server/common/proto/newcenter"
	"coin-server/common/values/env"
	"coin-server/pikaviewer/utils"
)

type MapLineInfo struct {
}

func (h *MapLineInfo) Info(mapId int64) (*models.AllLineInfo, error) {
	out := &newcenterpb.NewCenter_GetMapAllLinesResponse{}
	if err := utils.NATS.RequestWithOut(ctx.GetContext(), env.GetCenterServerId(), &newcenterpb.NewCenter_GetMapAllLinesRequest{MapId: mapId}, out); err != nil {
		return nil, utils.NewDefaultErrorWithMsg(err.Error())
	}
	return out.AllLineInfo, nil
}
