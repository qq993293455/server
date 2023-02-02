package relics

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	servicepb "coin-server/common/proto/service"
	"coin-server/game-server/event"
	"coin-server/rule"
)

func (s *Service) HandleRelicsSuitUpdate(ctx *ctx.Context, d *event.RelicsSuitUpdate) *errmsg.ErrMsg {
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.Relics_RelicsSuitUpdatePush{Suit: d.RelicsSuit})
	return nil
}

func (s *Service) HandleRelicsUpdate(ctx *ctx.Context, d *event.RelicsUpdate) *errmsg.ErrMsg {
	if !d.IsNewRelics {
		return nil
	}
	// 新获得遗物，更新属性加成及技能
	for _, relics := range d.Relics {
		updateRelicsAttr(ctx, relics)
		err := updateRelicsSkill(ctx, relics)
		if err != nil {
			return err
		}
		err = updateSuit(ctx, relics)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) HandleTaskUpdate(ctx *ctx.Context, d *event.TargetUpdate) *errmsg.ErrMsg {
	reader := rule.MustGetReader(ctx)
	funcMap := reader.GetRelicsAttrFunc()
	if relicsIds, exist := funcMap[d.Typ]; exist {
		return s.updateRelicsFuncAttr(ctx, relicsIds, d)
	}
	return nil
}
