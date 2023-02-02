package main

import (
	"context"
	"fmt"
	"os"
	"time"

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
	"coin-server/racingrank-server/dao"
	_ "coin-server/racingrank-server/env"
	"coin-server/racingrank-server/service"
	match2 "coin-server/racingrank-server/service/match"

	"go.uber.org/zap"
)

func main() {
	if err := ulimit.SetRLimit(); err != nil {
		panic(err)
	}
	serverId := values.ServerId(0)
	serverType := models.ServerType_RacingRankServer
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_RacingRankServer.String(), serverId)},
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

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	dao.Init(cnf)
	match2.Init(cnf, urls, serverId, serverType, log)
	match2.InitOwnerCron(log)
	match2.Start()

	s := service.NewService(urls, log, serverId, serverType)
	go func() {
		s.Serve()
	}()

	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		match2.Close(ctx, log)
	})
}
