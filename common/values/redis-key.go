package values

type RedisKeyType string

const (
	KeyValue RedisKeyType = "k"
	Hash     RedisKeyType = "h"
	List     RedisKeyType = "l"
	Set      RedisKeyType = "s"
	ZSET     RedisKeyType = "z"
)
