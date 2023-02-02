package plane

import (
	"fmt"
	"strconv"

	"coin-server/common/proto/cppbattle"
	"coin-server/game-server/service/user/db"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/plane-dungeon/dao"
	"coin-server/game-server/util/trans"
	"coin-server/rule"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewPlaneService(
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
	svc.svc.RegisterFunc("开启位面副本", svc.Start)
	// svc.svc.RegisterFunc("位面副本结束", svc.Finish)

	svc.svc.RegisterEvent("位面副本战斗推送", svc.PlaneDungeonBattlePush)

}

// Start 开启位面副本
func (svc *Service) Start(c *ctx.Context, req *servicepb.PlaneDungeon_PlaneDungeonStartRequest) (*servicepb.PlaneDungeon_PlaneDungeonStartResponse, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)

	if c.BattleServerId == 0 || c.BattleMapId == 0 {
		return nil, errmsg.NewInternalErr("invalid battleId or MapId")
	}
	mapCnf, ok0 := reader.MapScene.GetMapSceneById(c.BattleMapId)
	if !ok0 || mapCnf.MapType != values.Integer(models.BattleType_HangUp) {
		return nil, errmsg.NewInternalErr("not in hangup map")
	}
	// 获取突破配置
	cfg, ok := reader.PlaneDungeon.GetPlaneDungeonById(req.Id)
	if !ok {
		return nil, errmsg.NewInternalErr("plane_dungeon not found: " + strconv.Itoa(int(req.Id)))
	}
	if len(cfg.MonsterGroupInfo) == 0 {
		panic(fmt.Sprintf("no monsters! %d", req.Id))
	}

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

	medicines, err := svc.BagService.GetMedicineMsg(c, c.RoleId, cfg.MapScene)
	if err != nil {
		return nil, err
	}

	d, err00 := db.GetBattleSetting(c)
	if err00 != nil {
		return nil, err00
	}

	out1 := &cppbattle.CPPBattle_EnterPlaneDungeonResponse{}
	err = svc.svc.GetNatsClient().RequestWithOut(c, c.BattleServerId, &cppbattle.CPPBattle_EnterPlaneDungeonRequest{
		Id:  req.Id,
		Typ: int64(req.Typ),
		Sbp: &models.SingleBattleParam{
			Role:             role,
			Heroes:           cppHeroes,
			MonsterGroupInfo: cfg.MonsterGroupInfo,
			CountDown:        cfg.Duration,
			Medicine:         medicines,
			AutoSoulSkill:    d.Data.AutoSoulSkill,
		},
		RewardItems: cfg.DropReward,
		TreeName:    cfg.Bt,
		MapScenceId: cfg.MapScene,
		Duration:    cfg.Duration,
	}, out1)

	return &servicepb.PlaneDungeon_PlaneDungeonStartResponse{}, nil
}

// Finish 位面副本结束
func (svc *Service) Finish(c *ctx.Context, req *servicepb.PlaneDungeon_PlaneDungeonFinishRequest) (*servicepb.PlaneDungeon_PlaneDungeonFinishResponse, *errmsg.ErrMsg) {
	cfg, ok := rule.MustGetReader(c).PlaneDungeon.GetPlaneDungeonById(req.Id)
	if !ok {
		return nil, errmsg.NewErrDungeonNotExist()
	}

	var rewards map[int64]int64
	if req.IsSuccess {
		plane, err := dao.GetPlaneDungeon(c)
		if err != nil {
			return nil, err
		}
		// 只能获得一次奖励
		if !plane.Finished[req.Id] {
			plane.Finished[req.Id] = true
			rewards = cfg.DropReward
			_, err := svc.BagService.AddManyItem(c, c.RoleId, rewards)
			if err != nil {
				return nil, err
			}
			dao.SavePlaneDungeon(c, plane)
		}

		svc.TaskService.UpdateTarget(c, c.RoleId, req.Typ, req.Id, 1)
	}

	return &servicepb.PlaneDungeon_PlaneDungeonFinishResponse{Rewards: rewards}, nil
}

func (svc *Service) PlaneDungeonBattlePush(c *ctx.Context, msg *servicepb.PlaneDungeon_PlaneDungeonBattlePush) {
	cfg, ok := rule.MustGetReader(c).PlaneDungeon.GetPlaneDungeonById(msg.Id)
	if !ok {
		return
	}

	var rewards map[int64]int64
	if msg.Result == 2 {
		plane, err := dao.GetPlaneDungeon(c)
		if err != nil {
			return
		}
		// 只能获得一次奖励
		if !plane.Finished[msg.Id] {
			plane.Finished[msg.Id] = true
			rewards = cfg.DropReward
			_, err := svc.BagService.AddManyItem(c, c.RoleId, rewards)
			if err != nil {
				c.Error("call PlaneDungeonBattlePush AddManyItem error", zap.Error(err))
				return
			}
			dao.SavePlaneDungeon(c, plane)
		}

		svc.TaskService.UpdateTarget(c, c.RoleId, models.TaskType(msg.Typ), msg.Id, 1)
	}
}
