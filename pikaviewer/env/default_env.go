package env

import (
	"os"

	"coin-server/common/values/env"
)

const (
	PIKA_VIEWER_ADDR   = "PIKA_VIEWER_ADDR"
	PAY_SERVER_ADDR    = "PAY_SERVER_ADDR"
	CLIENT_STATIC_FILE = "CLIENT_STATIC_FILE"
	BATTLE_LOG_DIR     = "BATTLE_LOG_DIR"
	SO_BUILD_DIR_DEV   = "SO_BUILD_DIR_DEV"
	SO_BUILD_DIR_PATCH = "SO_BUILD_DIR_PATCH"
	BETA_ADDR          = "BETA_ADDR"
)

var DefaultEnv = map[string]string{
	env.CONF_PATH:      "newhttp/",
	env.APP_NAME:       "请设置名称",
	PIKA_VIEWER_ADDR:   ":9991",
	PAY_SERVER_ADDR:    ":9992",
	CLIENT_STATIC_FILE: "./pikaviewer/static",
	BATTLE_LOG_DIR:     "/root/battle/log",                            // 53服务器上的路径
	SO_BUILD_DIR_DEV:   "/root/GitRepoes/AndroidPkg/battle",           // 53服务器上dev分支路径
	SO_BUILD_DIR_PATCH: "/root/GitRepoes/AndroidPkgPatch/battle",      // 53服务器上dev分支路径
	BETA_ADDR:          "https://13.251.152.46:30004/v1/version/info", // 灰度服获取客户端版本信息地址
}

func init() {
	SetDefaultEnv(DefaultEnv)
}

func SetDefaultEnv(defaultEnv map[string]string) {
	for k, v := range defaultEnv {
		if os.Getenv(k) == "" {
			err := os.Setenv(k, v)
			if err != nil {
				panic(err)
			}
		}
	}
	env.SetDefaultEnv()
	env.SetRuleEnv()
}
