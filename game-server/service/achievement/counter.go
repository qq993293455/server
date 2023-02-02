package achievement

/*import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/proto/models"
	protosvc "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/achievement/dao"
)

type CounterService struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
}

func NewCounterService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
) *CounterService {
	s := &CounterService{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
	}
	return s
}

func (svc *CounterService) Router() {
	svc.svc.RegisterFunc("添加或设置成就计数器", svc.AddOrSetCounterRequest)

	eventlocal.SubscribeEventLocal(svc.HandleCntChangeEvent)
}

func (svc *CounterService) AddOrSetCounterRequest(ctx *ctx.Context, req *protosvc.Achievement_AddOrSetCounterRequest) (*protosvc.Achievement_AddOrSetCounterResponse, *errmsg.ErrMsg) {
	return nil, svc.addOrSet(ctx, req.CountTyp, req.AchievementId, req.Cnt)
}

func (svc *CounterService) addOrSet(ctx *ctx.Context, countTyp protosvc.CountTyp, achievementId values.AchievementId, val values.Integer) *errmsg.ErrMsg {
	counter, err := dao.GetCounter(ctx)
	if err != nil {
		return err
	}
	switch countTyp {
	case protosvc.CountTyp_CTAdd:
		counter.Add(achievementId, val)
	case protosvc.CountTyp_CTSet:
		counter.Update(achievementId, val)
	}
	dao.SaveCounter(ctx, counter)
	return svc.AchievementService.HandleCounterUpdate(ctx, achievementId, counter.GetCnt(achievementId))
}

//---------------------------------------------------event------------------------------------------------------------//

func (svc *CounterService) HandleCntChangeEvent(ctx *ctx.Context, d *event.CounterCntChangeData) *errmsg.ErrMsg {
	return svc.addOrSet(ctx, d.CountTyp, d.AchievementId, d.Val)
}
*/
