package controller

import (
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin/binding"

	"github.com/gin-gonic/gin"
)

var payHandler = new(handler.Pay)

type Pay struct{}

func (c *Pay) PaySuccess(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.Pay{}
	if err := ctx.ShouldBindBodyWith(req, binding.JSON); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("invalid request parameter: " + err.Error()))
		return
	}
	if err := payHandler.PaySuccess(req); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}
