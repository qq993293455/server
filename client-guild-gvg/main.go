package main

import (
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"coin-server/client-guild-gvg/db"
	"coin-server/client-guild-gvg/env"
	"coin-server/client-guild-gvg/player"
	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/utils"
	env2 "coin-server/common/values/env"
	"coin-server/rule"

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
	cnf, err := consulkv.NewConfig(env2.GetString(env2.CONF_PATH), env2.GetString(env2.CONF_HOSTS), log)
	utils.Must(err)
	rule.Load(cnf)
	// UserIdIndex = 9999999 // 打开这一行可以让UserId 固定。注释的话，每次都是全新的UserId
	useOldData := env2.GetInteger(env.USE_OLD_DATA)
	dataPath := "./client-guild-gvg.db"
	var sd *db.DB
	if useOldData == 0 {
		if utils.CheckPathExists(dataPath) {
			err := os.Remove(dataPath)
			if err != nil {
				panic(err)
			}
		}
	}
	sd = db.NewDB(log, dataPath)

	sd.Run()

	var clients []*player.Client
	for i := 0; i < 1; i++ {
		c := player.NewClient(NewUserId(), log, 70)
		clients = append(clients, c)
		c.Connect("10.23.50.229:8071", log)
	}

	//var success, failed int64
	//
	//for _, v := range clients {
	//	e := v.WaitOver()
	//	if e == nil {
	//		success++
	//	} else {
	//		failed++
	//	}
	//}
	//log.Info("测试结束", zap.Int64("success", success), zap.Int64("failed", failed))
	log.Sync()
}
