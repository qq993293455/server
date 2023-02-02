package guild_gvg

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/util/trans"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	module     *module.Module
	log        *logger.Logger
}

func NewGuildGVGService(
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
		module:     module,
		log:        log,
	}
	return s
}

func (this_ *Service) Router() {
	h := this_.svc.Group(handler.GVGServerAuth)
	h.RegisterFunc("获取玩家工会GVG战斗数据", this_.GetGVGFightInfo)
}

func (this_ *Service) GetGVGFightInfo(c *ctx.Context, request *servicepb.GuildGVG_GetGVGFightInfoRequest) (*servicepb.GuildGVG_GetGVGFightInfoResponse, *errmsg.ErrMsg) {
	// 获取角色信息
	role, err := this_.module.GetRoleModelByRoleId(c, request.RoleId)
	if err != nil {
		return nil, err
	}

	// 获取英雄信息
	heroesFormation, err := this_.module.FormationService.GetDefaultHeroes(c, request.RoleId)
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
	heroes, err := this_.module.GetHeroes(c, request.RoleId, heroIds)
	if err != nil {
		return nil, err
	}
	equips, err := this_.module.GetManyEquipBagMap(c, request.RoleId, this_.module.GetHeroesEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	cppHeroes := trans.Heroes2CppHeroes(c, heroes, equips)
	if len(cppHeroes) == 0 {
		return nil, errmsg.NewErrHeroNotFound()
	}

	return &servicepb.GuildGVG_GetGVGFightInfoResponse{Sbp: &models.SingleBattleParam{
		Role:   role,
		Heroes: cppHeroes,
	}}, nil
}
