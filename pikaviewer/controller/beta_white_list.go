package controller

import (
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

var bwlHandler = new(handler.BetaWhiteList)

type BetaWhiteList struct {
	Device  string `json:"device" binding:"required"`
	Enable  bool   `json:"enable"`
	Comment string `json:"comment"`
}

func (c *BetaWhiteList) List(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	list, err := bwlHandler.List()
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(list)
}

func (c *BetaWhiteList) Save(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	if err := ctx.ShouldBindJSON(c); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("保存失败：" + err.Error()))
		return
	}
	if err := bwlHandler.Save(c.Device, c.Enable, c.Comment); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (c *BetaWhiteList) Del(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	if err := ctx.ShouldBindJSON(c); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("删除失败：" + err.Error()))
		return
	}
	if err := bwlHandler.Del(c.Device); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}
