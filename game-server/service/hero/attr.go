/*
	英雄战斗力计算
*/
package hero

import (
	"math"

	"coin-server/common/ctx"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/hero/rule"
	rulemodel "coin-server/rule/rule-model"

	"github.com/gogo/protobuf/proto"

	"go.uber.org/zap"
)

// 英雄的战力
func (svc *Service) refreshAllHeroAttr(
	ctx *ctx.Context,
	heroes []*pbdao.Hero,
	level values.Level,
	levelIndex values.Integer,
	base bool,
	equips map[values.EquipId]*models.Equipment,
	roleAttr []*models.RoleAttr,
	roleSkill []values.HeroSkillId,
) {
	svc.refreshAttrFromFashion(ctx, heroes)
	for _, hero := range heroes {
		svc.refreshHeroAttr(ctx, hero, level, levelIndex, base, equips, roleAttr, roleSkill, 0)
	}
}

func (svc *Service) refreshAllHeroAttrByGuildBless(
	ctx *ctx.Context,
	heroes []*pbdao.Hero,
	level values.Level,
	levelIndex values.Integer,
	base bool,
	equips map[values.EquipId]*models.Equipment,
	roleAttr []*models.RoleAttr,
	roleSkill []values.HeroSkillId,
	d *event.BlessActivated,
) {
	svc.refreshAttrFromFashion(ctx, heroes)
	svc.refreshAttrFromGuildBless(ctx, heroes, d)
	for _, hero := range heroes {
		svc.refreshHeroAttr(ctx, hero, level, levelIndex, base, equips, roleAttr, roleSkill, 0)
	}
}

// 单个英雄的战力( 英雄等级对应的属性战力)
func (svc *Service) refreshHeroAttr(
	ctx *ctx.Context,
	hero *pbdao.Hero,
	level values.Level,
	levelIndex values.Integer,
	base bool,
	equips map[values.EquipId]*models.Equipment,
	roleAttr []*models.RoleAttr,
	roleSkill []values.HeroSkillId,
	specialStoneAtk values.Integer,
) {
	if base {
		svc.refreshAttrFromBase(ctx, hero, levelIndex)
	}
	if equips != nil {
		svc.refreshAttrFromEquip(ctx, hero, equips)
	}
	if roleSkill != nil {
		svc.refreshSkillFromRole(ctx, hero, roleSkill)
	}
	if roleAttr != nil {
		svc.refreshAttrFromRole(hero, roleAttr)
	}
	svc.refreshAttrFromSoulContract(ctx, hero)

	priAttr := make(map[values.AttrId]values.Integer)
	secAttr := make(map[values.AttrId]values.Integer)

	priFixedAttrMap := make(map[values.Integer]map[values.AttrId]values.Integer)
	secFixedAttrMap := make(map[values.Integer]map[values.AttrId]values.Integer)
	priPercentAttrMap := make(map[values.Integer]map[values.AttrId]values.Integer)
	secPercentAttrMap := make(map[values.Integer]map[values.AttrId]values.Integer)
	// 需要发个客户端（在面板显示的）百分比属性
	percentAttrMap := make(map[values.Integer]map[values.AttrId]values.Integer)
	// 先计算出各个系统的一二级属性
	for ac, item := range hero.Attrs {
		if ac == primaryAttr || ac == secondaryAttr {
			continue
		}
		// 固定值
		for id, val := range item.Fixed {
			cfg, ok := rule.GetAttrById(ctx, id)
			if !ok {
				svc.log.Warn("attr config not found", zap.Int64("id", id))
				continue
			}
			if cfg.AdvancedType == enum.PrimaryAttr {
				if _, ok := priFixedAttrMap[ac]; !ok {
					priFixedAttrMap[ac] = map[values.AttrId]values.Integer{}
				}
				priFixedAttrMap[ac][id] += val
				priAttr[id] += val
			} else {
				if _, ok := secFixedAttrMap[ac]; !ok {
					secFixedAttrMap[ac] = map[values.AttrId]values.Integer{}
				}
				secFixedAttrMap[ac][id] += val
				secAttr[id] += val
			}
		}
		// 百分比
		for id, val := range item.Percent {
			cfg, ok := rule.GetAttrById(ctx, id)
			if !ok {
				svc.log.Warn("attr config not found", zap.Int64("id", id))
				continue
			}
			if cfg.ShowTpye == enum.Direct {
				if cfg.AdvancedType == enum.PrimaryAttr {
					if _, ok := priPercentAttrMap[ac]; !ok {
						priPercentAttrMap[ac] = map[values.AttrId]values.Integer{}
					}
					priPercentAttrMap[ac][id] += val
				} else {
					if _, ok := secPercentAttrMap[ac]; !ok {
						secPercentAttrMap[ac] = map[values.AttrId]values.Integer{}
					}
					secPercentAttrMap[ac][id] += val
				}
			} else {
				if _, ok := percentAttrMap[ac]; !ok {
					percentAttrMap[ac] = map[values.AttrId]values.Integer{}
				}
				percentAttrMap[ac][id] += val
			}
		}
	}
	// 判断各个系统是否有百分比加成，根据加成算出各个系统的最终属性，先处理一级属性加成，再处理二级属性加成
	// 处理一级属性的百分比加成（获得各个系统最终的一级属性）
	for ac, item := range priPercentAttrMap {
		for id, perVal := range item {
			attrVal := priAttr[id]
			val := values.Integer(math.Ceil(values.Float(perVal) * values.Float(attrVal) / 10000.0))
			if val > 0 {
				if _, ok := priFixedAttrMap[ac]; !ok {
					priFixedAttrMap[ac] = map[values.AttrId]values.Integer{}
				}
				priFixedAttrMap[ac][id] += val
				priAttr[id] += val
			}
		}
	}
	// 将上面的（priFixedAttrMap） 一级属性转化为二级属性
	for ac, item := range priFixedAttrMap {
		for id, val := range item {
			list := rule.GetAttrTransConfigById(ctx, id)
			for _, transItem := range list {
				if !svc.attrTransformVocationCheck(hero.Id, transItem.Limithero) {
					continue
				}
				if _, ok := secFixedAttrMap[ac]; !ok {
					secFixedAttrMap[ac] = map[values.AttrId]values.Integer{}
				}
				if transItem.Transtype == ttPercent {
					if _, ok := secPercentAttrMap[ac]; !ok {
						secPercentAttrMap[ac] = map[values.AttrId]values.Integer{}
					}
					secPercentAttrMap[ac][transItem.TransattrId] += val * transItem.Transnum
				} else {
					v := val * transItem.Transnum
					secFixedAttrMap[ac][transItem.TransattrId] += v
					secAttr[transItem.TransattrId] += v
				}
			}
		}
	}
	// 处理二级属性百分比加成（这一步后各个系统的二级属性为最终值）
	for ac, item := range secPercentAttrMap {
		for id, perVal := range item {
			attrVal := secAttr[id]
			val := values.Integer(math.Ceil(values.Float(perVal) * values.Float(attrVal) / 10000.0))
			if val > 0 {
				if _, ok := secFixedAttrMap[ac]; !ok {
					secFixedAttrMap[ac] = map[values.AttrId]values.Integer{}
				}
				secFixedAttrMap[ac][id] += val
				secAttr[id] += val
			}
		}
	}

	allAttr := make(map[values.Integer]map[values.AttrId]values.Float)
	for ac, item := range priFixedAttrMap {
		if _, ok := allAttr[ac]; !ok {
			allAttr[ac] = map[values.AttrId]values.Float{}
		}
		for id, val := range item {
			allAttr[ac][id] += values.Float(val)
		}
	}
	for ac, item := range secFixedAttrMap {
		if _, ok := allAttr[ac]; !ok {
			allAttr[ac] = map[values.AttrId]values.Float{}
		}
		for id, val := range item {
			allAttr[ac][id] += values.Float(val)
		}
	}
	percentAttr := make(map[values.AttrId]values.Integer) // 不需要转换直接以百分比的方式显示在面板上的（发给前端的值是万分比）
	for ac, item := range percentAttrMap {
		if _, ok := allAttr[ac]; !ok {
			allAttr[ac] = map[values.AttrId]values.Float{}
		}
		for id, val := range item {
			allAttr[ac][id] += values.Float(val)
			percentAttr[id] += val
		}
	}
	// skills := make(map[values.Integer][]values.HeroSkillId)
	// for st, item := range hero.Skill {
	// 	for _, skillId := range item.Skill {
	// 		if _, ok := skills[st]; !ok {
	// 			skills[st] = make([]values.HeroSkillId, 0)
	// 		}
	// 		skills[st] = append(skills[st], skillId)
	// 	}
	// }

	svc.log.Info("attrs", zap.Any("func", proto.MessageName(ctx.Req)), zap.String("roleId", ctx.RoleId), zap.Any("primary attr", priFixedAttrMap), zap.Any("secondary attr", secFixedAttrMap))

	hero.Attrs[primaryAttr] = &models.HeroAttr{Fixed: priAttr, Percent: percentAttr} // 直接显示的百分比属性存在primaryAttr的Percent里
	hero.Attrs[secondaryAttr] = &models.HeroAttr{Fixed: secAttr}

	svc.calcCombatValue(ctx, hero, level, allAttr, equips, specialStoneAtk)
}

// 基础属性（等级带来的）
func (svc *Service) refreshAttrFromBase(ctx *ctx.Context, hero *pbdao.Hero, level values.Level) {
	cfg, ok := rule.GetRoleLvConfigByLv(ctx, level)
	if !ok {
		svc.log.Warn("role_lv config not found", zap.Int64("level", level))
		return
	}
	if len(cfg.ParameterQuality) <= 0 {
		svc.log.Warn("role_lv.ParameterQuality is nil", zap.Int64("level", level))
		return
	}
	if hero.Attrs == nil {
		hero.Attrs = make(map[int64]*models.HeroAttr)
	}
	attrs := make(map[values.Integer]values.Integer, len(cfg.ParameterQuality))
	for id, val := range cfg.ParameterQuality {
		attrs[id] = val
	}
	hero.Attrs[levelAttr] = &models.HeroAttr{Fixed: attrs}
}

func (svc *Service) refreshAttrFromEquip(ctx *ctx.Context, hero *pbdao.Hero, equips map[values.EquipId]*models.Equipment) {
	if len(hero.EquipSlot) <= 0 {
		return
	}
	fixedMap := make(map[values.AttrId]values.Integer)
	percentMap := make(map[values.AttrId]values.Integer)
	equipItemId := make([]values.ItemId, 0)
	tempBuffMap := make(map[values.HeroBuffId]struct{})
	buff := make([]values.HeroBuffId, 0)
	for skillId := range hero.Skill {
		hero.Skill[skillId].Equip = map[int64]int64{}
	}
	for slotId, slot := range hero.EquipSlot {
		if slot == nil || slot.EquipItemId == 0 {
			continue
		}
		equipItemId = append(equipItemId, slot.EquipItemId)
		fixed, percent, skillMap := svc.oneEquipAttr(ctx, hero, slot, slotId, equips)
		for id, val := range fixed {
			fixedMap[id] += val
		}
		for id, val := range percent {
			percentMap[id] += val
		}
		for id := range skillMap {
			if _, ok := tempBuffMap[id]; !ok {
				tempBuffMap[id] = struct{}{}
				buff = append(buff, id)
			}
		}
	}
	// 装备共鸣属性
	if hero.Resonance == nil {
		hero.Resonance = map[int64]models.ResonanceStatus{}
	}
	for id, val := range hero.Resonance {
		if val != models.ResonanceStatus_RSActivated {
			continue
		}
		cfg, ok := rule.GetEquipResonance(ctx, id)
		if !ok {
			ctx.Warn("equip_resonance config not found", zap.Int64("id", id))
			continue
		}
		for attrId, attrVal := range cfg.AttributeValue {
			fixedMap[attrId] += attrVal
		}
	}
	hero.Attrs[equipAttr] = &models.HeroAttr{
		Fixed:   fixedMap,
		Percent: percentMap,
	}
	equipSetFixedMap, equipSetPercentMap, equipSetBuff := svc.getEquipSetAttr(ctx, equipItemId)
	hero.Attrs[equipSetAttr] = &models.HeroAttr{
		Fixed:   equipSetFixedMap,
		Percent: equipSetPercentMap,
	}
	if hero.Buff == nil {
		hero.Buff = make(map[int64]*models.HeroBuffItem, 0)
	}
	hero.Buff[equipBuffType] = &models.HeroBuffItem{Buff: buff}
	hero.Buff[equipSetBuffType] = &models.HeroBuffItem{Buff: equipSetBuff}
}

func (svc *Service) refreshSkillFromRole(ctx *ctx.Context, hero *pbdao.Hero, roleSkill []values.HeroSkillId) {
	// if hero.Skill == nil {
	// 	hero.Skill = map[int64]*models.HeroSkillItem{}
	// }
	// hero.Skill[roleSkillType] = &models.HeroSkillItem{Skill: roleSkill}
	for _, skillId := range roleSkill {
		skillCfg, ok := rule.GetSkillById(ctx, skillId)
		if !ok {
			svc.log.Warn("skill config not found", zap.Int64("id", skillCfg.Id))
			continue
		}
		if hero.Skill == nil {
			hero.Skill = map[int64]*pbdao.SkillLevel{}
		}
		skill, ok := hero.Skill[skillCfg.SkillBaseId]
		if !ok {
			skill = &pbdao.SkillLevel{}
		}
		skill.Role = skillId - skillCfg.SkillBaseId
		if skill.Role < 0 {
			skill.Role = 0
		}
		hero.Skill[skillCfg.SkillBaseId] = skill
	}
}

func (svc *Service) refreshAttrFromRole(hero *pbdao.Hero, roleAttr []*models.RoleAttr) {
	if hero.Attrs == nil {
		hero.Attrs = map[int64]*models.HeroAttr{}
	}
	for _, attr := range roleAttr {
		typ := BonusType2AttrType(attr.Typ)
		if typ == 0 {
			svc.log.Error("invalid attr type", zap.Int64("bonus type", values.Integer(attr.Typ)))
			continue
		}
		fixed := make(map[values.AttrId]values.Integer)
		percent := make(map[values.AttrId]values.Integer)
		for _, bonus := range attr.AttrFixed {
			for id, val := range bonus.Attr {
				fixed[id] += val
			}
		}
		for _, bonus := range attr.AttrPercent {
			for id, val := range bonus.Attr {
				percent[id] += val
			}
		}
		for _, bonus := range attr.HeroAttrFixed {
			if bonus.HeroId == hero.Id {
				for id, val := range bonus.Attr {
					fixed[id] += val
				}
			}
		}
		for _, bonus := range attr.HeroAttrPercent {
			if bonus.HeroId == hero.Id {
				for id, val := range bonus.Attr {
					percent[id] += val
				}
			}
		}
		hero.Attrs[typ] = &models.HeroAttr{Fixed: fixed, Percent: percent}
	}
}

func (svc *Service) refreshAttrFromSoulContract(ctx *ctx.Context, hero *pbdao.Hero) {
	if hero.SoulContract == nil {
		return
	}
	cfg, ok := rule.GetSoulContract(ctx, hero.Id, hero.SoulContract.Rank, hero.SoulContract.Level)
	if !ok {
		svc.log.Error("soul contract config not found",
			zap.Int64("heroId", hero.Id),
			zap.Int64("rank", hero.SoulContract.Rank),
			zap.Int64("level", hero.SoulContract.Level))
		return
	}
	if hero.Attrs == nil {
		hero.Attrs = map[int64]*models.HeroAttr{}
	}
	hero.Attrs[soulContractAttr] = &models.HeroAttr{Fixed: cfg.RankAttrReward}
}

func (svc *Service) refreshAttrFromFashion(ctx *ctx.Context, heroes []*pbdao.Hero) {
	now := timer.StartTime(ctx.StartTime).Unix()
	attrs := make(map[int64]int64)
	for _, hero := range heroes {
		for id, expire := range hero.Fashion.Data {
			if expire == -1 || (expire > 0 && expire > now) {
				cfg, ok := rule.GetFashion(ctx, id)
				if !ok {
					ctx.Error("fashion config not found", zap.String("role_id", ctx.RoleId), zap.Int64("id", id))
				}
				// 这里的配置都是固定值
				for attrId, val := range cfg.Attr {
					attrs[attrId] += val
				}
			}
		}
	}
	for i := 0; i < len(heroes); i++ {
		heroes[i].Attrs[fashionAttr] = &models.HeroAttr{
			Fixed: attrs,
		}
	}
}

func (svc *Service) refreshAttrFromGuildBless(ctx *ctx.Context, heroes []*pbdao.Hero, d *event.BlessActivated) {
	attrs := make(map[values.AttrId]values.Integer)
	// temp2 := make(map[values.AttrId]values.Integer)
	for _, id := range d.Activated {
		cfg, ok := rule.GetBlessById(ctx, id)
		if !ok {
			ctx.Warn("bless config not found", zap.Int64("id", id))
			continue
		}
		attrs[cfg.AttrId] += cfg.FunctionValue + cfg.AddValue
		// temp2[cfg.AttrId] += cfg.FunctionValue + cfg.AddValue
	}
	// temp := make(map[values.AttrId]values.Integer)
	maxPage := rule.GetMaxBlessPage(ctx)
	cfgs := rule.GetAllBless(ctx)
	for _, cfg := range cfgs {
		v1 := (d.Stage - 1) / maxPage // 商
		v2 := (d.Stage - 1) % maxPage // 余
		n := v1
		if cfg.PageId <= v2 {
			n++
		}
		// temp[cfg.AttrId] += cfg.FunctionValue*n + n*(n-1)*cfg.AddValue/2
		attrs[cfg.AttrId] += cfg.FunctionValue*n + n*(n-1)*cfg.AddValue/2

	}
	for i := 0; i < len(heroes); i++ {
		heroes[i].Attrs[guildBlessAttr] = &models.HeroAttr{
			Fixed: attrs,
		}
	}
}

func (svc *Service) calcCombatValue(
	ctx *ctx.Context,
	hero *pbdao.Hero,
	level values.Level,
	allAttr map[values.Integer]map[values.AttrId]values.Float,
	equips map[values.EquipId]*models.Equipment,
	specialStoneAtk values.Integer,
) {
	lvCfg, ok := rule.GetRoleLvConfigByLv(ctx, level)
	if !ok {
		svc.log.Warn("role_lv config not found", zap.Int64("level", level))
		return
	}

	combatValue := make(map[values.Integer]values.Integer)
	for ac, item := range allAttr {
		for id, val := range item {
			attrCfg, ok := rule.GetAttrById(ctx, id)
			if !ok {
				svc.log.Warn("attr config not found", zap.Int64("id", id))
				continue
			}
			typ := AttrType2CombatValueType(ac)
			if typ == 0 {
				svc.log.Error("invalid combat value type", zap.Int64("attr type", ac))
				continue
			}
			combatValue[typ] += values.Integer(math.Ceil(val * values.Float(attrCfg.PowerNum) * values.Float(lvCfg.PowerNum) / 10000))
		}
	}
	combatValue[values.Integer(models.CombatValueType_CVTEquip)] += svc.calcEquipSkillCombatValue(ctx, equips, hero, level)
	combatValue[values.Integer(models.CombatValueType_CVTSkill)] += svc.calcTalentSkillCombatValue(ctx, hero, lvCfg)
	var lastSpecialStoneAtk values.Integer
	if hero.CombatValue.Details != nil {
		lastSpecialStoneAtk = hero.CombatValue.Details[values.Integer(models.CombatValueType_CVTSpecialStoneAtk)]
	}
	// >0 表示装了特殊技能石
	if specialStoneAtk > 0 {
		combatValue[values.Integer(models.CombatValueType_CVTSpecialStoneAtk)] = specialStoneAtk
	} else if specialStoneAtk == -1 { // 和天赋那边约定好，如果穿过来的是-1，就表示特殊技能石带来的战斗时为0
		combatValue[values.Integer(models.CombatValueType_CVTSpecialStoneAtk)] = 0
	} else { // 其他情况用玩家当前身上的技能石的值
		combatValue[values.Integer(models.CombatValueType_CVTSpecialStoneAtk)] = lastSpecialStoneAtk
	}

	if hero.CombatValue == nil {
		hero.CombatValue = &models.CombatValue{}
	}
	var total values.Integer
	for _, val := range combatValue {
		total += val
	}
	// 技能战斗力包含特殊技能石的战斗力
	combatValue[values.Integer(models.CombatValueType_CVTSkill)] += combatValue[values.Integer(models.CombatValueType_CVTSpecialStoneAtk)]
	hero.CombatValue.Total = total
	hero.CombatValue.Details = combatValue
}

func (svc *Service) calcEquipSkillCombatValue(ctx *ctx.Context, equips map[values.EquipId]*models.Equipment, hero *pbdao.Hero, level values.Level) values.Integer {
	var cv values.Integer
	for id, equip := range equips {
		star, ok := svc.getSlotStart(hero, id)
		if !ok {
			continue
		}
		equipCfg, ok := rule.GetEquipByItemId(ctx, equip.ItemId)
		if !ok {
			svc.log.Warn("equip config not found", zap.Int64("id", equip.ItemId))
			continue
		}
		cv += svc.CalEquipScore(ctx, equip, star, level, equipCfg, hero.Id)
	}
	return cv
}

func (svc *Service) calcTalentSkillCombatValue(ctx *ctx.Context, hero *pbdao.Hero, lvCfg *rulemodel.RoleLv) values.Integer {
	var val values.Float
	for skillId, levelInfo := range hero.Skill {
		max := rule.GetMaxSkillId(ctx, skillId)
		skill := skillId
		if levelInfo != nil {
			skill += levelInfo.Talent
		}
		if skill > max {
			skill = max
		}
		skillCfg, ok := rule.GetSkillById(ctx, skill)
		if !ok {
			svc.log.Warn("skill config not found", zap.Int64("id", skill))
			continue
		}
		val += values.Float(skillCfg.PowerNum)
	}
	cv := values.Integer(math.Ceil(values.Float(lvCfg.SkillPowerNum) / 10000 * val))
	// buff带来的战斗力也属于归为技能里面
	for _, buff := range hero.TalentBuff {
		buffCfg, ok := rule.GetBuffById(ctx, buff)
		if !ok {
			svc.log.Warn("buff config not found", zap.Int64("id", buff))
			continue
		}
		cv += buffCfg.PowerNum
	}
	// 未装备技能的战力
	skills, err := svc.GetUnused(ctx, hero.Id)
	if err != nil {
		ctx.Error("GetUnused err", zap.String("role_id", ctx.RoleId), zap.Int64("hero_id", hero.Id))
	}
	for id, lv := range skills {
		max := rule.GetMaxSkillId(ctx, id)
		id += lv
		if id > max {
			id = max
		}
		skillCfg, ok := rule.GetSkillById(ctx, id)
		if !ok {
			svc.log.Warn("skill config not found", zap.Int64("id", id))
			continue
		}
		val += values.Float(skillCfg.PowerNum)
	}
	cv += values.Integer(math.Ceil(values.Float(lvCfg.SkillPowerNum) / 10000 * val))
	return cv
}

func (svc *Service) getSlotStart(hero *pbdao.Hero, equipId values.EquipId) (values.Integer, bool) {
	for _, slot := range hero.EquipSlot {
		if slot.EquipId == equipId {
			return slot.Star, true
		}
	}
	return 0, false
}
