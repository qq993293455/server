package achievement

import (
	"sort"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	protomodels "coin-server/common/proto/models"
	protosvc "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	event2 "coin-server/common/values/event"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/achievement/dao"
	achval "coin-server/game-server/service/achievement/values"
	"coin-server/game-server/util/trans"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
}

func NewAchievementService(
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
	module.AchievementService = s
	typAchievementRule := rule.MustGetReader(&ctx.Context{}).GetTaskTypAchieve()
	for typ, achiveMap := range typAchievementRule {
		for _, v := range achiveMap {
			s.Module.RegisterCondHandler(typ, v, event2.AllCount, s.RunningTotalHandler, nil)
		}
	}
	return s
}

// ---------------------------------------------------proto------------------------------------------------------------//

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取成就列表", svc.GetList)
	svc.svc.RegisterFunc("获取成就详细", svc.GetDetail)
	svc.svc.RegisterFunc("领取成就奖励", svc.Collect)

	/*svc.svc.RegisterFunc("作弊获得成就计数", svc.CheatAddCounter)
	svc.svc.RegisterFunc("作弊清除成就和计数", svc.CheatClear)*/
}

func (svc *Service) GetList(ctx *ctx.Context, _ *protosvc.Achievement_GetAchievementListRequest) (*protosvc.Achievement_GetAchievementListResponse, *errmsg.ErrMsg) {
	achVal, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	res := &protosvc.Achievement_GetAchievementListResponse{
		Point: achVal.GetPoint(),
		List:  make([]*protomodels.AchievementView, len(achVal.GetAll())),
	}
	idx := 0
	for _, ach := range achVal.GetAll() {
		res.List[idx] = ach.ToProtoView()
		idx++
	}
	sort.Slice(res.List, func(i, j int) bool {
		r1, r2 := res.List[i], res.List[j]
		if r1.HasUnread && !r2.HasUnread {
			return true
		}
		if !r1.HasUnread && r2.HasUnread {
			return false
		}
		if r1.HasUnread && r2.HasUnread {
			if r1.DoneTime == r2.DoneTime {
				return r1.Id < r2.Id
			}
			return r1.DoneTime > r2.DoneTime
		}
		if !r1.IsFinished && r2.IsFinished {
			return true
		}
		if r1.IsFinished && !r2.IsFinished {
			return false
		}
		return r1.Id < r2.Id
	})
	return res, nil
}

func (svc *Service) GetDetail(ctx *ctx.Context, req *protosvc.Achievement_GetAchievementDetailRequest) (*protosvc.Achievement_GetAchievementDetailResponse, *errmsg.ErrMsg) {
	achVal, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	info, err := achVal.GetDetail(req.AchievementId)
	if err != nil {
		return nil, err
	}
	return &protosvc.Achievement_GetAchievementDetailResponse{
		Info: info.ToProtoView(),
	}, nil
}

func (svc *Service) Collect(ctx *ctx.Context, req *protosvc.Achievement_CollectRequest) (*protosvc.Achievement_CollectResponse, *errmsg.ErrMsg) {
	achInfo, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	detail, err := achInfo.GetDetail(req.AchievementId)
	if err != nil {
		return nil, err
	}
	achievementRule := rulemodel.GetReader().AchievementGear()
	var (
		detailRule *rulemodel.AchievementList
		exist      bool
	)
	if _, exist = achievementRule[req.AchievementId]; !exist {
		return nil, errmsg.NewErrAchievementNotExist()
	}
	if detailRule, exist = achievementRule[req.AchievementId][req.Gear]; !exist {
		return nil, errmsg.NewErrAchievementNotExist()
	}
	if !detail.Gears.IsDone(req.Gear) {
		return nil, errmsg.NewErrAchievementNotFinished()
	}
	if detail.Gears.IsCollected(req.Gear) {
		return nil, errmsg.NewErrAchievementAlreadyCollect()
	}
	if req.Gear != detail.CollectedGear+1 {
		return nil, errmsg.NewErrAchievementAlreadyCollect()
	}
	detail.Gears.Collect(req.Gear)
	detail.CollectedGear = req.Gear
	achInfo.AddPoint(detailRule.AchievementPoint)
	reward := detailRule.Reward
	_, err = svc.Module.BagService.AddManyItem(ctx, ctx.RoleId, reward)
	if err != nil {
		return nil, err
	}
	dao.Save(ctx, achInfo)
	return &protosvc.Achievement_CollectResponse{
		Items:    trans.ItemMapToProto(reward),
		CurrGear: req.Gear,
	}, nil
}

// ---------------------------------------------------func------------------------------------------------------------//

func (svc *Service) IsDone(ctx *ctx.Context, typ values.AchievementId, gear values.Integer) bool {
	if gear <= 0 {
		return false
	}
	acInfo, err := dao.Get(ctx)
	if err != nil {
		return false
	}
	typInfo, err := acInfo.GetDetail(typ)
	if err != nil {
		return false
	}
	return gear <= typInfo.CurrGear
}

func (svc *Service) CurrGear(ctx *ctx.Context, typ values.AchievementId) (values.Integer, *errmsg.ErrMsg) {
	acInfo, err := dao.Get(ctx)
	if err != nil {
		return 0, err
	}
	typInfo, err := acInfo.GetDetail(typ)
	if err != nil {
		return 0, err
	}
	return typInfo.CurrGear, nil
}

func (svc *Service) handleDetail(ctx *ctx.Context, cnt values.Integer, detail *achval.AchievementDetail) bool {
	if cnt <= detail.CurrCnt {
		return false
	}
	if detail.IsFinished() {
		if cnt > detail.CurrCnt {
			detail.CurrCnt = cnt
			return true
		}
		return false
	}
	detail.CurrCnt = cnt
	currGear := detail.CurrGear
	achievementRule := rulemodel.GetReader().AchievementGear()
	for len(achievementRule[detail.AchievementId][currGear].TaskTypeParam) >= 3 && cnt >= achievementRule[detail.AchievementId][currGear].TaskTypeParam[2] {
		detail.Gears.Done(currGear)
		currGear++
		if _, exist := achievementRule[detail.AchievementId][currGear]; !exist {
			break
		}
		detail.CurrGear = currGear
		detail.DoneTime = timer.Now().Unix()
		ctx.PushMessage(&protosvc.Achievement_DonePush{
			Info: detail.ToProtoView(),
		})
	}
	return true
}

func (svc *Service) RunningTotalHandler(ctx *ctx.Context, d *event.TargetUpdate, _ any) *errmsg.ErrMsg {
	achInfo, err := dao.Get(ctx)
	if err != nil {
		return err
	}
	typAchievementRule := rule.MustGetReader(ctx).GetTaskTypAchieve()
	changeFlag := false
	for achiveId, itemId := range typAchievementRule[d.Typ] {
		if d.Id != itemId {
			continue
		}
		detail, err := achInfo.GetDetail(achiveId)
		if err != nil {
			return err
		}
		if hasChange := svc.handleDetail(ctx, d.Count, detail); hasChange {
			changeFlag = true
		}
	}
	if changeFlag {
		dao.Save(ctx, achInfo)
	}
	return nil
}

// ---------------------------------------------------cheat------------------------------------------------------------//

/*func (svc *Service) CheatAddCounter(ctx *ctx.Context, req *protosvc.Achievement_CheatAddCounterRequest) (*protosvc.Achievement_CheatAddCounterResponse, *errmsg.ErrMsg) {
	ctx.PublishEventLocal(&event.CounterCntChangeData{
		AchievementId: req.AchievementId,
		CountTyp:      protosvc.CountTyp_CTAdd,
		Val:           req.Cnt,
	})
	return nil, nil
}

func (svc *Service) CheatClear(ctx *ctx.Context, _ *protosvc.Achievement_CheatClearRequest) (*protosvc.Achievement_CheatClearResponse, *errmsg.ErrMsg) {
	achInfo, err := dao.Get(ctx)
	if err != nil {
		return nil, err
	}
	achInfo.CheatClear(1)
	counter, err := dao.GetCounter(ctx)
	if err != nil {
		return nil, err
	}
	counter.Update(1, 0)
	dao.Save(ctx, achInfo)
	dao.SaveCounter(ctx, counter)
	return nil, nil
}*/
