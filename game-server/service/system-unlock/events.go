package system_unlock

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/system-unlock/dao"
)

func (s *Service) HandleUnlock(ctx *ctx.Context, d *event.TargetUpdate, args any) *errmsg.ErrMsg {
	unlock, err := dao.GetSysUnlock(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	arg, _ := args.(int64)
	unlocked, ok := unlock.Unlock[arg]
	if !ok {
		return nil
	}
	if unlocked {
		return nil
	}
	unlock.Unlock[arg] = true
	// GetSysUnlockCache().SetUnlock(ctx, models.SystemType(arg))
	dao.SaveSysUnlock(ctx, unlock)
	ctx.PublishEventLocal(&event.SystemUnlock{SystemId: []values.SystemId{values.SystemId(arg)}})
	return nil
}

func (s *Service) HandleSystemUnlock(ctx *ctx.Context, d *event.SystemUnlock) *errmsg.ErrMsg {
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.SystemUnlock_SystemUnlockPush{SystemId: d.SystemId})
	return nil
}

func (s *Service) HandleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	if d.IsRegister {
		unlock, err := dao.GetSysUnlock(ctx, ctx.RoleId)
		if err != nil {
			return err
		}
		unlock.Unlock = getDefaultUnlock(ctx)
		dao.SaveSysUnlock(ctx, unlock)
		// cache.init(ctx, unlock.Unlock)
		return nil
	}
	// unlock, err := dao.GetSysUnlock(ctx, ctx.RoleId)
	// if err != nil {
	// 	return err
	// }
	// cache.init(ctx, unlock.Unlock)
	return nil
}
