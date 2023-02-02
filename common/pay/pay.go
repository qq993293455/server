package pay

import (
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
	"github.com/go-redis/redis/v8"
)

const PayQueueKey = "pay_queue"

func GetKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.Pay, values.Hash, roleId)
}

func GetPayRedis() redis.Cmdable {
	return redisclient.GetDefaultRedis()
}

func GetPayQueueRedis() redis.Cmdable {
	return redisclient.GetDefaultRedis()
}
