package controller

import (
	"log"
	"net/http"

	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

var gitlabHandler = new(handler.Gitlab)

type Gitlab struct {
}

func (c *Gitlab) Members(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	// projectId, _ := strconv.Atoi(ctx.Param("pid"))
	// if projectId <= 0 {
	//	resp.Send(utils.NewDefaultErrorWithMsg("无效的项目"))
	//	return
	// }
	members, err := gitlabHandler.Members()
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(members)
}

func (c *Gitlab) Modify(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.Gitlab{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("参数有误，请检查"))
		return
	}
	success, err := gitlabHandler.ModifyAccessLevel(req)
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(success)
}

func (c *Gitlab) RestoreAccessLevel(ctx *gin.Context) {
	req := &handler.RestoreForm{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.Status(http.StatusBadRequest)
		log.Println("RestoreAccessLevel ShouldBindJSON err:", err)
		return
	}
	if err := gitlabHandler.RestoreAccessLevel(req); err != nil {
		ctx.Status(http.StatusBadRequest)
		log.Println("gitlabHandler.RestoreAccessLevel err:", err)
		return
	}
	ctx.Status(http.StatusOK)
}
