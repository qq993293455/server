package bag

import (
	"math"
	"math/rand"
	"sort"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/bag/dao"
	"coin-server/game-server/service/bag/rule"
	rule2 "coin-server/game-server/service/bag/rule"
	rulemodel "coin-server/rule/rule-model"

	"go.uber.org/zap"
)

func (s *Service) addEquip(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) ([]*models.Equipment, *errmsg.ErrMsg) {
	// err := checkAndUpdate(ctx, 1)
	// if err != nil {
	//	return err
	// }
	// cfg, ok := rule2.GetEquipById(ctx, itemId)
	// if !ok {
	//	return errmsg.NewInternalErr("equip not found")
	// }
	// equip := &models.Equipment{
	//	EquipId: xid.New().String(),
	//	ItemId:  itemId,
	//	Level:   99, // TODO
	//	Affix:   nil,
	// }
	//
	// itemCfg, ok := rule2.GetItemById(ctx, itemId)
	// if !ok {
	//	return errmsg.NewInternalErr("item not found")
	// }
	// if cfg.AttributeNum > 0 {
	//	equip.Affix = s.genAffix(ctx, cfg, itemCfg.Quality)
	// }
	// dao.AddEquip(ctx, roleId, EquipModel2Dao(equip))
	//
	// ctx.PublishEventLocal(&event.EquipUpdate{
	//	RoleId: roleId,
	//	Equips: []*models.Equipment{equip},
	// })

	return s.addManyEquip(ctx, roleId, map[values.ItemId]values.Integer{itemId: 1})
}

func (s *Service) addManyEquip(ctx *ctx.Context, roleId values.RoleId, items map[values.ItemId]values.Integer) ([]*models.Equipment, *errmsg.ErrMsg) {
	cnt := values.Integer(0)
	for _, v := range items {
		cnt += v
	}
	// err := addBagLength(ctx, cnt)
	// if err != nil {
	// 	return nil, err
	// }
	equipId, err := dao.GetEquipId(ctx, roleId)
	if err != nil {
		return nil, err
	}
	equips := make([]*models.Equipment, 0)
	totalCnt := int64(0)
	for itemId, count := range items {
		for i := 0; i < int(count); i++ {
			equip, _, err := s.GenEquip(ctx, itemId)
			if err != nil {
				return nil, err
			}
			equipId.EquipId++
			equip.EquipId = strconv.Itoa(int(equipId.EquipId))
			equips = append(equips, equip)
			totalCnt++
		}
	}
	dao.SaveManyEquip(ctx, roleId, EquipModels2Dao(equips))
	dao.SaveEquipId(ctx, equipId)
	// dao.SaveEquipBrief(ctx, roleId, EquipManyModel2EquipmentBrief(equips)...)

	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskGainEquipmentNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskGainEquipmentNumAcc),
			Id:      0,
			Cnt:     totalCnt,
			Replace: false,
		},
		models.TaskType_TaskGainEquipmentNum: {
			Typ:     values.Integer(models.TaskType_TaskGainEquipmentNum),
			Id:      0,
			Cnt:     totalCnt,
			Replace: false,
		},
	}
	s.UpdateTargets(ctx, ctx.RoleId, tasks)
	ctx.PublishEventLocal(&event.EquipUpdate{
		New:    true,
		RoleId: roleId,
		Items:  items,
		Equips: equips,
	})

	return equips, nil
}

func (s *Service) UnlockEquipAffix(ctx *ctx.Context, star values.Integer, equip *models.Equipment) bool {
	unlockedMap := rule.GetUnlockEquipAffix(ctx, star)
	for i := 0; i < len(equip.Detail.Affix); i++ {
		if _, ok := unlockedMap[i]; ok {
			equip.Detail.Affix[i].Active = true
		} else {
			equip.Detail.Affix[i].Active = false
		}
	}

	return true
}

func (s *Service) refreshEquipLight(ctx *ctx.Context, equip *models.Equipment) bool {
	var change bool
	cfg, ok := rule.GetEquipById(ctx, equip.ItemId)
	if !ok {
		return false
	}
	if len(cfg.QualityNum) <= 0 || len(cfg.QualityEffects) <= 0 {
		return false
	}
	// 服务启动的时候做了检查：QualityNum列里所有的key在QualityEffects列一定存在
	list := make([]*EquipQualityNum, 0, len(cfg.QualityNum))
	for q, num := range cfg.QualityNum {
		list = append(list, &EquipQualityNum{
			Quality: q,
			Num:     num,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Quality > list[j].Quality
	})
	light := values.Integer(-1)
	for _, item := range list {
		if s.getEquipAffixCount(equip, item.Quality) >= item.Num {
			light = cfg.QualityEffects[item.Quality]
			break
		}
	}
	if light == -1 {
		return false
	}
	if equip.Detail.LightEffect != light {
		equip.Detail.LightEffect = light
		change = true
	}
	return change
}

// 获取当前装备词条品质>=q的词条总数量（发光逻辑由原来的已激活词条改为全部词条）
func (s *Service) getEquipAffixCount(equip *models.Equipment, q values.Quality) values.Integer {
	var count values.Integer
	for _, affix := range equip.Detail.Affix {
		if affix.Quality >= q {
			count++
		}
	}
	return count
}

func (s *Service) GenEquip(ctx *ctx.Context, itemId values.ItemId) (*models.Equipment, values.Integer, *errmsg.ErrMsg) {
	cfg, ok := rule2.GetEquipById(ctx, itemId)
	if !ok {
		s.log.Error("equip config not found", zap.Int64("item_id", itemId))
		return nil, 0, errmsg.NewInternalErr("equip config not found")
	}

	equip := &models.Equipment{
		// EquipId: xid.New().String(),
		ItemId: itemId,
		// BaseScore: 0,
		Level:  cfg.QualityEquip,
		HeroId: 0,
		Detail: &models.EquipmentDetail{
			Score: 0,
			Affix: nil,
			// ForgeId:     "",
			ForgeName:   "",
			LightEffect: 0,
		},
	}
	if cfg.AttributeNum > 0 {
		var err *errmsg.ErrMsg
		equip.Detail.Affix, err = s.genAffix(cfg)
		if err != nil {
			return nil, 0, err
		}
	}
	s.refreshEquipLight(ctx, equip)
	s.UnlockEquipAffix(ctx, enum.InitEquipStar, equip)

	// equip.BaseScore = s.BagService.CalEquipScore(ctx, equip, 0, 1, cfg, 0)
	itemCfg, ok := rule2.GetItemById(ctx, itemId)
	if ok {
		s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGetQualityEquipmentAcc, itemCfg.Quality, 1)
		s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGetQualityEquipment, itemCfg.Quality, 1)
	}
	return equip, itemCfg.Quality, nil
}

func (s *Service) genAffix(cfg *rulemodel.Equip) ([]*models.Affix, *errmsg.ErrMsg) {
	list := make([]*models.Affix, 0)
	var err *errmsg.ErrMsg
	// 先处理EquipSkill，优先级最高
	entryCfg := s.getEquipSkillEquipEntry(cfg.EquipSkill)
	// EquipSkill里必出的词缀
	list, _, err = s.genCertainAffix(entryCfg, int(cfg.AttributeNum))
	if err != nil {
		return nil, err
	}
	if len(list) < int(cfg.AttributeNum) {
		var affix *models.Affix
		// EquipSkill只出一个，如果未配置必出，那么继续按权重逻辑在非必须的列表里出一个即可
		if len(list) == 0 && len(entryCfg) > 0 {
			affix, entryCfg, err = s.GenOneAffix(entryCfg, 0, 0)
			if err != nil {
				return nil, err
			}
			if affix != nil {
				list = append(list, affix)
			}
		}
		// Attribute里的词缀
		if len(list) < int(cfg.AttributeNum) {
			entryCfg = rule2.GetEquipEntryCfgByAttribute(cfg.Attribute)
			if len(entryCfg) == 0 {
				s.log.Warn("equip_entry not exist", zap.Int64("group", cfg.Attribute))
				return nil, errmsg.NewInternalErr("equip_entry not exist")
			}

			// Attribute里必出的词缀
			affixList, newEntryCfg, err1 := s.genCertainAffix(entryCfg, int(cfg.AttributeNum)-len(list))
			if err1 != nil {
				return nil, err1
			}
			list = append(list, affixList...)
			if len(list) < int(cfg.AttributeNum) {
				// 100%出现品质的词缀（>=cfg.AttributeQuality即可）
				affix, newEntryCfg, err = s.GenOneAffix(newEntryCfg, cfg.AttributeQuality, 0)
				if err != nil {
					return nil, err
				}
				if affix != nil {
					list = append(list, affix)
				}
				start := len(list)
				for i := start; i < int(cfg.AttributeNum); i++ {
					affix, newEntryCfg, err = s.GenOneAffix(newEntryCfg, 0, 0)
					if err != nil {
						return nil, err
					}
					if affix != nil {
						list = append(list, affix)
					}
				}
			}
		}
	}
	// 词缀按品质从低到高排序，均为未激活状态，待装备星级提升时再处理是否激活
	sort.Slice(list, func(i, j int) bool {
		return list[i].Quality < list[j].Quality
	})

	// unlockList := rule2.UnlockEquipAffixCfg(ctx)
	// // 无技能词缀
	// if skillAffix == nil {
	//	return list
	// }
	// // 将技能词缀放在对应的解锁星级位置
	// index := 0
	// for i := 0; i < len(unlockList); i++ {
	//	if skillAffixUnlockStarLevel == unlockList[i] {
	//		index = i
	//		break
	//	}
	// }
	// list = append(list[0:index+1], list[index:]...)
	// list[index] = skillAffix

	return list, nil
}

// 必出的词缀（entryweight=0）
func (s *Service) genCertainAffix(entryCfg []rulemodel.EquipEntry, count int) ([]*models.Affix, []rulemodel.EquipEntry, *errmsg.ErrMsg) {
	list := make([]*models.Affix, 0)
	cfg := make([]rulemodel.EquipEntry, 0)
	for _, affixCfg := range entryCfg {
		if affixCfg.Entryweight == 0 {
			if len(list) < count {
				q, err := s.GenAffixQuality(affixCfg.Qualityweight, 0)
				if err != nil {
					return nil, nil, err
				}
				model, err := s.getAffixModel(affixCfg, q)
				if err != nil {
					return nil, nil, err
				}
				list = append(list, model)
			} else {
				cfg = append(cfg, affixCfg)
			}
		} else {
			cfg = append(cfg, affixCfg)
		}
	}

	return list, cfg, nil
}

func (s *Service) getEquipSkillEquipEntry(cfg map[values.Integer]values.Integer) []rulemodel.EquipEntry {
	if len(cfg) <= 0 {
		return nil
	}
	var list []rulemodel.EquipEntry
	for id, prob := range cfg {
		v := rand.Int63n(10001)
		if v <= prob {
			entryCfg := rule2.GetEquipEntryCfgByAttribute(id)
			if len(entryCfg) == 0 {
				s.log.Warn("equip_entry not exist", zap.Int64("group", id))
				continue
			}
			list = append(list, entryCfg...)
		}
	}
	return list
}

// GenOneAffix
// minQuality 最低品质限制，0表示不限制
// quality 品质，0表示随机
func (s *Service) GenOneAffix(entryCfg []rulemodel.EquipEntry, minQuality, quality values.Integer) (*models.Affix, []rulemodel.EquipEntry, *errmsg.ErrMsg) {
	var total values.Integer
	for _, item := range entryCfg {
		total += item.Entryweight
	}
	if total <= 0 {
		return nil, nil, errmsg.NewInternalErr("invalid total")
	}
	var affixCfg rulemodel.EquipEntry
	val := rand.Int63n(total)
	var index = -1
	for i, item := range entryCfg {
		if val < item.Entryweight {
			affixCfg = item
			index = i
			break
		}
		val -= item.Entryweight
	}
	if index == -1 {
		return nil, entryCfg, nil
	}
	entryCfg = append(entryCfg[:index], entryCfg[index+1:]...)
	if quality == 0 {
		var err *errmsg.ErrMsg
		quality, err = s.GenAffixQuality(affixCfg.Qualityweight, minQuality)
		if err != nil {
			return nil, nil, err
		}
	}
	affix, err := s.getAffixModel(affixCfg, quality)
	if err != nil {
		return nil, nil, err
	}
	return affix, entryCfg, nil
}

func (s *Service) getAffixModel(affixCfg rulemodel.EquipEntry, quality values.Integer) (*models.Affix, *errmsg.ErrMsg) {
	isSkill, isPercent, affixVal, err := s.getAffixValue(quality, affixCfg)
	if err != nil {
		return nil, err
	}
	affixModel := &models.Affix{
		AffixId:   affixCfg.Id,
		Quality:   quality,
		Active:    false,
		AttrId:    affixCfg.Attr,
		Bonus:     map[int64]int64{},
		IsPercent: isPercent,
	}
	if isSkill {
		affixModel.BuffId = affixVal
	} else {
		affixModel.AffixValue = affixVal
	}
	return affixModel, nil
}

func (s *Service) GenAffixQuality(qualityWeight map[values.Integer]values.Integer, minQuality values.Integer) (values.Integer, *errmsg.ErrMsg) {
	totalQualityWeight := values.Integer(0)
	weightMap := make(map[values.Integer]values.Integer)
	list := make([]values.Integer, 0)
	for q, v := range qualityWeight {
		// 不能低于minQuality品质
		if q >= minQuality {
			totalQualityWeight += v
			weightMap[q] = v
			list = append(list, q)
		}
	}
	if totalQualityWeight <= 0 {
		return 0, errmsg.NewInternalErr("invalid totalQualityWeight")
	}
	var quality values.Integer
	val := rand.Int63n(totalQualityWeight)
	for q, v := range weightMap {
		if val < v {
			quality = q
			break
		}
		val -= v
	}
	// 若未取到正确的品质，则随机取一个
	if quality == 0 {
		quality = list[rand.Intn(len(list))]
	}
	return quality, nil
}

func (s *Service) getAffixValue(quality values.Integer, cfg rulemodel.EquipEntry) (bool, bool, values.Integer, *errmsg.ErrMsg) {
	var isPercent bool
	if cfg.Attr == 0 {
		// 在规则配置里程序启动的时候已经检查了所有qualityweight配置如果为技能的时候在skillquality或SkillLv列一定能找到
		// 先判断skillquality列，如果有值则认为是技能（buff）
		if len(cfg.Skillquality) > 0 {
			skill, err := s.genAffixSkill(quality, cfg)
			if err != nil {
				return false, isPercent, skill, err
			}
			return true, isPercent, skill, nil
		}
		// 对技能等级加成
		return true, isPercent, 0, nil
	}
	min := values.Integer(0)
	max := values.Integer(0)
	if len(cfg.Qualitysection) > 0 {
		max = cfg.Qualitysection[quality] + 1 // 需包含最大值
		if v, ok := cfg.Qualitysection[quality-1]; ok {
			min = v
		}
	} else {
		isPercent = true
		max = cfg.QualityPer[quality] + 1 // 需包含最大值
		if v, ok := cfg.QualityPer[quality-1]; ok {
			min = v
		}
	}
	min++
	// 策划可能配出min>max的情况，遇到这种情况交换两个值
	if min >= max {
		max, min = min, max
		// 可能有2个值相等的情况，max+1保证随机数不panic
		max += 1
	}
	if max-min <= 0 {
		return false, false, 0, errmsg.NewInternalErr("invalid val")
	}
	val := values.Float(rand.Int63n(max-min) + min)
	return false, isPercent, cfg.Lowvalue + values.Integer(math.Ceil(values.Float(cfg.Maxvalue-cfg.Lowvalue)*(val/10000))), nil
}

func (s *Service) genAffixSkill(quality values.Integer, cfg rulemodel.EquipEntry) (values.Integer, *errmsg.ErrMsg) {
	skill, ok := cfg.Skillquality[quality]
	if ok {
		return skill, nil
	}
	// 若未取到则随机取一个
	list := make([]values.Integer, 0)
	for _, skillId := range cfg.Skillquality {
		list = append(list, skillId)
	}
	if len(list) <= 0 {
		return 0, errmsg.NewInternalErr("invalid val")
	}
	return list[rand.Intn(len(list))], nil
}

func (s *Service) DelEquipment(ctx *ctx.Context, equipId ...values.EquipId) *errmsg.ErrMsg {
	if len(equipId) <= 0 {
		return nil
	}
	if err := dao.DelManyEquip(ctx, ctx.RoleId, equipId...); err != nil {
		return nil
	}
	dao.DelEquipBrief(ctx, ctx.RoleId, equipId...)

	// err := subBagLength(ctx, values.Integer(len(equipId)))
	// if err != nil {
	// 	return err
	// }

	if err := s.afterSub(ctx, ctx.RoleId, values.Integer(len(equipId))); err != nil {
		return err
	}

	list := make([]values.EquipId, 0, len(equipId))
	list = append(list, equipId...)
	ctx.PublishEventLocal(&event.EquipDestroyed{
		RoleId:  ctx.RoleId,
		EquipId: list,
	})
	return nil
}

func (s *Service) CalEquipScore(
	ctx *ctx.Context,
	equip *models.Equipment,
	star, roleLevel values.Integer,
	equipCfg *rulemodel.Equip,
	heroId values.HeroId,
) values.Integer {
	// lvCfg, ok := rule.GetRoleLvConfigByLv(ctx, roleLevel)
	// if !ok {
	// 	s.log.Warn("role_lv config not found", zap.Int64("level", roleLevel))
	// 	return 0
	// }
	affixInfo := s.getEquipAffixInfo(ctx, equip.Detail)
	if affixInfo == nil {
		s.log.Warn("equip affix info is nil", zap.String("equipId", equip.EquipId))
		return 0
	}
	var total values.Integer
	// 装备属性战斗力已经在属性那边算过了，所以这里不需要（也不能）再计算属性的战斗力
	// attrs := s.getEquipAttrs(ctx, affixInfo, star, equipCfg, heroId)
	// for id, val := range attrs {
	// 	attrCfg, ok := rule.GetAttrById(ctx, id)
	// 	if !ok {
	// 		s.log.Warn("attr config not found", zap.Int64("id", id))
	// 		continue
	// 	}
	// 	total += values.Integer(math.Ceil(val * values.Float(attrCfg.PowerNum) * values.Float(lvCfg.PowerNum) / 10000))
	// }
	// buff和天赋获得的战力都是直接加上配置的数值即可
	for _, buffId := range affixInfo.Buff {
		buffCfg, ok := rule.GetBuffById(ctx, buffId)
		if !ok {
			s.log.Warn("buff config not found", zap.Int64("id", buffId))
			continue
		}
		total += buffCfg.PowerNum
	}
	for _, id := range affixInfo.Talent {
		talentCfg, ok := rule.GetTalentById(ctx, id)
		if !ok {
			s.log.Warn("talent config not found", zap.Int64("id", id))
			continue
		}
		total += talentCfg.EquipScore
	}
	return total
}

func (s *Service) GetEquipId(ctx *ctx.Context) (*pbdao.EquipId, *errmsg.ErrMsg) {
	return dao.GetEquipId(ctx, ctx.RoleId)
}

func (s *Service) SaveEquipId(ctx *ctx.Context, id *pbdao.EquipId) {
	dao.SaveEquipId(ctx, id)
}

type equipStarToValue struct {
	Range []values.Integer
	Value []values.Integer
}

func (s *Service) getEquipBaseAttr(star values.Integer, equipCfg *rulemodel.Equip) map[values.AttrId]values.Float {
	attrs := make(map[values.AttrId]values.Float)
	if len(equipCfg.AttrId) > 0 {
		starToValueMap := make(map[values.AttrId]*equipStarToValue)
		for i, attrId := range equipCfg.AttrId {
			attrs[attrId] = values.Float(equipCfg.AttrValue[i])
			starRange := make([]values.Integer, 0)
			value := make([]values.Integer, 0)
			if star > 0 {
				for j := 0; j < len(equipCfg.AttrStarRange[i]); j++ {
					if j == 0 && star <= equipCfg.AttrStarRange[i][j] {
						starRange = append(starRange, equipCfg.AttrStarRange[i][j])
						value = append(value, equipCfg.Attr[i][j])
						break
					}
					if star >= equipCfg.AttrStarRange[i][j] {
						starRange = append(starRange, equipCfg.AttrStarRange[i][j])
						value = append(value, equipCfg.Attr[i][j])
					}
					if star < equipCfg.AttrStarRange[i][j] {
						starRange = append(starRange, equipCfg.AttrStarRange[i][j])
						value = append(value, equipCfg.Attr[i][j])
						break
					}
				}
				starToValueMap[attrId] = &equipStarToValue{
					Range: starRange,
					Value: value,
				}
			}
		}
		for attrId, item := range starToValueMap {
			for i := 0; i < len(item.Range); i++ {
				v := item.Range[i]
				if i == 0 {
					if star >= v {
						attrs[attrId] += values.Float(v * item.Value[i])
					} else if i == 0 {
						attrs[attrId] += values.Float(star * item.Value[i])
					}
				} else if star > item.Range[i] {
					attrs[attrId] += values.Float((item.Range[i] - item.Range[i-1]) * item.Value[i])
				} else {
					attrs[attrId] += values.Float((star - item.Range[i-1]) * item.Value[i])
				}
			}
			if star > item.Range[len(item.Range)-1] {
				attrs[attrId] += values.Float((star - item.Range[len(item.Range)-1]) * item.Value[len(item.Range)-1])
			}
		}
	}
	return attrs
}

func (s *Service) getEquipAffixInfo(ctx *ctx.Context, equipDetail *models.EquipmentDetail) *EquipAffixInfo {
	if equipDetail == nil {
		return nil
	}
	fixedMap := make(map[values.AttrId]values.Float)
	percentMap := make(map[values.AttrId]values.Float)
	buff := make([]values.HeroBuffId, 0)
	talent := make([]values.TalentId, 0)
	talentMap := make(map[values.TalentId]struct{})
	for _, affix := range equipDetail.Affix {
		if !affix.Active || affix.AffixId <= 0 {
			continue
		}
		if affix.AttrId > 0 {
			if affix.IsPercent {
				percentMap[affix.AttrId] += values.Float(affix.AffixValue)
				for _, val := range affix.Bonus {
					percentMap[affix.AttrId] += values.Float(val)
				}
			} else {
				fixedMap[affix.AttrId] += values.Float(affix.AffixValue)
				for _, val := range affix.Bonus {
					fixedMap[affix.AttrId] += values.Float(val)
				}
			}
		} else if affix.BuffId > 0 {
			buff = append(buff, affix.BuffId)
		} else {
			cfg, ok := rule.GetEquipEntryById(ctx, affix.AffixId)
			if !ok {
				s.log.Warn("EquipEntry config not found", zap.Int64("id", affix.AffixId))
				continue
			}
			for _, talentId := range cfg.SkillId {
				if _, ok := talentMap[talentId]; !ok {
					talentMap[talentId] = struct{}{}
					talent = append(talent, talentId)
				} else {
					s.log.Warn("duplicate talentId in EquipEntry", zap.Int64("affixId", affix.AffixId), zap.Int64("talentId", talentId))
				}
			}
		}
	}
	return &EquipAffixInfo{
		FixedAttr:   fixedMap,
		PercentAttr: percentMap,
		Buff:        buff,
		Talent:      talent,
	}
}

func (s *Service) getEquipAttrs(ctx *ctx.Context, affixInfo *EquipAffixInfo, star values.Integer, equipCfg *rulemodel.Equip, heroId values.HeroId) map[values.AttrId]values.Float {
	baseAttr := s.getEquipBaseAttr(star, equipCfg)
	affixFixedAttr := affixInfo.FixedAttr
	affixPercentAttr := affixInfo.PercentAttr
	for id, val := range baseAttr {
		affixFixedAttr[id] += val
	}

	priAttr := make(map[values.AttrId]values.Float)
	secAttr := make(map[values.AttrId]values.Float)

	priPercentAttrMap := make(map[values.AttrId]values.Float)
	secPercentAttrMap := make(map[values.AttrId]values.Float)
	for id, val := range affixFixedAttr {
		cfg, ok := rule.GetAttrById(ctx, id)
		if !ok {
			s.log.Warn("attr config not found", zap.Int64("id", id))
			continue
		}
		if cfg.AdvancedType == enum.PrimaryAttr {
			priAttr[id] += val
		} else {
			secAttr[id] += val
		}
	}
	percentAttrMap := make(map[values.AttrId]values.Float)
	// 装备目前没有百分比（这里先留着，万一策划后面要加）
	for id, val := range affixPercentAttr {
		cfg, ok := rule.GetAttrById(ctx, id)
		if !ok {
			s.log.Warn("attr config not found", zap.Int64("id", id))
			continue
		}
		if cfg.ShowTpye == enum.Direct {
			if cfg.AdvancedType == enum.PrimaryAttr {
				priPercentAttrMap[id] += val
			} else {
				secPercentAttrMap[id] += val
			}
		} else {
			percentAttrMap[id] += values.Float(val) / 1000
		}
	}
	// 一级属性百分比加成
	for id, perVal := range priPercentAttrMap {
		attrVal := priAttr[id]
		val := perVal * attrVal / 10000.0
		if val > 0 {
			priAttr[id] += val
		}
	}
	// 属性转换（一级属性转换为二级属性）
	for id, val := range priAttr {
		list := rule.GetAttrTransConfigById(ctx, id)
		for _, transItem := range list {
			if !s.attrTransformVocationCheck(heroId, transItem.Limithero) {
				continue
			}
			if transItem.Transtype == 1 {
				secPercentAttrMap[transItem.TransattrId] += val * values.Float(transItem.Transnum)
			} else {
				v := val * values.Float(transItem.Transnum)
				secAttr[transItem.TransattrId] += v
			}
		}
	}
	// 处理二级属性百分比加成（这一步后各个系统的二级属性为最终值）
	for id, perVal := range secPercentAttrMap {
		attrVal := secAttr[id]
		val := perVal * (attrVal / 10000.0)
		if val > 0 {
			secAttr[id] += val
		}
	}
	allAttr := make(map[values.AttrId]values.Float)
	for id, val := range priAttr {
		allAttr[id] += val
	}
	for id, val := range secAttr {
		allAttr[id] += val
	}
	for id, val := range percentAttrMap {
		allAttr[id] += val
	}

	return allAttr
}

func (s *Service) attrTransformVocationCheck(heroId values.HeroId, list []values.Integer) bool {
	// 没填或者填0就是对所有职业都生效
	if heroId == 0 || len(list) == 0 || (len(list) == 1 && list[0] == 0) {
		return true
	}
	for _, v := range list {
		if heroId == v {
			return true
		}
	}
	return false
}

// GetEquipTalentBonus 装备对技能等级加成
func (s *Service) GetEquipTalentBonus(ctx *ctx.Context, equip *models.Equipment, onlyActive, takeDown bool) map[values.HeroSkillId]values.Level {
	if equip.Detail == nil {
		return nil
	}
	ret := make(map[values.HeroSkillId]values.Level)
	for _, affix := range equip.Detail.Affix {
		if onlyActive && !affix.Active {
			continue
		}
		cfg, ok := rule.GetEquipEntryById(ctx, affix.AffixId)
		if !ok {
			s.log.Warn("EquipEntry config not found", zap.Int64("id", affix.AffixId))
			continue
		}
		if len(cfg.SkillId) > 0 { // 对单个或多个技能
			for _, id := range cfg.SkillId {
				if takeDown {
					ret[id] -= cfg.SkillLv[affix.Quality]
				} else {
					ret[id] += cfg.SkillLv[affix.Quality]
				}
			}
		}
	}
	return ret
}

// func (s *Service) calEquipTalentScore(ctx *ctx.Context, equip *models.Equipment, heroId values.HeroId) values.Integer {
// 	var score values.Integer
// 	data := s.GetEquipTalentBonus(ctx, equip, heroId, false)
// 	for _, item := range data {
// 		for _, bonus := range item {
// 			score += bonus.Talent.EquipScore * bonus.Level
// 		}
// 	}
// 	return score
// }
