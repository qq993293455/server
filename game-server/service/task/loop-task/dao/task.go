package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetTask(ctx *ctx.Context, roleId values.RoleId) (*dao.LoopTask, *errmsg.ErrMsg) {
	ret := &dao.LoopTask{RoleId: roleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), ret)
	if err != nil {
		return nil, err
	}
	if !ok {
		ret.Tasks = make([]*models.LoopTask, 0)
	}
	return ret, nil
}

func SaveTask(ctx *ctx.Context, task *dao.LoopTask) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), task)
	return
}

func GetLastLogin(ctx *ctx.Context, roleId values.RoleId) (*dao.TaskLogin, *errmsg.ErrMsg) {
	ret := &dao.TaskLogin{RoleId: roleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	return ret, err
}

func SaveLastLogin(ctx *ctx.Context, data *dao.TaskLogin) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}

func GetTaskStageReward(ctx *ctx.Context, roleId values.RoleId) (*dao.TaskStageReward, *errmsg.ErrMsg) {
	ret := &dao.TaskStageReward{RoleId: roleId}
	has, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		SaveTaskStageReward(ctx, ret)
	}
	if ret.Daily == nil {
		ret.Daily = map[int64]bool{}
	}
	if ret.Weekly == nil {
		ret.Weekly = map[int64]bool{}
	}
	return ret, err
}

func SaveTaskStageReward(ctx *ctx.Context, data *dao.TaskStageReward) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}
