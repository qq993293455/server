package main

import (
	"fmt"
	"net/netip"
	"os"

	"coin-server/common/idgenerate"
	"coin-server/common/redisclient"
	"coin-server/new-center-server/service/edge"
	"coin-server/rule"

	"coin-server/common/consulkv"
	"coin-server/common/core"
	"coin-server/common/logger"
	"coin-server/common/pprof"
	"coin-server/common/proto/models"
	self_ip "coin-server/common/self-ip"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	_ "coin-server/new-center-server/env"
	"coin-server/new-center-server/service"
	"coin-server/new-center-server/web"

	"go.uber.org/zap"
)

func main() {
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_NewCenterServer.String(), env.GetServerId())},
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_NewCenterServer,
			"serverId":   env.GetServerId(),
		},
		Development: true,
		Discard:     false,
	})
	go web.StartWeb(log)
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)
	self_ip.InitLan(cnf)
	redisclient.Init(cnf)
	idgenerate.Init(redisclient.GetCommonRedis())
	rule.Load(cnf)
	edge.InitLineConfig(cnf, log)
	ip, err := netip.ParseAddr(self_ip.SelfIpLan)
	if err != nil {
		panic(err)
	}
	addr := netip.AddrPortFrom(ip, 34567)
	s := service.NewCenterService(log, addr.String(), urls, env.GetServerId())
	err = s.Run()
	if err != nil {
		panic(err)
	}
	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		s.Close()
	})
}
