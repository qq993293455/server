package hero

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/hero/dao"
	"coin-server/game-server/service/hero/rule"

	"go.uber.org/zap"
)

func (svc *Service) Activate(ctx *ctx.Context, req *servicepb.Hero_ActivateResonanceRequest) (*servicepb.Hero_ActivateResonanceResponse, *errmsg.ErrMsg) {
	hero, ok, err := dao.NewHero(ctx.RoleId).GetOne(ctx, req.HeroOriginId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	if hero.Resonance == nil {
		return nil, errmsg.NewErrEquipNoResonanceToActivate()
	}
	status, ok := hero.Resonance[req.ResonanceId]
	if !ok {
		return nil, errmsg.NewErrEquipNoResonanceToActivate()
	}
	if status != models.ResonanceStatus_RSReached {
		return nil, errmsg.NewErrEquipNoResonanceToActivate()
	}

	_, ok = rule.GetEquipResonance(ctx, req.ResonanceId)
	if !ok {
		ctx.Error("equip_resonance config not found", zap.Int64("id", req.ResonanceId))
		return nil, errmsg.NewInternalErr("equip_resonance config not found")
	}
	hero.Resonance[req.ResonanceId] = models.ResonanceStatus_RSActivated
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getHeroEquippedEquipId(hero)...)
	if err != nil {
		return nil, err
	}
	svc.refreshHeroAttr(ctx, hero, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill, 0)
	if err := dao.NewHero(ctx.RoleId).Save(ctx, hero); err != nil {
		return nil, err
	}
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: []*event.HeroAttrUpdateItem{{
			IsSkillChange: false,
			Hero:          svc.dao2model(ctx, hero),
		}},
	})
	return &servicepb.Hero_ActivateResonanceResponse{
		Hero: svc.dao2model(ctx, hero),
	}, nil
}

func (svc *Service) refreshEquipResonanceStatus(ctx *ctx.Context, hero *pbdao.Hero) {
	if hero.EquipSlot == nil {
		return
	}
	dataMap := rule.GetEquipResonanceByHero(ctx, hero.Id)
	if len(dataMap) <= 0 {
		return
	}
	if hero.Resonance == nil {
		hero.Resonance = map[int64]models.ResonanceStatus{}
	}
	for level, resonance := range dataMap {
		status := hero.Resonance[resonance.Id]
		if status == models.ResonanceStatus_RSActivated {
			continue
		}
		if svc.resonanceReached(hero, level) {
			hero.Resonance[resonance.Id] = models.ResonanceStatus_RSReached
		}
	}
}

func (svc *Service) resonanceReached(hero *pbdao.Hero, level values.Integer) bool {
	if hero.EquipSlot == nil {
		return false
	}
	var count int
	for _, slot := range hero.EquipSlot {
		if slot.Star >= level {
			count++
		}
	}
	// 7个部位全达成
	return count >= 7
}
