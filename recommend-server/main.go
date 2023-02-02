package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	self_ip "coin-server/common/self-ip"
	_ "coin-server/recommend-server/env"
	"coin-server/recommend-server/service"

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
	serverId := env.GetServerId()
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_RecommendServer.String(), serverId)},
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_RecommendServer,
			"serverId":   serverId,
		},
		Development: true,
		Discard:     false,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)

	redisclient.Init(cnf)
	self_ip.InitLan(cnf)

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	rand.Seed(time.Now().Unix())
	s := service.NewService(urls, log, serverId, models.ServerType_RecommendServer)
	go func() {
		s.Serve()
	}()

	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Stop()
	})
}
