package hero

import (
	"strconv"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/cppbattle"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/fashion_service"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/hero/dao"
	"coin-server/game-server/service/hero/rule"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

func (svc *Service) DressFashion(ctx *ctx.Context, req *servicepb.Hero_DressFashionRequest) (*servicepb.Hero_DressFashionResponse, *errmsg.ErrMsg) {
	heroDao := dao.NewHero(ctx.RoleId)
	hero, ok, err := heroDao.GetOne(ctx, req.HeroOriginId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrHeroNotFound()
	}
	if hero.Fashion.Dressed == req.FashionId {
		return &servicepb.Hero_DressFashionResponse{}, nil
	}
	defaultFashion := rule.GetDefaultFashion(ctx, req.HeroOriginId)
	if req.FashionId != defaultFashion {
		expire, ok := hero.Fashion.Data[req.FashionId]
		if !ok {
			return nil, errmsg.NewErrHeroFashionNotActivated()
		}
		now := timer.StartTime(ctx.StartTime).Unix()
		if expire > 0 && expire <= now {
			return nil, errmsg.NewErrHeroFashionNotActivated()
		}
	}
	hero.Fashion.Dressed = req.FashionId
	if err := heroDao.Save(ctx, hero); err != nil {
		return nil, err
	}
	ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: []*models.Hero{svc.dao2model(ctx, hero)},
	})
	if err := svc.sync2battle(ctx, hero.BuildId, req.FashionId); err != nil {
		return nil, err
	}
	return &servicepb.Hero_DressFashionResponse{}, nil
}

func (svc *Service) FashionExpiredCheck(ctx *ctx.Context, role *pbdao.Role) *errmsg.ErrMsg {
	heroDao := dao.NewHero(ctx.RoleId)
	heroes, err := heroDao.Get(ctx) // TODO 待优化（登录的时候已经从数据库里取过一次了）
	if err != nil {
		return err
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return err
	}
	now := timer.StartTime(ctx.StartTime).Unix()
	list := make([]*pbdao.Hero, 0)
	expired := make([]values.FashionId, 0)
	for _, hero := range heroes {
		if hero.Fashion == nil {
			hero.Fashion = &models.HeroFashion{}
		}
		if hero.Fashion.Data == nil {
			hero.Fashion.Data = make(map[int64]int64)
		}
		for id, expire := range hero.Fashion.Data {
			if expire == -1 {
				continue
			}
			if expire <= now {
				delete(hero.Fashion.Data, id)
				if hero.Fashion.Dressed == id {
					hero.Fashion.Dressed = rule.GetDefaultFashion(ctx, hero.Id)
				}
				list = append(list, hero)
				expired = append(expired, id)
			}
		}
	}
	if len(expired) <= 0 {
		return nil
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getAllHeroEquippedEquipId(heroes)...)
	if err != nil {
		return err
	}
	svc.refreshAllHeroAttr(ctx, heroes, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill)
	if err := svc.sendFashionExpiredMail(ctx, role.Language, expired...); err != nil {
		return err
	}
	return heroDao.Save(ctx, heroes...)
}

func (svc *Service) ActivateFashion(ctx *ctx.Context, id values.ItemId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	cfg, ok := rule.GetFashionActivate(ctx, id)
	if !ok {
		ctx.Error("fashion activate config not found", zap.String("role_id", ctx.RoleId), zap.Int64("id", id))
		return nil, errmsg.NewInternalErr("fashion activate config not found")
	}
	fashionCfg, ok := rule.GetFashion(ctx, cfg.FashionId)
	if !ok {
		ctx.Error("fashion config not found", zap.String("role_id", ctx.RoleId), zap.Int64("id", cfg.FashionId))
		return nil, errmsg.NewInternalErr("fashion config not found")
	}
	heroDao := dao.NewHero(ctx.RoleId)
	heroes, err := heroDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	var hero *pbdao.Hero
	for _, d := range heroes {
		if d.Id == fashionCfg.Hero {
			hero = d
			break
		}
	}
	if hero == nil {
		return nil, errmsg.NewErrHeroNotFound()
	}
	if hero.Fashion == nil {
		hero.Fashion = &models.HeroFashion{}
	}
	if hero.Fashion.Data == nil {
		hero.Fashion.Data = make(map[int64]int64)
	}
	expire, ok := hero.Fashion.Data[cfg.FashionId]
	// 已有永久时装，转换为道具
	if expire == -1 {
		if len(cfg.TransferItem) > 0 {
			if _, err := svc.AddManyItem(ctx, ctx.RoleId, cfg.TransferItem); err != nil {
				return nil, err
			}
		}
		return cfg.TransferItem, nil
	}
	now := timer.StartTime(ctx.StartTime).Unix()
	// 已存在，直接续期，不用增加属性
	if ok {
		if cfg.Duration == -1 {
			expire = cfg.Duration
			svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskOwnFashionCnt, 0, 1)
			svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskActiveFashion, cfg.FashionId, 1)
			if fashionCfg.IsDefault != 1 {
				svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGetForeverFashion, 0, 1)
			}
		} else {
			// 防止有已过期但未删除的时装
			if expire < now {
				expire = now
			}
			expire += cfg.Duration
		}
		hero.Fashion.Data[cfg.FashionId] = expire
		if err := heroDao.Save(ctx, hero); err != nil {
			return nil, err
		}
		if err = svc.send2FashionServer(ctx, hero.Id, cfg.FashionId, expire); err != nil {
			return nil, err
		}
		ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
			Hero: []*models.Hero{svc.dao2model(ctx, hero)},
		})
		return map[values.ItemId]values.Integer{id: 1}, nil
	}

	// 激活皮肤，同时更新属性和战斗力
	if cfg.Duration == -1 {
		expire = cfg.Duration
		svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskOwnFashionCnt, 0, 1)
		svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskActiveFashion, cfg.FashionId, 1)
		if fashionCfg.IsDefault != 1 {
			svc.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGetForeverFashion, 0, 1)
		}
	} else {
		expire = timer.StartTime(ctx.StartTime).Add(time.Second * time.Duration(cfg.Duration)).Unix()
	}
	hero.Fashion.Data[cfg.FashionId] = expire
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getAllHeroEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	svc.refreshAllHeroAttr(ctx, heroes, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill)

	heroModels := make([]*models.Hero, 0)
	heroAttrUpdateItem := make([]*event.HeroAttrUpdateItem, 0)
	for _, d := range heroes {
		model := svc.dao2model(ctx, d)
		heroModels = append(heroModels, model)
		heroAttrUpdateItem = append(heroAttrUpdateItem, &event.HeroAttrUpdateItem{
			IsSkillChange: false,
			Hero:          model,
		})
	}

	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: heroAttrUpdateItem,
	})
	ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: heroModels,
	})
	if err := heroDao.Save(ctx, heroes...); err != nil {
		return nil, err
	}
	err = svc.send2FashionServer(ctx, hero.Id, cfg.FashionId, expire)
	return map[values.ItemId]values.Integer{id: 1}, err
}

func (svc *Service) FashionExpired(ctx *ctx.Context, req *servicepb.Hero_FashionExpiredRequest) (*servicepb.Hero_FashionExpiredResponse, *errmsg.ErrMsg) {
	ctx.Debug("FashionExpired",
		zap.String("role_id", ctx.RoleId),
		zap.Int64("fashion_id", req.FashionId),
		zap.Int64("hero_id", req.HeroId),
		zap.Int64("time", timer.StartTime(ctx.StartTime).Unix()),
	)
	if ctx.ServerType != models.ServerType_FashionServer {
		ctx.Error("[FashionExpired] invalid server type request", zap.Int32("type", int32(ctx.ServerType)))
		return nil, errmsg.NewProtocolErrorInfo("invalid server type request")
	}
	heroDao := dao.NewHero(ctx.RoleId)
	heroes, err := heroDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	var hero *pbdao.Hero
	for _, d := range heroes {
		if d.Id == req.HeroId {
			hero = d
			break
		}
	}
	if hero == nil {
		return nil, errmsg.NewErrHeroNotFound()
	}
	// 判断一下是不是真的过期了
	expire, ok := hero.Fashion.Data[req.FashionId]
	if !ok {
		return &servicepb.Hero_FashionExpiredResponse{}, nil
	}
	if expire == -1 {
		return &servicepb.Hero_FashionExpiredResponse{}, nil
	}
	if expire > timer.StartTime(ctx.StartTime).Unix() {
		return &servicepb.Hero_FashionExpiredResponse{}, nil
	}

	delete(hero.Fashion.Data, req.FashionId)
	if hero.Fashion.Dressed == req.FashionId {
		hero.Fashion.Dressed = rule.GetDefaultFashion(ctx, req.HeroId)
	}

	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	roleAttr, roleSkill, err := svc.GetRoleAttrAndSkill(ctx)
	if err != nil {
		return nil, err
	}
	equips, err := svc.GetManyEquipBagMap(ctx, ctx.RoleId, svc.getAllHeroEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	svc.refreshAllHeroAttr(ctx, heroes, role.Level, role.LevelIndex, false, equips, roleAttr, roleSkill)

	// 给玩家发时装过期邮件
	if err := svc.sendFashionExpiredMail(ctx, role.Language, req.FashionId); err != nil {
		return nil, err
	}
	heroModels := make([]*models.Hero, 0)
	heroAttrUpdateItem := make([]*event.HeroAttrUpdateItem, 0)
	for _, d := range heroes {
		model := svc.dao2model(ctx, d)
		heroModels = append(heroModels, model)
		heroAttrUpdateItem = append(heroAttrUpdateItem, &event.HeroAttrUpdateItem{
			IsSkillChange: false,
			Hero:          model,
		})
	}
	ctx.PublishEventLocal(&event.HeroAttrUpdate{
		Data: heroAttrUpdateItem,
	})
	ctx.PushMessage(&servicepb.Hero_HeroUpdatePush{
		Hero: heroModels,
	})
	ctx.PushMessage(&servicepb.Hero_FashionExpiredPush{
		HeroOriginId: req.HeroId,
		FashionId:    req.FashionId,
	})
	if err = heroDao.Save(ctx, heroes...); err != nil {
		return nil, err
	}
	if err := svc.sync2battle(ctx, hero.BuildId, req.FashionId); err != nil {
		return nil, err
	}
	return &servicepb.Hero_FashionExpiredResponse{}, err
}

func (svc *Service) sendFashionExpiredMail(ctx *ctx.Context, language values.Integer, list ...values.FashionId) *errmsg.ErrMsg {
	if len(list) <= 0 {
		return nil
	}
	id := values.Integer(enum.FashionExpiredId)
	var expiredAt values.Integer
	cfg, ok := rule.GetMailConfigTextId(ctx, id)
	if ok {
		expiredAt = timer.StartTime(ctx.StartTime).Add(time.Second * time.Duration(cfg.Overdue)).UnixMilli()
	}
	mailList := make([]*models.Mail, 0, len(list))
	for _, fashionId := range list {
		mailList = append(mailList, &models.Mail{
			Id:        xid.New().String(),
			Type:      models.MailType_MailTypeSystem,
			TextId:    id,
			ExpiredAt: expiredAt,
			Args:      []string{svc.getFashionName(ctx, fashionId, language)},
		})
	}
	return svc.BatchAdd(ctx, ctx.RoleId, mailList)
}

func (svc *Service) getFashionName(ctx *ctx.Context, fashionId values.FashionId, language values.Integer) string {
	cfg, ok := rule.GetFashion(ctx, fashionId)
	if !ok {
		ctx.Warn("fashion config not found", zap.Int64("id", fashionId))
		return ""
	}
	id, _ := strconv.Atoi(cfg.Language1)
	languageCfg, ok := rule.GetLanguageBackend(ctx, values.Integer(id))
	if !ok {
		ctx.Warn("language backend config not found", zap.String("id", cfg.Language1))
		return ""
	}
	switch language {
	case enum.DefaultLanguage:
		return languageCfg.En
	case 40:
		return languageCfg.Cn
	case 41:
		return languageCfg.Hk
	default:
		return languageCfg.En
	}
}

func (svc *Service) send2FashionServer(ctx *ctx.Context, heroId values.HeroId, fashionId values.FashionId, expiredAt values.Integer) *errmsg.ErrMsg {
	if expiredAt > 0 {
		_, err := svc.svc.GetNatsClient().Request(ctx, 0, &fashion_service.Fashion_FashionTimerUpdateRequest{
			HeroId:    heroId,
			FashionId: fashionId,
			ExpiredAt: expiredAt,
		})
		return err
	}
	_, err := svc.svc.GetNatsClient().Request(ctx, 0, &fashion_service.Fashion_FashionTimerRemoveRequest{
		FashionId: fashionId,
	})
	return err
}

func (svc *Service) sync2battle(ctx *ctx.Context, heroId values.HeroId, fashionId values.FashionId) *errmsg.ErrMsg {
	battleSrvId, err := svc.GetCurBattleSrvId(ctx)
	if err != nil {
		return err
	}
	if battleSrvId > 0 {
		return svc.svc.GetNatsClient().Publish(battleSrvId, ctx.ServerHeader, &cppbattle.CPPBattle_HeroFashionPush{
			BattleServerId: battleSrvId,
			RoleId:         ctx.RoleId,
			HeroConfigId:   heroId,
			Fashion:        fashionId,
		})
	}
	return nil
}
