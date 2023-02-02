package utils

import (
	"coin-server/common/logger"
	"coin-server/common/natsclient"
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

var NATS *natsclient.ClusterClient

func InitNats(urls []string, serverId values.ServerId, serverType models.ServerType, log *logger.Logger) {
	NATS = natsclient.NewClusterClient(serverType, serverId, urls, log)
}
