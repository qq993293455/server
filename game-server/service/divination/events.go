package divination

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/game-server/event"
	"coin-server/game-server/service/divination/dao"
	"coin-server/game-server/util/trans"
)

func (svc *Service) HandleRoleLvUpEvent(ctx *ctx.Context, e *event.UserLevelChange) *errmsg.ErrMsg {
	if !e.IsAdvance {
		return nil
	}
	div, err := svc.getDivination(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	div.AvailableCount = div.TotalCount
	dao.Save(ctx, div)
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.Divination_ResetDivinationPush{Info: trans.DivinationD2M(div)})
	return nil
}

func (svc *Service) HandleExtraSkillTypTotal(ctx *ctx.Context, e *event.ExtraSkillTypTotal) *errmsg.ErrMsg {
	if e.TypId != models.EntrySkillType_ESTDivinationNumAdd {
		return nil
	}
	div, err := svc.addTotal(ctx, ctx.RoleId, e.ThisAdd)
	if err != nil {
		return err
	}
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.Divination_ResetDivinationPush{Info: trans.DivinationD2M(div)})
	return nil
}
