package enum

import (
	"coin-server/common/redisclient"

	"github.com/go-redis/redis/v8"
)

const (
	TopRankId    = "top_rank"
	TopRankLimit = "top_rank_limit"
)

func GetRedisClient() redis.Cmdable {
	return redisclient.GetDefaultRedis()
}
