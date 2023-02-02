package controller

import (
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

type WhiteList struct {
	Device  string `json:"device" binding:"required"`
	Enable  bool   `json:"enable"`
	Comment string `json:"comment"`
}

func (c *WhiteList) List(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	list, err := wlHandler.List()
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(list)
}

func (c *WhiteList) Save(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	if err := ctx.ShouldBindJSON(c); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("保存失败：" + err.Error()))
		return
	}
	if err := wlHandler.Save(c.Device, c.Enable, c.Comment); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (c *WhiteList) Del(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	if err := ctx.ShouldBindJSON(c); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("删除失败：" + err.Error()))
		return
	}
	if err := wlHandler.Del(c.Device); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}
