package main

import (
	"context"
	"os"

	"coin-server/common/consulkv"
	"coin-server/common/core"
	"coin-server/common/iggsdk"
	"coin-server/common/logger"
	"coin-server/common/pprof"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	self_ip "coin-server/common/self-ip"
	"coin-server/common/ulimit"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	_ "coin-server/edge-server/env"
	ipdomain "coin-server/edge-server/ip-domain"
	"coin-server/edge-server/service"

	"go.uber.org/zap"
)

func main() {

	err := ulimit.SetOpenCoreDump()
	if err != nil {
		panic(err)
	}
	lo := &logger.Options{
		Console: "stdout",
		// FilePath:   []string{fmt.Sprintf("./%s.log", models.ServerType_GameServer.String())},
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_EdgeServer,
		},
		Development: true,
		Discard:     false,
	}

	log := logger.MustNewAsync(zap.DebugLevel, lo)
	logger.SetDefaultLogger(log)

	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	redisclient.Init(cnf)
	id, err := redisclient.GetLockerRedis().IncrBy(context.Background(), "__EdgeServerID__", 10).Result()
	if err != nil {
		panic(err)
	}
	log = log.WithFields(zap.Int64("serverId", id))
	logger.SetDefaultLogger(log)

	iggsdk.Init("0.0.1", log, cnf)
	utils.Must(err)
	self_ip.Init(cnf)
	ipdomain.Init(cnf)

	s := service.NewEdgeService(log, urls, os.Getenv(env.BATTLE_SERVER_PATH), id)
	err = s.Run()
	if err != nil {
		panic(err)
	}

	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Close()
	})

}
