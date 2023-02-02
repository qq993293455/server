package main

import (
	"flag"

	_ "coin-server/visual-battle-client/env"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/utils"
	"coin-server/common/values/env"

	"coin-server/rule"

	"go.uber.org/zap"
)

var addr = flag.String("addr", "127.0.0.1:9876", "new battle service visual address")
var mapId = flag.String("map_id", "5001", "new battle service enter map id")
var scale = flag.Float64("scale", 20, "显示缩放倍数")

func main() {
	flag.Parse()
	log := logger.MustNew(zap.DebugLevel, &logger.Options{
		Console:    "stdout",
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": "VisualClient",
		},
		Development: true,
	})
	logger.SetDefaultLogger(log)
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	rule.Load(cnf)
	v := NewVisual(*addr, log)
	v.startRender(*mapId, *scale)
}
