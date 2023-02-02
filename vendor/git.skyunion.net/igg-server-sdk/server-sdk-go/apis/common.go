package apis

import "git.skyunion.net/igg-server-sdk/server-sdk-go/iggreq"

// GlobalInit 全局初始化
func GlobalInit(gatewayToken, gatewaySecret string) {
	iggreq.ConfigInstance().ConfigGateway(gatewayToken, gatewaySecret)
}

// SetUserAgent 设置 API 网关请求User-Agent 标头信息
func SetUserAgent(productName, productVersion string) {
	iggreq.ConfigInstance().ConfigProduct(productName, productVersion)
}
