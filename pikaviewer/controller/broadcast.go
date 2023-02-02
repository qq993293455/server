package controller

import (
	"strconv"

	ctx2 "coin-server/common/ctx"
	"coin-server/common/im"
	"coin-server/pikaviewer/utils"
	"github.com/gin-gonic/gin/binding"

	"github.com/gin-gonic/gin"
)

type Broadcast struct {
	ParseType int `json:"parse_type" binding:"required"`
	Time      int `json:"time"` // 毫秒级时间戳
}

func (c *Broadcast) Broadcast(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	if err := ctx.ShouldBindJSON(c); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("广播失败：" + err.Error()))
		return
	}
	if c.ParseType != im.ParseTypeClientUpdate && c.ParseType != im.ParseTypeNoticeOperator {
		resp.Send(utils.NewDefaultErrorWithMsg("广播失败：无效的ParseType"))
		return
	}
	msg := &im.Message{
		Type:       im.MsgTypeBroadcast,
		RoleID:     "admin",
		RoleName:   "admin",
		Content:    strconv.Itoa(c.Time),
		ParseType:  c.ParseType,
		IsMarquee:  false,
		IsVolatile: true,
	}
	if err := im.DefaultClient.SendMessage(ctx2.GetContext(), msg); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

type Notice struct {
	Data string `json:"data"` // 公告内容 示例：{"eng": {"title": "en title", "content": "en content"}, "chs": {"title": "简中标题", "content": "简中内容"}}
}

func (c *Notice) Broadcast(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	if err := ctx.ShouldBindBodyWith(c, binding.JSON); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("广播失败：" + err.Error()))
		return
	}
	msg := &im.Message{
		Type:       im.MsgTypeBroadcast,
		RoleID:     "admin",
		RoleName:   "admin",
		Content:    c.Data,
		ParseType:  im.ParseTypeNoticeOperator,
		IsMarquee:  false,
		IsVolatile: true,
	}
	if err := im.DefaultClient.SendMessage(ctx2.GetContext(), msg); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}
