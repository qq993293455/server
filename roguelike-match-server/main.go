package main

import (
	"fmt"
	"os"

	"coin-server/common/idgenerate"
	"coin-server/common/im"
	self_ip "coin-server/common/self-ip"
	_ "coin-server/roguelike-match-server/env"
	"coin-server/roguelike-match-server/service"
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
	serverId := env.GetRoguelikeServerId()
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_RoguelikeMatchServer.String(), serverId)},
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_RoguelikeMatchServer,
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
	im.Init(cnf)
	idgenerate.Init(redisclient.GetCommonRedis())
	self_ip.InitLan(cnf)

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	s := service.NewService(urls, log, serverId, models.ServerType_RoguelikeMatchServer)
	go func() {
		s.Serve()
	}()

	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Stop()
	})
}
