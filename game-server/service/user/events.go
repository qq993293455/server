package user

import (
	"fmt"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/iggsdk"
	"coin-server/common/proto/dao"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/syncrole"
	"coin-server/common/utils"
	"coin-server/common/utils/generic/slices"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/user/db"
	rule2 "coin-server/game-server/service/user/rule"
	"coin-server/rule"

	"go.uber.org/zap"

	"github.com/gogo/protobuf/proto"
)

func (this_ *Service) HandleLoginEvent(c *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	if d.IsRegister {
		return nil
	}
	_, _, err := this_.CheckAvatarExpired(c)
	return err
}

func (this_ *Service) HandleAttrUpdateToRole(ctx *ctx.Context, d *event.AttrUpdateToRole) *errmsg.ErrMsg {
	if len(d.AttrFixed) == 0 && len(d.AttrPercent) == 0 {
		return nil
	}
	roleAttr, err := db.GetRoleAttrByType(ctx, ctx.RoleId, d.Typ)
	if err != nil {
		return err
	}
	r := rule.MustGetReader(ctx)
	if len(d.AttrFixed) != 0 {
		for k, v := range d.AttrFixed {
			cfg, ok := r.Attr.GetAttrById(k)
			if !ok {
				continue
			}
			find := false
			if d.HeroId == 0 {
				for _, attr := range roleAttr.AttrFixed {
					if attr.AdvancedType == cfg.AdvancedType {
						find = true
						if d.IsCover {
							attr.Attr[k] = v
						} else {
							attr.Attr[k] += v
						}
					}
				}
				if !find {
					roleAttr.AttrFixed = append(roleAttr.AttrFixed, &models.AttrBonus{
						AdvancedType: cfg.AdvancedType,
						Attr:         map[values.AttrId]values.Integer{k: v},
					})
				}
			} else {
				for _, attr := range roleAttr.HeroAttrFixed {
					if attr.AdvancedType == cfg.AdvancedType && attr.HeroId == d.HeroId {
						find = true
						if d.IsCover {
							attr.Attr[k] = v
						} else {
							attr.Attr[k] += v
						}
					}
				}
				if !find {
					roleAttr.HeroAttrFixed = append(roleAttr.HeroAttrFixed, &models.HeroAttrBonus{
						AdvancedType: cfg.AdvancedType,
						HeroId:       d.HeroId,
						Attr:         map[values.AttrId]values.Integer{k: v},
					})
				}
			}
		}
	}
	if len(d.AttrPercent) != 0 {
		for k, v := range d.AttrPercent {
			cfg, ok := r.Attr.GetAttrById(k)
			if !ok {
				continue
			}
			find := false
			if d.HeroId == 0 {
				for _, attr := range roleAttr.AttrPercent {
					if attr.AdvancedType == cfg.AdvancedType {
						find = true
						if d.IsCover {
							attr.Attr[k] = v
						} else {
							attr.Attr[k] += v
						}
					}
				}
				if !find {
					roleAttr.AttrPercent = append(roleAttr.AttrPercent, &models.AttrBonus{
						AdvancedType: cfg.AdvancedType,
						Attr:         map[values.AttrId]values.Integer{k: v},
					})
				}
			} else {
				for _, attr := range roleAttr.HeroAttrPercent {
					if attr.AdvancedType == cfg.AdvancedType && attr.HeroId == d.HeroId {
						find = true
						if d.IsCover {
							attr.Attr[k] = v
						} else {
							attr.Attr[k] += v
						}
					}
				}
				if !find {
					roleAttr.HeroAttrPercent = append(roleAttr.HeroAttrPercent, &models.HeroAttrBonus{
						AdvancedType: cfg.AdvancedType,
						HeroId:       d.HeroId,
						Attr:         map[values.AttrId]values.Integer{k: v},
					})
				}
			}
		}
	}
	db.SaveRoleAttr(ctx, ctx.RoleId, roleAttr)
	ctx.PublishEventLocal(&event.RoleAttrUpdate{
		Typ:             d.Typ,
		AttrFixed:       roleAttr.AttrFixed,
		AttrPercent:     roleAttr.AttrPercent,
		HeroAttrFixed:   roleAttr.HeroAttrFixed,
		HeroAttrPercent: roleAttr.HeroAttrPercent,
	})
	return nil
}

func (this_ *Service) HandleLogoutEvent(c *ctx.Context, d event.Logout) *errmsg.ErrMsg {
	c.Info("user logout")
	return nil
}

// HandleRoleSkillUpdate 技能更新
func (this_ *Service) HandleRoleSkillUpdate(ctx *ctx.Context, d *event.RoleSkillUpdate) *errmsg.ErrMsg {
	if d.OldSkill == 0 && d.NewSkill == 0 {
		return nil
	}
	skill, err := db.GetRoleSkill(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if d.OldSkill != 0 {
		for i := range skill.SkillId {
			if skill.SkillId[i] == d.OldSkill {
				skill.SkillId = append(skill.SkillId[:i], skill.SkillId[i+1:]...)
				break
			}
		}
	}
	if d.NewSkill != 0 {
		skill.SkillId = append(skill.SkillId, d.NewSkill)
	}
	db.SaveRoleSkill(ctx, skill)
	ctx.PublishEventLocal(&event.RoleSkillUpdateFinish{
		Skill: skill.SkillId,
	})
	return nil
}

func (this_ *Service) HandleRoleLvUpEvent(c *ctx.Context, e *event.UserLevelChange) *errmsg.ErrMsg {
	c.PushMessage(&lessservicepb.User_UserLevelChangePush{
		Level:          e.Level,
		LevelIndex:     e.LevelIndex,
		LevelIncr:      e.Incr,
		LevelIndexIncr: e.LevelIndexIncr,
	})
	if !e.IsAdvance {
		return nil
	}
	if e.Incr == 1 {
		this_.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskRoleLvlUpCnt, 0, 1)
	}
	err := this_.titleUnlock(c)
	if err != nil {
		return err
	}
	return nil
}

func (this_ *Service) HeroUpdate(ctx *ctx.Context, _ *event.HeroAttrUpdate) *errmsg.ErrMsg {
	heroes, err := this_.GetAllHero(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	role, err := db.GetRole(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	var combatValue values.Integer
	for _, hero := range heroes {
		if hero.CombatValue == nil {
			continue
		}
		combatValue += hero.CombatValue.Total
	}
	if role.Power != combatValue {
		old := role.Power
		role.Power = combatValue
		// 战斗力变化时同步战斗力至百人排行榜
		if err := this_.syncTopRank(ctx, role.Title, role.Title, role.Power); err != nil {
			return err
		}
		if role.Power > 1000000000 {
			iggsdk.GetAlarmIns().SendRes(fmt.Sprintf("玩家 %d 战斗力超过10亿", utils.Base34DecodeString(ctx.RoleId)))
		}
		if role.Power > role.HighestPower {
			role.HighestPower = role.Power
		}
		db.SaveRole(ctx, role)
		this_.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskPower, 0, role.Power, true)
		ctx.PublishEventLocal(&event.UserCombatValueChange{
			RoleId: ctx.RoleId,
			Value:  combatValue,
		})
		reqMsgName := proto.MessageName(ctx.Req)
		if reqMsgName != "less_service.User.RoleLoginRequest" {
			ctx.PushMessage(&servicepb.User_UserCombatValueChangePush{
				Old: old,
				Cur: role.Power,
			})
		}
		utils.Must(syncrole.Update(ctx, role))
	}
	return nil
}

func (this_ *Service) HandleReadPointAdd(c *ctx.Context, evt *event.RedPointAdd) *errmsg.ErrMsg {
	rp, err := db.GetReadPoint(c, evt.RoleId)
	if err != nil {
		return err
	}
	val := rp.RedPoints[evt.Key]
	val += evt.Val
	if val < 0 {
		val = 0
	}
	rp.RedPoints[evt.Key] = val
	db.SaveRedPoint(c, rp)
	c.PushMessageToRole(c.RoleId, &lessservicepb.User_ReadPointChangePush{
		RedPoints: rp.RedPoints,
	})
	return nil
}

func (this_ *Service) HandleReadPointChange(c *ctx.Context, evt *event.RedPointChange) *errmsg.ErrMsg {
	rp, err := db.GetReadPoint(c, evt.RoleId)
	if err != nil {
		return err
	}
	val := rp.RedPoints[evt.Key]
	val = evt.Val
	if val < 0 {
		val = 0
	}
	rp.RedPoints[evt.Key] = val
	db.SaveRedPoint(c, rp)
	c.PushMessageToRole(c.RoleId, &lessservicepb.User_ReadPointChangePush{
		RedPoints: rp.RedPoints,
	})
	return nil
}

func (this_ *Service) HandleEntrySpecialAddition(c *ctx.Context, e *event.EntrySpecialAddition) *errmsg.ErrMsg {
	// TODO 等挂机收益（金币和经验）功能从战斗服临时背包挪到游戏服来
	if !slices.In(enum.EntryListIdle(), e.EntryId) {
		return nil
	}

	return nil
}

func (this_ *Service) HandleTitleChange(ctx *ctx.Context, d *event.UserTitleChange) *errmsg.ErrMsg {
	return this_.syncTopRank(ctx, d.LastTitle, d.CurrentTitle, d.CombatValue)
}

func (this_ *Service) HandleRecentChatAdd(ctx *ctx.Context, d *event.UserRecentChatAdd) *errmsg.ErrMsg {
	return this_.addRecentChatIds(ctx, d.MyRoleId, d.TarRoleId)
}

func (this_ *Service) HandleExtraSkillAdd(c *ctx.Context, d *event.ExtraSkillTypAdd) *errmsg.ErrMsg {
	cnt, err := db.GetExtraSkill(c, c.RoleId)
	if err != nil {
		return err
	}
	tc := d.Cnt
	logicId := d.LogicId
	typId := int64(d.TypId)
	if data, exit := cnt.Data[typId]; !exit {
		cnt.Data[typId] = &dao.ExtraSkillTypCntDetail{
			Cnt: map[int64]int64{
				logicId: d.Cnt,
			},
		}
	} else {
		data.Cnt[logicId] += d.Cnt
		tc = data.Cnt[logicId]
	}
	db.SaveExtraSkill(c, cnt)
	c.PublishEventLocal(&event.ExtraSkillTypTotal{
		TypId:    d.TypId,
		LogicId:  d.LogicId,
		TotalCnt: tc,
		ThisAdd:  d.Cnt,
		ValueTyp: d.ValueTyp,
	})
	return nil
}

func (this_ *Service) HandleRoguelikeExtra(c *ctx.Context, d *event.ExtraSkillTypTotal) *errmsg.ErrMsg {
	if d.TypId != models.EntrySkillType_ESTRoguelikeChallengeNumAdd {
		return nil
	}
	cnt, err := db.GetRLCnt(c, c.RoleId)
	if err != nil {
		return err
	}
	cnt.ExtraCnt = d.TotalCnt
	db.SaveRLCnt(c, cnt)
	return nil
}

func (this_ *Service) HandleUserRechargeChange(c *ctx.Context, d *event.RechargeSuccEvt) *errmsg.ErrMsg {
	c.PushMessage(&servicepb.User_RechargeChangePush{
		Old: d.Old,
		New: d.New,
	})
	return nil
}

func (this_ *Service) HandleNormalPaySuccess(ctx *ctx.Context, d *event.PaySuccess) *errmsg.ErrMsg {
	isNormal, ok := rule2.IsNormalPayItem(ctx, d.PcId)
	if !ok {
		ctx.Error("charge config not found", zap.Int64("pc_id", d.PcId), zap.String("role_id", ctx.RoleId))
		return errmsg.NewInternalErr("charge config not found")
	}
	if isNormal {
		item, id, err := this_.StellargemShopBuy(ctx, d.PcId)
		if err != nil {
			return err
		}
		ctx.PushMessage(&servicepb.Pay_NormalSuccessPush{
			PcId:         d.PcId,
			ShopNormalId: id,
			Items:        item,
		})
	}
	return nil
}
