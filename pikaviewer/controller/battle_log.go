package controller

import (
	"os"

	"coin-server/pikaviewer/env"
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

var blHandler = new(handler.BattleLog)

type BattleLog struct {
	Name []string `json:"name"`
}

func (c *BattleLog) FileList(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	files, err := blHandler.FileList()
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(files)
}

func (c *BattleLog) Delete(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	if err := ctx.ShouldBindJSON(c); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("无效的文件：" + err.Error()))
		return
	}
	if err := blHandler.Delete(c.Name); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (c *BattleLog) Download(ctx *gin.Context) {
	contentType := "text/plain"
	// name := time.Now().Format("2006-01-02-15-04") + ".txt"
	// fileContentDisposition := "attachment;filename=\"" + name + "\""
	ctx.Header("Content-Type", contentType)
	// ctx.Header("Content-Disposition", fileContentDisposition)
	ctx.File(os.Getenv(env.BATTLE_LOG_DIR) + "/" + ctx.Param("name"))
}
