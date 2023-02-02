package v1

import (
	"net/http"
	"os"
	"strings"

	"coin-server/common/values/env"
	"coin-server/pikaviewer/controller"
	"coin-server/pikaviewer/global"
	"coin-server/pikaviewer/utils"

	"github.com/gin-gonic/gin"
)

var (
	system           = new(controller.System)
	restartCtrl      = controller.NewRestart()
	userCtrl         = new(controller.User)
	routesCtrl       = new(controller.Routes)
	playerCtrl       = new(controller.Player)
	mailCtrl         = new(controller.Mail)
	announcementCtrl = new(controller.Announcement)
	battleLogCtrl    = new(controller.BattleLog)
	buildCtrl        = new(controller.Build)
	versionCtrl      = new(controller.Version)
	broadcastCtrl    = new(controller.Broadcast)
	noticeCtrl       = new(controller.Notice)
	whiteListCtrl    = new(controller.WhiteList)
	gitlabCtrl       = new(controller.Gitlab)
	cdKeyCtrl        = new(controller.CdKey)
	bwlCtrl          = new(controller.BetaWhiteList)
	mapLineInfoCtrl  = new(controller.MapLineInfo)
	statCtrl         = new(controller.Stat)
)

func NewV1Router(r *gin.Engine) {
	router := r.Group("v1")
	router.GET("name", system.Name)

	router.GET("routes", Logged(), routesCtrl.List)

	query := router.Group("query")
	query.Use(Logged())
	{
		qc := controller.NewQuery()
		query.POST("data", qc.Do)
	}
	// node := router.Group("node")
	// {
	// 	nc := controller.NewNodes()
	// 	node.GET("module", nc.ModuleModes)
	// 	node.GET("pika", nc.PikaNodes)
	// }

	restart := router.Group("restart")
	restart.Use(Logged())
	{

		restart.GET("pid", restartCtrl.Pid)
		restart.GET("do", restartCtrl.Restart)
		restart.GET("log", restartCtrl.ReadLog)
		restart.GET("/sync/map", restartCtrl.SyncMap)
		restart.GET("/overwrite/dev", restartCtrl.OverwriteDev)
	}

	admin := router.Group("admin")
	user := admin.Group("user")
	user.POST("login", userCtrl.Login)

	player := admin.Group("player")
	player.Use(Logged())
	player.GET("/info", playerCtrl.PlayerInfo)
	player.GET("/mail/info", playerCtrl.MailInfo)
	player.POST("/mail/send", block(), mailCtrl.Send)
	player.POST("/mail/delete", block(), mailCtrl.Delete)
	player.GET("/mail/entire", block(), mailCtrl.EntireMail)
	player.GET("/kick/off", playerCtrl.KickOffUser)

	announcement := router.Group("announcement")
	announcement.GET("/:version", announcementCtrl.GetPB)
	announcement.Use(Logged())
	announcement.GET("/list", announcementCtrl.List)
	announcement.POST("/save", announcementCtrl.Save)
	announcement.POST("/del", announcementCtrl.Del)

	battleLog := router.Group("battle/log")
	battleLog.Use(Logged())
	battleLog.GET("/files", battleLogCtrl.FileList)
	battleLog.POST("/del", battleLogCtrl.Delete)
	battleLog.GET("/download/:name", battleLogCtrl.Download)

	build := router.Group("build")
	build.Use(Logged())
	build.POST("/do", buildCtrl.Build)
	build.GET("/log", buildCtrl.ReadLog)

	version := router.Group("version")
	version.GET("/info", versionCtrl.Get)
	version.POST("/save", Logged(), versionCtrl.Save)
	version.POST("/file", Logged(), versionCtrl.UploadVersionFile)
	version.POST("/zip", Logged(), versionCtrl.UploadHotUpdateFile)

	broadcast := router.Group("broadcast")
	broadcast.Use(Logged())
	broadcast.POST("/send", broadcastCtrl.Broadcast)

	// 游戏公告接口
	notice := router.Group("notice")
	// notice.Use(Logged())
	notice.POST("/send", noticeCtrl.Broadcast)

	whiteList := router.Group("whitelist")
	whiteList.Use(Logged())
	whiteList.GET("/list", whiteListCtrl.List)
	whiteList.POST("/save", whiteListCtrl.Save)
	whiteList.POST("/del", whiteListCtrl.Del)

	gitlab := router.Group("gitlab")
	gitlab.POST("/restore", GitlabAuth(), gitlabCtrl.RestoreAccessLevel)
	gitlab.Use(Logged())
	gitlab.GET("/members", gitlabCtrl.Members)
	gitlab.POST("/modify", gitlabCtrl.Modify)

	cdKey := router.Group("cdkey")
	cdKey.Use(Logged(), block())
	cdKey.POST("/save", cdKeyCtrl.Save)
	cdKey.GET("/deactive", cdKeyCtrl.DeActive)

	bwl := router.Group("beta/whitelist")
	bwl.Use(Logged())
	bwl.GET("/list", bwlCtrl.List)
	bwl.POST("/save", bwlCtrl.Save)
	bwl.POST("/del", bwlCtrl.Del)

	mapLineInfo := router.Group("map/line/info")
	mapLineInfo.Use(Logged())
	mapLineInfo.GET("", mapLineInfoCtrl.Info)

	stat := router.Group("stat")
	stat.GET("record", statCtrl.Record)
}

func Logged() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorization := ctx.GetHeader("Authorization")
		temp := strings.Split(authorization, " ")
		response := utils.NewResponse(ctx)
		if len(temp) != 2 || temp[1] == "" {
			ctx.Abort()
			response.Send(utils.NewDefaultErrorWithMsg("登录已过期，请重新登录"))
		} else {
			utils.SetToken(ctx, temp[1])
			ctx.Next()
		}
	}
}

func GitlabAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("X-Gitlab-Token")
		if token != "d822321fa6194cb788c059520a8e177c" {
			ctx.Abort()
			ctx.Status(http.StatusForbidden)
		} else {
			ctx.Next()
		}
	}
}

func block() gin.HandlerFunc {
	appName := os.Getenv(env.APP_NAME)
	return func(ctx *gin.Context) {
		var ok bool
		for _, item := range global.GMAPIWhiteList {
			if item == appName {
				ok = true
				break
			}
		}
		if ok {
			ctx.Next()
		} else {
			response := utils.NewResponse(ctx)
			ctx.Abort()
			response.Send(utils.NewDefaultErrorWithMsg("Forbidden"))
		}
	}
}
