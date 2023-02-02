package mail

import (
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"

	"github.com/go-redis/redis/v8"
)

const EntireMailKey = "entire_mail"

func GetMailRedis() redis.Cmdable {
	return redisclient.GetDefaultRedis()
}

func GetMailKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.Mail, values.Hash, roleId)
}
