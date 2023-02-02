package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetThreeDayData(ctx *ctx.Context) (*dao.ThreeDay, *errmsg.ErrMsg) {
	td := &dao.ThreeDay{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), td)
	if err != nil {
		return nil, err
	}
	if !ok {
		td = &dao.ThreeDay{
			RoleId:        ctx.RoleId,
			LoginDay:      0,
			LastLoginTime: 0,
			Receive:       nil,
			ReceiveTime:   0,
		}
	}
	if td.Receive == nil {
		td.Receive = map[int64]int64{}
	}
	return td, nil
}

func SaveThreeDayData(ctx *ctx.Context, td *dao.ThreeDay) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), td)
}
