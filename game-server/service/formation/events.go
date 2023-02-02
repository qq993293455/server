package formation

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/cppbattle"
	"coin-server/common/proto/models"
	"coin-server/game-server/event"
	"coin-server/game-server/service/formation/dao"
	"coin-server/game-server/util/trans"
	"coin-server/rule"
)

func (this_ *Service) HandlerHeroAttrUpdate(c *ctx.Context, d *event.HeroAttrUpdate) *errmsg.ErrMsg {
	if len(d.Data) == 0 {
		return nil
	}

	formations, err := this_.Get(c, c.RoleId)
	if err != nil {
		return err
	}
	change := false
	defaultChange := false
	heroChange := false
	for idx := range d.Data {
		data := d.Data[idx]
		heroId := data.Hero.Id
		rh, ok := rule.MustGetReader(c).RowHero.GetRowHeroById(heroId)
		if !ok {
			return errmsg.NewErrHeroNotFound()
		}

		for i, v := range formations.Assembles {
			if v.Hero_0 == heroId {
				change = true // 属性有改变
			}
			if v.Hero_1 == heroId {
				change = true // 属性有改变
			}
			if v.HeroOrigin_0 == rh.OriginId && v.Hero_0 != heroId {
				change = true
				v.Hero_0 = heroId
				heroChange = true // 英雄有改变
			}
			if v.HeroOrigin_1 == rh.OriginId && v.Hero_1 != heroId {
				change = true
				v.Hero_1 = heroId
				heroChange = true // 英雄有改变
			}
			if change && int64(i) == formations.DefaultIndex {
				defaultChange = true // 改到了默认队伍
			}
		}
	}

	if change {
		dao.Save(c, formations)
	}

	if defaultChange {
		for _, h := range d.Data {
			if len(h.Hero.Skill) == 0 {
				return errmsg.NewInternalErr("SkillIds empty")
			}
		}

		if heroChange {
			err = this_.onDefaultChange(c, formations.Assembles, formations.DefaultIndex, 2)
			if err != nil {
				return err
			}
		} else {
			heroesModel := make([]*models.Hero, 0, len(d.Data))
			for _, item := range d.Data {
				heroesModel = append(heroesModel, item.Hero)
			}
			equips, err := this_.module.GetManyEquipBagMap(c, c.RoleId, this_.module.GetHeroesEquippedEquipId(heroesModel)...)
			if err != nil {
				return err
			}
			heroes := make([]*cppbattle.CPPBattle_HeroUpdate, len(d.Data))
			for idx := range d.Data {

				equip := make(map[int64]int64)
				equipLightEffect := make(map[int64]int64)
				for slot, item := range d.Data[idx].Hero.EquipSlot {
					if item.EquipItemId == 0 {
						equip[slot] = -1
						continue
					}
					equip[slot] = item.EquipItemId
					if equipModel, ok := equips[item.EquipId]; ok && equipModel.Detail != nil {
						if equipModel.Detail.LightEffect > 0 {
							equipLightEffect[slot] = equipModel.Detail.LightEffect
						}
					}
				}
				heroes[idx] = &cppbattle.CPPBattle_HeroUpdate{
					ConfigId:         d.Data[idx].Hero.Id,
					Attr:             d.Data[idx].Hero.Attrs,
					Equip:            equip,
					SkillIds:         d.Data[idx].Hero.Skill,
					IsSkillChange:    d.Data[idx].IsSkillChange,
					Buff:             d.Data[idx].Hero.Buff,
					TalentBuff:       d.Data[idx].Hero.TalentBuff,
					EquipLightEffect: equipLightEffect,
					// Fashion:          d.Data[idx].Hero.Fashion.Dressed,
					Fashion: trans.GetNewModelIdByFashionId(c, d.Data[idx].Hero.Fashion.Dressed),
				}
				if heroes[idx].Fashion == 0 {
					panic("fashion == 0")
				}
				if len(heroes[idx].SkillIds) == 0 {
					panic("skills empty")
				}
			}

			battleServerId, err1 := this_.module.GetCurBattleSrvId(c)
			if err1 != nil {
				return err1
			}
			if battleServerId > 0 {
				role, err := this_.module.GetRoleByRoleId(c, c.RoleId)
				if err != nil {
					return err
				}
				return this_.svc.GetNatsClient().Publish(battleServerId, c.ServerHeader, &cppbattle.CPPBattle_HeroAttrUpdatePush{
					BattleServerId: battleServerId,
					ObjId:          c.RoleId,
					Heroes:         heroes,
					Title:          role.Title,
					Level:          role.Level,
				})
			}
		}
	}
	return nil
}
