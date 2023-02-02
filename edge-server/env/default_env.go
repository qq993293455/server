package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"coin-server/common/values"
	"coin-server/common/values/env"
)

const (
	EDGE_TYPE   = "EDGE_TYPE"   // 0 等待center调度（默认）  1 是挂机，2 是其他
	WORK_DIR    = "WORK_DIR"    // 工作目录
	CENTER_ADDR = "CENTER_ADDR" // 中心服地址

	EDGE_FIXED_PORT   = "EDGE_FIXED_PORT"   // 是否是固定端口
	EDGE_FIXED_PORTS  = "EDGE_FIXED_PORTS"  // 被指定的固定端口
	EDGE_PORTS_START  = "EDGE_PORTS_START"  // 端口开始
	EDGE_PORTS_END    = "EDGE_PORTS_END"    // 端口结束
	EDGE_TOTAL_WEIGHT = "EDGE_TOTAL_WEIGHT" // 手动指定EDGE的总权重
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
		env.CONF_PATH:          "newhttp/",
		env.CONF_FORMAT:        "json",
		env.APP_NAME:           "edge-server",
		env.APP_MODE:           "RELEASE",
		env.SERVER_ID:          "0",
		env.PPROF_ADDR:         "0.0.0.0:7101",
		env.ERROR_CODE_STACK:   "1",
		env.BATTLE_SERVER_PATH: "./vsbuild/battle/x64/Debug/battle_visual_new.exe",
		EDGE_TYPE:              "0",
		WORK_DIR:               "../battle",
		CENTER_ADDR:            "",
		EDGE_FIXED_PORT:        "0",
		EDGE_PORTS_START:       "31000",
		EDGE_PORTS_END:         "60000",
		EDGE_FIXED_PORTS:       "",
		EDGE_TOTAL_WEIGHT:      "0",
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

func GetInt64(v string) int64 {
	is := strings.TrimSpace(os.Getenv(v))
	if is == "" {
		panic(fmt.Sprintf("os env '%s' is empty", v))
	}
	i, err := strconv.Atoi(is)
	if err != nil {
		panic(fmt.Sprintf("os env '%s' atoi error:%s", v, err.Error()))
	}
	return int64(i)
}

func GetUint16(v string) uint16 {
	return uint16(GetInt64(v))
}

func CenterServerId() values.ServerId {
	return env.GetCenterServerId()
}
