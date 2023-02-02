package handler

import (
	"context"

	redis "coin-server/common/redisclient"
	utils2 "coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/pikaviewer/utils"

	goredis "github.com/go-redis/redis/v8"
)

type Query struct {
	Way string `json:"way" binding:"required"`
	//Module   string `json:"module"`
	Instance string `json:"instance"`
	Key      string `json:"key" binding:"required"`
	//Type     string `json:"type" binding:"required"`
	Field   string `json:"field"`
	KeyInfo *KeyInfo
	Client  goredis.Cmdable
}

type KeyInfo struct {
	Module values.ModuleName
	Type   values.RedisKeyType
	RoleId values.RoleId
}

func NewQuery() *Query {
	return &Query{}
}

func (q *Query) Do() (interface{}, error) {
	if err := q.extractFromRedisKey(); err != nil {
		return nil, err
	}
	if err := q.getClient(); err != nil {
		return nil, err
	}
	return q.query()
}

func (q *Query) extractFromRedisKey() error {
	module, typ, roleId, err := utils2.ExtractFromRedisKey(q.Key)
	if err != nil {
		return err
	}
	q.KeyInfo = &KeyInfo{
		Module: module,
		Type:   typ,
		RoleId: roleId,
	}
	return nil
}

func (q *Query) getClient() error {
	q.Client = redis.GetDefaultRedis()
	return nil
	//if q.Way == "module" {
	//	//_ctx := &ctx.Context{
	//	//	ServerHeader: &models.ServerHeader{
	//	//		RoleId: q.KeyInfo.RoleId,
	//	//	},
	//	//	Context: context.Background(),
	//	//	Module:  string(q.KeyInfo.Module),
	//	//}
	//	//q.Client = redis.GetWithContext(ctx)
	//	// TODO module细化
	//	q.Client = redis.GetUserRedis()
	//	return nil
	//}
	//
	//client, ok := redis.GetInstanceByName(q.Instance)
	//if !ok {
	//	return utils.NewErrWithMsg("未找到对应的实例")
	//}
	//q.Client = client
	//return nil
}

func (q *Query) query() (interface{}, error) {
	ctx := context.Background()
	switch q.KeyInfo.Type {
	case values.KeyValue:
		cmd := q.Client.Get(ctx, q.Key)
		b, err := cmd.Bytes()
		if err == goredis.Nil {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		parser := NewParser(b, nil)
		return parser.KV(), nil
	case values.Hash:
		if q.Field == "" {
			cmd := q.Client.HGetAll(ctx, q.Key)
			result, err := cmd.Result()
			if err == goredis.Nil {
				return nil, nil
			}
			if err != nil {
				return nil, err
			}
			parser := NewParser(nil, result)
			return parser.Hash(), err
		}
		cmd := q.Client.HGet(ctx, q.Key, q.Field)
		b, err := cmd.Bytes()
		if err == goredis.Nil {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		parser := NewParser(b, nil)
		return parser.KV(), nil
	default:
		return nil, utils.NewErrWithMsg("不支持的数据类型")
	}
}
