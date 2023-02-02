package main

import (
	"os"
	"strings"

	"coin-server/common/consulkv"
	"coin-server/common/core"
	"coin-server/common/logger"
	"coin-server/common/metrics"
	"coin-server/common/pprof"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	self_ip "coin-server/common/self-ip"
	"coin-server/common/serverids"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	_ "coin-server/gateway-tcp/env"
	"coin-server/gateway-tcp/service"

	"go.uber.org/zap"
)

func main() {
	serverId := env.GetServerId()
	tcpListen := env.GetTCPAddr()
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console: "stdout",
		// FilePath:   []string{fmt.Sprintf("./%s.log", models.ServerType_GatewayStdTcp.String())},
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_GatewayStdTcp,
			"serverId":   serverId,
		},
		Development: true,
		// Discard:     true,
	})
	logger.SetDefaultLogger(log)
	log.Info("TCPADDR", zap.String("addr", tcpListen))
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	redisclient.Init(cnf)
	self_ip.Init(cnf)
	serverids.Init(cnf)

	// 监控
	metrics.Init(strings.ToLower(models.ServerType_GatewayStdTcp.String()), env.GetServerId())
	metrics.Start(log, os.Getenv(env.METRICS_ADDR))

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)
	s := service.NewService(serverId, tcpListen, urls, log, true)
	s.Run()

	pprof.Start(log, env.GetString(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Close()
	})
}
