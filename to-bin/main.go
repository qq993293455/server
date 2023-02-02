package main

import (
	_ "coin-server/to-bin/env"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	"coin-server/rule"
	"coin-server/to-bin/gen"
	"go.uber.org/zap"
)

func main() {
	log := logger.MustNew(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": -1,
			"serverId":   0,
		},
		Development: true,
		Discard:     false,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	rule.Load(cnf)
	gen.LoadRule()
}
