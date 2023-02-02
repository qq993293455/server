package client_version

import (
	"coin-server/common/redisclient"

	"github.com/go-redis/redis/v8"
)

const (
	ClientVersionKey = "CLIENT_VERSION"
	AuditVersionKey  = "AUDIT_VERSION"
)

func GetClientVersionRedis() redis.Cmdable {
	return redisclient.GetDefaultRedis()
}
