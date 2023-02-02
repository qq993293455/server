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
	self_ip "coin-server/common/self-ip"
	_ "coin-server/common/statistical2"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	"coin-server/statistical-server/config"
	"coin-server/statistical-server/consumer"
	"coin-server/statistical-server/database"
	_ "coin-server/statistical-server/env"
	"coin-server/statistical-server/kafka"

	"go.uber.org/zap"
)

func main() {
	name := "StatisticalServer"
	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", name, env.GetServerId())},
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": name,
			"serverId":   -1,
		},
		Development: true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	config.Init(cnf)
	kafka.Init()
	kafka.InitLogin()
	database.InitMysql()
	//database.InitClickHouse()
	database.InitSQL()
	self_ip.InitLan(cnf)

	consumer.Start()

	log.Info("start statistical server success")
	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := consumer.Close(ctx); err != nil {
			log.Error("close consumer error", zap.Error(err))
		} else {
			log.Info("close consumer success")
		}
	})
}
