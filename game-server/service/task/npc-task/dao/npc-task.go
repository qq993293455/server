package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetTask(ctx *ctx.Context, roleId values.RoleId) (*dao.NpcTask, *errmsg.ErrMsg) {
	ret := &dao.NpcTask{RoleId: roleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), ret)
	if err != nil {
		return nil, err
	}
	if !ok || ret.Tasks == nil {
		ret.Tasks = map[int64]*models.NpcTask{}
	}

	return ret, nil
}

func SaveTask(ctx *ctx.Context, task *dao.NpcTask) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), task)
	return
}

func GetNpcTaskUnlock(ctx *ctx.Context, roleId values.RoleId, taskId values.TaskId) (*dao.NpcTaskUnlock, *errmsg.ErrMsg) {
	unlock := &dao.NpcTaskUnlock{TaskId: taskId}
	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getNpcTaskUnlockKey(roleId), unlock)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, err
	}
	return unlock, err
}

func SaveNpcTaskUnlock(ctx *ctx.Context, roleId values.RoleId, task *dao.NpcTaskUnlock) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getNpcTaskUnlockKey(roleId), task)
}

func getNpcTaskUnlockKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.NpcTaskUnlock, values.Hash, roleId)
}
