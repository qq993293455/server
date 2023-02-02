package consul

import (
	"os"
	"strings"

	"coin-server/common/values/env"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

//支持多个文件的配置
var ConfigCenter *configCenter

type configCenter struct {
	vs map[string]*viper.Viper
}

func InitConfig() {
	if ConfigCenter != nil {
		return
	}
	config := &configCenter{
		vs: make(map[string]*viper.Viper, 0),
	}
	ConfigCenter = config
	return
}

func (cfg *configCenter) RegisterConfig(name string, path string) {
	v := viper.New()
	cfg.vs[name] = v
	providers := strings.Split(os.Getenv(env.CONF_HOSTS), ",")
	var err error
	for _, provider := range providers {
		err = v.AddRemoteProvider(os.Getenv(env.CONF_TYPE), provider, os.Getenv(env.CONF_PATH)+path)
		if err != nil {
			continue
		}
	}
	v.SetConfigType(os.Getenv(env.CONF_FORMAT))
	err = v.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
	cfg.startWatch(v)
}

func (cfg *configCenter) GetConfig(name string) *viper.Viper {
	config, ok := cfg.vs[name]
	if !ok {
		err := errors.New("ConfigCenter: no exist config")
		panic(err)
	}
	return config
}

func (cfg *configCenter) GetConfigWithType(name string, typ string) *viper.Viper {
	config, ok := cfg.vs[name]
	if !ok {
		err := errors.New("ConfigCenter: no exist config")
		panic(err)
	}
	if typ == "" {
		typ = os.Getenv(env.CONF_FORMAT)
	}
	config.SetConfigType(typ)
	return config
}

func (cfg *configCenter) startWatch(v *viper.Viper) {
	err := v.WatchRemoteConfigOnChannel()
	if err != nil {
		panic(err)
	}
}
