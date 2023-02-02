package iggreq

import (
	"fmt"
	"runtime"
)

// IAPIConfig 配置接口
type IAPIConfig interface {
	GatewayConfig() GatewayConfiguration
	ConfigGateway(token, secret string)
	ConfigProduct(name, version string)
	ConfigSDKVersion(version string)
}

type apiConfig struct {
	gatewayToken   string
	gatewaySecret  string
	productName    string
	productVersion string
	sdkVersion     string
}

var (
	gConfig apiConfig
)

func init() {
	gConfig = apiConfig{
		productName:    "GameServer",
		productVersion: "unknown",
	}
}

func ConfigInstance() IAPIConfig {
	return &gConfig
}

func (cfg *apiConfig) GatewayConfig() GatewayConfiguration {
	return NewGatewayConfiguration(cfg.gatewayToken, cfg.gatewaySecret)
}

func (cfg *apiConfig) ConfigGateway(token, secret string) {
	cfg.gatewayToken = token
	cfg.gatewaySecret = secret
}

func (cfg *apiConfig) ConfigProduct(name, version string) {
	cfg.productName = name
	cfg.productVersion = version
}

func (cfg *apiConfig) ConfigSDKVersion(version string) {
	cfg.sdkVersion = version
}
func (cfg *apiConfig) GetUserAgent() string {
	return fmt.Sprintf("%s/%s (OS %s %s;%s;) IGGServerSDK/%s (Go)",
		gConfig.productName,
		gConfig.productVersion,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
		gConfig.sdkVersion)
}
