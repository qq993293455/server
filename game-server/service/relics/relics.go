package relics

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/relics/dao"
	"coin-server/rule"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewRelicsService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
		log:        log,
	}
	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("获取遗物套装", s.GetRelicsSuitRequest)
	s.svc.RegisterFunc("遗物升级", s.LevelUpgradeRequest)
	s.svc.RegisterFunc("遗物升星", s.StarUpgradeRequest)
	s.svc.RegisterFunc("遗物合成", s.RelicsCompose)
	s.svc.RegisterFunc("遗物红点已读", s.RelicsRead)

	s.svc.RegisterFunc("作弊遗物满级", s.CheatRelicsLevelStarMax)

	eventlocal.SubscribeEventLocal(s.HandleRelicsSuitUpdate)
	eventlocal.SubscribeEventLocal(s.HandleRelicsUpdate)
	eventlocal.SubscribeEventLocal(s.HandleTaskUpdate)
}

func (s *Service) GetRelicsSuitRequest(ctx *ctx.Context, _ *servicepb.Relics_GetRelicsSuitRequest) (*servicepb.Relics_GetRelicsSuitResponse, *errmsg.ErrMsg) {
	relicsSuit, err := dao.GetRelicsSuit(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.Relics_GetRelicsSuitResponse{Suit: RelicsDao2Models(relicsSuit)}, nil
}

func (s *Service) LevelUpgradeRequest(ctx *ctx.Context, req *servicepb.Relics_RelicsLevelUpgradeRequest) (*servicepb.Relics_RelicsLevelUpgradeResponse, *errmsg.ErrMsg) {
	relics, err := s.BagService.GetRelicsById(ctx, ctx.RoleId, req.RelicsId)
	if err != nil {
		return nil, err
	}
	err = s.levelUpgrade(ctx, relics)
	if err != nil {
		return nil, err
	}
	s.BagService.UpdateRelics(ctx, ctx.RoleId, relics)
	// 遗物升级打点
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskLevelUpRelicsNum: {
			Typ:     values.Integer(models.TaskType_TaskLevelUpRelicsNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskLevelUpRelicsNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskLevelUpRelicsNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	s.TaskService.UpdateTargets(ctx, ctx.RoleId, tasks)
	return &servicepb.Relics_RelicsLevelUpgradeResponse{}, nil
}

func (s *Service) StarUpgradeRequest(ctx *ctx.Context, req *servicepb.Relics_RelicsStarUpgradeRequest) (*servicepb.Relics_RelicsStarUpgradeResponse, *errmsg.ErrMsg) {
	relics, err := s.BagService.GetRelicsById(ctx, ctx.RoleId, req.RelicsId)
	if err != nil {
		return nil, err
	}
	err = s.starUpgrade(ctx, relics)
	if err != nil {
		return nil, err
	}
	err = updateSuit(ctx, relics)
	if err != nil {
		return nil, err
	}
	reader := rule.MustGetReader(ctx)
	relicsCnf, has := reader.Relics.GetRelicsById(relics.RelicsId)
	if !has {
		return nil, errmsg.NewErrBagRelicsNotExist()
	}
	funcAttrCnf, has := reader.RelicsFunctionAttr.GetRelicsFunctionAttrById(relicsCnf.RelicsFunctionAttrId)
	if has {
		mulit := values.Float(1) + values.Float(relics.Star*funcAttrCnf.StarsCoefficient)/10000
		addFuncAttr := map[values.AttrId]values.Integer{}
		if relics.FuncAttr == nil {
			relics.FuncAttr = map[values.AttrId]values.Integer{}
		}
		for attrKey, val := range funcAttrCnf.AddAttr {
			newVal := values.Integer(values.Float(val*relics.FuncTimes) * mulit)
			addFuncAttr[attrKey] += newVal - relics.FuncAttr[attrKey]
			relics.FuncAttr[attrKey] = newVal
		}
		ctx.PublishEventLocal(&event.AttrUpdateToRole{
			Typ:       models.AttrBonusType_TypeRelicsFunc,
			AttrFixed: addFuncAttr,
		})
	}
	s.BagService.UpdateRelics(ctx, ctx.RoleId, relics)
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskRelicsStartNum: {
			Typ:     values.Integer(models.TaskType_TaskRelicsStartNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskRelicsStartNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskRelicsStartNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	s.TaskService.UpdateTargets(ctx, ctx.RoleId, tasks)
	return &servicepb.Relics_RelicsStarUpgradeResponse{}, nil
}

func (s *Service) RelicsCompose(c *ctx.Context, req *servicepb.Relics_RelicsComposeRequest) (*servicepb.Relics_RelicsComposeResponse, *errmsg.ErrMsg) {
	relics, err := s.BagService.GetRelicsById(c, c.RoleId, req.RelicsId)
	if relics != nil {
		return nil, errmsg.NewErrRelicsExist()
	}
	if err.ErrMsg != errmsg.NewErrBagRelicsNotExist().ErrMsg {
		return nil, err
	}
	cfg, has := rule.MustGetReader(c).Relics.GetRelicsById(req.RelicsId)
	if !has {
		return nil, errmsg.NewErrBagRelicsNotExist()
	}
	fragment := cfg.FragmentRelics
	if len(fragment) != 2 {
		return nil, errmsg.NewErrBagRelicsNotExist()
	}
	itemCnt, err := s.BagService.GetItem(c, c.RoleId, fragment[0])
	if err != nil {
		return nil, err
	}
	if itemCnt < fragment[1] {
		return nil, errmsg.NewErrRelicsFragmentNotEnough()
	}
	if err = s.BagService.ExchangeManyItem(c, c.RoleId, map[values.ItemId]values.Integer{req.RelicsId: 1}, map[values.ItemId]values.Integer{fragment[0]: fragment[1]}); err != nil {
		return nil, err
	}
	return &servicepb.Relics_RelicsComposeResponse{}, nil
}

func (s *Service) RelicsRead(c *ctx.Context, _ *servicepb.Relics_RelicsReadRequest) (*servicepb.Relics_RelicsReadResponse, *errmsg.ErrMsg) {
	relics, err := s.BagService.GetRelics(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	for _, r := range relics {
		if r.IsNew {
			r.IsNew = false
		}
	}
	s.BagService.UpdateMultiRelics(c, c.RoleId, relics)
	return &servicepb.Relics_RelicsReadResponse{}, nil
}

func (s *Service) CheatRelicsLevelStarMax(c *ctx.Context, _ *servicepb.Relics_CheatRelicsLevelMaxRequest) (*servicepb.Relics_CheatRelicsLevelMaxResponse, *errmsg.ErrMsg) {
	relics, err := s.BagService.GetRelics(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	reader := rule.MustGetReader(c)
	for _, relic := range relics {
		item, ok := reader.Item.GetItemById(relic.RelicsId)
		if !ok {
			return nil, errmsg.NewErrBagRelicsNotExist()
		}
		lvMax := reader.Relics.Level(item.Quality)
		for currLv := relic.Level + 1; currLv < lvMax; currLv++ {
			relic.Level++
			updateRelicsAttr(c, relic)
		}
		cfg, ok := reader.Relics.GetRelicsById(relic.RelicsId)
		starMax := values.Integer(len(cfg.StarsCost))
		for currStar := relic.Star; currStar < starMax; currStar++ {
			relic.Star++
			updateRelicsSkill(c, relic)
			updateSuit(c, relic)
		}
		s.BagService.UpdateRelics(c, c.RoleId, relic)
	}
	return &servicepb.Relics_CheatRelicsLevelMaxResponse{}, nil
}

func (s *Service) levelUpgrade(ctx *ctx.Context, relics *models.Relics) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	item, ok := r.Item.GetItemById(relics.RelicsId)
	if !ok {
		return errmsg.NewErrBagRelicsNotExist()
	}
	lockLvl, cost, has := r.Relics.LevelUpCost(item.Quality, relics.Level)
	if !has {
		return errmsg.NewErrRelicsLevelMax()
	}
	role, err := s.Module.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if role.Level < lockLvl {
		return errmsg.NewErrRoleLevelNotEnough()
	}
	err = s.BagService.SubManyItem(ctx, ctx.RoleId, cost)
	if err != nil {
		return err
	}
	relics.Level++
	updateRelicsAttr(ctx, relics)
	return nil
}

func (s *Service) starUpgrade(ctx *ctx.Context, relics *models.Relics) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.Relics.GetRelicsById(relics.RelicsId)
	if !ok {
		return nil
	}
	if int(relics.Star) >= len(cfg.StarsCost) {
		return errmsg.NewErrRelicsStarMax()
	}
	cost := map[values.ItemId]values.Integer{}
	for i := 0; i < len(cfg.StarsCost[relics.Star]); i += 2 {
		cost[cfg.StarsCost[relics.Star][i]] = cfg.StarsCost[relics.Star][i+1]
	}
	err := s.BagService.SubManyItem(ctx, ctx.RoleId, cost)
	if err != nil {
		return err
	}
	relics.Star++
	err = updateRelicsSkill(ctx, relics)
	if err != nil {
		return err
	}
	return nil
}

func updateSuit(ctx *ctx.Context, relics *models.Relics) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.Relics.GetRelicsById(relics.RelicsId)
	if !ok {
		return nil
	}
	suit, err := dao.GetRelicsSuitById(ctx, ctx.RoleId, cfg.SuitId)
	if err != nil {
		return err
	}
	if suit == nil {
		suit = &pbdao.RelicsSuit{
			SuitId: cfg.SuitId,
			Level:  0,
			Star:   DefaultRelicsStar,
		}
	}
	// 不是新获得遗物才会升星
	if relics.Star != DefaultRelicsStar {
		suit.Star++
		// 套装升级且不是最后一级
		if int(suit.Level+1) < len(cfg.SuitSkill) && suit.Star >= cfg.SuitSkill[suit.Level][0] {
			suit.Level++
			addSkill(ctx, cfg.SuitSkill[suit.Level][1])
		}
	}
	dao.SaveRelicsSuit(ctx, ctx.RoleId, suit)
	ctx.PublishEventLocal(&event.RelicsSuitUpdate{
		RelicsSuit: RelicsDao2Model(suit),
	})
	return nil
}

func (s *Service) updateRelicsFuncAttr(ctx *ctx.Context, relicsIds []values.Integer, d *event.TargetUpdate) *errmsg.ErrMsg {
	reader := rule.MustGetReader(ctx)
	manyRelics, err := s.BagService.GetManyRelics(ctx, ctx.RoleId, relicsIds)
	if err != nil {
		return err
	}
	addFuncAttr := map[values.AttrId]values.Integer{}
	updateRelics := make([]*models.Relics, 0, len(manyRelics))
	for _, relics := range manyRelics {
		relicsCnf, has := reader.Relics.GetRelicsById(relics.RelicsId)
		if !has {
			continue
		}
		funcAttrCnf, has := reader.RelicsFunctionAttr.GetRelicsFunctionAttrById(relicsCnf.RelicsFunctionAttrId)
		if !has {
			continue
		}
		if models.TaskType(funcAttrCnf.RelicsType[0]) != d.Typ {
			continue
		}
		if funcAttrCnf.RelicsType[1] != d.Id {
			continue
		}
		upper := d.Count / funcAttrCnf.RelicsType[2]
		if upper > funcAttrCnf.AddUpperLimit {
			upper = funcAttrCnf.AddUpperLimit
		}
		if upper != relics.FuncTimes {
			mulit := values.Float(1) + values.Float(relics.Star*funcAttrCnf.StarsCoefficient)/10000
			if relics.FuncAttr == nil {
				relics.FuncAttr = map[values.AttrId]values.Integer{}
			}
			for attrKey, val := range funcAttrCnf.AddAttr {
				newVal := values.Integer(values.Float(val*upper) * mulit)
				addFuncAttr[attrKey] += newVal - relics.FuncAttr[attrKey]
				relics.FuncAttr[attrKey] = newVal
			}
			relics.FuncTimes = upper
			updateRelics = append(updateRelics, relics)
		}
	}
	if len(updateRelics) > 0 {
		s.BagService.UpdateMultiRelics(ctx, ctx.RoleId, updateRelics)
		ctx.PublishEventLocal(&event.AttrUpdateToRole{
			Typ:       models.AttrBonusType_TypeRelicsFunc,
			AttrFixed: addFuncAttr,
		})
	}
	return nil
}

// 新获得、升级会变更遗物属性
func updateRelicsAttr(ctx *ctx.Context, relics *models.Relics) {
	var (
		fixed   map[values.AttrId]values.Integer
		percent map[values.AttrId]values.Integer
	)
	r := rule.MustGetReader(ctx)
	cfg, _ := r.Relics.GetRelicsById(relics.RelicsId)
	// 首次获得
	if relics.Level == DefaultRelicsLevel {
		for i := 0; i < len(cfg.AttrId); i++ {
			if cfg.AttrPer[i] == 1 {
				if fixed == nil {
					fixed = map[values.AttrId]values.Integer{}
				}
				fixed[cfg.AttrId[i]] = cfg.AttrValue[i]
			} else {
				if percent == nil {
					percent = map[values.AttrId]values.Integer{}
				}
				percent[cfg.AttrId[i]] = cfg.AttrValue[i]
			}
		}
		ctx.PublishEventLocal(&event.AttrUpdateToRole{
			Typ:         models.AttrBonusType_TypeRelics,
			AttrFixed:   fixed,
			AttrPercent: percent,
		})
		return
	}

	for i, attrRange := range cfg.AttrStarRange {
		var growth values.Integer
		for j := range attrRange {
			if relics.Level <= attrRange[j] {
				growth = cfg.Attr[i][j]
				break
			}
		}
		if relics.Attr == nil {
			relics.Attr = map[int64]int64{}
		}
		relics.Attr[cfg.AttrId[i]] += growth
		if cfg.AttrPer[i] == 1 {
			if fixed == nil {
				fixed = map[values.AttrId]values.Integer{}
			}
			fixed[cfg.AttrId[i]] = growth
		} else {
			if percent == nil {
				percent = map[values.AttrId]values.Integer{}
			}
			percent[cfg.AttrId[i]] = growth
		}
	}
	ctx.PublishEventLocal(&event.AttrUpdateToRole{
		Typ:         models.AttrBonusType_TypeRelics,
		AttrFixed:   fixed,
		AttrPercent: percent,
	})
	return
}

// 新获得、升星会变更遗物技能
func updateRelicsSkill(ctx *ctx.Context, relics *models.Relics) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.Relics.GetRelicsById(relics.RelicsId)
	if !ok {
		return nil
	}

	// 新获得
	if relics.Star == DefaultRelicsStar {
		for _, skillId := range cfg.RelicsSkill {
			addSkill(ctx, skillId)
		}
		return nil
	}

	// 升星
	lastStar := relics.Star - 1
	for _, skillId := range cfg.RelicsSkill {
		skillCfg, has := r.RelicsSkill.GetRelicsSkillById(skillId)
		if !has {
			return nil
		}
		// todo: 特殊类型的技能
		if len(skillCfg.Typ) == 0 {
			return nil
		}
		// 获得技能
		if skillCfg.Typ[0] == 1 {
			/*ctx.PublishEventLocal(&event.RoleSkillUpdate{
				OldSkill: skillCfg.StarsValue()[lastStar],
				NewSkill: skillCfg.StarsValue()[relics.Star],
			})*/
		} else if skillCfg.Typ[0] == values.Integer(models.EntrySkillType_ESTAttrAdd) || skillCfg.Typ[0] == values.Integer(models.EntrySkillType_ESTHeroAttrAdd) {
			var (
				fixed   map[values.AttrId]values.Integer
				percent map[values.AttrId]values.Integer
			)
			if skillCfg.Value == 1 {
				fixed = map[values.AttrId]values.Integer{}
				fixed[skillCfg.Typ[1]] = skillCfg.StarsValue[relics.Star] - skillCfg.StarsValue[lastStar]
			}
			if skillCfg.Value == 2 {
				percent = map[values.AttrId]values.Integer{}
				percent[skillCfg.Typ[1]] = skillCfg.StarsValue[relics.Star] - skillCfg.StarsValue[lastStar]
			}
			evt := &event.AttrUpdateToRole{
				Typ:         models.AttrBonusType_TypeRelics,
				AttrFixed:   fixed,
				AttrPercent: percent,
			}
			if skillCfg.Typ[0] == values.Integer(models.EntrySkillType_ESTHeroAttrAdd) {
				evt.HeroId = skillCfg.Typ[2]
			}
			ctx.PublishEventLocal(evt)
		} else {
			// 非属性效果
			ctx.PublishEventLocal(&event.ExtraSkillTypAdd{
				TypId:    models.EntrySkillType(skillCfg.Typ[0]),
				LogicId:  skillCfg.Typ[1],
				ValueTyp: skillCfg.Value,
				Cnt:      skillCfg.StarsValue[relics.Star] - skillCfg.StarsValue[lastStar],
			})
		}
	}
	return nil
}

func addSkill(ctx *ctx.Context, skillId values.Integer) {
	r := rule.MustGetReader(ctx)
	skillCfg, has := r.RelicsSkill.GetRelicsSkillById(skillId)
	if !has {
		return
	}
	// todo: 特殊类型的技能
	if len(skillCfg.Typ) == 0 {
		return
	}
	// 获得技能
	if skillCfg.Typ[0] == 1 {
		/*ctx.PublishEventLocal(&event.RoleSkillUpdate{
			OldSkill: 0,
			NewSkill: skillCfg.StarsValue()[0],
		})*/
	} else if skillCfg.Typ[0] == values.Integer(models.EntrySkillType_ESTAttrAdd) || skillCfg.Typ[0] == values.Integer(models.EntrySkillType_ESTHeroAttrAdd) {
		var (
			fixed   map[values.AttrId]values.Integer
			percent map[values.AttrId]values.Integer
		)
		if skillCfg.Value == 1 {
			fixed = map[values.AttrId]values.Integer{}
			fixed[skillCfg.Typ[1]] = skillCfg.StarsValue[0]
		}
		if skillCfg.Value == 2 {
			percent = map[values.AttrId]values.Integer{}
			percent[skillCfg.Typ[1]] = skillCfg.StarsValue[0]
		}
		evt := &event.AttrUpdateToRole{
			Typ:         models.AttrBonusType_TypeRelics,
			AttrFixed:   fixed,
			AttrPercent: percent,
		}
		if skillCfg.Typ[0] == values.Integer(models.EntrySkillType_ESTHeroAttrAdd) {
			evt.HeroId = skillCfg.Typ[2]
		}
		ctx.PublishEventLocal(evt)
	} else {
		ctx.PublishEventLocal(&event.ExtraSkillTypAdd{
			TypId:    models.EntrySkillType(skillCfg.Typ[0]),
			LogicId:  skillCfg.Typ[1],
			ValueTyp: skillCfg.Value,
			Cnt:      skillCfg.StarsValue[0],
		})
	}
}
