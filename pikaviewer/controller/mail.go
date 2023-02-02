package controller

import (
	"fmt"

	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin/binding"

	"github.com/gin-gonic/gin"
)

var mailHandler = new(handler.Mail)

type Mail struct {
}

func (c *Mail) Send(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.SendForm{}
	if err := ctx.ShouldBindBodyWith(req, binding.JSON); err != nil {
		fmt.Println(err)
		resp.Send(utils.NewDefaultErrorWithMsg("参数有误，请检查"))
		return
	}
	failed, err := mailHandler.Send(req)
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(failed)
}

func (c *Mail) Delete(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.DeleteForm{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		fmt.Println(err)
		resp.Send(utils.NewDefaultErrorWithMsg("参数有误，请检查"))
		return
	}
	err := mailHandler.Delete(req)
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (c *Mail) EntireMail(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	data, err := mailHandler.GetEntireMail()
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(data)
}
