package dao

import (
	"context"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetStatus(ctx *ctx.Context) (*dao.RacingRankStatus, *errmsg.ErrMsg) {
	status := &dao.RacingRankStatus{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), status)
	if err != nil {
		return nil, err
	}
	if !ok {
		status = &dao.RacingRankStatus{
			RoleId:       ctx.RoleId,
			Enrolled:     false,
			NextRefresh:  0,
			Season:       0,
			HighestRank:  -1,
			RewardedRank: nil,
			EnrollTime:   0,
			EndTime:      0,
		}
	}
	if status.RewardedRank == nil {
		status.RewardedRank = []int64{}
	}
	return status, nil
}

func SaveStatus(ctx *ctx.Context, status *dao.RacingRankStatus) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), status)
}

func SaveStatusImmediately(status *dao.RacingRankStatus) *errmsg.ErrMsg {
	o := orm.GetOrm(context.Background())
	o.SetPB(redisclient.GetDefaultRedis(), status)
	return o.Do()
}
