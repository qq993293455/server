package values

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"
)

const AnyStoneSuitTyp = 99

type TalentI interface {
	RoleId() values.RoleId
	CommonPoints() values.Integer
	Gets() []*models.HeroTalent
	RuneLevelUp(c *ctx.Context, rune *models.TalentRune, oblation []*models.TalentRune) (values.Integer, []values.RuneId, *errmsg.ErrMsg)
	RuneLevelUpByDust(c *ctx.Context, rune *models.TalentRune, cnt values.Integer) (values.Integer, values.Integer, *errmsg.ErrMsg)
	UnlockPlate(c *ctx.Context, configId, plateIdx, loc values.Integer) (*models.HeroTalent, *errmsg.ErrMsg)
	InlayRune(c *ctx.Context, configId, plateIdx, loc values.Integer, rune *models.TalentRune, isInlay bool) (*models.TalentPlate, *errmsg.ErrMsg)
	SkillLevelUp(c *ctx.Context, configId, skillId, lvl values.Integer) (map[values.Integer]values.Integer, *models.SkillDetail, *errmsg.ErrMsg)
	SkillHandleRoleLvlUp(c *ctx.Context, roleLvl values.Integer) (values.Integer, []*models.HeroTalent)
	SkillHandleHeroRoleLvlUp(c *ctx.Context, roleLvl values.Integer, configId values.Integer) (values.Integer, *models.HeroTalent)
	SkillChoose(c *ctx.Context, configId values.Integer, idx, skillId, roleLvl values.Integer) (values.Integer, *models.HeroTalent, *errmsg.ErrMsg)
	Reset(c *ctx.Context, configId, plateIdx values.Integer) (*models.TalentPlate, []values.RuneId, *errmsg.ErrMsg)
	GainCommonP(num values.Integer)
	GainParticularP(configId, num values.Integer)
	InlayStone(c *ctx.Context, configId values.Integer, talentRule *rule_model.RoleSkill, stoneRule *rule_model.Skillstone, holeIdx int) (values.Integer, *models.SkillDetail, *errmsg.ErrMsg)
	RemoveStone(c *ctx.Context, configId, skillId values.Integer, holeIdx int) (values.Integer, *models.SkillDetail, *errmsg.ErrMsg)
	RemoveAllStone(c *ctx.Context, configId, skillId values.Integer) (map[values.ItemId]values.Integer, *models.SkillDetail, *errmsg.ErrMsg)
	CalAttr(c *ctx.Context, configId values.Integer) (*models.TalentAdvance, *errmsg.ErrMsg)
	CalSkill(c *ctx.Context, configId values.Integer) (*models.SkillAdvance, *errmsg.ErrMsg)
	CalAllSkill(c *ctx.Context) (map[values.HeroId]*models.SkillAdvance, *errmsg.ErrMsg)
	LockList() []values.Integer
	Lock(stoneId values.Integer)
	Unlock(stoneId values.Integer)
	IsLock(stoneId values.Integer) bool
	Init(c *ctx.Context, originId values.HeroId, roleLvl values.Integer) (values.Integer, *errmsg.ErrMsg)
	IsFirstUpdate(talentId, lvl values.Integer) bool
	FirstUpdate(talentId, lvl values.Integer)
	UpdateExtraLvl(c *ctx.Context, heroId values.HeroId, data map[values.TalentId]values.Level) map[values.HeroBuildId]map[values.TalentId]values.Level
	CalAllUnusedSkill(c *ctx.Context) (map[values.HeroId]map[values.Integer]values.Integer, *errmsg.ErrMsg)
	CalUnusedSkill(c *ctx.Context, configId values.Integer) (map[values.Integer]values.Integer, *errmsg.ErrMsg)
	ToDao() *dao.Talent
}

type talent struct {
	values *dao.Talent
}

func NewTalent(val *dao.Talent) TalentI {
	return &talent{
		values: val,
	}
}

func (s *talent) RoleId() values.RoleId {
	return s.values.RoleId
}

func (s *talent) CommonPoints() values.Integer {
	return s.values.CommonPoints
}

func (s *talent) Gets() []*models.HeroTalent {
	return s.values.Ht
}

func (s *talent) Init(c *ctx.Context, originId values.HeroId, roleLvl values.Integer) (values.Integer, *errmsg.ErrMsg) {
	originIdx := -1
	for idx, v := range s.values.Ht {
		if v != nil && v.OriginId == originId {
			originIdx = idx
		}
	}
	reader := rule.MustGetReader(c)
	configIds := reader.DeriveHeroMap(originId)
	var originHeroId values.HeroId = 0
	if len(configIds) > 0 {
		originHeroId = configIds[0]
	}
	if originIdx == -1 {
		slot, has := reader.KeyValue.GetMapInt64Int64("RoleSkillSlot")
		if !has {
			return 0, errmsg.NewErrTalentIllegal()
		}
		rowHero, has := reader.RowHero.GetRowHeroById(originHeroId)
		if !has {
			return 0, errmsg.NewErrTalentIllegal()
		}
		ht := &models.HeroTalent{
			OriginId:     originId,
			ConfigId:     originHeroId,
			ChosenSkills: make([]values.Integer, len(slot)),
			Each:         make([]*models.TalentPlate, 0, len(rowHero.TalentPlateId)),
		}
		for _, plateId := range rowHero.TalentPlateId {
			plate, has := reader.Talentplate.GetTalentplateById(plateId)
			if !has {
				return 0, errmsg.NewErrTalentIllegal()
			}
			if len(plate.ShapeData) != len(plate.UnlockData) {
				return 0, errmsg.NewErrTalentIllegal()
			}
			each := make(map[values.Integer]*models.TalentDetail)
			unlockNum := values.Integer(0)
			for idx, shape := range plate.ShapeData {
				isUnlock := false
				if plate.UnlockData[idx] == 1 {
					isUnlock = true
					unlockNum++
				}
				loc := s.posToInt(shape)
				each[loc] = &models.TalentDetail{
					Loc:      loc,
					IsUnlock: isUnlock,
				}
			}
			ht.Each = append(ht.Each, &models.TalentPlate{
				UnlockNum: unlockNum,
				Plate:     each,
			})
		}
		s.values.Ht = append(s.values.Ht, ht)
		originIdx = len(s.values.Ht) - 1
	}
	skillLvl, _ := s.SkillHandleRoleLvlUp(c, roleLvl)
	return skillLvl, nil
}

func (s *talent) posToInt(pos []values.Integer) values.Integer {
	return pos[0]*1000 + pos[1]
}

func (s *talent) intToPos(loc values.Integer) []values.Integer {
	x := loc / 1000
	y := loc % 1000
	if y >= 500 {
		y -= 1000
		x++
	} else if y <= -500 {
		y += 1000
		x--
	}
	return []values.Integer{x, y}
}

func (s *talent) RuneLevelUp(c *ctx.Context, rune *models.TalentRune, oblation []*models.TalentRune) (values.Integer, []values.RuneId, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	talentRule, has := reader.Talent.GetTalentById(rune.TalentId)
	if !has {
		return 0, nil, errmsg.NewErrTalentNotExist()
	}
	lvlUpCfg := s.getTalentLvlUpCfg(c, talentRule)
	if lvlUpCfg == nil {
		return 0, nil, errmsg.NewErrTalentIllegal()
	}
	getExpCfg, has := reader.KeyValue.GetIntegerArray("TalentRuneGetExp")
	if !has {
		return 0, nil, errmsg.NewErrTalentIllegal()
	}
	exchangeCfg, has := reader.KeyValue.GetInt64("TalentRuneLevelExchangeExp")
	if !has {
		return 0, nil, errmsg.NewErrTalentIllegal()
	}
	delIds := make([]values.RuneId, 0, len(oblation))
	for _, ob := range oblation {
		obRule, ok := reader.Talent.GetTalentById(ob.TalentId)
		if !ok {
			return 0, nil, errmsg.NewErrTalentNotExist()
		}
		expCfgIdx := len(obRule.RuneShape) - 1
		if expCfgIdx >= len(getExpCfg) {
			return 0, nil, errmsg.NewErrTalentIllegal()
		}
		exp := (getExpCfg[expCfgIdx] + ob.AccExp) * exchangeCfg / 10000
		for {
			if rune.Lvl == talentRule.TalentLevelUpperLimit {
				break
			}
			var lvlUpNeed values.Integer
			if rune.Lvl-1 >= values.Integer(len(lvlUpCfg)) {
				lvlUpNeed = lvlUpCfg[len(lvlUpCfg)-1]
			} else {
				lvlUpNeed = lvlUpCfg[rune.Lvl-1]
			}
			lvlUpNeed = lvlUpNeed - rune.CurrExp
			if exp < lvlUpNeed {
				// 不足升级
				rune.CurrExp += exp
				rune.AccExp += exp
				break
			}
			// 升到下一级
			rune.Lvl++
			exp -= lvlUpNeed
			rune.CurrExp = 0
			rune.AccExp += lvlUpNeed
		}
		delIds = append(delIds, ob.RuneId)
		if rune.Lvl == talentRule.TalentLevelUpperLimit {
			break
		}
	}
	if rune.PlateIdx > 0 {
		for _, v := range s.values.Ht {
			if v != nil && v.ConfigId == talentRule.BuildId {
				if rune.PlateIdx-1 >= values.Integer(len(v.Each)) {
					return 0, nil, errmsg.NewErrPlateNotEnough()
				}
				plate := v.Each[rune.PlateIdx-1]
				for _, shape := range talentRule.RuneShape {
					if r, exist := plate.Plate[rune.Loc+s.posToInt(shape)]; exist {
						r.Lvl = rune.Lvl
					}
				}
			}
		}
	}
	return talentRule.BuildId, delIds, nil
}

func (s *talent) RuneLevelUpByDust(c *ctx.Context, rune *models.TalentRune, cnt values.Integer) (values.Integer, values.Integer, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	talentRule, has := reader.Talent.GetTalentById(rune.TalentId)
	if !has {
		return 0, 0, errmsg.NewErrTalentNotExist()
	}
	lvlUpCfg := s.getTalentLvlUpCfg(c, talentRule)
	if lvlUpCfg == nil {
		return 0, 0, errmsg.NewErrTalentIllegal()
	}
	exp := cnt
	for {
		if rune.Lvl == talentRule.TalentLevelUpperLimit {
			break
		}
		var lvlUpNeed values.Integer
		if rune.Lvl-1 >= values.Integer(len(lvlUpCfg)) {
			lvlUpNeed = lvlUpCfg[len(lvlUpCfg)-1]
		} else {
			lvlUpNeed = lvlUpCfg[rune.Lvl-1]
		}
		lvlUpNeed = lvlUpNeed - rune.CurrExp
		if exp < lvlUpNeed {
			// 不足升级
			rune.CurrExp += exp
			rune.AccExp += exp
			exp = 0
			break
		}
		// 升到下一级
		rune.Lvl++
		exp -= lvlUpNeed
		rune.CurrExp = 0
		rune.AccExp += lvlUpNeed
	}
	if rune.PlateIdx > 0 {
		for _, v := range s.values.Ht {
			if v != nil && v.ConfigId == talentRule.BuildId {
				if rune.PlateIdx-1 >= values.Integer(len(v.Each)) {
					return 0, 0, errmsg.NewErrPlateNotEnough()
				}
				plate := v.Each[rune.PlateIdx-1]
				for _, shape := range talentRule.RuneShape {
					if r, exist := plate.Plate[rune.Loc+s.posToInt(shape)]; exist {
						r.Lvl = rune.Lvl
					}
				}
			}
		}
	}
	return talentRule.BuildId, cnt - exp, nil
}

func (s *talent) InlayRune(c *ctx.Context, configId, plateIdx, loc values.Integer, rune *models.TalentRune, isInlay bool) (*models.TalentPlate, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RowHero.GetRowHeroById(configId)
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	if isInlay {
		// 镶嵌
		if rune.PlateIdx > 0 {
			return nil, errmsg.NewErrPlateAlreadyInlay()
		}
		talentRule, has := reader.Talent.GetTalentById(rune.TalentId)
		if !has {
			return nil, errmsg.NewErrTalentNotExist()
		}
		for _, v := range s.values.Ht {
			if v != nil && v.OriginId == cfg.OriginId {
				if plateIdx-1 >= values.Integer(len(v.Each)) {
					return nil, errmsg.NewErrPlateNotEnough()
				}
				plate := v.Each[plateIdx-1]
				for _, shape := range talentRule.RuneShape {
					currLoc := loc + s.posToInt(shape)
					if block, exist := plate.Plate[currLoc]; !exist {
						return nil, errmsg.NewErrPlateNotEnough()
					} else {
						if block.RuneId != "" {
							return nil, errmsg.NewErrPlateAlreadyInlay()
						}
						block.RuneId = rune.RuneId
						block.TalentId = rune.TalentId
						block.AnchorLoc = loc
						block.Lvl = rune.Lvl
						rune.PlateIdx = plateIdx
						rune.Loc = loc
					}
				}
				return plate, nil
			}
		}
		return nil, errmsg.NewErrHeroNotFound()
	} else {
		// 取
		if rune.PlateIdx == 0 {
			return nil, errmsg.NewErrPlateNotInlay()
		}
		for _, v := range s.values.Ht {
			if v != nil && v.OriginId == cfg.OriginId {
				if plateIdx-1 >= values.Integer(len(v.Each)) {
					return nil, errmsg.NewErrPlateNotEnough()
				}
				plate := v.Each[plateIdx-1]
				if block, exist := plate.Plate[loc]; !exist {
					return nil, errmsg.NewErrPlateNotEnough()
				} else {
					if block.RuneId == "" || block.RuneId != rune.RuneId {
						return nil, errmsg.NewErrPlateNotInlay()
					}
					talentRule, has := reader.Talent.GetTalentById(block.TalentId)
					if !has {
						return nil, errmsg.NewErrTalentNotExist()
					}
					anchorLoc := block.AnchorLoc
					for _, shape := range talentRule.RuneShape {
						currLoc := anchorLoc + s.posToInt(shape)
						if currBlock, ok := plate.Plate[currLoc]; !ok {
							return nil, errmsg.NewErrPlateNotEnough()
						} else {
							if currBlock.RuneId == "" || currBlock.RuneId != rune.RuneId {
								return nil, errmsg.NewErrPlateNotInlay()
							}
							currBlock.RuneId = ""
							currBlock.TalentId = 0
							currBlock.AnchorLoc = 0
							currBlock.Lvl = 0
							rune.PlateIdx = 0
							rune.Loc = 0
						}
					}
				}
				return plate, nil
			}
		}
		return nil, errmsg.NewErrHeroNotFound()
	}
}

func (s *talent) UnlockPlate(c *ctx.Context, configId, plateIdx, loc values.Integer) (*models.HeroTalent, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RowHero.GetRowHeroById(configId)
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	for _, v := range s.values.Ht {
		if v != nil && v.OriginId == cfg.OriginId {
			if plateIdx-1 >= values.Integer(len(v.Each)) {
				return nil, errmsg.NewErrPlateNotEnough()
			}
			plate := v.Each[plateIdx-1]
			if block, exist := plate.Plate[loc]; !exist {
				return nil, errmsg.NewErrPlateNotEnough()
			} else {
				if block.IsUnlock {
					return nil, errmsg.NewErrPlateAlreadyUnlock()
				}
				plateUnlockCfg, ok := reader.KeyValue.GetIntegerArray("TalentGridUnlock")
				if !ok {
					return nil, errmsg.NewErrTalentIllegal()
				}
				var cost values.Integer
				if plate.UnlockNum >= values.Integer(len(plateUnlockCfg)) {
					cost = plateUnlockCfg[len(plateUnlockCfg)-1]
				} else {
					cost = plateUnlockCfg[plate.UnlockNum]
				}
				if cost+v.UsedPoints > s.values.CommonPoints+v.ParticularPoints {
					return nil, errmsg.NewErrPointsNotEnough()
				}
				block.IsUnlock = true
				v.UsedPoints += cost
				plate.UnlockNum++
			}
			return v, nil
		}
	}
	return nil, errmsg.NewErrHeroNotFound()
}

func (s *talent) getTalentLvlUpCfg(c *ctx.Context, talentRule *rule_model.Talent) []values.Integer {
	reader := rule.MustGetReader(c)
	var res []values.Integer
	var has bool
	switch len(talentRule.RuneShape) {
	case 1:
		res, has = reader.KeyValue.GetIntegerArray("TalentRuneUpgradeCostWhite")
		if !has {
			return nil
		}
	case 2:
		res, has = reader.KeyValue.GetIntegerArray("TalentRuneUpgradeCostBlue")
		if !has {
			return nil
		}
	case 3:
		res, has = reader.KeyValue.GetIntegerArray("TalentRuneUpgradeCostPurple")
		if !has {
			return nil
		}
	case 4:
		res, has = reader.KeyValue.GetIntegerArray("TalentRuneUpgradeCostOrange")
		if !has {
			return nil
		}
	default:
		return nil
	}
	return res
}

func (s *talent) SkillLevelUp(c *ctx.Context, configId, skillId, lvl values.Integer) (map[values.Integer]values.Integer, *models.SkillDetail, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	roleSkillRule, has := reader.RoleSkill.GetRoleSkillById(skillId)
	if !has {
		return nil, nil, errmsg.NewErrTalentNotExist()
	}
	cfg, ok := reader.RowHero.GetRowHeroById(configId)
	if !ok {
		return nil, nil, errmsg.NewErrHeroNotFound()
	}
	for _, v := range s.values.Ht {
		if v != nil && v.OriginId == cfg.OriginId {
			if v.Skills == nil {
				v.Skills = map[int64]*models.SkillDetail{}
			}
			if v.ConfigId != roleSkillRule.RoleId {
				return nil, nil, errmsg.NewErrTalentIllegal()
			}
			if skill, exist := v.Skills[skillId]; !exist {
				// 未激活
				return nil, nil, errmsg.NewErrTalentNotExist()
			} else {
				if skill.Lvl == roleSkillRule.LevelLimit {
					return nil, nil, errmsg.NewErrTalentNoLevel()
				}
				skill.Lvl++
				skillLvCfg, ok := reader.SkillLv.GetSkillLvById(roleSkillRule.Id, skill.Lvl)
				if !ok {
					return nil, nil, errmsg.NewErrTalentNotExist()
				}
				for preSkill, lv := range skillLvCfg.UnlockSkillLv {
					if existSkill, exist := v.Skills[preSkill]; !exist {
						return nil, nil, errmsg.NewErrPreTalentNotActive()
					} else if existSkill.Lvl < lv {
						return nil, nil, errmsg.NewErrPreTalentNotActive()
					}
				}
				if lvl < skillLvCfg.UnlockLv {
					return nil, nil, errmsg.NewErrRoleLevelNotEnough()
				}
				return skillLvCfg.TalentPreconditions, skill, nil
			}

		}
	}
	return nil, nil, errmsg.NewErrHeroNotFound()
}

func (s *talent) SkillHandleRoleLvlUp(c *ctx.Context, roleLvl values.Integer) (values.Integer, []*models.HeroTalent) {
	reader := rule.MustGetReader(c)
	roleSkillList := reader.RoleSkill.List()
	var skillLvlUp = values.Integer(0)
	var changeMap = map[values.HeroId]bool{}
	for _, roleSkill := range roleSkillList {
		if len(roleSkill.SkillHoleUnlockRoleLv) != len(roleSkill.SkillHoleId) {
			continue
		}
		skillLv, has := reader.SkillLv.GetSkillLvById(roleSkill.Id, 1)
		if !has {
			continue
		}
		if skillLv.UnlockLv <= roleLvl {
			for _, v := range s.values.Ht {
				if v != nil && v.ConfigId == roleSkill.RoleId {
					if v.Skills == nil {
						v.Skills = map[int64]*models.SkillDetail{}
					}
					if _, exist := v.Skills[roleSkill.Id]; exist {
						break
					}
					canUnlock := true
					for preSkillId, preLvl := range skillLv.UnlockSkillLv {
						if preDetail, exist := v.Skills[preSkillId]; !exist {
							canUnlock = false
							break
						} else if preDetail.Lvl < preLvl {
							canUnlock = false
							break
						}
					}
					if !canUnlock {
						break
					}
					sd := &models.SkillDetail{
						SkillId: roleSkill.Id,
						Lvl:     1,
						Holes:   make([]*models.Hole, len(roleSkill.SkillHoleId)),
					}
					for idx, gemTyp := range roleSkill.SkillHoleId {
						sd.Holes[idx] = &models.Hole{
							HoleType: gemTyp,
						}
					}
					v.Skills[roleSkill.Id] = sd
					skillLvlUp++
					changeMap[v.OriginId] = true
					break
				}
			}
		}
	}
	if len(changeMap) > 0 {
		var changeHero = make([]*models.HeroTalent, 0, len(changeMap))
		for _, v := range s.values.Ht {
			if v != nil && changeMap[v.OriginId] {
				changeHero = append(changeHero, v)
			}
		}
		return skillLvlUp, changeHero
	}
	return skillLvlUp, nil
}

func (s *talent) SkillHandleHeroRoleLvlUp(c *ctx.Context, roleLvl, configId values.Integer) (values.Integer, *models.HeroTalent) {
	reader := rule.MustGetReader(c)
	roleSkillList := reader.RoleSkill.List()
	var skillLvlUp = values.Integer(0)
	isChange := false
	for _, roleSkill := range roleSkillList {
		if roleSkill.RoleId != configId {
			continue
		}
		if len(roleSkill.SkillHoleUnlockRoleLv) != len(roleSkill.SkillHoleId) {
			continue
		}
		skillLv, has := reader.SkillLv.GetSkillLvById(roleSkill.Id, 1)
		if !has {
			continue
		}
		if skillLv.UnlockLv <= roleLvl {
			for _, v := range s.values.Ht {
				if v != nil && v.ConfigId == roleSkill.RoleId {
					if v.Skills == nil {
						v.Skills = map[int64]*models.SkillDetail{}
					}
					if _, exist := v.Skills[roleSkill.Id]; exist {
						break
					}
					canUnlock := true
					for preSkillId, preLvl := range skillLv.UnlockSkillLv {
						if preDetail, exist := v.Skills[preSkillId]; !exist {
							canUnlock = false
							break
						} else if preDetail.Lvl < preLvl {
							canUnlock = false
							break
						}
					}
					if !canUnlock {
						break
					}
					sd := &models.SkillDetail{
						SkillId: roleSkill.Id,
						Lvl:     1,
						Holes:   make([]*models.Hole, len(roleSkill.SkillHoleId)),
					}
					for idx, gemTyp := range roleSkill.SkillHoleId {
						sd.Holes[idx] = &models.Hole{
							HoleType: gemTyp,
						}
					}
					v.Skills[roleSkill.Id] = sd
					skillLvlUp++
					isChange = true
					break
				}
			}
		}
	}
	if isChange {
		var changeHero *models.HeroTalent
		for _, v := range s.values.Ht {
			if v != nil && v.ConfigId == configId {
				changeHero = v
				break
			}
		}
		return skillLvlUp, changeHero
	}
	return skillLvlUp, nil
}

func (s *talent) SkillChoose(c *ctx.Context, configId values.Integer, idx, skillId, lvl values.Integer) (values.Integer, *models.HeroTalent, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RowHero.GetRowHeroById(configId)
	if !ok {
		return 0, nil, errmsg.NewErrHeroNotFound()
	}
	for _, v := range s.values.Ht {
		if v != nil && v.OriginId == cfg.OriginId {
			if v.ChosenSkills == nil || len(v.ChosenSkills) < int(idx) {
				return 0, nil, errmsg.NewErrSkillEquipNotActive()
			}
			slot, has := reader.KeyValue.GetMapInt64Int64("RoleSkillSlot")
			if !has {
				return 0, nil, errmsg.NewErrTalentIllegal()
			}
			if lvl < slot[idx+1] {
				return 0, nil, errmsg.NewErrRoleLevelNotEnough()
			}
			chooseChangeCnt := values.Integer(0)
			if skillId == 0 {
				if v.ChosenSkills[int(idx)] > 0 {
					chooseChangeCnt = -1
				}
				v.ChosenSkills[int(idx)] = 0
				return chooseChangeCnt, v, nil
			}
			if _, exist := v.Skills[skillId]; !exist {
				return 0, v, errmsg.NewErrTalentNotExist()
			}
			if v.ChosenSkills[int(idx)] == 0 {
				chooseChangeCnt = 1
			}
			v.ChosenSkills[int(idx)] = skillId
			return chooseChangeCnt, v, nil
		}
	}
	return 0, nil, errmsg.NewErrHeroNotFound()
}

func (s *talent) Reset(c *ctx.Context, configId, plateIdx values.Integer) (*models.TalentPlate, []values.RuneId, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RowHero.GetRowHeroById(configId)
	if !ok {
		return nil, nil, errmsg.NewErrHeroNotFound()
	}
	for _, v := range s.values.Ht {
		if v != nil && v.OriginId == cfg.OriginId {
			if plateIdx-1 >= values.Integer(len(v.Each)) {
				return nil, nil, errmsg.NewErrPlateNotEnough()
			}
			plate := v.Each[plateIdx-1]
			runeIds := make([]values.RuneId, 0, len(plate.Plate))
			for _, block := range plate.Plate {
				if block.RuneId != "" {
					runeIds = append(runeIds, block.RuneId)
					block.RuneId = ""
					block.TalentId = 0
					block.AnchorLoc = 0
				}
			}
			return plate, runeIds, nil
		}
	}
	return nil, nil, errmsg.NewErrHeroNotFound()
}

func (s *talent) GainCommonP(num values.Integer) {
	s.values.CommonPoints += num
}

func (s *talent) GainParticularP(configId, num values.Integer) {
	for _, v := range s.values.Ht {
		if v != nil && v.OriginId == configId {
			v.ParticularPoints += num
			return
		}
	}
	s.values.Ht = append(s.values.Ht, &models.HeroTalent{
		OriginId:         configId,
		ParticularPoints: num,
	})
}

func (s *talent) IsFirstUpdate(talentId, lvl values.Integer) bool {
	if oldLvl, exist := s.values.FirstUpdateM[talentId]; exist {
		if lvl > oldLvl {
			return true
		}
		return false
	}
	return true
}

func (s *talent) FirstUpdate(talentId, lvl values.Integer) {
	if lvl > s.values.FirstUpdateM[talentId] {
		s.values.FirstUpdateM[talentId] = lvl
	}
}

func (s *talent) InlayStone(c *ctx.Context, configId values.Integer, skillRule *rule_model.RoleSkill, stoneRule *rule_model.Skillstone, holeIdx int) (values.Integer, *models.SkillDetail, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RowHero.GetRowHeroById(configId)
	if !ok {
		return 0, nil, errmsg.NewErrHeroNotFound()
	}
	for _, v := range s.values.Ht {
		if v != nil && v.OriginId == cfg.OriginId {
			if t, exist := v.Skills[skillRule.Id]; !exist {
				return 0, nil, errmsg.NewErrTalentNotExist()
			} else {
				if holeIdx > len(t.Holes) {
					return 0, nil, errmsg.NewErrHoleNotMatch()
				}
				if holeIdx < len(skillRule.SkillHoleUnlockRoleLv) && t.Lvl < skillRule.SkillHoleUnlockRoleLv[holeIdx] {
					return 0, nil, errmsg.NewErrRoleLevelNotEnough()
				}
				if t.Holes[holeIdx].HoleType != AnyStoneSuitTyp && t.Holes[holeIdx].HoleType != stoneRule.SkillStoneShape {
					return 0, nil, errmsg.NewErrHoleNotMatch()
				}
				currStoneId := t.Holes[holeIdx].StoneId
				t.Holes[holeIdx].StoneId = stoneRule.Id
				return currStoneId, t, nil
			}
		}
	}
	return 0, nil, errmsg.NewErrHeroNotFound()
}

func (s *talent) RemoveStone(c *ctx.Context, configId, skillId values.Integer, holeIdx int) (values.Integer, *models.SkillDetail, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RowHero.GetRowHeroById(configId)
	if !ok {
		return 0, nil, errmsg.NewErrHeroNotFound()
	}
	for _, v := range s.values.Ht {
		if v != nil && v.OriginId == cfg.OriginId {
			if t, exist := v.Skills[skillId]; !exist {
				return 0, nil, errmsg.NewErrTalentNotExist()
			} else {
				if holeIdx > len(t.Holes) {
					return 0, nil, errmsg.NewErrHoleNotMatch()
				}
				if t.Holes[holeIdx].StoneId == 0 {
					return 0, nil, errmsg.NewErrStoneNotInHole()
				}
				id := t.Holes[holeIdx].StoneId
				t.Holes[holeIdx].StoneId = 0
				return id, t, nil
			}
		}
	}
	return 0, nil, errmsg.NewErrHeroNotFound()
}

func (s *talent) RemoveAllStone(c *ctx.Context, configId, skillId values.Integer) (map[values.ItemId]values.Integer, *models.SkillDetail, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RowHero.GetRowHeroById(configId)
	if !ok {
		return nil, nil, errmsg.NewErrHeroNotFound()
	}
	for _, v := range s.values.Ht {
		if v != nil && v.OriginId == cfg.OriginId {
			if t, exist := v.Skills[skillId]; !exist {
				return nil, nil, errmsg.NewErrTalentNotExist()
			} else {
				m := map[values.ItemId]values.Integer{}
				for holeIdx := range t.Holes {
					if t.Holes[holeIdx].StoneId != 0 {
						stoneRule, has := rule.MustGetReader(c).Skillstone.GetSkillstoneById(t.Holes[holeIdx].StoneId)
						if !has {
							continue
						}
						m[stoneRule.ItemId] += 1
						t.Holes[holeIdx].StoneId = 0
					}
				}
				return m, t, nil
			}
		}
	}
	return nil, nil, errmsg.NewErrHeroNotFound()
}

func (s *talent) LockList() []values.Integer {
	if len(s.values.LockStoneIds) > 0 {
		res := make([]values.Integer, 0, len(s.values.LockStoneIds))
		for id := range s.values.LockStoneIds {
			res = append(res, id)
		}
		return res
	}
	return nil
}

func (s *talent) Lock(stoneId values.Integer) {
	if s.values.LockStoneIds == nil {
		s.values.LockStoneIds = map[int64]bool{}
	}
	s.values.LockStoneIds[stoneId] = true
}

func (s *talent) Unlock(stoneId values.Integer) {
	if s.values.LockStoneIds != nil {
		delete(s.values.LockStoneIds, stoneId)
	}
}

func (s *talent) IsLock(stoneId values.Integer) bool {
	if s.values.LockStoneIds == nil {
		return false
	}
	return s.values.LockStoneIds[stoneId]
}

func (s *talent) UpdateExtraLvl(c *ctx.Context, heroId values.HeroId, data map[values.HeroSkillId]values.Level) map[values.HeroBuildId]map[values.HeroSkillId]values.Level {
	handled := map[values.HeroBuildId]map[values.HeroSkillId]values.Level{}
	reader := rule.MustGetReader(c)
	for skillId, lvl := range data {
		_, ok := reader.Skill.GetSkillById(skillId)
		if !ok {
			continue
		}
		if _, exist := handled[heroId]; !exist {
			handled[heroId] = map[values.HeroSkillId]values.Level{}
		}
		handled[heroId][skillId] = lvl
	}
	for _, t := range s.values.Ht {
		if extraMap, exist := handled[t.OriginId]; exist {
			for skillId, lvl := range extraMap {
				if t.ExtraSkills == nil {
					t.ExtraSkills = map[int64]int64{}
				}
				t.ExtraSkills[skillId] += lvl
			}
		}
	}
	return handled
}

func (s *talent) CalAllSkill(c *ctx.Context) (map[values.HeroId]*models.SkillAdvance, *errmsg.ErrMsg) {
	res := map[values.HeroId]*models.SkillAdvance{}
	for _, t := range s.values.Ht {
		eachRes, err := s.CalSkill(c, t.ConfigId)
		if err != nil {
			return nil, err
		}
		res[t.ConfigId] = eachRes
	}
	return res, nil
}

func (s *talent) CalSkill(c *ctx.Context, configId values.Integer) (*models.SkillAdvance, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RowHero.GetRowHeroById(configId)
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	for _, t := range s.values.Ht {
		if t.OriginId == cfg.OriginId {
			res := &models.SkillAdvance{
				Skills: map[int64]*models.SkillStoneAdvance{},
				StoneAttr: &models.TalentAttr{
					AttrFixed:   map[int64]int64{},
					AttrPercent: map[int64]int64{},
				},
			}
			for _, skillId := range t.ChosenSkills {
				skill, exist := t.Skills[skillId]
				if !exist || skill == nil || skill.Lvl == 0 {
					continue
				}
				ssa := &models.SkillStoneAdvance{
					Stones: make([]values.Integer, 0, len(skill.Holes)),
				}
				for _, h := range skill.Holes {
					if h.StoneId > 0 {
						ssa.Stones = append(ssa.Stones, h.StoneId)
						stoneCfg, ok := reader.Skillstone.GetSkillstoneById(h.StoneId)
						if ok {
							for _, attr := range stoneCfg.SkillStoneAttrFixed {
								if len(attr) == 2 {
									res.StoneAttr.AttrFixed[attr[0]] += attr[1]
								}
							}
							for _, attr := range stoneCfg.SkillStoneAttrPercentage {
								if len(attr) == 2 {
									res.StoneAttr.AttrPercent[attr[0]] += attr[1]
								}
							}
							if stoneCfg.PowerNum > 0 {
								res.SpecialStoneAtk += stoneCfg.PowerNum
							}
						}
					}
				}
				res.Skills[skill.SkillId+skill.Lvl-1] = ssa
			}
			soulSkillId := reader.GetSoulSkill()[configId]
			if soulSkillId > 0 {
				soulSkill, exist := t.Skills[soulSkillId]
				if exist && soulSkill != nil && soulSkill.Lvl > 0 {
					res.Skills[soulSkill.SkillId+soulSkill.Lvl-1] = &models.SkillStoneAdvance{}
				}
			}
			unusedSkills := map[values.Integer]values.Integer{}
			for skillId, skillDetail := range t.Skills {
				if _, exist := res.Skills[skillId]; !exist {
					unusedSkills[skillId] = skillDetail.Lvl
				}
			}
			if res.SpecialStoneAtk == 0 {
				res.SpecialStoneAtk = -1
			}
			res.UnusedSkills = unusedSkills
			return res, nil
		}
	}
	return nil, nil
}

func (s *talent) CalAllUnusedSkill(c *ctx.Context) (map[values.HeroId]map[values.Integer]values.Integer, *errmsg.ErrMsg) {
	res := map[values.HeroId]map[values.Integer]values.Integer{}
	for _, t := range s.values.Ht {
		each, err := s.CalUnusedSkill(c, t.OriginId)
		if err != nil {
			return nil, err
		}
		res[t.OriginId] = each
	}
	return res, nil
}

func (s *talent) CalUnusedSkill(c *ctx.Context, configId values.Integer) (map[values.Integer]values.Integer, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	for _, t := range s.values.Ht {
		if t.OriginId == configId {
			used := map[values.Integer]bool{}
			for _, skillId := range t.ChosenSkills {
				skill, exist := t.Skills[skillId]
				if !exist || skill == nil || skill.Lvl == 0 {
					continue
				}
				used[skillId] = true
			}
			soulSkillId := reader.GetSoulSkill()[configId]
			unusedSkills := map[values.Integer]values.Integer{}
			for skillId, skillDetail := range t.Skills {
				if _, exist := used[skillId]; !exist && skillId != soulSkillId {
					unusedSkills[skillId] = skillDetail.Lvl
				}
			}
			return unusedSkills, nil
		}
	}
	return nil, nil
}

func (s *talent) CalAttr(c *ctx.Context, configId values.Integer) (*models.TalentAdvance, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RowHero.GetRowHeroById(configId)
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	for _, t := range s.values.Ht {
		if t.OriginId == cfg.OriginId {
			res := &models.TalentAdvance{
				Attr: &models.TalentAttr{
					AttrFixed:   map[int64]int64{},
					AttrPercent: map[int64]int64{},
				},
				TalentBuff: make([]values.Integer, 0, 8),
			}
			for _, each := range t.Each {
				for _, detail := range each.Plate {
					if detail.RuneId != "" {
						tRule, ok := reader.Talent.GetTalentById(detail.TalentId)
						if !ok {
							return nil, errmsg.NewErrTalentNotExist()
						}
						if len(tRule.TalentSkillBuff) > 0 {
							for _, buffId := range tRule.TalentSkillBuff {
								realBuffId := buffId + detail.Lvl - 1
								_, ok = reader.Buff.GetBuffById(realBuffId)
								if ok {
									res.TalentBuff = append(res.TalentBuff, realBuffId)
								}
							}
						}
						if tRule.TalentAttrId != 0 {
							atRule, has := reader.TalentAttr.GetTalentAttrById(tRule.TalentAttrId + detail.Lvl - 1)
							if has {
								for _, fix := range atRule.TalentAttrFixed {
									if len(fix) == 2 {
										res.Attr.AttrFixed[fix[0]] += fix[1]
									}
								}
								for _, per := range atRule.TalentAttrFixedPercentage {
									if len(per) == 2 {
										res.Attr.AttrPercent[per[0]] += per[1]
									}
								}
							}
						}
					}
				}
			}
			return res, nil
		}
	}
	return nil, nil
}

func (s *talent) ToDao() *dao.Talent {
	return s.values
}
