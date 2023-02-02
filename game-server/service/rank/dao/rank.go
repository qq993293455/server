package dao

import (
	"fmt"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/values"

	"github.com/gogo/protobuf/proto"
)

const (
	rankRedisKey = "rank_hash"

	criticalCntPerShard = 1000
)

func CreateRank(ctx *ctx.Context, rankId values.RankId) *errmsg.ErrMsg {
	rankShards := &dao.Rank{RankId: rankId}
	has, err := ctx.NewOrm().GetPB(redisclient.GetCommonRedis(), rankShards)
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	shardKey := getRedisShardKey(rankId, 0)
	rankShards.Shards = []string{shardKey}
	ctx.NewOrm().SetPB(redisclient.GetCommonRedis(), rankShards)
	return nil
}

func CreateValue(ctx *ctx.Context, value *models.RankValue) *errmsg.ErrMsg {
	rankShards := &dao.Rank{RankId: value.RankId}
	has, err := ctx.NewOrm().GetPB(redisclient.GetCommonRedis(), rankShards)
	if err != nil {
		return err
	}
	var shardKey string
	if !has {
		// 当前不存在rankId对应的分片索引
		value.Shard = 0
		shardKey = getRedisShardKey(value.RankId, value.Shard)
		rankShards.Shards = []string{shardKey}
	} else if needShard(rankShards) {
		// 如果需要扩充分片
		value.Shard = int64(len(rankShards.Shards))
		shardKey = getRedisShardKey(value.RankId, value.Shard)
		rankShards.Shards = append(rankShards.Shards, shardKey)
	} else {
		value.Shard = int64(len(rankShards.Shards)) - 1
		shardKey = getRedisShardKey(value.RankId, value.Shard)
	}
	rankShards.RecordCnt++
	ctx.NewOrm().SetPB(redisclient.GetCommonRedis(), rankShards)
	byt, e := proto.Marshal(rankShards)
	if e != nil {
		return errmsg.NewProtocolError(e)
	}
	redisclient.GetCommonRedis().HSet(ctx, shardKey, value.OwnerId, byt)
	return nil
}

func GetValuesByRankId(ctx *ctx.Context, rankId values.RankId) []*models.RankValue {
	rankShards := &dao.Rank{RankId: rankId}
	has, err := ctx.NewOrm().GetPB(redisclient.GetCommonRedis(), rankShards)
	if err != nil || !has {
		return nil
	}
	res := make([]*models.RankValue, 0)
	for _, shardKey := range rankShards.Shards {
		data, err := redisclient.GetCommonRedis().HGetAll(ctx, shardKey).Result()
		if err != nil {
			return nil
		}
		for _, v := range data {
			each := &models.RankValue{}
			if err = proto.Unmarshal([]byte(v), each); err != nil {
				return nil
			}
			if each.OwnerId != "" {
				res = append(res, each)
			}
		}
	}
	return res
}

func SaveValue(ctx *ctx.Context, value *models.RankValue) {
	redisclient.GetCommonRedis().HSet(ctx, getRedisShardKey(value.RankId, value.Shard), value.OwnerId, value)
}

func DeleteValue(ctx *ctx.Context, value *models.RankValue) {
	redisclient.GetCommonRedis().HDel(ctx, getRedisShardKey(value.RankId, value.Shard), value.OwnerId)
}

func getRedisShardKey(rankId values.RankId, shard values.Integer) string {
	return fmt.Sprintf("%s:%s:%d", rankRedisKey, rankId, shard)
}

func needShard(s *dao.Rank) bool {
	if (int(s.RecordCnt) / len(s.Shards)) > criticalCntPerShard {
		return true
	}
	return false
}
