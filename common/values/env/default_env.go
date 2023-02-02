package env

import (
	"os"

	"coin-server/common/ulimit"
)

var defaultEnv = map[string]string{
	CONF_TYPE:            "consul",
	CONF_HOSTS:           "10.23.20.53:8500,10.23.20.53:8501,10.23.20.53:8502",
	CONF_PATH:            "config/default/",
	CONF_FORMAT:          "yaml",
	HTTP_ADDR:            ":8070",
	TCP_ADDR:             ":8071",
	LOG_STDOUT:           "0",
	LOG_LEVEL:            "DEBUG",
	LOG_MAX_SIZE:         "200",
	LOG_MAX_BACKUP:       "20",
	LOG_UTC_TIME:         "1",
	NONCORE_LOG:          "1",
	FORMAT_LOG_OPEN:      "1",
	FORMAT_LOG_STDOUT:    "1",
	FORMAT_LOG_FILE:      "FormatLog",
	TIME_ZONE:            "8",
	APP_NAME:             "server-demo",
	APP_MODE:             "RELEASE",
	GAME_ID:              "120",
	SERVER_ID:            "1",
	OPEN_TRACEING:        "0",
	CENTER_SERVER_ID:     "1",
	OPEN_GM_HANDLER:      "0",
	ROGUE_LIKE_SERVER_ID: "1",
	OPEN_MIDDLE_ERROR:    "1",
	GUILD_GVG_SERVER_ID:  "1",
}

func init() {
	if err := ulimit.SetRLimit(); err != nil {
		panic(err)
	}
}

// SetDefaultEnv 设置默认的环境变量
// defaultEnv的键为环境变量的key，值为它对应的默认值
func SetDefaultEnv() {
	for k, v := range defaultEnv {
		if os.Getenv(k) == "" {
			err := os.Setenv(k, v)
			if err != nil {
				panic(err)
			}
		}
	}
}
