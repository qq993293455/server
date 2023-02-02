package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

const recordKey = "genRankServer"

func GetGenRecord(ctx *ctx.Context) (*dao.GenRankRecord, *errmsg.ErrMsg) {
	res := &dao.GenRankRecord{Id: recordKey}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.GenRankRecord{
			Id:  recordKey,
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	}
	return res, nil
}

func SaveGenRecord(ctx *ctx.Context, v *dao.GenRankRecord) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), v)
}
