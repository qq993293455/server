package env

import (
	"os"
	"strings"

	"coin-server/common/values/env"
)

// 为了让环境变量最先执行，所以采用全局函数的方式
var setDefaultEnvErr = setDefaultEnv()

func init() {
	if setDefaultEnvErr != nil {
		panic(setDefaultEnvErr)
	}
}

var (
	XLSX_DIR             = "XLSX_DIR"
	TABLE_BIN_PATHS      = "TABLE_BIN_PATHS"
	PROTO_DIR            = "PROTO_DIR"
	SCHEME_PATH          = "SCHEME_PATH"
	CPP_CONFIG_CODE_PATH = "CPP_CONFIG_CODE_PATH"
)

func setDefaultEnv() error {

	for k, v := range map[string]string{
		env.CONF_PATH:        "newhttp/",
		env.CONF_FORMAT:      "json",
		env.APP_NAME:         "excel-to-bin",
		env.PPROF_ADDR:       "0.0.0.0:6665",
		env.ERROR_CODE_STACK: "1",
		XLSX_DIR:             "../share/excel/xlsx",
		SCHEME_PATH:          "./excel-to-bin/scheme.json",
		PROTO_DIR:            "../share/proto/proto/configcpp",
		TABLE_BIN_PATHS:      "../battle/res/table.bin.txt;../client/Assets/AssetsPackage/Remoted/Dynamic/CppTableBin/table.bin.txt",
		CPP_CONFIG_CODE_PATH: "../battle/src_new/battle/define/config.stream.hpp",
	} {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if strings.TrimSpace(os.Getenv(k)) == "" {
			err := os.Setenv(k, v)
			if err != nil {
				return err
			}
		}
	}
	env.SetDefaultEnv()
	env.SetRuleEnv()
	return nil
}
