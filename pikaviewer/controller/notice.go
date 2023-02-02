package controller

import (
	"encoding/json"

	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var noticeHandler = new(handler.GameNotice)

type GameNotice struct {
}

func (c *GameNotice) Publish(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.GameNotice{}
	if err := ctx.ShouldBindBodyWith(req, binding.JSON); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("发布失败：" + err.Error()))
		return
	}
	temp := make(map[int64]string)
	if err := json.Unmarshal([]byte(req.Title), &temp); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("标题数据格式错误：" + err.Error()))
		return
	}
	if err := json.Unmarshal([]byte(req.Content), &temp); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("内容数据格式错误：" + err.Error()))
		return
	}
	if err := noticeHandler.Publish(req); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

type GameNoticeDelForm struct {
	Id string `json:"id" binding:"required"`
}

func (c *GameNotice) Delete(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &GameNoticeDelForm{}
	if err := ctx.ShouldBindBodyWith(req, binding.JSON); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("删除失败：" + err.Error()))
		return
	}
	if err := noticeHandler.Delete(req.Id); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}
