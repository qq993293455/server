package iggsdk

import (
	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"git.skyunion.net/igg-server-sdk/server-sdk-go/apis"
)

func Init(productVersion string, log *logger.Logger, cnf *consulkv.Config) {
	gatewayToken := "d34f2ed9b76cc5dc03fae2de8af567e3"
	gatewaySecret := "CqktUaWHgHhg8uup2aoZoroklreHJBet5NZxBn6q"

	productName := "ROW"

	//////////////////////////////////////////
	// 初始化 SDK
	apis.GlobalInit(gatewayToken, gatewaySecret)
	apis.SetUserAgent(productName, productVersion)

	InitAccountIns()
	InitAlarmIns(log, cnf)
	InitPushIns(cnf)
	InitIpIns()
}
