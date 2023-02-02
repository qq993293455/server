package equip_forge

import (
	"sort"
	"strconv"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/iggsdk"
	"coin-server/common/im"
	"coin-server/common/logger"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/statistical"
	models2 "coin-server/common/statistical/models"
	"coin-server/common/timer"
	wr "coin-server/common/utils/weightedrand"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/common/values/enum/Notice"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/equip-forge/dao"
	"coin-server/game-server/service/equip-forge/rule"
	values2 "coin-server/game-server/service/equip-forge/values"
	rule2 "coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
	"github.com/rs/xid"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewEquipForgeService(
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
		log:        log,
		Module:     module,
	}
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取打造信息", svc.Info)
	svc.svc.RegisterFunc("打造一件装备", svc.ForgeOne)

	// 作弊器
	svc.svc.RegisterFunc("获取所有打造图纸", svc.CheatAddAllRecipe)
	svc.svc.RegisterFunc("修改打造等级", svc.CheatUpdateForgeLevel)
}

func (svc *Service) Info(ctx *ctx.Context, _ *servicepb.EquipForge_EquipForgeInfoRequest) (*servicepb.EquipForge_EquipForgeInfoResponse, *errmsg.ErrMsg) {
	data, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &servicepb.EquipForge_EquipForgeInfoResponse{
		ForgeInfo: &models.EquipForge{
			Level:      data.Level,
			Exp:        data.Exp,
			ForgeCount: data.ForgeCount,
		},
	}, nil
}

func (svc *Service) ForgeOne(ctx *ctx.Context, req *servicepb.EquipForge_ForgeRequest) (*servicepb.EquipForge_ForgeResponse, *errmsg.ErrMsg) {
	recipe, ok := rule.GetFixedRecipeInfo(ctx, req.RecipeId)
	if !ok {
		return nil, errmsg.NewErrForgeRecipeNotExist()
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if role.Level < recipe.Level {
		return nil, errmsg.NewErrRoleLevelNotEnough()
	}
	info, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	minimumGuarantee, ok := rule.GetFixedNormalMinimumGuaranteeCount(ctx)
	if !ok {
		return nil, errmsg.NewInternalErr("minimum guarantee not found")
	}
	levelCfg, ok := rule.GetLevelConfig(ctx, info.Level)
	if !ok {
		return nil, errmsg.NewInternalErr("forge level config not found")
	}
	bagConfig, err := svc.GetBagConfig(ctx)
	if err != nil {
		return nil, err
	}

	if len(info.ForgeCount) <= 0 {
		info.ForgeCount = make(map[values.Integer]values.Integer)
	}
	if _, ok := info.ForgeCount[req.RecipeId]; !ok {
		info.ForgeCount[req.RecipeId] = 0
	}
	info.ForgeCount[req.RecipeId]++
	isMinimumGuarantee := info.ForgeCount[req.RecipeId] >= minimumGuarantee
	if isMinimumGuarantee {
		info.ForgeCount[req.RecipeId] = 0
	}
	info.TotalForgeCount++

	products, err := svc.beforeForge(ctx, info, recipe, levelCfg, req.SupplementId, isMinimumGuarantee)
	if err != nil {
		return nil, err
	}
	list := svc.handleBoxProb(products)
	boxId := svc.randomBoxId(list)
	ret, err := svc.UseItemCase(ctx, boxId, 1, nil)
	if err != nil {
		return nil, err
	}
	if len(ret) <= 0 {
		svc.log.Error("ForgeOne UseItemCase ret is empty", zap.Int64("boxId", boxId))
		return nil, errmsg.NewInternalErr("forge result is empty")
	}
	var itemId values.ItemId
	for id := range ret {
		itemId = id
		break
	}
	equip, q, err := svc.GenEquip(ctx, itemId)
	if err != nil {
		return nil, err
	}
	// 记录锻造者
	// equip.Detail.ForgeId = ctx.RoleId
	if q >= 4 {
		equip.Detail.ForgeName = role.Nickname
	}

	if bagConfig.Config.Capacity > bagConfig.Config.CapacityOccupied {
		equipId, err := svc.GetEquipId(ctx)
		if err != nil {
			return nil, err
		}
		equipId.EquipId++
		equip.EquipId = strconv.Itoa(int(equipId.EquipId))
		if err := svc.SaveEquipment(ctx, ctx.RoleId, equip); err != nil {
			return nil, err
		}
		bagConfig.Config.CapacityOccupied++
		svc.SaveBagConfig(ctx, bagConfig)
		svc.SaveEquipId(ctx, equipId)
		// ctx.PushMessageToRole(ctx.RoleId, &servicepb.Bag_EquipGotPush{Equipments: []*models.Equipment{equip}})
		ctx.PublishEventLocal(&event.EquipUpdate{
			New:    true,
			RoleId: ctx.RoleId,
			Items:  map[values.ItemId]values.Integer{equip.ItemId: 1},
			Equips: []*models.Equipment{equip},
		})
	} else {
		id := values.Integer(enum.ItemReissueId)
		var expiredAt values.Integer
		cfg, ok := rule.GetMailConfigTextId(ctx, id)
		if ok {
			expiredAt = timer.StartTime(ctx.StartTime).Add(time.Second * time.Duration(cfg.Overdue)).UnixMilli()
		}
		if err := svc.Add(ctx, ctx.RoleId, &models.Mail{
			Id:        xid.New().String(),
			Type:      models.MailType_MailTypeSystem,
			TextId:    id,
			ExpiredAt: expiredAt,
			Attachment: []*models.Item{{
				ItemId: equip.ItemId,
				Count:  1,
			}},
		}); err != nil {
			return nil, err
		}
	}
	// svc.SaveEquipmentBrief(ctx, ctx.RoleId, equip)

	if err := svc.addForgeExp(ctx, info, levelCfg); err != nil {
		return nil, err
	}
	dao.Save(ctx, info)

	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskBuildEquipAcc: {
			Typ:     values.Integer(models.TaskType_TaskBuildEquipAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskBuildEquip: {
			Typ:     values.Integer(models.TaskType_TaskBuildEquip),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskBuildQualityEquipAcc: {
			Typ:     values.Integer(models.TaskType_TaskBuildQualityEquipAcc),
			Id:      q,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskBuildQualityEquip: {
			Typ:     values.Integer(models.TaskType_TaskBuildQualityEquip),
			Id:      q,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskBuildTargetEquip: {
			Typ:     values.Integer(models.TaskType_TaskBuildTargetEquip),
			Id:      itemId,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskGainEquipmentNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskGainEquipmentNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskGainEquipmentNum: {
			Typ:     values.Integer(models.TaskType_TaskGainEquipmentNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	svc.UpdateTargets(ctx, ctx.RoleId, tasks)

	statistical.Save(ctx.NewLogServer(), &models2.Forge{
		IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
		EventTime: timer.Now(),
		GwId:      statistical.GwId(),
		RoleId:    ctx.RoleId,
		Quality:   q,
		EquipId:   equip.EquipId,
		ItemId:    itemId,
	})

	if q > 4 { // 打造出橙色以上装备时发跑马灯公告
		err = svc.ImService.SendNotice(ctx, im.ParseTypeNoticeEquip, Notice.Equip, role.Nickname, equip.ItemId)
		if err != nil {
			svc.log.Error("svc.ImService.SendNotice error", zap.Error(err))
		}
	}

	return &servicepb.EquipForge_ForgeResponse{
		Equip:     equip,
		ForgeInfo: &models.EquipForge{Level: info.Level, Exp: info.Exp, ForgeCount: info.ForgeCount},
	}, nil
}

func (svc *Service) beforeForge(
	ctx *ctx.Context,
	info *pbdao.EquipForge,
	recipe *rulemodel.ForgeFixedRecipe,
	levelCfg *rulemodel.ForgeLevel,
	supplement values.ItemId,
	isMinimumGuarantee bool,
) (map[values.Quality]*values2.Product, *errmsg.ErrMsg) {
	// 这里已改为角色等级判断
	// 判断打造等级对应可打造的图纸等级是否能满足当前图纸等级
	// if levelCfg.RecipeLevel < recipe.Level {
	//	return nil, errmsg.NewErrForgeLevelNotEnough()
	// }
	cost := map[values.ItemId]values.Integer{
		recipe.Id: 1, // 图纸
		enum.Gold: recipe.Gold,
	}
	for id, count := range recipe.MaterialsList {
		cost[id] += count
	}
	products := make(map[values.Quality]*values2.Product)
	for _, item := range recipe.Products {
		p := &values2.Product{
			BoxId:   item[0],
			Quality: item[1],
			Weight:  item[2],
		}
		products[p.Quality] = p
	}
	var cfg *rulemodel.ForgeSupplement
	// 补充材料一次消耗一个（如果玩家选择了）
	if supplement > 0 {
		cost[supplement] += 1
		var ok bool
		cfg, ok = rule.GetSupplement(ctx, supplement)
		if !ok {
			return nil, errmsg.NewErrForgeInvalidSupplement()
		}
		// 填充材料加成
		for quality, prob := range cfg.EquipQualityProbability {
			p, ok := products[quality]
			if ok {
				p.Weight += prob
				products[quality] = p
			}
		}
	}
	// 扣除打造消耗
	if err := svc.SubManyItem(ctx, ctx.RoleId, cost); err != nil {
		return nil, err
	}
	// 保底加成
	if isMinimumGuarantee {
		for quality, prob := range recipe.MinimumGuaranteeProbability {
			p, ok := products[quality]
			if ok {
				p.Weight += prob
				products[quality] = p
			}
		}
	}
	// 先判断的打造的等级是否超过了加成等级上限，如果超过则不加成
	if recipe.Level > rule.GetEquipForgeMaxLevel(ctx) {
		return products, nil
	}
	// 打造等级加成
	levelBonus := rule.GetForgeLevelBonus(ctx, info.Level, recipe.Level)
	// 打造等级对打造100%会有加成，如果没有则表配置有误，需要返回error
	if len(levelBonus) <= 0 {
		svc.log.Error("forge level bonus not found",
			zap.Int64("level", info.Level),
			zap.Int64("recipe_level", recipe.Level),
		)
		return products, nil
	}
	for quality, prob := range levelBonus {
		p, ok := products[quality]
		if ok {
			p.Weight += prob
			products[quality] = p
		}
	}
	return products, nil
}

func (svc *Service) randomBoxId(list []*values2.Product) values.Integer {
	choice := make([]*wr.Choice[values.Integer, values.Integer], 0, len(list))
	for _, product := range list {
		choice = append(choice, &wr.Choice[values.Integer, values.Integer]{
			Item:   product.BoxId,
			Weight: product.Weight,
		})
	}
	chooser, _ := wr.NewChooser(choice...)
	boxId := chooser.Pick()
	if boxId <= 0 {
		boxId = list[len(list)-1].BoxId
	}
	return boxId
}

func (svc *Service) handleBoxProb(products map[values.Quality]*values2.Product) []*values2.Product {
	list := make([]*values2.Product, 0, len(products))
	for _, product := range products {
		list = append(list, product)
	}
	// 按品质从高到低排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Quality > list[j].Quality
	})
	// 按品质从高到低，移除超过100%的品质，并把品质概率总和归为100%
	var totalProb int64
	max := values.Integer(10000)
	newList := make([]*values2.Product, 0)
	for _, product := range list {
		if product.Weight <= 0 {
			continue
		}
		if product.Weight >= max {
			totalProb = max
			product.Weight = max
			newList = append(newList, product)
			break
		} else if totalProb < max {
			totalProb += product.Weight
			if totalProb > max {
				product.Weight -= totalProb - max
				totalProb = max
			}
			newList = append(newList, product)
		}
	}
	// 概率不足100%，提高最低品质的概率直至100%
	if totalProb < max {
		newList[len(newList)-1].Weight += max - totalProb
	}
	// 概率从低到高排序
	sort.Slice(newList, func(i, j int) bool {
		if newList[i].Weight == newList[j].Weight {
			return newList[i].Quality < newList[j].Quality
		} else {
			return newList[i].Weight < newList[j].Weight
		}
	})
	return newList
}

func (svc *Service) fixedForging(boxId values.Integer) {

}

func (svc *Service) addForgeExp(ctx *ctx.Context, info *pbdao.EquipForge, levelCfg *rulemodel.ForgeLevel) *errmsg.ErrMsg {
	exp := rule.GetForgeOnceExp(ctx)
	if exp <= 0 {
		return nil
	}
	maxLevel := rule.GetMaxForgeLevel(ctx)
	if info.Level >= maxLevel {
		return nil
	}
	info.Exp += exp
	if info.Exp >= levelCfg.Exp {
		info.Exp -= levelCfg.Exp
		info.Level++
	}
	return nil
}

func (svc *Service) CheatAddAllRecipe(ctx *ctx.Context, req *servicepb.EquipForge_CheatAddAllRecipeRequest) (*servicepb.EquipForge_CheatAddAllRecipeResponse, *errmsg.ErrMsg) {
	list := rule2.MustGetReader(ctx).ForgeFixedRecipe.List()
	items := make(map[values.ItemId]values.Integer, len(list))
	for _, recipe := range list {
		items[recipe.Id] = req.Num
	}
	if _, err := svc.AddManyItem(ctx, ctx.RoleId, items); err != nil {
		return nil, err
	}
	return &servicepb.EquipForge_CheatAddAllRecipeResponse{}, nil
}

func (svc *Service) CheatUpdateForgeLevel(ctx *ctx.Context, req *servicepb.EquipForge_CheatUpdateForgeLevelRequest) (*servicepb.EquipForge_CheatUpdateForgeLevelResponse, *errmsg.ErrMsg) {
	level := req.Level
	maxLevel := rule.GetMaxForgeLevel(ctx)
	if level > maxLevel {
		level = maxLevel
	}
	if level <= 0 {
		level = 1
	}
	info, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	info.Exp = 0
	info.Level = level
	dao.Save(ctx, info)
	return &servicepb.EquipForge_CheatUpdateForgeLevelResponse{
		ForgeInfo: &models.EquipForge{
			Level:      info.Level,
			Exp:        info.Exp,
			ForgeCount: info.ForgeCount,
		},
	}, nil
}
