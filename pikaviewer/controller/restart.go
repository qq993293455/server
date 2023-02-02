package controller

import (
	"os"

	"coin-server/common/values/env"
	"coin-server/pikaviewer/handler"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

var restartHandler = handler.NewRestart(
	"./pikaviewer/log/battle-restart.log",
	"./pikaviewer/log/logic-restart.log",
	"./pikaviewer/log/dungeon-restart.log",
	"./pikaviewer/log/dungeon-match-restart.log",

	"./pikaviewer/log/logic.log",
	"./pikaviewer/log/battle.log",
	"./pikaviewer/log/dungeon.log",
	"./pikaviewer/log/dungeon-match.log",
)

type Restart struct{}

func NewRestart() *Restart {
	return &Restart{}
}

func (r *Restart) Pid(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	battle := restartHandler.BattleServerPid()
	logic := restartHandler.LogicServerPid()
	dungeon := restartHandler.DungeonServerPid()
	dungeonMatch := restartHandler.DungeonMatchServerPid()
	resp.Send(gin.H{
		"battle":       battle,
		"logic":        logic,
		"dungeon":      dungeon,
		"dungeonMatch": dungeonMatch,
	})
}

func (r *Restart) Restart(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	typ := utils.ServerType(ctx.Query("type"))
	ip := ctx.ClientIP()
	tag := getTag()
	if typ == utils.Battle {
		resp.Send(restartHandler.RestartBattleServer(ip, tag))
	} else if typ == utils.Logic {
		resp.Send(restartHandler.RestartLogicServer(ip, tag))
	} else if typ == utils.Dungeon {
		resp.Send(restartHandler.RestartDungeonServer(ip, tag))
	} else if typ == utils.DungeonMatch {
		resp.Send(restartHandler.RestartDungeonMatchServer(ip, tag))
	} else {
		resp.Send(gin.H{
			"code": 1,
			"msg":  "type error",
		})
	}
}

func (r *Restart) ReadLog(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	tag := getTag()
	data := restartHandler.ReadLog(utils.ServerType(ctx.Query("type")), tag)
	resp.Send(data)
}

func getTag() string {
	return os.Getenv(env.APP_NAME)
}

func (r *Restart) SyncMap(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	if err := restartHandler.SyncMap(getTag(), ctx.ClientIP()); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}

func (r *Restart) OverwriteDev(ctx *gin.Context) {
	resp := utils.NewResponse(ctx)
	if err := restartHandler.OverwriteDev(getTag(), ctx.ClientIP()); err != nil {
		resp.Send(err)
		return
	}
	resp.Send()
}
