package personalboss

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/im"
	daopb "coin-server/common/proto/dao"
	modelspb "coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/redisclient"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/utils/generic/maps"
	"coin-server/common/utils/imutil"
	"coin-server/common/utils/percent"
	"coin-server/common/values"
	"coin-server/common/values/enum/AttrId"
	"coin-server/game-server/module"
	values2 "coin-server/game-server/service/journey/values"
	"coin-server/game-server/service/personal-boss/dao"
	rule2 "coin-server/game-server/service/personal-boss/rule"
	"coin-server/game-server/service/user/db"
	"coin-server/game-server/util"
	"coin-server/game-server/util/trans"
	"coin-server/rule"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/xid"
)

type Service struct {
	serverId   values.ServerId
	serverType modelspb.ServerType
	svc        *service.Service
	*module.Module
}

func NewPersonalBossService(
	serverId values.ServerId,
	serverType modelspb.ServerType,
	svc *service.Service,
	module *module.Module,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
	}
	s.Module.PersonalBossService = s
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取个人BOSS信息", svc.PersonalBossInfo)
	svc.svc.RegisterFunc("开启个人BOSS战斗", svc.PersonalBossStart)
	svc.svc.RegisterFunc("个人BOSS战斗结束", svc.PersonalBossFinish)
	svc.svc.RegisterFunc("领取最终奖励", svc.DrawPassReward)
	svc.svc.RegisterFunc("保存助战效果方案", svc.SaveHelperCards)
	svc.svc.RegisterFunc("获取我的助战信息", svc.PersonalBossHelperInfo)
	svc.svc.RegisterFunc("向聊天频道发送助力申请", svc.SendHelpMessage)
	svc.svc.RegisterFunc("助力其他玩家", svc.HelpPersonalBoss)
	svc.svc.RegisterFunc("钻石购买助战点数", svc.BuyPersonalBossPoint)

}

func (svc *Service) genBossList(c *ctx.Context) []*modelspb.PersonalBoss {
	reader := rule.MustGetReader(c)

	floors := reader.PersonalBossNumberFloor.List()
	buffs := reader.PersonalBossBuff.List()
	ret := make([]*modelspb.PersonalBoss, 0, len(floors))

	bosses := rule2.GetPersonalBossLibrary()
	for _, floorCfg := range floors {
		bossCfg := bosses[rand.Intn(len(bosses))]
		rand.Shuffle(len(buffs), func(i, j int) {
			buffs[i], buffs[j] = buffs[j], buffs[i]
		})
		pb := &modelspb.PersonalBoss{
			Floor:         floorCfg.Id,
			BossLibraryId: bossCfg.Id,
			BossBuffIds:   make([]int64, 0, int(floorCfg.BossBuffNum)),
		}
		for i := 0; i < int(floorCfg.BossBuffNum); i++ {
			pb.BossBuffIds = append(pb.BossBuffIds, buffs[i].BossBuffId)
		}
		ret = append(ret, pb)
	}

	return ret
}

func (svc *Service) getGlobalPersonalBossInfo(c *ctx.Context) (*daopb.GlobalPersonalBossInfo, *errmsg.ErrMsg) {
	global, err := dao.GetGlobalPersonalBossInfo(c)
	if err != nil {
		return nil, err
	}
	day := rule2.MustGetPersonalBossEventResetTime()

	if len(global.BossList) == 0 || global.ResetAt < timer.Now().UnixMilli() {
		global.BossList = svc.genBossList(c)
		global.ResetAt = util.DefaultTodayRefreshTime().AddDate(0, 0, day).UnixMilli()
		dao.SaveGlobalPersonalBossInfo(c, global)
	}

	return global, nil
}

func (svc *Service) resetPersonalBossHelperInfo(c *ctx.Context, roleId values.RoleId) *errmsg.ErrMsg {
	pbh, err := svc.getPersonalBossHelperInfo(c, roleId)
	if err != nil {
		return err
	}
	pbh.Pbh = &modelspb.PersonalBossHelperInfo{
		TotalPoint: rule2.MustGetInitialValueOfHelperPoints(),
	}
	pbh.Share = nil
	dao.SavePersonalBossHelperInfo(c, pbh)
	return nil
}

func (svc *Service) getPersonalBossInfo(c *ctx.Context, roleId values.RoleId) (*daopb.GlobalPersonalBossInfo, *daopb.PersonalBossInfo, *errmsg.ErrMsg) {
	global, err := svc.getGlobalPersonalBossInfo(c)
	if err != nil {
		return nil, nil, err
	}

	info, err := dao.GetPersonalBossInfo(c, roleId)
	if err != nil {
		return nil, nil, err
	}
	if info.Pbi == nil {
		info.Pbi = &modelspb.PersonalBossInfo{}
	}

	if info.Pbi.CurFloor == 1 { // TODO 临时 修正已有问题的账号
		info.Pbi.IsCompleted = false
		info.Pbi.IsDrawPassReward = false
	}

	if info.Pbi.ResetAt < global.ResetAt {
		if info.Pbi.IsCompleted && !info.Pbi.IsDrawPassReward { // 未领取最终奖励 邮件发放
			err := svc.sendPassRewardMail(c, info.Pbi.PassRewardId)
			if err != nil {
				return nil, nil, err
			}
		}

		reader := rule.MustGetReader(c)
		role, err := svc.UserService.GetRoleByRoleId(c, roleId)
		if err != nil {
			return nil, nil, err
		}
		info.Pbi.ResetAt = global.ResetAt
		info.Pbi.CurFloor = 1
		info.Pbi.IsCompleted = false
		info.Pbi.IsDrawPassReward = false

		// 根据等级获取boss属性
		monsterAttr := reader.MonsterAttr.GetPersonalBossMonsterAttr()
		ma, ok := monsterAttr[role.Level]
		if !ok {
			panic(errors.New(fmt.Sprintf("PersonalBossMonsterAttr lv %d not found", role.Level)))
		}
		info.Pbi.MonsterAttrId = ma.Id

		// 重置助战信息
		info.Pbi.Helper = &modelspb.PersonalBossHelper{
			TotalPoint: rule2.MustGetInitialValueOfHelperPoints(),
			UsedPoint:  0,
			Cards:      map[int64]int64{},
		}

		info.Pbi.RoleLv = role.Level
		for _, cfg := range reader.PersonalBossPassingReward.List() {
			if len(cfg.PlayerLevelRange) < 2 {
				panic(errors.New(fmt.Sprintf("PersonalBossPassingReward PlayerLevelRange config failed. id: %d", cfg.Id)))
			}
			if cfg.PlayerLevelRange[1] == -1 {
				cfg.PlayerLevelRange[1] = math.MaxInt64
			}
			if role.Level >= cfg.PlayerLevelRange[0] && role.Level <= cfg.PlayerLevelRange[1] {
				info.Pbi.PassRewardId = cfg.Id
				break
			}
		}

		err = svc.resetPersonalBossHelperInfo(c, roleId)
		if err != nil {
			return nil, nil, err
		}

		dao.SavePersonalBossInfo(c, info)
	}

	return global, info, nil
}

func (svc *Service) sendPassRewardMail(c *ctx.Context, passRewardId values.Integer) *errmsg.ErrMsg {
	cfg, ok := rule.MustGetReader(c).PersonalBossPassingReward.GetPersonalBossPassingRewardById(passRewardId)
	if !ok {
		panic(fmt.Sprintf("PersonalBossPassingReward config not found. id: %d", passRewardId))
	}
	items := make([]*modelspb.Item, 0, len(cfg.ActivityReward))

	for id, cnt := range cfg.ActivityReward {
		items = append(items, &modelspb.Item{ItemId: id, Count: cnt})
	}

	mail := &modelspb.Mail{
		Type:       modelspb.MailType_MailTypeSystem,
		TextId:     100032,
		Attachment: items,
	}
	err := svc.MailService.Add(c, c.RoleId, mail)
	if err != nil {
		return err
	}

	return nil
}

func (svc *Service) getPersonalBossHelperInfo(c *ctx.Context, roleId values.RoleId) (*daopb.PersonalBossHelperInfo, *errmsg.ErrMsg) {
	err := c.DRLock(redisclient.GetLocker(), helpLock+roleId)
	if err != nil {
		return nil, err
	}
	data, err := dao.GetPersonalBossHelperInfo(c, roleId)
	if data.Share != nil && data.Share.HelperMap == nil {
		data.Share.HelperMap = map[string]bool{}
	}
	if data.Pbh == nil {
		data.Pbh = &modelspb.PersonalBossHelperInfo{
			TotalPoint: rule2.MustGetInitialValueOfHelperPoints(),
		}
	}
	return data, nil
}

func (svc *Service) GetPersonalBossResetRemainSec(c *ctx.Context) (values.Integer, *errmsg.ErrMsg) {
	global, err := svc.getGlobalPersonalBossInfo(c)
	if err != nil {
		return -1, err
	}
	return global.ResetAt/1000 - timer.Now().Unix(), nil
}

func (svc *Service) PersonalBossInfo(c *ctx.Context, _ *servicepb.PersonalBoss_PersonalBossInfoRequest) (*servicepb.PersonalBoss_PersonalBossInfoResponse, *errmsg.ErrMsg) {
	global, info, err := svc.getPersonalBossInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	pbh, err := svc.getPersonalBossHelperInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	info.Pbi.Helper.TotalPoint = pbh.Pbh.TotalPoint
	info.Pbi.BossList = global.BossList
	return &servicepb.PersonalBoss_PersonalBossInfoResponse{Info: info.Pbi}, nil
}

// 生成怪物信息
func (svc *Service) genMonsterInfo(c *ctx.Context, info *daopb.PersonalBossInfo, boss *modelspb.PersonalBoss) *modelspb.MonsterInfo {
	reader := rule.MustGetReader(c)

	blCfg, ok := reader.PersonalBossLibrary.GetPersonalBossLibraryById(boss.BossLibraryId)
	if !ok {
		panic(fmt.Sprintf("PersonalBossHelperCards config not found. id: %d", boss.BossLibraryId))
	}
	attrCfg, ok := reader.MonsterAttr.GetMonsterAttrById(info.Pbi.MonsterAttrId)
	if !ok {
		panic(fmt.Sprintf("PersonalBossHelperCards config not found. id: %d", boss.BossLibraryId))
	}
	floorCfg, ok := reader.PersonalBossNumberFloor.GetPersonalBossNumberFloorById(info.Pbi.CurFloor)
	if !ok {
		panic(errors.New(fmt.Sprintf("PersonalBossNumberFloor config not found. id: %d", info.Pbi.CurFloor)))
	}
	pblCfg := rule2.GetPersonalBossLvByRoleLv(info.Pbi.CurFloor, info.Pbi.RoleLv)

	attr := maps.Copy(attrCfg.ParameterQuality)
	// boss 属性计算
	attr[AttrId.Hp] = percent.Addition(attr[AttrId.Hp], blCfg.HpCoefficient)
	attr[AttrId.Hp] = percent.Addition(attr[AttrId.Hp], pblCfg.BossHp)
	attr[AttrId.Atk] = percent.Addition(attr[AttrId.Atk], blCfg.AtkCoefficient)
	attr[AttrId.Atk] = percent.Addition(attr[AttrId.Atk], pblCfg.BossAtk)
	attr[AttrId.Def] = percent.Addition(attr[AttrId.Def], blCfg.DefCoefficient)
	attr[AttrId.Def] = percent.Addition(attr[AttrId.Def], pblCfg.BossDef)

	return &modelspb.MonsterInfo{
		MonsterId: blCfg.BossId,
		Attr:      attr,
		AreaId:    floorCfg.BossBirthPoint,
		MonsterLv: info.Pbi.RoleLv,
	}
}

// PersonalBossStart 开启个人BOSS战斗
func (svc *Service) PersonalBossStart(c *ctx.Context, _ *servicepb.PersonalBoss_PersonalBossStartRequest) (*servicepb.PersonalBoss_PersonalBossStartResponse, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)

	global, info, err := svc.getPersonalBossInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if info.Pbi.IsCompleted {
		return nil, errmsg.NewErrPersonalBossCompleted()
	}
	boss := global.BossList[info.Pbi.CurFloor-1]

	monsterInfo := svc.genMonsterInfo(c, info, boss)

	// 获取角色信息
	role, err := svc.Module.GetRoleModelByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	// 获取英雄信息
	heroesFormation, err := svc.Module.FormationService.GetDefaultHeroes(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	heroIds := make([]int64, 0, 2)
	if heroesFormation.Hero_0 > 0 && heroesFormation.HeroOrigin_0 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_0)
	}
	if heroesFormation.Hero_1 > 0 && heroesFormation.HeroOrigin_1 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_1)
	}
	heroes, err := svc.Module.GetHeroes(c, c.RoleId, heroIds)
	if err != nil {
		return nil, err
	}
	equips, err := svc.GetManyEquipBagMap(c, c.RoleId, svc.GetHeroesEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	cppHeroes := trans.Heroes2CppHeroes(c, heroes, equips)
	if len(cppHeroes) == 0 {
		return nil, errmsg.NewErrHeroNotFound()
	}

	// 获取玩家的buff
	buffs := make([]int64, 0)
	for k := range info.Pbi.Helper.Cards {
		cfg, ok := reader.PersonalBossHelperCards.GetPersonalBossHelperCardsById(k)
		if !ok {
			panic(fmt.Sprintf("PersonalBossHelperCards config not found. id: %d", k))
		}
		buffs = append(buffs, cfg.CardBuffId)
	}
	if len(buffs) > 0 {
		for _, hero := range cppHeroes {
			hero.BuffIds = append(hero.BuffIds, buffs...)
		}
	}

	// 获取BOSS的buff
	monsterBuffs := make([]*modelspb.BuffInfo, 0)
	for _, id := range boss.BossBuffIds {
		monsterBuffs = append(monsterBuffs, &modelspb.BuffInfo{BuffId: id, Overlay: 1})
	}

	curBattleInfo, err := svc.BattleService.GetCurrBattleInfo(c, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err != nil {
		return nil, err
	}
	// 吃药信息
	medicines, err := svc.BagService.GetMedicineMsg(c, c.RoleId, curBattleInfo.HungupMapId)
	if err != nil {
		return nil, err
	}

	d, err00 := db.GetBattleSetting(c)
	if err00 != nil {
		return nil, err00
	}

	return &servicepb.PersonalBoss_PersonalBossStartResponse{
		BattleId: -10000, // 直接写死-10000 客户端必须，对于服务器无意义
		Sbp: &modelspb.SingleBattleParam{
			Role:          role,
			Heroes:        cppHeroes,
			CountDown:     rule2.MustGetPersonalBossEventCountdownTime(),
			MonsterBuffs:  monsterBuffs,
			Monsters:      []*modelspb.MonsterInfo{monsterInfo},
			Medicine:      medicines,
			AutoSoulSkill: d.Data.AutoSoulSkill,
		},
	}, nil
}

// PersonalBossFinish 个人BOSS战斗结束
func (svc *Service) PersonalBossFinish(c *ctx.Context, req *servicepb.PersonalBoss_PersonalBossFinishRequest) (*servicepb.PersonalBoss_PersonalBossFinishResponse, *errmsg.ErrMsg) {
	if !req.IsSuccess {
		return &servicepb.PersonalBoss_PersonalBossFinishResponse{}, nil
	}

	global, info, err := svc.getPersonalBossInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if info.Pbi.IsCompleted {
		return nil, errmsg.NewErrPersonalBossCompleted()
	}

	reader := rule.MustGetReader(c)
	pblCfg := rule2.GetPersonalBossLvByRoleLv(info.Pbi.CurFloor, info.Pbi.RoleLv)

	rewards := reader.GetDropList(pblCfg.DropListId)
	_, err = svc.BagService.AddManyItem(c, c.RoleId, rewards)
	if err != nil {
		return nil, err
	}
	if info.Pbi.CurFloor == global.BossList[len(global.BossList)-1].Floor {
		info.Pbi.IsCompleted = true
		svc.TaskService.UpdateTarget(c, c.RoleId, modelspb.TaskType_TaskEndPersonalBossRoundAcc, 0, 1)
		svc.TaskService.UpdateTarget(c, c.RoleId, modelspb.TaskType_TaskEndPersonalBossRound, 0, 1)
	} else {
		info.Pbi.CurFloor++
	}
	err = svc.Module.JourneyService.AddToken(c, c.RoleId, values2.JourneyPersonalBoss, 1)
	if err != nil {
		return nil, err
	}

	dao.SavePersonalBossInfo(c, info)
	svc.TaskService.UpdateTarget(c, c.RoleId, modelspb.TaskType_TaskPersonalBossKill, 0, 1)
	svc.TaskService.UpdateTarget(c, c.RoleId, modelspb.TaskType_TaskJoinPersonBossAcc, 0, 1)
	svc.TaskService.UpdateTarget(c, c.RoleId, modelspb.TaskType_TaskJoinPersonBoss, 0, 1)
	return &servicepb.PersonalBoss_PersonalBossFinishResponse{Rewards: rewards}, nil
}

func (svc *Service) DrawPassReward(c *ctx.Context, _ *servicepb.PersonalBoss_DrawPassRewardRequest) (*servicepb.PersonalBoss_DrawPassRewardResponse, *errmsg.ErrMsg) {
	_, info, err := svc.getPersonalBossInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if !info.Pbi.IsCompleted {
		return nil, errmsg.NewErrPersonalBossNotCompleted()
	}
	if info.Pbi.IsDrawPassReward {
		return nil, errmsg.NewErrPersonalBossAlreadyDrawPassReward()
	}
	cfg, ok := rule.MustGetReader(c).PersonalBossPassingReward.GetPersonalBossPassingRewardById(info.Pbi.PassRewardId)
	if !ok {
		panic(fmt.Sprintf("PersonalBossPassingReward config not found. id: %d", info.Pbi.PassRewardId))
	}
	info.Pbi.IsDrawPassReward = true

	_, err = svc.BagService.AddManyItem(c, c.RoleId, cfg.ActivityReward)
	if err != nil {
		return nil, err
	}
	dao.SavePersonalBossInfo(c, info)
	return &servicepb.PersonalBoss_DrawPassRewardResponse{Rewards: cfg.ActivityReward}, nil
}

func (svc *Service) SaveHelperCards(c *ctx.Context, req *servicepb.PersonalBoss_SaveHelperCardsRequest) (*servicepb.PersonalBoss_SaveHelperCardsResponse, *errmsg.ErrMsg) {
	cards, total := rule2.CalCardsPoint(req.Cards)

	_, info, err := svc.getPersonalBossInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	pbh, err := svc.getPersonalBossHelperInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	if pbh.Pbh.TotalPoint < total { // 助战点不足
		return nil, errmsg.NewErrPersonalBossPointNotEnough()
	}

	info.Pbi.Helper.Cards = cards
	info.Pbi.Helper.UsedPoint = total
	dao.SavePersonalBossInfo(c, info)
	return &servicepb.PersonalBoss_SaveHelperCardsResponse{}, nil
}

func (svc *Service) PersonalBossHelperInfo(c *ctx.Context, _ *servicepb.PersonalBoss_PersonalBossHelperInfoRequest) (*servicepb.PersonalBoss_PersonalBossHelperInfoResponse, *errmsg.ErrMsg) {
	pbh, err := svc.getPersonalBossHelperInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.PersonalBoss_PersonalBossHelperInfoResponse{Info: pbh.Pbh}, nil
}

func (svc *Service) SendHelpMessage(c *ctx.Context, req *servicepb.PersonalBoss_SendHelpMessageRequest) (*servicepb.PersonalBoss_SendHelpMessageResponse, *errmsg.ErrMsg) {
	pbh, err := svc.getPersonalBossHelperInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	role, err := svc.UserService.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if pbh.Pbh.Cd >= timer.Now().UnixMilli() {
		return nil, errmsg.NewErrPersonalBossHelpShareCd()
	}
	cd := rule2.MustGetPersonalBossSharingInterval()
	pbh.Pbh.Cd = timer.Now().UnixMilli() + cd*1000

	if pbh.Share == nil {
		global, err := svc.getGlobalPersonalBossInfo(c)
		if err != nil {
			return nil, err
		}
		pbh.Share = &modelspb.HelpMessage{
			Id: xid.New().String(),
			Role: &modelspb.HelperRole{
				RoleId:      role.RoleId,
				Nickname:    role.Nickname,
				Level:       role.Level,
				AvatarId:    role.AvatarId,
				AvatarFrame: role.AvatarFrame,
			},
			BeHelpedCount: 0,
			Records:       make([]*modelspb.HelperRecord, 0),
			Expires:       global.ResetAt,
		}
		dao.SavePersonalBossHelperInfo(c, pbh)
	}

	content, _ := jsoniter.MarshalToString(&HelpShare{
		RoleId:    role.RoleId,
		HelpMsgId: pbh.Share.Id,
	})
	for i, ok := range req.Channels {
		if !ok {
			continue
		}
		switch i {
		case 0: // 世界
			utils.Must(im.DefaultClient.SendMessage(c, &im.Message{
				Type:      im.MsgTypeBroadcast,
				RoleID:    role.RoleId,
				RoleName:  role.Nickname,
				Content:   content,
				ParseType: im.ParseTypePersonalBossHelp,
				Extra:     imutil.GetIMRoleInfoExtra(role),
			}))
		case 1: // 同语种
			utils.Must(im.DefaultClient.SendMessage(c, &im.Message{
				Type:      im.MsgTypeRoom,
				RoleID:    role.RoleId,
				RoleName:  role.Nickname,
				TargetID:  strconv.Itoa(int(role.Language)),
				Content:   content,
				ParseType: im.ParseTypePersonalBossHelp,
				Extra:     imutil.GetIMRoleInfoExtra(role),
			}))
		case 2: // 公会
			guildId, err := svc.GuildService.GetGuildIdByRole(c)
			if err != nil {
				return nil, err
			}
			if guildId == "" {
				continue
			}
			utils.Must(im.DefaultClient.SendMessage(c, &im.Message{
				Type:      im.MsgTypeRoom,
				RoleID:    role.RoleId,
				RoleName:  role.Nickname,
				TargetID:  guildId,
				Content:   content,
				ParseType: im.ParseTypePersonalBossHelp,
				Extra:     imutil.GetIMRoleInfoExtra(role),
			}))
		}
	}

	return &servicepb.PersonalBoss_SendHelpMessageResponse{Cd: pbh.Pbh.Cd}, nil
}

// HelpPersonalBoss 助力其他玩家
func (svc *Service) HelpPersonalBoss(c *ctx.Context, req *servicepb.PersonalBoss_HelpPersonalBossRequest) (*servicepb.PersonalBoss_HelpPersonalBossResponse, *errmsg.ErrMsg) {
	// 被助战对象的判断
	target, err := svc.getPersonalBossHelperInfo(c, req.RoleId)
	if err != nil {
		return nil, err
	}
	if target.Share == nil || target.Share.Id != req.HelpMsgId || target.Share.Expires < timer.Now().UnixMilli() {
		return nil, errmsg.NewErrPersonalBossHelpMsgExpired()
	}
	if c.RoleId == req.RoleId {
		return &servicepb.PersonalBoss_HelpPersonalBossResponse{Hm: target.Share}, nil
	}
	if _, ok := target.Share.HelperMap[c.RoleId]; ok { // 已助战
		return &servicepb.PersonalBoss_HelpPersonalBossResponse{Hm: target.Share}, nil
	}
	limit := rule2.MustGetPersonalBossHelpByPointsNumber() // 可被助战的次数
	if target.Share.BeHelpedCount >= limit {
		return &servicepb.PersonalBoss_HelpPersonalBossResponse{Hm: target.Share}, nil
	}
	limit = rule2.MustGetPersonalBossGetHelpPointsUpperLimit() // 获得助战点上限
	if target.Pbh.TotalPoint >= limit {
		return &servicepb.PersonalBoss_HelpPersonalBossResponse{Hm: target.Share}, nil
	}

	// 助战人的判断
	my, err := svc.getPersonalBossHelperInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	limit = rule2.MustGetPersonalBossHelpPointsNumber() // 可助战的次数
	if my.Pbh.HelpCount >= limit {
		return &servicepb.PersonalBoss_HelpPersonalBossResponse{Hm: target.Share}, nil
	}

	target.Share.BeHelpedCount++
	point := rule2.MustGetPersonalBossHelpPoints()
	// TODO 好友和公会成员额外加1点
	target.Pbh.TotalPoint += point
	c.PushMessageToRole(req.RoleId, &servicepb.PersonalBoss_AddHelpPointPush{
		Point: point,
		Total: target.Pbh.TotalPoint,
	})
	my.Pbh.HelpCount++

	targetRole, err := svc.UserService.GetRoleByRoleId(c, req.RoleId)
	if err != nil {
		return nil, err
	}
	role, err := svc.UserService.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	reward := map[int64]int64{rule2.RandHelpReward(role.Level): 1}
	_, err = svc.BagService.AddManyItem(c, c.RoleId, reward)
	if err != nil {
		return nil, err
	}

	record := &modelspb.HelperRecord{
		Role: &modelspb.HelperRole{
			RoleId:      role.RoleId,
			Nickname:    role.Nickname,
			Level:       role.Level,
			AvatarId:    role.AvatarId,
			AvatarFrame: role.AvatarFrame,
		},
		Target: &modelspb.HelperRole{
			RoleId:      targetRole.RoleId,
			Nickname:    targetRole.Nickname,
			Level:       targetRole.Level,
			AvatarId:    targetRole.AvatarId,
			AvatarFrame: targetRole.AvatarFrame,
		},
		Reward:   reward,
		Point:    point,
		HelpTime: timer.Now().UnixMilli(),
	}

	target.Pbh.Records = append(target.Pbh.Records, record)
	target.Share.Records = append(target.Share.Records, record)
	target.Share.HelperMap[c.RoleId] = true
	my.Pbh.Records = append(my.Pbh.Records, record)

	dao.SavePersonalBossHelperInfo(c, target)
	dao.SavePersonalBossHelperInfo(c, my)
	return &servicepb.PersonalBoss_HelpPersonalBossResponse{Hm: target.Share, MyHelpCount: my.Pbh.HelpCount}, nil
}

// BuyPersonalBossPoint 直接购买助战点
func (svc *Service) BuyPersonalBossPoint(c *ctx.Context, _ *servicepb.PersonalBoss_BuyPersonalBossPointRequest) (*servicepb.PersonalBoss_BuyPersonalBossPointResponse, *errmsg.ErrMsg) {
	my, err := svc.getPersonalBossHelperInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	limit := rule2.MustGetPersonalBossGetHelpPointsUpperLimit() // 获得助战点上限
	if my.Pbh.TotalPoint >= limit {
		return nil, errmsg.NewErrPersonalBossPointUpperLimit()
	}

	price := rule2.MustGetPersonalBossHelpPointsExchangePrice()
	err = svc.BagService.SubManyItem(c, c.RoleId, price)
	if err != nil {
		return nil, err
	}

	my.Pbh.TotalPoint++
	dao.SavePersonalBossHelperInfo(c, my)
	c.PushMessage(&servicepb.PersonalBoss_AddHelpPointPush{
		Point: 1,
		Total: my.Pbh.TotalPoint,
	})
	return &servicepb.PersonalBoss_BuyPersonalBossPointResponse{}, nil
}
