package controller

import (
	"strconv"

	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin/binding"

	"github.com/gin-gonic/gin"
)

var playerHandler = new(handler.Player)

type Player struct {
}

func (c *Player) PlayerInfo(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	name := ctx.Query("name")
	if name == "" {
		resp.Send(utils.NewDefaultErrorWithMsg("请输入要查询的关键字"))
		return
	}
	data, total, err := playerHandler.GetPlayerInfo(name)
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(gin.H{
		"list":  data,
		"total": total,
	})
}

func (c *Player) MailInfo(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	uid, _ := strconv.Atoi(ctx.Query("uid"))
	if uid <= 0 {
		resp.Send(utils.NewDefaultErrorWithMsg("无效的UID"))
		return
	}
	data, err := playerHandler.MailInfo(uint64(uid))
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(data)
}

func (c *Player) KickOffUser(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	uid, _ := strconv.Atoi(ctx.Query("uid"))
	if uid <= 0 {
		resp.Send(utils.NewDefaultErrorWithMsg("无效的UID"))
		return
	}
	sec, _ := strconv.Atoi(ctx.Query("sec"))
	if err := playerHandler.KickOffUser(uid, int64(sec), 0); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

type KickPlayerReq struct {
	IggId  int   `json:"igg_id" binding:"required"`
	Sec    int64 `json:"sec"`
	Status int64 `json:"status"`
}

func (c *Player) KickPlayer(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &KickPlayerReq{}
	if err := ctx.ShouldBindBodyWith(req, binding.JSON); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("invalid request parameter: " + err.Error()))
		return
	}
	if req.Status != 0 && req.Status != 2 {
		resp.Send(utils.NewDefaultErrorWithMsg("invalid request parameter: status (0,2)"))
		return
	}
	if err := playerHandler.KickOffUser(req.IggId, req.Sec, req.Status); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

type GetPlayerInfoForSOPReq struct {
	Val     int  `json:"val" binding:"required"`
	IsIggId bool `json:"is_igg_id"`
}

func (c *Player) GetPlayerInfoForSOP(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &GetPlayerInfoForSOPReq{}
	if err := ctx.ShouldBindBodyWith(req, binding.JSON); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("invalid request parameter: " + err.Error()))
		return
	}
	if req.Val <= 0 {
		resp.Send(utils.NewDefaultErrorWithMsg("invalid val"))
		return
	}
	data, err := playerHandler.GetPlayerInfoForSOP(req.Val, req.IsIggId)
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(data)
}

type BanPlayerReq struct {
	IggId    int   `json:"igg_id" binding:"required"`
	Duration int64 `json:"duration"`
}

func (c *Player) BanPlayer(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &BanPlayerReq{}
	if err := ctx.ShouldBindBodyWith(req, binding.JSON); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("invalid request parameter: " + err.Error()))
		return
	}
	if err := playerHandler.BanPlayer(req.IggId, req.Duration); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

type HandleChat struct {
	IggId int `json:"igg_id" binding:"required"`
	Sec   int `json:"sec"`
}

func (c *Player) BanChat(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &HandleChat{}
	if err := ctx.ShouldBindBodyWith(req, binding.JSON); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("invalid request parameter: " + err.Error()))
		return
	}
	if req.Sec <= 0 {
		resp.Send(utils.NewDefaultErrorWithMsg("invalid request parameter: sec"))
		return
	}
	if err := playerHandler.BanChat(req.IggId, req.Sec); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (c *Player) UnBanChat(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &HandleChat{}
	if err := ctx.ShouldBindBodyWith(req, binding.JSON); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("invalid request parameter: " + err.Error()))
		return
	}
	if err := playerHandler.UnBanChat(req.IggId); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (c *Player) OnlineCount(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	count, err := playerHandler.OnlineCount()
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(count)
}

func (c *Player) Total(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	count, err := playerHandler.Total()
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(count)
}

func (c *Player) UpdatePlayerCurrency(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.UpdatePlayerCurrencyReq{}
	if err := ctx.ShouldBindBodyWith(req, binding.JSON); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("invalid request parameter: " + err.Error()))
		return
	}
	if err := playerHandler.UpdatePlayerCurrency(req); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}
