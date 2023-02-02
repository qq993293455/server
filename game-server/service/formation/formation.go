package formation

import (
	"fmt"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/handler"
	"coin-server/common/logger"
	"coin-server/common/proto/cppbattle"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/sensitive"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/formation/dao"
	"coin-server/game-server/util/trans"
	"coin-server/rule"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	module     *module.Module
	log        *logger.Logger
}

func NewFormationService(
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
	s.module.FormationService = s
	return s
}

// IsUnlock 中间件，判断是否解锁编队
func (this_ *Service) IsUnlock(next handler.HandleFunc) handler.HandleFunc {
	return func(c *ctx.Context) *errmsg.ErrMsg {
		return next(c)
	}
}

func (this_ *Service) Router() {
	h := this_.svc.Group(this_.IsUnlock)
	h.RegisterFunc("获取编队信息", this_.HandleGet)
	h.RegisterFunc("新增一个编队", this_.HandleNewAssemble)
	h.RegisterFunc("保存编队", this_.HandleSave)

	eventlocal.SubscribeEventLocal(this_.HandlerHeroAttrUpdate)
}

func (this_ *Service) HandleGet(c *ctx.Context, _ *servicepb.Formation_GetRequest) (*servicepb.Formation_GetResponse, *errmsg.ErrMsg) {
	data, err := this_.Get(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.Formation_GetResponse{
		Assembles:    data.Assembles,
		DefaultIndex: data.DefaultIndex,
	}, nil
}

func (this_ *Service) GetDefaultHeroes(c *ctx.Context, roleId values.RoleId) (*models.Assemble, *errmsg.ErrMsg) {
	data, err := this_.Get(c, roleId)
	if err != nil {
		return nil, err
	}
	da := data.Assembles[data.DefaultIndex]
	return da, nil
}

func (this_ *Service) heroIsChange(c *ctx.Context, heroId values.HeroId) (bool, bool, *errmsg.ErrMsg) {
	data, err := this_.Get(c, c.RoleId)
	if err != nil {
		return false, false, err
	}
	rh, ok := rule.MustGetReader(c).RowHero.GetRowHeroById(heroId)
	if !ok {
		return false, false, errmsg.NewErrHeroNotFound()
	}

	change := false
	defaultChange := false
	for i, v := range data.Assembles {
		if v.HeroOrigin_0 == rh.OriginId && v.Hero_0 != heroId {
			change = true
			v.Hero_0 = heroId
		}
		if v.HeroOrigin_1 == rh.OriginId && v.Hero_1 != heroId {
			change = true
			v.Hero_1 = heroId
		}
		if change && int64(i) == data.DefaultIndex {
			defaultChange = true
		}
	}

	return change, defaultChange, nil
}

func (this_ *Service) HeroChange(c *ctx.Context, heroId values.HeroId) *errmsg.ErrMsg {
	data, err := this_.Get(c, c.RoleId)
	if err != nil {
		return err
	}
	rh, ok := rule.MustGetReader(c).RowHero.GetRowHeroById(heroId)
	if !ok {
		return errmsg.NewErrHeroNotFound()
	}

	change := false
	defaultChange := false
	for i, v := range data.Assembles {
		if v.HeroOrigin_0 == rh.OriginId && v.Hero_0 != heroId {
			change = true
			v.Hero_0 = heroId
		}
		if v.HeroOrigin_1 == rh.OriginId && v.Hero_1 != heroId {
			change = true
			v.Hero_1 = heroId
		}
		if change && int64(i) == data.DefaultIndex {
			defaultChange = true
		}
	}
	if defaultChange {
		err = this_.onDefaultChange(c, data.Assembles, data.DefaultIndex, 2)
		if err != nil {
			return err
		}
	}
	if change {
		dao.Save(c, data)
	}
	return nil
}

func (this_ *Service) HandleNewAssemble(c *ctx.Context, _ *servicepb.Formation_NewRequest) (*servicepb.Formation_NewResponse, *errmsg.ErrMsg) {
	data, err := this_.Get(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	formationCost, ok := rule.MustGetReader(c).KeyValue.GetItem("FormationCost")
	if !ok {
		panic("not found key_value: FormationCost")
	}
	if formationCost.Count > 0 {
		err = this_.module.SubItem(c, c.RoleId, formationCost.ItemId, formationCost.Count)
		if err != nil {
			return nil, err
		}
	}

	al := len(data.Assembles)
	formationLimit, ok := rule.MustGetReader(c).KeyValue.GetInt64("FormationLimit")
	if !ok {
		formationLimit = 5
	}
	if int64(al) >= formationLimit {
		return nil, errmsg.NewErrFormationMax()
	}
	index := al + 1
	data.Assembles = append(data.Assembles, &models.Assemble{
		Hero_0: 0,
		Hero_1: 0,
		Name:   fmt.Sprintf("编队%d", index),
	})
	dao.Save(c, data)
	return &servicepb.Formation_NewResponse{
		Assembles:    data.Assembles,
		DefaultIndex: data.DefaultIndex,
	}, nil
}

func (this_ *Service) Get(c *ctx.Context, roleId string) (*daopb.Formation, *errmsg.ErrMsg) {
	data, ok, err := dao.Get(c, roleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		bhs := rule.MustGetReader(c).GetBeginHeroes()
		if len(bhs) < 1 {
			return nil, errmsg.NewErrFormationHeroNotEnough()
		}
		if len(bhs) == 1 {
			bhs = append(bhs, []int64{0, 0})
		}
		data = dao.Create(c, c.RoleId, bhs[0][0], bhs[0][1], bhs[1][0], bhs[1][1])
	}
	return data, nil
}

func (this_ *Service) onDefaultChange(c *ctx.Context, data []*models.Assemble, defaultIndex int64, changeType int64) *errmsg.ErrMsg {
	u, err := this_.module.GetUserById(c, c.UserId)
	if err != nil {
		return err
	}
	scene, ok := rule.MustGetReader(c).MapScene.GetMapSceneById(u.MapId)
	if ok && u.BattleServerId > 0 {
		a := data[defaultIndex]
		if scene.MapType == int64(models.BattleType_HangUp) ||
			scene.MapType == int64(models.BattleType_BossHall) ||
			scene.MapType == int64(models.BattleType_UnionBoss) {
			heroIds := make([]int64, 0, 2)
			if a.Hero_0 != 0 {
				heroIds = append(heroIds, a.HeroOrigin_0)
			}
			if a.Hero_1 != 0 {
				heroIds = append(heroIds, a.HeroOrigin_1)
			}
			heroes, err := this_.module.GetHeroes(c, c.RoleId, heroIds)
			if err != nil {
				return err
			}
			equips, err := this_.module.GetManyEquipBagMap(c, c.RoleId, this_.module.GetHeroesEquippedEquipId(heroes)...)
			if err != nil {
				return err
			}
			cppHeroes := make([]*models.HeroForBattle, 0, len(heroes))
			for _, h := range heroes {
				hero := &models.HeroForBattle{}
				equip := make(map[int64]int64)
				equipLightEffect := make(map[int64]int64)
				for slot, item := range h.EquipSlot {
					if item.EquipItemId == 0 {
						equip[slot] = -1
						continue
					}
					equip[slot] = item.EquipItemId
					if equipModel, ok := equips[item.EquipId]; ok && equipModel.Detail != nil {
						if equipModel.Detail.LightEffect > 0 {
							equipLightEffect[slot] = equipModel.Detail.LightEffect
						}
					}
				}
				hero.Equip = equip
				hero.Attr = h.Attrs
				hero.ConfigId = h.Id
				hero.SkillIds = h.Skill
				hero.BuffIds = h.Buff
				hero.TalentBuff = h.TalentBuff
				hero.EquipLightEffect = equipLightEffect
				// hero.Fashion = h.Fashion.Dressed
				hero.Fashion = trans.GetNewModelIdByFashionId(c, h.Fashion.Dressed)
				cppHeroes = append(cppHeroes, hero)
			}
			role, err := this_.module.GetRoleModelByRoleId(c, c.RoleId)
			if err != nil {
				return err
			}
			out := &cppbattle.CPPBattle_CPPChangeFormationResponse{}
			err = this_.svc.GetNatsClient().RequestWithOut(c, u.BattleServerId, &cppbattle.CPPBattle_CPPChangeFormationRequest{
				Heroes:     cppHeroes,
				Role:       role,
				ChangeType: changeType,
			}, out)
			if err != nil {
				return err
			}
			switch out.Result {
			case 1:
				return errmsg.NewErrFormationHeroDead()
			case 2, 3:
				return nil
			case 4:
				return errmsg.NewErrFormationSwitchHeroCD()
			default:
			}
		}
	}
	return nil
}

func (this_ *Service) HandleSave(c *ctx.Context, req *servicepb.Formation_SaveRequest) (*servicepb.Formation_SaveResponse, *errmsg.ErrMsg) {
	data, err := this_.Get(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	this_.log.Info("HandleSave Get", zap.Any("data", data))
	now := timer.Now().UnixMilli()
	formationEditCD, ok := rule.MustGetReader(c).KeyValue.GetInt64("FormationEditCD")
	if !ok {
		formationEditCD = 30000
	}
	if now-data.SetDefaultTime < formationEditCD {
		return nil, errmsg.NewErrFormationSetDefaultCD()
	}

	if len(data.Assembles) != len(req.Assembles) {
		return nil, errmsg.NewErrInvalidRequestParam()
	}

	if req.DefaultIndex < 0 || req.DefaultIndex >= int64(len(req.Assembles)) {
		return nil, errmsg.NewErrInvalidRequestParam()
	}

	change := false
	defaultChange := false
	for i, v := range data.Assembles {
		nv := req.Assembles[i]
		if int64(i) == req.DefaultIndex && nv.Hero_0 == 0 && nv.Hero_1 == 0 {
			return nil, errmsg.NewErrInvalidRequestParam()
		}

		// if v.Hero_0 != 0 && nv.Hero_0 == 0 {
		//	return nil, errmsg.NewErrInvalidRequestParam()
		// }
		// if v.Hero_1 != 0 && nv.Hero_1 == 0 {
		//	return nil, errmsg.NewErrInvalidRequestParam()
		// }
		if nv.Hero_0 != 0 && nv.Hero_1 != 0 {
			if nv.Hero_0 == nv.Hero_1 {
				return nil, errmsg.NewErrInvalidRequestParam()
			}
		}
		if nv.Hero_1 != 0 && nv.Hero_0 == 0 {
			return nil, errmsg.NewErrInvalidRequestParam()
		}
		if nv.Hero_0 != 0 {
			rh, ok := rule.MustGetReader(c).RowHero.GetRowHeroById(nv.Hero_0)
			if !ok {
				return nil, errmsg.NewErrInvalidRequestParam()
			}
			nv.HeroOrigin_0 = rh.OriginId
		}
		if nv.Hero_1 != 0 {
			rh, ok := rule.MustGetReader(c).RowHero.GetRowHeroById(nv.Hero_1)
			if !ok {
				return nil, errmsg.NewErrInvalidRequestParam()
			}
			nv.HeroOrigin_1 = rh.OriginId
		}

		if v.Hero_0 != nv.Hero_0 || v.Hero_1 != nv.Hero_1 {
			change = true
			if int64(i) == data.DefaultIndex {
				defaultChange = true
			}
		}
		if v.Name != nv.Name {
			if !sensitive.TextValid(nv.Name) {
				return nil, errmsg.NewErrSensitive()
			}
			change = true
		}
	}
	if data.DefaultIndex != req.DefaultIndex {
		change = true
		defaultChange = true
	}
	if defaultChange {
		this_.log.Info("onDefaultChange", zap.Any("req", req))
		err = this_.onDefaultChange(c, req.Assembles, req.DefaultIndex, 1)
		if err != nil {
			return nil, err
		}
	}

	if change {
		data.Assembles = req.Assembles
		data.DefaultIndex = req.DefaultIndex
		data.SetDefaultTime = now
		dao.Save(c, data)
	}
	return &servicepb.Formation_SaveResponse{}, nil
}
