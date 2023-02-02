package main

import (
	"fmt"
	"os"

	"coin-server/common/consulkv"
	"coin-server/common/core"
	"coin-server/common/im"
	"coin-server/common/ipregion"
	"coin-server/common/logger"
	"coin-server/common/orm"
	_ "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	_ "coin-server/common/proto/models"
	_ "coin-server/common/proto/service"
	redis "coin-server/common/redisclient"
	"coin-server/common/statistical"
	"coin-server/common/statistical2"
	"coin-server/common/ulimit"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	"coin-server/pikaviewer/docs"
	selfEnv "coin-server/pikaviewer/env"
	"coin-server/pikaviewer/global"
	v1 "coin-server/pikaviewer/router/v1"
	"coin-server/pikaviewer/selector"
	utils2 "coin-server/pikaviewer/utils"
	"coin-server/pikaviewer/web"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	if err := ulimit.SetRLimit(); err != nil {
		panic(err)
	}
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		FilePath:   []string{fmt.Sprintf("./%s.log", "PikaViewer")},
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_GMServer,
			"serverId":   -1,
		},
		Development: true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log, true)
	utils.Must(err)
	redis.Init(cnf)
	orm.InitMySQL(cnf)
	im.Init(cnf)
	statistical.Init(cnf)
	statistical2.Init(cnf)
	utils2.InitGitlab()
	ipregion.Init()
	selector.Init(cnf)

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)
	utils2.InitNats(urls, 0, models.ServerType_GMServer, log)
	global.Init(cnf)

	gin.SetMode(gin.ReleaseMode)
	gmServer()
	apiServer()

	core.WaitClose(func() {})
}

func gmServer() {
	r := gin.Default()

	web.Init(r)

	cc := cors.DefaultConfig()
	cc.AllowHeaders = append(cc.AllowHeaders, "Jwt", "server", "authorization")
	cc.AllowAllOrigins = false
	cc.AllowCredentials = true
	cc.AllowOrigins = []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://10.23.50.200:8080",
		"http://10.23.50.62:8080",
		"http://10.23.33.204:8080",
		"http://10.23.20.53:9999",
		"http://10.23.50.229:9999",
	}
	r.Use(cors.New(cc))

	v1.NewV1Router(r)
	addr := os.Getenv(selfEnv.PIKA_VIEWER_ADDR)
	go func() {
		logger.DefaultLogger.Info("start pika_viewer", zap.String("addr", addr))
		if err := r.RunTLS(addr, "./pikaviewer/cer.crt", "./pikaviewer/private.key"); err != nil {
			panic(err)
		}
	}()
}

func apiServer() {
	r := gin.Default()
	v1.ApiRouter(r)
	docs.Serve(r)
	addr := os.Getenv(selfEnv.PAY_SERVER_ADDR)
	go func() {
		logger.DefaultLogger.Info("Listening and serving HTTP on", zap.String("addr", addr))
		if err := r.Run(addr); err != nil {
			panic(err)
		}
	}()
}
