package shop

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/proto/models"
	protosvc "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/roguelike/dao"
	"coin-server/rule"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
}

func NewRoguelikeService(
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
	return s
}

// ---------------------------------------------------proto------------------------------------------------------------//

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取肉鸽完成副本", svc.GetRLDone)

	svc.svc.RegisterFunc("作弊解锁所有肉鸽副本", svc.CheatAllRlDoneRequest)

	eventlocal.SubscribeEventLocal(svc.HandleRlFinish)
}

func (svc *Service) GetRLDone(c *ctx.Context, _ *protosvc.Roguelike_GetRLDoneRequest) (*protosvc.Roguelike_GetRLDoneResponse, *errmsg.ErrMsg) {
	data, err := dao.GetRlDone(c)
	if err != nil {
		return nil, err
	}
	return &protosvc.Roguelike_GetRLDoneResponse{
		DoneMap: data.DoneMap,
	}, nil
}

func (svc *Service) HandleRlFinish(c *ctx.Context, evt *event.RLFinishEvent) *errmsg.ErrMsg {
	if evt.IsSucc {
		data, err := dao.GetRlDone(c)
		if err != nil {
			return err
		}
		cfg, has := rule.MustGetReader(c).RoguelikeDungeon.GetRoguelikeDungeonById(values.Integer(evt.RoguelikeId))
		if !has || cfg == nil {
			return nil
		}
		lvlId := cfg.DungeonLv[0] * 1000
		if !data.DoneMap[lvlId] {
			data.DoneMap[lvlId] = true
			dao.SaveRlDone(c, data)
			c.PushMessageToRole(c.RoleId, &protosvc.Roguelike_RLDoneChangePush{
				DoneMap: data.DoneMap,
			})
		}
	}
	return nil
}

// ---------------------------------------------------cheat------------------------------------------------------------//

func (svc *Service) CheatAllRlDoneRequest(c *ctx.Context, _ *protosvc.Roguelike_CheatAllRlDoneRequest) (*protosvc.Roguelike_CheatAllRlDoneResponse, *errmsg.ErrMsg) {
	data, err := dao.GetRlDone(c)
	if err != nil {
		return nil, err
	}
	ruleList := rule.MustGetReader(c).RoguelikeDungeon.List()
	for _, rl := range ruleList {
		data.DoneMap[rl.Id] = true
	}
	dao.SaveRlDone(c, data)
	c.PushMessageToRole(c.RoleId, &protosvc.Roguelike_RLDoneChangePush{
		DoneMap: data.DoneMap,
	})
	return &protosvc.Roguelike_CheatAllRlDoneResponse{}, nil
}
