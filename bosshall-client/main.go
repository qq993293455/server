package main

import (
	"strconv"
	"sync/atomic"
	"time"

	"coin-server/bosshall-client/client"
	_ "coin-server/bosshall-client/env"
	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	"coin-server/rule"

	"go.uber.org/zap"

	"go.uber.org/zap/zapcore"
)

var UserIdIndex = time.Now().UnixNano()

func NewUserId() string {
	uid := atomic.AddInt64(&UserIdIndex, 1)
	return strconv.Itoa(int(uid))
}

func main() {

	log := logger.MustNew(zapcore.InfoLevel, &logger.Options{
		Console:     "stdout",
		FilePath:    nil,
		RemoteAddr:  nil,
		InitFields:  nil,
		Development: true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	rule.Load(cnf)
	// UserIdIndex = 9999999 // 打开这一行可以让UserId 固定。注释的话，每次都是全新的UserId
	var clients []*client.Client
	for i := 0; i < 200; i++ {
		c := client.NewClient(NewUserId(), log, 506001, 70)
		clients = append(clients, c)
		c.Connect("10.23.50.229:8071", log)
	}

	var success, failed int64

	for _, v := range clients {
		e := v.WaitOver()
		if e == nil {
			success++
		} else {
			failed++
		}
	}
	log.Info("压测结束", zap.Int64("success", success), zap.Int64("failed", failed))
	log.Sync()
}
