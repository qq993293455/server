package sevenDays

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
	discordDao "coin-server/game-server/service/discord/dao"
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

func NewDiscordService(
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
	module.DiscordService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("查询是否可以领取", this_.GetDiscordInfo)
	this_.svc.RegisterFunc("领取奖品", this_.ReceiveReward)
}

// msg proc ============================================================================================================================================================================================================
func (this_ *Service) GetDiscordInfo(ctx *ctx.Context, req *servicepb.Discord_DiscordInfoRequest) (*servicepb.Discord_DiscordInfoResponse, *errmsg.ErrMsg) {
	isReceived, _, err := this_.CanReceive(ctx)
	if err != nil {
		return nil, err
	}
	return &servicepb.Discord_DiscordInfoResponse{
		IsReceived: isReceived,
	}, nil
}

func (this_ *Service) ReceiveReward(ctx *ctx.Context, req *servicepb.Discord_DiscordReceiveRewardRequest) (*servicepb.Discord_DiscordReceiveRewardResponse, *errmsg.ErrMsg) {
	isReceived, discordNum, err := this_.CanReceive(ctx)
	if err != nil {
		return nil, err
	}
	if isReceived {
		return nil, errmsg.NewErrDiscordReceived()
	}

	rewards, err := GetDiscordReward(ctx)
	if err != nil {
		return nil, err
	}

	ret := &servicepb.Discord_DiscordReceiveRewardResponse{}
	for i := 0; i < len(rewards); i += 2 {
		item := &models.Item{
			ItemId: rewards[i],
			Count:  rewards[i+1],
		}

		err = this_.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
		if err != nil {
			this_.log.Error("Discord ReceiveReward AddManyItemPb err",
				zap.Any("err info", err), zap.Any("role", ctx.RoleId), zap.Any("itemid", item.ItemId), zap.Any("item cnt", item.Count))
			break
		}
		ret.Rewards = append(ret.Rewards, item)
	}

	infos := discordDao.GetDiscordInfo(ctx, ctx.RoleId)
	if infos == nil {
		infos = &dao.DiscordData{
			RoleId: ctx.RoleId,
		}
	}
	infos.RewardVersion = append(infos.RewardVersion, discordNum)
	discordDao.SaveDiscordInfo(ctx, infos)
	this_.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskReceiveDiscord, 0, 1)
	return ret, nil
}

func (this_ *Service) CanReceive(ctx *ctx.Context) (bool, int64, *errmsg.ErrMsg) {
	discordNum, err := GetDiscordNum(ctx)
	if err != nil {
		return false, 0, err
	}

	isReceived := false
	infos := discordDao.GetDiscordInfo(ctx, ctx.RoleId)
	if infos != nil {
		for _, v := range infos.RewardVersion {
			if v == discordNum {
				isReceived = true
				break
			}
		}
	}
	return isReceived, discordNum, nil
}

func GetDiscordNum(ctx *ctx.Context) (int64, *errmsg.ErrMsg) {
	discordNum, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("DiscordNum")
	if !ok {
		return 0, errmsg.NewErrDiscordConfig()
	}
	return discordNum, nil
}

func GetDiscordReward(ctx *ctx.Context) ([]int64, *errmsg.ErrMsg) {
	discordReward, ok := rule.MustGetReader(ctx).KeyValue.GetIntegerArray("DiscordReward")
	if !ok {
		return nil, errmsg.NewErrDiscordConfig()
	}

	if len(discordReward)%2 != 0 {
		return nil, errmsg.NewErrDiscordConfig()
	}

	return discordReward, nil
}
