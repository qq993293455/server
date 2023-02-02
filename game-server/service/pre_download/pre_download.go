package pre_download

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/module"
	PreDownloadDao "coin-server/game-server/service/pre_download/dao"
	"coin-server/rule"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	log *logger.Logger
}

func NewPreDownloadService(
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
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("查询是否可以领取", this_.GetPreDownloadInfo)
	this_.svc.RegisterFunc("领取奖品", this_.ReceiveReward)
}

// msg proc ============================================================================================================================================================================================================
func (this_ *Service) GetPreDownloadInfo(ctx *ctx.Context, req *servicepb.PreDownload_PreDownloadInfoRequest) (*servicepb.PreDownload_PreDownloadInfoResponse, *errmsg.ErrMsg) {
	info := PreDownloadDao.GetPreDownloadInfo(ctx, ctx.RoleId)
	if info == nil {
		info = &dao.PreDownloadData{
			RoleId:     ctx.RoleId,
			IsReceived: false,
		}
	}

	return &servicepb.PreDownload_PreDownloadInfoResponse{
		IsReceived: info.IsReceived,
	}, nil
}

func (this_ *Service) ReceiveReward(ctx *ctx.Context, req *servicepb.PreDownload_PreDownloadReceiveRewardRequest) (*servicepb.PreDownload_PreDownloadReceiveRewardResponse, *errmsg.ErrMsg) {
	info := PreDownloadDao.GetPreDownloadInfo(ctx, ctx.RoleId)
	if info == nil {
		info = &dao.PreDownloadData{
			RoleId:     ctx.RoleId,
			IsReceived: false,
		}
	}

	if info.IsReceived {
		return nil, errmsg.NewErrPreDownloadReceived()
	}

	rewards, err := GetPreDownloadReward(ctx)
	if err != nil {
		return nil, err
	}

	ret := &servicepb.PreDownload_PreDownloadReceiveRewardResponse{}
	for k, v := range rewards {
		item := &models.Item{
			ItemId: k,
			Count:  v,
		}

		err = this_.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
		if err != nil {
			this_.log.Error("PreDownload ReceiveReward AddManyItemPb err",
				zap.Any("err info", err), zap.Any("role", ctx.RoleId), zap.Any("itemid", item.ItemId), zap.Any("item cnt", item.Count))
			break
		}
		ret.Rewards = append(ret.Rewards, item)
	}

	info.IsReceived = true
	PreDownloadDao.SavePreDownloadInfo(ctx, info)

	return ret, nil
}

func GetPreDownloadReward(ctx *ctx.Context) (map[values.Integer]values.Integer, *errmsg.ErrMsg) {
	preDownload, ok := rule.MustGetReader(ctx).KeyValue.GetMapInt64Int64("DownloadRewards")
	if !ok {
		return nil, errmsg.NewErrPreDownloadConfig()
	}
	return preDownload, nil
}
