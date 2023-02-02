package npc

import (
	"strconv"

	"coin-server/game-server/service/user/db"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/npc/dao"
	"coin-server/game-server/util/trans"
	"coin-server/rule"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
}

func NewNpcService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
	}
	module.NpcService = s
	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("进行对话", s.NpcTalkRequest)
	s.svc.RegisterFunc("开始NPC战斗", s.BattleStart)
	s.svc.RegisterFunc("完成NPC战斗", s.BattleFinish)

	eventlocal.SubscribeEventLocal(s.HandleTalkEvent)
}

func (s *Service) NpcTalkRequest(ctx *ctx.Context, req *servicepb.Npc_TalkRequest) (*servicepb.Npc_TalkResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.NpcDialogue.GetNpcDialogueById(req.DialogId)
	if !ok {
		return nil, nil
	}
	ndr := r.NpcDialogue.GetNpcDialogRelation(req.DialogId)
	e := &event.Talk{
		DialogId:     req.DialogId,
		HeadDialogId: ndr.HeadDialogId,
		IsEnd:        ndr.IsEnd,
		TaskId:       req.TaskId,
		Kind:         req.Kind,
		Typ:          req.Typ,
	}
	if cfg.Typ != 1 {
		ctx.PublishEventLocal(e)
		return nil, nil
	}

	opt := r.NpcDialogue.GetNpcDialogOpts(req.DialogId)
	if len(opt) == 0 {
		return nil, nil
	}

	var reward map[values.ItemId]values.Integer
	for idx := range opt {
		if opt[idx].Id == req.Opt {
			reward = opt[idx].OptAward
			e.OptIdx = values.Integer(idx)
			break
		}
	}
	ctx.PublishEventLocal(e)
	if reward == nil {
		return &servicepb.Npc_TalkResponse{}, nil
	}

	// 有奖励的情况
	talk, err := dao.GetNpcTalk(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	talk.TalkReward[req.DialogId] = true
	_, err = s.BagService.AddManyItem(ctx, ctx.RoleId, reward)
	if err != nil {
		return nil, err
	}
	dao.SaveNpcTalk(ctx, talk)
	return &servicepb.Npc_TalkResponse{}, nil
}

// BattleStart 开启NPC副本
func (s *Service) BattleStart(c *ctx.Context, req *servicepb.Npc_NPCDungeonStartRequest) (*servicepb.Npc_NPCDungeonStartResponse, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)

	// 获取突破配置
	cfg, ok := reader.NpcChallengeDungeon.GetNpcChallengeDungeonById(req.Id)
	if !ok {
		return nil, errmsg.NewInternalErr("npc_challenge_dungeon not found: " + strconv.Itoa(int(req.Id)))
	}

	// 获取角色信息
	role, err := s.Module.GetRoleModelByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	// 获取英雄信息
	heroesFormation, err := s.Module.FormationService.GetDefaultHeroes(c, c.RoleId)
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
	heroes, err := s.Module.GetHeroes(c, c.RoleId, heroIds)
	if err != nil {
		return nil, err
	}
	equips, err := s.GetManyEquipBagMap(c, c.RoleId, s.GetHeroesEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	cppHeroes := trans.Heroes2CppHeroes(c, heroes, equips)
	if len(cppHeroes) == 0 {
		return nil, errmsg.NewErrHeroNotFound()
	}

	d, err00 := db.GetBattleSetting(c)
	if err00 != nil {
		return nil, err00
	}

	return &servicepb.Npc_NPCDungeonStartResponse{
		BattleId: -10000, // 直接写死-10000 客户端必须，对于服务器无意义
		Sbp: &models.SingleBattleParam{
			Role:             role,
			Heroes:           cppHeroes,
			MonsterGroupInfo: cfg.MonsterGroupInfo,
			CountDown:        cfg.Duration,
			AutoSoulSkill:    d.Data.AutoSoulSkill,
		},
	}, nil
}

// BattleFinish NPC挑战副本结束
func (s *Service) BattleFinish(c *ctx.Context, req *servicepb.Npc_NPCDungeonFinishRequest) (*servicepb.Npc_NPCDungeonFinishResponse, *errmsg.ErrMsg) {
	c.PublishEventLocal(&event.NpcDungeonFinish{
		RoleId:    c.RoleId,
		DungeonId: req.Id,
		IsSuccess: req.IsSuccess,
		TaskType:  req.Typ,
	})
	if req.IsSuccess {
		s.TaskService.UpdateTarget(c, c.RoleId, req.Typ, req.Id, 1)
	}
	return &servicepb.Npc_NPCDungeonFinishResponse{}, nil
}
