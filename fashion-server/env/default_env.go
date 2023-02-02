package env

import (
	"os"
	"strings"

	"coin-server/common/values/env"
)

var setDefaultEnvErr = setDefaultEnv()

func init() {
	if setDefaultEnvErr != nil {
		panic(setDefaultEnvErr)
	}
}

func setDefaultEnv() error {
	for k, v := range map[string]string{
		env.CONF_PATH:        "newhttp/",
		env.CONF_FORMAT:      "json",
		env.APP_NAME:         "racingrank-server",
		env.APP_MODE:         "DEBUG",
		env.PPROF_ADDR:       "0.0.0.0:6160",
		env.ERROR_CODE_STACK: "1",
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

func IsDebug() bool {
	mode := os.Getenv(env.APP_MODE)
	return mode == "DEBUG"
}
