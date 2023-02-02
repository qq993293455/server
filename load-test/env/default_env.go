package env

import (
	"os"
	"strconv"
	"strings"

	"coin-server/common/values"
	"coin-server/common/values/env"
)

var DefaultEnv = map[string]string{
	env.CONF_PATH:                    "newhttp/",
	env.CONF_FORMAT:                  "json",
	env.APP_NAME:                     "battle-server",
	env.APP_MODE:                     "RELEASE",
	env.SERVER_ID:                    "11",
	env.PPROF_ADDR:                   "0.0.0.0:6066",
	env.ERROR_CODE_STACK:             "1",
	env.LOCUST_MASTER_HOST:           "127.0.0.1",
	env.LOCUST_MASTER_PORT:           "5557",
	env.LOCUST_USE_SAVE_USER:         "false",
	env.LOCUST_TARGET_NET_TYPE:       "tcp",
	env.LOCUST_TARGET_SERVER_ADDR:    "127.0.0.1:8071",
	env.LOCUST_TARGET_SERVER_ID:      "1",
	env.LOCUST_TARGET_LESS_SERVER_ID: "1",
}

var _ = SetDefaultEnv(DefaultEnv)

// SetDefaultEnv 设置默认的环境变量
// defaultEnv的键为环境变量的key，值为它对应的默认值
func SetDefaultEnv(defaultEnv map[string]string) bool {
	for k, v := range defaultEnv {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if strings.TrimSpace(os.Getenv(k)) == "" {
			err := os.Setenv(k, v)
			if err != nil {
				panic(err)
			}
		}
	}
	env.SetDefaultEnv()
	env.SetRuleEnv()
	return true
}

func UseSaveUser() bool {
	str := os.Getenv(env.LOCUST_USE_SAVE_USER)
	if strings.ToLower(str) == "true" {
		return true
	}
	return false
}

func GetTargetServerId() values.ServerId {
	str := os.Getenv(env.LOCUST_TARGET_SERVER_ID)
	if str == "" {
		return 0
	}
	v, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return values.ServerId(v)
}

func GetTargetLessServerId() values.ServerId {
	str := os.Getenv(env.LOCUST_TARGET_LESS_SERVER_ID)
	if str == "" {
		return 0
	}
	v, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return values.ServerId(v)
}
