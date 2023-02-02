package main

import (
	"fmt"
	"os"

	"coin-server/common/consulkv"
	"coin-server/common/core"
	"coin-server/common/logger"
	"coin-server/common/pprof"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/statistical2"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/env"
	_ "coin-server/pay-server/env"
	"coin-server/pay-server/service"
	"go.uber.org/zap"
)

func main() {
	serverId := values.ServerId(0)
	serverType := models.ServerType_PayServer
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_PayServer.String(), serverId)},
		InitFields: map[string]interface{}{
			"serverType": serverType,
			"serverId":   serverId,
		},
		Development: true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	redisclient.Init(cnf)
	statistical2.Init(cnf)

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	s := service.NewService(urls, log, serverId, serverType)
	go func() {
		s.Serve()
	}()

	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Stop()
	})
}
