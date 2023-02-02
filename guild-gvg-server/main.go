package main

import (
	"fmt"
	"os"

	"coin-server/common/consulkv"
	"coin-server/common/core"
	"coin-server/common/im"
	"coin-server/common/logger"
	"coin-server/common/mysqlclient"
	"coin-server/common/pprof"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/tracing"
	"coin-server/common/ulimit"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	_ "coin-server/guild-gvg-server/env"
	"coin-server/guild-gvg-server/service"
	"coin-server/rule"

	"go.uber.org/zap"
)

func main() {
	if err := ulimit.SetRLimit(); err != nil {
		panic(err)
	}
	serverId := env.GetServerId()
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_GVGGuildServer.String(), serverId)},
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_GVGGuildServer,
			"serverId":   serverId,
		},
		Development: true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	tracing.Init(cnf, serverId, models.ServerType_GVGGuildServer)
	redisclient.Init(cnf)
	rule.Load(cnf)
	im.Init(cnf)
	mysqlCnf := &mysqlclient.Config{}
	utils.Must(cnf.Unmarshal("gvg/mysql", mysqlCnf))
	mysqlCnf.Database = ""
	mysql, err := mysqlclient.NewClient(mysqlCnf, log)
	utils.Must(err)

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	s := service.NewService(mysql, urls, log, serverId, models.ServerType_GVGGuildServer, true, true)
	go func() {
		s.Serve()
	}()

	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Close()
	})
}
