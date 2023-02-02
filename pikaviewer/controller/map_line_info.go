package controller

import (
	"strconv"

	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

var mapLineInfoHandler = new(handler.MapLineInfo)

type MapLineInfo struct {
}

func (c *MapLineInfo) Info(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	mapId, _ := strconv.Atoi(ctx.Query("map"))
	if mapId <= 0 {
		resp.Send(utils.NewDefaultErrorWithMsg("无效的地图id"))
		return
	}
	data, err := mapLineInfoHandler.Info(int64(mapId))
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(data)
}
