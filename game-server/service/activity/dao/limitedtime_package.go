// 限时弹窗礼包

package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/game-server/util"
)

func GetLimitedTimePackages(ctx *ctx.Context) (*daopb.LimitedTimePackages, *errmsg.ErrMsg) {
	data := &daopb.LimitedTimePackages{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		data = &daopb.LimitedTimePackages{
			RoleId: ctx.RoleId,
		}
	}
	if data.Packages == nil {
		data.Packages = map[int64]*models.LimitedTimePackage{}
	}
	return data, nil
}

func SaveLimitedTimePackages(ctx *ctx.Context, data *daopb.LimitedTimePackages) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}

func GetLimitedTimePackageLocks(ctx *ctx.Context) (*daopb.LimitedTimePackageLocks, *errmsg.ErrMsg) {
	data := &daopb.LimitedTimePackageLocks{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		data = &daopb.LimitedTimePackageLocks{
			RoleId:  ctx.RoleId,
			ResetAt: util.DefaultNextRefreshTime().UnixMilli(),
		}
	}
	if data.Locks == nil {
		data.Locks = map[int64]int64{}
	}
	return data, nil
}

func SaveLimitedTimePackageLocks(ctx *ctx.Context, data *daopb.LimitedTimePackageLocks) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}
