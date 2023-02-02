package main

import (
	"fmt"
	"os"

	self_ip "coin-server/common/self-ip"
	"coin-server/common/values"
	_ "coin-server/gen-rank-server/env"
	"coin-server/gen-rank-server/service"
	"coin-server/rule"

	"coin-server/common/consulkv"
	"coin-server/common/core"
	"coin-server/common/logger"
	"coin-server/common/pprof"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values/env"

	"go.uber.org/zap"
)

func main() {
	serverId := values.ServerId(0)
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_GenRankServer.String(), serverId)},
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_GenRankServer,
			"serverId":   serverId,
		},
		Development: true,
		Discard:     false,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	rule.Load(cnf)
	redisclient.Init(cnf)
	self_ip.InitLan(cnf)

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	s := service.NewService(urls, log, serverId, models.ServerType_GenRankServer)
	go func() {
		s.Serve()
	}()

	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Stop()
	})
}
