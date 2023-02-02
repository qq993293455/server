package main

import (
	"flag"
	"fmt"
	"os"

	"coin-server/common/core"
	"coin-server/common/pprof"
	self_ip "coin-server/common/self-ip"
	_ "coin-server/role-state-server/env"
	"coin-server/role-state-server/service"
	"coin-server/role-state-server/service/state/raft"
	"coin-server/rule"

	"go.uber.org/zap"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/env"
)

var (
	clusterServerId = flag.String("s", "0", "serverId")
	tcpAddr         = flag.String("t", "127.0.0.1:9601", "tcpAddr")
	bootstrap       = flag.Bool("b", false, "bootstrap")
	join            = flag.Bool("j", false, "join")
	dbPath          = flag.String("d", "../db", "dbPath")
)

func main() {
	serverId := values.ServerId(0)
	serverType := models.ServerType_RoleStateServer
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_RoleStateServer.String(), serverId)},
		InitFields: map[string]interface{}{
			"serverType": serverType,
			"serverId":   serverId,
		},
		Development: true,
		Discard:     false,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	rule.Load(cnf)
	self_ip.InitLan(cnf)

	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	flag.Parse()
	s := service.NewService(urls, log, serverId, serverType, &raft.Options{
		ServerId:  *clusterServerId,
		TcpAddr:   *tcpAddr,
		Bootstrap: *bootstrap,
		Join:      *join,
		DBPath:    *dbPath,
	})
	go func() {
		s.Serve()
	}()

	pprof.Start(log, os.Getenv(env.PPROF_ADDR)+*clusterServerId)
	core.WaitClose(func() {
		s.Stop()
	})
}
