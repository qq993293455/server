package main

import (
	"fmt"
	"os"

	_ "coin-server/activity-server/env"
	"coin-server/activity-server/service"
	"coin-server/activity-server/service/activity"
	"coin-server/common/consulkv"
	"coin-server/common/core"
	"coin-server/common/logger"
	"coin-server/common/pprof"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	self_ip "coin-server/common/self-ip"
	"coin-server/common/tracing"
	"coin-server/common/ulimit"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/env"
	"coin-server/rule"

	"go.uber.org/zap"
)

func main() {
	if err := ulimit.SetRLimit(); err != nil {
		panic(err)
	}
	serverId := values.ServerId(0)
	serverType := models.ServerType_ActivityServer
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_ActivityServer.String(), serverId)},
		InitFields: map[string]interface{}{
			"serverType": serverType,
			"serverId":   serverId,
		},
		Development: true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	tracing.Init(cnf, serverId, serverType)
	redisclient.Init(cnf)
	self_ip.InitLan(cnf)
	rule.Load(cnf)

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	// activity.Init(cnf, urls, serverId, serverType, log)
	activity.InitOwnerCron(log)
	// activity.Start()

	s := service.NewService(urls, log, serverId, serverType)
	go func() {
		s.Serve()
	}()

	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Stop()
	})
}
