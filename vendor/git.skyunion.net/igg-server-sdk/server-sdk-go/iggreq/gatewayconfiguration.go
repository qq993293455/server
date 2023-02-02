package iggreq

// GatewayConfiguration 网关配置
type GatewayConfiguration struct {
	GatewayAddress string
	GatewayToken   string
	GatewaySecret  string
}

func NewGatewayConfiguration(token, secret string) GatewayConfiguration {
	return GatewayConfiguration{
		GatewayAddress: GatewayInternalURL,
		GatewayToken:   token,
		GatewaySecret:  secret,
	}
}
