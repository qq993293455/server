package main

import (
	"fmt"
	"os"
	"strings"

	_ "coin-server/arena-server/env"
	"coin-server/arena-server/service"
	"coin-server/common/ArenaRule"
	"coin-server/common/consulkv"
	"coin-server/common/core"
	"coin-server/common/ctx"
	"coin-server/common/logger"
	"coin-server/common/metrics"
	"coin-server/common/pprof"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	self_ip "coin-server/common/self-ip"
	"coin-server/common/tracing"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	"coin-server/rule"

	"go.uber.org/zap"
)

func main() {
	serverId := env.GetArenaServerId()
	if serverId == 0 {
		serverId = int64(models.ArenaType_ArenaType_Default)
	}
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ArenaType_ArenaType_Default.String(), serverId)},
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_ArenaServer,
			"serverId":   serverId,
		},
		Development: true,
		//Discard:     true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	tracing.Init(cnf, serverId, models.ServerType_ArenaServer)
	redisclient.Init(cnf)
	self_ip.InitLan(cnf)
	//dao.Init(cnf)
	rule.Load(cnf)
	ArenaRule.Init(cnf)

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	s := service.NewService(urls, log, serverId, models.ServerType_ArenaServer)
	s.Serve()

	pprof.Start(log, ArenaRule.GetPprofAddr())
	metrics.Init(strings.ToLower(models.ServerType_ArenaServer.String()), serverId)
	metrics.Start(log, os.Getenv(env.METRICS_ADDR))

	core.WaitClose(func() {
		ctxx := ctx.GetContext()
		s.Stop(ctxx)
		ctxx.NewOrm().Do()
	})
}
