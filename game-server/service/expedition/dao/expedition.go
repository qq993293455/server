package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/timer"
)

func Get(ctx *ctx.Context) (*dao.Expedition, bool, *errmsg.ErrMsg) {
	e := &dao.Expedition{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), e)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		e = &dao.Expedition{
			RoleId: ctx.RoleId,
			Task:   map[string]*models.ExpeditionTask{},
			Execution: &dao.Execution{
				ExtraRestoreCount: map[int64]int64{},
				LastRecoverTime:   timer.StartTime(ctx.StartTime).Unix(),
			},
			NormalSlot: 0,
			ExtraSlot:  0,
			MustCount:  0,
			Refresh:    &dao.ExpeditionRefresh{},
		}
	}
	if e.Task == nil {
		e.Task = map[string]*models.ExpeditionTask{}
	}
	if e.Execution == nil {
		e.Execution = &dao.Execution{}
	}
	if e.Execution.ExtraRestoreCount == nil {
		e.Execution.ExtraRestoreCount = map[int64]int64{}
	}
	if e.Refresh == nil {
		e.Refresh = &dao.ExpeditionRefresh{}
	}
	return e, ok, nil
}

func Save(ctx *ctx.Context, e *dao.Expedition) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), e)
}
