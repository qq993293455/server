package controller

import (
	"strconv"
	"time"

	"coin-server/common/proto/models"
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

var (
	announcementHandler = new(handler.Announcement)
	wlHandler           = new(handler.WhiteList)
)

type Announcement struct {
	Id string `json:"id" binding:"required"`
}

var allType = []string{"maintenance", "normal", "force_update"}

func (c *Announcement) List(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	list, err := announcementHandler.Find()
	if err != nil {
		resp.Send(err)
		return
	}
	resp.Send(list)
}

func (c *Announcement) Save(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	req := &handler.Announcement{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("保存失败：" + err.Error()))
		return
	}
	if err := announcementHandler.Save(req); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (c *Announcement) GetPB(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	version := ctx.Param("version")
	typ := ctx.Query("type")
	language, _ := strconv.Atoi(ctx.Query("language"))
	deviceId := ctx.Query("device")
	web := ctx.Query("web")
	isGM := web == "1"
	data := make(map[string]*models.Announcement)
	now := time.Now().UnixMilli()
	var hasMaintenance bool
	if typ != "" {
		a, err := announcementHandler.GetPB(typ, version, int64(language), isGM)
		if err != nil {
			resp.Send(err)
			return
		}
		data[typ] = a
	} else {
		for _, t := range allType {
			a, err := announcementHandler.GetPB(t, version, int64(language), isGM)
			if err != nil {
				resp.Send(err)
				return
			}
			data[t] = a
		}
	}
	if !isGM {
		for t, el := range data {
			if el == nil {
				continue
			}
			if el.EndTime > 0 && el.EndTime <= now {
				data[t] = nil
			} else if t == "maintenance" {
				hasMaintenance = true
			}
		}
	}
	if hasMaintenance {
		ok, err := wlHandler.IsInWhiteList(deviceId)
		if err != nil {
			resp.Send(err)
			return
		}
		// 白名单，不下发维护公告
		if ok {
			data["maintenance"] = nil
		}
	}
	resp.Send(data)
}

func (c *Announcement) Del(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	if err := ctx.ShouldBindJSON(c); err != nil {
		resp.Send(utils.NewDefaultErrorWithMsg("无效的id"))
		return
	}
	if err := announcementHandler.Del(c.Id); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}
