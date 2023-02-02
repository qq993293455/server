package main

import (
	"strings"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/utils"
	"coin-server/common/values/env"
	"coin-server/excel-to-bin/cpb"
	env2 "coin-server/excel-to-bin/env"
	"coin-server/excel-to-bin/parse"
	"coin-server/excel-to-bin/tobin"
	"coin-server/rule"

	"go.uber.org/zap"
)

func main() {
	log := logger.MustNew(zap.DebugLevel, &logger.Options{
		Console: "stdout",
		// FilePath:   []string{fmt.Sprintf("./%s.log", models.ServerType_GameServer.String())},
		RemoteAddr: nil,
		InitFields: map[string]interface{}{
			"serverType": "excel-to-bin",
			"serverId":   0,
		},
		Development: true,
		Discard:     false,
	})
	logger.SetDefaultLogger(log)
	parse.ReadJsonConfig(env.GetString(env2.SCHEME_PATH))
	tisMap := parse.GetAllExcelField(env.GetString(env2.XLSX_DIR))
	fb := cpb.CreateProtoBuffer(tisMap, func(s string) string {
		return s
	})

	cpb.Write(fb, env.GetString(env2.PROTO_DIR))
	fb1 := cpb.CreateProtoBuffer(tisMap, func(s string) string {
		return utils.ToSnake(s)
	})
	cnf, err := consulkv.NewConfig(env.GetString(env.CONF_PATH), env.GetString(env.CONF_HOSTS), log)
	utils.Must(err)
	rule.Load(cnf)

	r := utils.GetRuleName()
	logger.DefaultLogger.Info("loading latest rule", zap.String("branch", r))
	req := rule.NewRequest()
	if err := req.Get(r, ""); err != nil {
		panic(err)
	}
	binPathS := env.GetString(env2.TABLE_BIN_PATHS)
	pathS := strings.Split(binPathS, ";")
	tobin.WriteData(fb1, req.Result.Rule, tisMap, pathS)
}
