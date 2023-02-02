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
	googleQuestDao "coin-server/game-server/service/google_quest/dao"
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

func NewGoogleQuestService(
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
	this_.svc.RegisterFunc("查询是否可以领取", this_.GetGoogleQuestInfo)
	this_.svc.RegisterFunc("领取奖品", this_.ReceiveReward)
}

// msg proc ============================================================================================================================================================================================================
func (this_ *Service) GetGoogleQuestInfo(ctx *ctx.Context, req *servicepb.GoogleQuest_GoogleQuestInfoRequest) (*servicepb.GoogleQuest_GoogleQuestInfoResponse, *errmsg.ErrMsg) {
	isReceived, _, err := this_.CanReceive(ctx)
	if err != nil {
		return nil, err
	}
	return &servicepb.GoogleQuest_GoogleQuestInfoResponse{
		IsReceived: isReceived,
	}, nil
}

func (this_ *Service) ReceiveReward(ctx *ctx.Context, req *servicepb.GoogleQuest_GoogleQuestReceiveRewardRequest) (*servicepb.GoogleQuest_GoogleQuestReceiveRewardResponse, *errmsg.ErrMsg) {
	isReceived, googleQuestNum, err := this_.CanReceive(ctx)
	if err != nil {
		return nil, err
	}
	if isReceived {
		return nil, errmsg.NewErrGoogleQuestReceived()
	}

	rewards, err := GetGoogleQuestReward(ctx)
	if err != nil {
		return nil, err
	}

	ret := &servicepb.GoogleQuest_GoogleQuestReceiveRewardResponse{}
	for i := 0; i < len(rewards); i += 2 {
		item := &models.Item{
			ItemId: rewards[i],
			Count:  rewards[i+1],
		}

		err = this_.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
		if err != nil {
			this_.log.Error("GoogleQuest ReceiveReward AddManyItemPb err",
				zap.Any("err info", err), zap.Any("role", ctx.RoleId), zap.Any("itemid", item.ItemId), zap.Any("item cnt", item.Count))
			break
		}
		ret.Rewards = append(ret.Rewards, item)
	}

	infos := googleQuestDao.GetGoogleQuestInfo(ctx, ctx.RoleId)
	if infos == nil {
		infos = &dao.GoogleQuestData{
			RoleId: ctx.RoleId,
		}
	}
	infos.VersionIds = append(infos.VersionIds, googleQuestNum)
	googleQuestDao.SaveGoogleQuestInfo(ctx, infos)

	return ret, nil
}

func (this_ *Service) CanReceive(ctx *ctx.Context) (bool, int64, *errmsg.ErrMsg) {
	googleQuestNum, err := GetGoogleQuestNum(ctx)
	if err != nil {
		return false, 0, err
	}

	isReceived := false
	infos := googleQuestDao.GetGoogleQuestInfo(ctx, ctx.RoleId)
	if infos != nil {
		for _, v := range infos.VersionIds {
			if v == googleQuestNum {
				isReceived = true
				break
			}
		}
	}
	return isReceived, googleQuestNum, nil
}

func GetGoogleQuestNum(ctx *ctx.Context) (int64, *errmsg.ErrMsg) {
	googleQuestNum, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("GoogleQuestionNum")
	if !ok {
		return 0, errmsg.NewErrGoogleQuestConfig()
	}
	return googleQuestNum, nil
}

func GetGoogleQuestReward(ctx *ctx.Context) ([]int64, *errmsg.ErrMsg) {
	googleQuestReward, ok := rule.MustGetReader(ctx).KeyValue.GetIntegerArray("GoogleQuestionReward")
	if !ok {
		return nil, errmsg.NewErrGoogleQuestConfig()
	}

	if len(googleQuestReward)%2 != 0 {
		return nil, errmsg.NewErrGoogleQuestConfig()
	}

	return googleQuestReward, nil
}
