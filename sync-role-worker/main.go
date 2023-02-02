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
	_ "coin-server/common/statistical"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	"coin-server/sync-role-worker/dao"
	_ "coin-server/sync-role-worker/env"
	"coin-server/sync-role-worker/worker"

	"go.uber.org/zap"
)

func main() {
	name := "SyncRoleWorker"
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
	log.Info("CONF_HOSTS: " + env.GetString(env.CONF_HOSTS))
	dao.Init(cnf)
	worker.Init(cnf)

	worker.Start()

	log.Info("start sync-role-worker success")
	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	core.WaitClose(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := worker.Close(ctx); err != nil {
			log.Error("close sync-role-worker error", zap.Error(err))
		} else {
			log.Info("close sync-role-worker success")
		}
	})
}
