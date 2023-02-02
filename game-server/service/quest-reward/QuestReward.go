package questReward

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
	QuestRewardDao "coin-server/game-server/service/quest-reward/dao"
	"coin-server/rule"
)

const (
	mailId = 100011
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	log *logger.Logger
}

func NewQuestRewardService(
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
	module.QuestRewardService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("查询是否可以领取", this_.GetQuestRewardInfo)
	this_.svc.RegisterFunc("领取奖品", this_.ReceiveReward)
}

// msg proc ============================================================================================================================================================================================================
func (this_ *Service) GetQuestRewardInfo(ctx *ctx.Context, req *servicepb.QuestReward_QuestRewardInfoRequest) (*servicepb.QuestReward_QuestRewardInfoResponse, *errmsg.ErrMsg) {
	info := QuestRewardDao.GetQuestRewardInfo(ctx, ctx.RoleId)
	if info == nil {
		info = &dao.QuestRewardData{
			RoleId:     ctx.RoleId,
			IsReceived: false,
		}
	}

	return &servicepb.QuestReward_QuestRewardInfoResponse{
		IsReceived: info.IsReceived,
	}, nil
}

func (this_ *Service) ReceiveReward(ctx *ctx.Context, req *servicepb.QuestReward_QuestRewardReceiveRewardRequest) (*servicepb.QuestReward_QuestRewardReceiveRewardResponse, *errmsg.ErrMsg) {
	info := QuestRewardDao.GetQuestRewardInfo(ctx, ctx.RoleId)
	if info == nil {
		info = &dao.QuestRewardData{
			RoleId:     ctx.RoleId,
			IsReceived: false,
		}
	}

	if info.IsReceived {
		return nil, errmsg.NewErrQuestRewardReceived()
	}

	rewards, err := GetQuestRewardReward(ctx)
	if err != nil {
		return nil, err
	}

	ret := &servicepb.QuestReward_QuestRewardReceiveRewardResponse{}
	var items []*models.Item
	for i := 0; i < len(rewards); i += 2 {
		item := &models.Item{
			ItemId: rewards[i],
			Count:  rewards[i+1],
		}

		// err = this_.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
		// if err != nil {
		// 	this_.log.Error("QuestReward ReceiveReward AddManyItemPb err",
		// 		zap.Any("err info", err), zap.Any("role", ctx.RoleId), zap.Any("itemid", item.ItemId), zap.Any("item cnt", item.Count))
		// 	break
		// }
		items = append(items, item)
		ret.Rewards = append(ret.Rewards, item)
	}

	this_.SendItem(ctx, items)
	info.IsReceived = true
	QuestRewardDao.SaveQuestRewardInfo(ctx, info)

	return ret, nil
}

func (this_ *Service) SendItem(ctx *ctx.Context, items []*models.Item) (bool, *errmsg.ErrMsg) {
	if err := this_.MailService.Add(ctx, ctx.RoleId, &models.Mail{
		Type:       models.MailType_MailTypeSystem,
		TextId:     mailId,
		Args:       []string{},
		Attachment: items,
	}); err != nil {
		return false, err
	}
	return true, nil
}

func GetQuestRewardReward(ctx *ctx.Context) ([]int64, *errmsg.ErrMsg) {
	questReward, ok := rule.MustGetReader(ctx).KeyValue.GetIntegerArray("QuestReward")
	if !ok {
		return nil, errmsg.NewErrQuestRewardConfig()
	}

	if len(questReward)%2 != 0 {
		return nil, errmsg.NewErrQuestRewardConfig()
	}

	return questReward, nil
}
