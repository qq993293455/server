package global

import (
	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/utils"
	"go.uber.org/zap"

	"stathat.com/c/consistent"
)

var (
	Config            *ConfigMain
	GatewayConsistent *consistent.Consistent
	GMAPIWhiteList    []string // 后台web页面白名单
)

type ConfigMain struct {
	*ApiConfig
	GateWay map[string]int64
}

type ApiConfig struct {
	EnableDebug bool     `json:"enable_debug,omitempty"`
	EnableSign  bool     `json:"enable_sign,omitempty"`
	SignKey     string   `json:"sign_key,omitempty"`
	IPWhiteList []string `json:"ip_white_list,omitempty"`
}

func Init(consul *consulkv.Config) {
	api := &ApiConfig{}
	gateway := make(map[string]int64)
	utils.Must(consul.Unmarshal("gm/api", api))
	utils.Must(consul.Unmarshal("gateway", &gateway))
	Config = &ConfigMain{
		ApiConfig: api,
		GateWay:   gateway,
	}

	c := consistent.New()
	for item := range gateway {
		c.Add(item)
	}
	GatewayConsistent = c

	var wl []string
	if err := consul.Unmarshal("gm/white_list", &wl); err != nil {
		logger.DefaultLogger.Warn("unmarshal consul gm/white_list err", zap.Error(err))
	}
	GMAPIWhiteList = wl
}
