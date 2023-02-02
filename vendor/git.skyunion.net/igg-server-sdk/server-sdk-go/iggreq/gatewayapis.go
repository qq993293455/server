package iggreq

const (
	// GatewayURL gateway host scheme
	GatewayURL = "http://192.243.46.246"
	// GatewayInternalURL Another gateway host scheme
	GatewayInternalURL = "https://apis-internal.skyunion.net"

	// AlarmAPIPath alarm api path on gateway
	AlarmAPIPath = "/ea/alarm/sendAlarm"
	// AlarmSecCode Alarm api query security code
	AlarmSecCode = "rt5{jnZ@1"

	// IPQueryAPIPath IP query api path on gateway
	IPQueryAPIPath = "/common/ipquery/get"

	// TokenVerifyAPIPath Access token verify api path on gateway
	TokenVerifyAPIPath     = "/ums/member/access_token/verify"
	TokenCertsAPIPath      = "/ums/member/access_token/certs"
	TokenMinimumVerAPIPath = "/ums/member/access_token/get_minimum_version"

	// PushSendMsgAPIPath send push message api path on gateways
	PushSendMsgAPIPath = "/common/push/send_msg"
	// PushSendMsgLoginAPIPath push abnormal login message api path on gateway
	PushSendMsgLoginAPIPath = "/common/push/send_msg_for_login"

	// MessageGatewaySendMsgAPIPath message gateway send message api path on gateway
	MessageGatewaySendMsgAPIPath = "/ea/message/v1/sendMessage"

	KVStorageAPIPushPath      = "/storage/game_storage/push"
	KVStorageAPICDNPushPath   = "/storage/game_storage/protracted/cdn/push"
	KVStorageAPICDNDeletePath = "/storage/game_storage/protracted/cdn/delete"
	KVStorageAPIKVPushPath    = "/storage/game_storage/protracted/kv/push"
	KVStorageAPIKVPullPath    = "/storage/game_storage/protracted/kv/pull"
	KVStorageSecCode          = "Jxjd34#K"
)
