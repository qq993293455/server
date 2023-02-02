package controller

import (
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

var buildHandler = new(handler.Build)

type Build struct {
}

func (c *Build) Build(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.Build{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("build err：" + err.Error()))
		return
	}
	if err := buildHandler.Build(req.Branch); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (c *Build) ReadLog(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	resp.Send(buildHandler.ReadLog())
}
