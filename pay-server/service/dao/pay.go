package dao

import (
	"context"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/pay"
	"coin-server/common/proto/dao"
	"coin-server/common/values"
)

func SavePay(roleId values.RoleId, data *dao.Pay) *errmsg.ErrMsg {
	db := orm.GetOrm(ctx.GetContext())
	db.HSetPB(pay.GetPayRedis(), pay.GetKey(roleId), data)
	return db.Do()
}

func Save2Queue(data *dao.PayQueue) error {
	return pay.GetPayQueueRedis().RPush(context.Background(), pay.PayQueueKey, data.ToSave()).Err()
}
