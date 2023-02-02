package tower

/*
import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/tower/dao"
	"fmt"
)

type TowerDefaultSerivce struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	log *logger.Logger
}

func NewTowerService(
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
	module.TowerService = s
	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("获取爬塔数据", s.GetTowerInfoRequest)
	s.svc.RegisterFunc("获取下一层数据", s.GetNextChallengeDataRequest)
}

// ---------------------------------------------------proto------------------------------------------------------------//

func (s *Service) GetTowerInfoRequest(ctx *ctx.Context, _ *servicepb.Tower_GetTowerInfoRequest) (*servicepb.Tower_GetTowerInfoResponse, *errmsg.ErrMsg) {
	dao_tower, err := dao.GetTowerData(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	fmt.Println(dao_tower)
	// var data *servicepb.Tower_GetTowerInfoResponse = new(servicepb.Tower_GetTowerInfoResponse)
	// data.DataMap = dao_tower.Data
	return &servicepb.Tower_GetTowerInfoResponse{
		DataMap: dao_tower.Data,
	}, nil
}

func (s *Service) GetNextChallengeDataRequest(ctx *ctx.Context, _ *servicepb.Tower_GetNextChallengeDataRequest) (*servicepb.Tower_GetNextChallengeDataResponse, *errmsg.ErrMsg) {
	dao_tower, err := dao.GetTowerData(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	fmt.Println(dao_tower)
	return &servicepb.Tower_GetNextChallengeDataResponse{}, nil
}
*/
