package main

import (
	"fmt"

	"coin-server/common/consulkv"
	"coin-server/common/idgenerate"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/ulimit"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	"coin-server/load-test/core"
	"coin-server/load-test/module"

	"go.uber.org/zap"
)

func main() {
	err := ulimit.SetRLimit()
	utils.Must(err)
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		FilePath:   []string{fmt.Sprintf("./%s.log", models.ServerType_LoadTest.String())},
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_LoadTest,
			"serverId":   env.GetServerId(),
		},
		Development: true,
		Discard:     true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	redisclient.Init(cnf)
	idgenerate.Init(redisclient.GetCommonRedis())
	module.Init()
	core.RunBoomer(env.GetString(env.LOCUST_MASTER_HOST), int(env.GetInteger(env.LOCUST_MASTER_PORT)), module.GetModuleTasks())
}
