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

func setDefaultEnv() error {

	for k, v := range map[string]string{
		env.CONF_PATH:   "newhttp/",
		env.CONF_FORMAT: "json",
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
