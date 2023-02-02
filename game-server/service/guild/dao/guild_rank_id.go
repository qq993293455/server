package dao

import (
	"context"

	"coin-server/common/errmsg"
	"coin-server/common/redisclient"
	"coin-server/common/values"
	"coin-server/common/values/enum"
)

func GetRankId() (values.RankId, *errmsg.ErrMsg) {
	ctx := context.Background()
	c := redisclient.GetDefaultRedis()
	rankId, err := c.Get(ctx, enum.GuildRankKey).Result()
	if err != nil {
		if err == redisclient.Nil {
			return "", nil
		}
		return "", errmsg.NewErrorDB(err)
	}
	return rankId, nil
}

func SaveRankId(rankId values.RankId) *errmsg.ErrMsg {
	ctx := context.Background()
	c := redisclient.GetDefaultRedis()
	if err := c.Set(ctx, enum.GuildRankKey, rankId, 0).Err(); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}
