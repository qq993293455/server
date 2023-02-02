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
	self_ip "coin-server/common/self-ip"
	"coin-server/common/tracing"
	"coin-server/common/ulimit"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/env"
	_ "coin-server/rank-server/env"
	"coin-server/rank-server/service"
	"coin-server/rank-server/service/rank/dao"

	"go.uber.org/zap"
)

func main() {
	if err := ulimit.SetRLimit(); err != nil {
		panic(err)
	}
	serverId := values.ServerId(0)
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_RankServer.String(), serverId)},
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_RankServer,
			"serverId":   serverId,
		},
		Development: true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	tracing.Init(cnf, serverId, models.ServerType_RankServer)
	redisclient.Init(cnf)
	self_ip.InitLan(cnf)
	dao.Init(cnf)

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	s := service.NewService(urls, log, serverId, models.ServerType_RankServer)
	go func() {
		s.Serve()
	}()

	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Stop()
	})
}
