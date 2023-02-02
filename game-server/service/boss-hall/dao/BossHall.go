package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetBossHallData(ctx *ctx.Context, roleId values.RoleId) (*dao.BossHallData, *errmsg.ErrMsg) {
	ret := &dao.BossHallData{RoleId: roleId}
	has, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if err != nil {
		return nil, err
	}
	if !has {
		ret = &dao.BossHallData{
			RoleId: roleId,
			Info: &models.BossHallInfo{
				NextRefreshTime: 0,
				JoinTimes:       0,
				KillTimes:       0,
			},
		}
	}

	return ret, nil
}

func SaveBossHallData(ctx *ctx.Context, data *dao.BossHallData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}

func GetBossKillJoinData(ctx *ctx.Context, roleId values.RoleId) (*dao.BossKillJoinData, *errmsg.ErrMsg) {
	ret := &dao.BossKillJoinData{RoleId: roleId}
	has, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if err != nil {
		return nil, err
	}
	if !has {
		ret = &dao.BossKillJoinData{
			RoleId:          roleId,
			NextRefreshTime: 0,
		}
	}

	return ret, nil
}

func SaveBossKillJoinData(ctx *ctx.Context, data *dao.BossKillJoinData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}
