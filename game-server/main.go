package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"coin-server/common/ArenaRule"
	"coin-server/common/consulkv"
	"coin-server/common/core"
	"coin-server/common/idgenerate"
	"coin-server/common/iggsdk"
	"coin-server/common/im"
	"coin-server/common/logger"
	_ "coin-server/common/map-data"
	"coin-server/common/metrics"
	"coin-server/common/orm"
	"coin-server/common/pprof"
	"coin-server/common/proto/models"
	"coin-server/common/racingrank"
	"coin-server/common/redisclient"
	self_ip "coin-server/common/self-ip"
	"coin-server/common/sensitive"
	"coin-server/common/statistical"
	"coin-server/common/statistical2"
	"coin-server/common/syncrole"
	tendissync "coin-server/common/tendis-sync"
	"coin-server/common/tracing"
	"coin-server/common/ulimit"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	_ "coin-server/game-server/env"
	"coin-server/game-server/service"
	"coin-server/rule"

	"go.uber.org/zap"
)

func main() {
	if err := ulimit.SetRLimit(); err != nil {
		panic(err)
	}
	serverId := env.GetServerId()

	log := logger.MustNewAsync(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		FilePath:   []string{fmt.Sprintf("./%s.%d.log", models.ServerType_GameServer.String(), serverId)},
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": models.ServerType_GameServer,
			"serverId":   serverId,
		},
		Development: true,
		// Discard:     true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	tracing.Init(cnf, serverId, models.ServerType_GameServer)
	tendissync.Init(cnf, serverId)
	redisclient.Init(cnf)
	orm.InitMySQL(cnf)
	idgenerate.Init(redisclient.GetCommonRedis())
	im.Init(cnf)
	statistical.Init(cnf)
	statistical2.Init(cnf)
	rule.Load(cnf)
	self_ip.InitLan(cnf)
	syncrole.Init(cnf)   // 同步角色信息
	sensitive.Init(cnf)  // 敏感词过滤
	racingrank.Init(cnf) // 竞速赛匹配
	ArenaRule.Init(cnf)
	iggsdk.Init("0.0.1", log, cnf)

	rand.Seed(time.Now().UnixNano() + serverId)
	var urls []string
	err = cnf.Unmarshal("nats-cluster", &urls)
	utils.Must(err)

	s := service.NewService(urls, log, serverId, models.ServerType_GameServer)
	s.Serve()

	pprof.Start(log, os.Getenv(env.PPROF_ADDR))
	metrics.Init(strings.ToLower(models.ServerType_GameServer.String()), env.GetServerId())
	metrics.Start(log, os.Getenv(env.METRICS_ADDR))

	core.WaitClose(func() {
		s.Stop()
		syncrole.Close()
		racingrank.Close()
	})
}
