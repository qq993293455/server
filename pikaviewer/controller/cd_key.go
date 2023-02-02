package controller

import (
	"strconv"

	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"
	"github.com/gin-gonic/gin"
)

var (
	cdKeyHandler = new(handler.CdKeyGen)
)

type CdKey struct {
	Id string `json:"id" binding:"required"`
}

func (c *CdKey) Save(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.CdKeyGen{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("保存失败：" + err.Error()))
		return
	}
	if err := cdKeyHandler.Save(req); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (c *CdKey) DeActive(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	batchId := ctx.Query("batch_id")
	batchIdInt, err := strconv.ParseInt(batchId, 10, 64)
	if err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("批次号错误：" + err.Error()))
		return
	}
	if err = cdKeyHandler.DeActive(batchIdInt); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("保存失败：" + err.Error()))
		return
	}
	resp.Send()
}
