package MoonthlyCard

import (
	"fmt"
	"strconv"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	mDao "coin-server/game-server/service/moonthly_card/dao"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"

	"go.uber.org/zap"
)

const (
	dailyMailId     = 100014
	purchaseMailId  = 100015
	subscribeMailId = 100016
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewMoonthlyCardService(
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
	module.MoothlyCardService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("获取月卡情况", this_.MoonthlyCardInfo)
	this_.svc.RegisterFunc("购买月卡", this_.MoonthlyCardBuy)
	this_.svc.RegisterFunc("领取月卡", this_.MoonthlyCardReceive)
	this_.svc.RegisterFunc("获取配置", this_.MoonthlyCardCnf)

	this_.svc.RegisterFunc("gm修改月卡开始时间", this_.CheatMoonthlyCardStartTime)

	eventlocal.SubscribeEventLocal(this_.HandleTaskChange)
	eventlocal.SubscribeEventLocal(this_.HandleRoleLoginEvent)
	eventlocal.SubscribeEventLocal(this_.HandleDailyPaySuccess)
}

// msg proc ============================================================================================================================================================================================================
func (this_ *Service) CheatMoonthlyCardStartTime(ctx *ctx.Context, req *servicepb.MoonthlyCard_CheatMoonthlyCardStartTimeRequest) (*servicepb.MoonthlyCard_CheatMoonthlyCardStartTimeResponse, *errmsg.ErrMsg) {
	data := this_.Update(ctx)
	aData, ok := data.Infos[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrMoonthlyCardNoStart()
	}
	cardInfo, ok := aData.ActivationCards[req.Id]
	if !ok {
		return nil, errmsg.NewErrMoonthlyCardNoCard()
	}

	cardInfo.StartTime = req.StartTime

	mDao.SaveSevenDaysInfo(ctx, data)

	return &servicepb.MoonthlyCard_CheatMoonthlyCardStartTimeResponse{
		Info: cardInfo,
	}, nil
}

func (this_ *Service) MoonthlyCardInfo(ctx *ctx.Context, req *servicepb.MoonthlyCard_MoonthlyCardInfoRequest) (*servicepb.MoonthlyCard_MoonthlyCardInfoResponse, *errmsg.ErrMsg) {
	data := this_.Update(ctx)
	return &servicepb.MoonthlyCard_MoonthlyCardInfoResponse{
		Infos: data.Infos,
	}, nil
}

func (this_ *Service) monthlyCardBuyByPayId(ctx *ctx.Context, payId, expireTime values.Integer) *errmsg.ErrMsg {
	reader := rule.MustGetReader(ctx)
	list := reader.ActivityMonthlycard.List()
	var cardCfgId values.Integer = 0
	isFind := false
	isSpecial := false
	number := values.Integer(0)
	for _, cfg := range list {
		for idx, pId := range cfg.Price {
			if pId == payId {
				cardCfgId = cfg.Id
				isFind = true
				number = cfg.PurchaseOptions[idx]
				break
			}
		}
		if isFind {
			break
		}
		for idx, pId := range cfg.FirstPrice {
			if pId == payId {
				cardCfgId = cfg.Id
				isFind = true
				number = cfg.PurchaseOptions[idx]
				break
			}
		}
		if isFind {
			break
		}
		for idx, pId := range cfg.Price2 {
			if pId == payId {
				cardCfgId = cfg.Id
				isFind = true
				isSpecial = true
				number = cfg.PurchaseOptions[idx]
				break
			}
		}
		if isFind {
			break
		}
	}
	if cardCfgId == 0 {
		return errmsg.NewErrMoonthlyCardNoCard()
	}
	cardCfg, ok := reader.ActivityMonthlycard.GetActivityMonthlycardById(cardCfgId)
	if !ok {
		return errmsg.NewErrMoonthlyCardNoCard()
	}
	_, err := this_.MoonthlyCardBuy(ctx, &servicepb.MoonthlyCard_MoonthlyCardBuyRequest{
		Id:         cardCfg.Id,
		Type:       models.MoonthlyCardType(cardCfg.ActivateType),
		ActivityId: cardCfg.ActivityId,
		Number:     number,
		ExpireTime: expireTime,
		IsSpecial:  isSpecial,
	})
	return err
}

func (this_ *Service) MoonthlyCardBuy(ctx *ctx.Context, req *servicepb.MoonthlyCard_MoonthlyCardBuyRequest) (*servicepb.MoonthlyCard_MoonthlyCardBuyResponse, *errmsg.ErrMsg) {
	data := this_.Update(ctx)

	aData, ok := data.Infos[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrMoonthlyCardNoStart()
	}

	isFind := false
	index := 0
	for idx, id := range aData.CanActivationCards {
		if id == req.Id {
			index = idx
			isFind = true
			break
		}
	}

	if !isFind {
		return nil, errmsg.NewErrMoonthlyCardNotPurchase()
	}

	mCnf, ok := rule.MustGetReader(ctx).ActivityMonthlycard.GetActivityMonthlycardById(req.Id)
	if !ok {
		ctx.Error("ActivityMonthlycard.GetActivityMonthlycardById not find", zap.Any("msg", req))
		return nil, errmsg.NewErrMoonthlyCardConfig()
	}

	if mCnf.ActivateType != int64(req.Type) {
		return nil, errmsg.NewErrMoonthlyCardPurchaseType()
	}
	if req.Type == models.MoonthlyCardType_Subscribe && !req.IsSpecial && req.ExpireTime == 0 {
		return nil, errmsg.NewErrMoonthlyCardPurchaseType()
	}

	isFind = false
	for _, cnt := range mCnf.PurchaseOptions {
		if cnt == req.Number {
			isFind = true
			break
		}
	}

	if !isFind {
		return nil, errmsg.NewErrMoonthlyCardPurchaseNumber()
	}

	isNewBuy := false
	_, ok = aData.ActivationCards[req.Id]
	if !ok {
		aData.ActivationCards[req.Id] = &models.MoonthlyCardInfo{
			Id:             mCnf.Id,
			StartTime:      timer.Now().Unix(),
			ReceiveTimes:   0,
			RemainingTimes: 0,
			CardType:       req.Type,
			CanReceive:     false,
		}
		isNewBuy = true
	}
	cardInfo := aData.ActivationCards[req.Id]
	if req.Type == models.MoonthlyCardType_Subscribe && !req.IsSpecial {
		todayBegin := this_.Module.RefreshService.GetCurrDayFreshTime(ctx)
		cardInfo.StartTime = timer.Now().Unix()
		cardInfo.RemainingTimes = (req.ExpireTime-todayBegin.Unix())/86400 + 1
		cardInfo.ReceiveTimes = 0
		if !isNewBuy && !cardInfo.CanReceive {
			cardInfo.ReceiveTimes++
		}
		cardInfo.IsSub = true
	} else {
		cardInfo.RemainingTimes += mCnf.Duration * req.Number
	}

	for i := 0; i < int(req.Number); i++ {
		if req.Type == models.MoonthlyCardType_Subscribe {
			mailId := mCnf.Mail3
			if req.IsSpecial {
				mailId = mCnf.Mail2
			}
			this_.SendRewardMail(ctx, mailId, mCnf.RenewalReward, 1, nil)
		} else {
			this_.SendRewardMail(ctx, mCnf.Mail2, mCnf.PurchaseReward, 1, nil)
		}
	}

	if mCnf.Timeliness == 2 {
		aData.InvalidCards = append(aData.InvalidCards, req.Id)
		aData.CanActivationCards = append(aData.CanActivationCards[:index], aData.CanActivationCards[index+1:]...)
	}

	_, ok = aData.PurchaseTimes[req.Id]
	if !ok {
		aData.PurchaseTimes[req.Id] = &models.MoonthlyCardPurchaseTimes{}
	}
	if aData.PurchaseTimes[req.Id].DetailTimes == nil {
		aData.PurchaseTimes[req.Id].DetailTimes = make(map[int64]int64)
	}

	aData.PurchaseTimes[req.Id].DetailTimes[req.Number] += req.Number
	this_.RefreshMoothlyCard(ctx, data)
	mDao.SaveSevenDaysInfo(ctx, data)
	ctx.PushMessage(&servicepb.MoonthlyCard_MonthlyCardBuyPush{
		Id:     req.Id,
		Type:   req.Type,
		Number: req.Number,
		Info:   aData,
	})
	return &servicepb.MoonthlyCard_MoonthlyCardBuyResponse{}, nil
}

func (this_ *Service) MoonthlyCardReceive(ctx *ctx.Context, req *servicepb.MoonthlyCard_MoonthlyCardReceiveRequest) (*servicepb.MoonthlyCard_MoonthlyCardReceiveResponse, *errmsg.ErrMsg) {
	data := this_.Update(ctx)
	aData, ok := data.Infos[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrMoonthlyCardNoStart()
	}
	cardInfo := aData.ActivationCards[req.Id]
	if cardInfo == nil || !cardInfo.CanReceive {
		return nil, errmsg.NewErrMoonthlyCardNoReward()
	}

	mCnf, ok := rule.MustGetReader(ctx).ActivityMonthlycard.GetActivityMonthlycardById(req.Id)
	if !ok {
		ctx.Error("ActivityMonthlycard.GetActivityMonthlycardById not find", zap.Any("msg", req))
		return nil, errmsg.NewErrMoonthlyCardConfig()
	}

	var rewardList []*models.Item
	for i := 0; i < len(mCnf.DailyReward); i += 2 {
		item := &models.Item{
			ItemId: mCnf.DailyReward[i],
			Count:  mCnf.DailyReward[i+1],
		}

		err := this_.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
		if err != nil {
			ctx.Error("MoonthlyCardReceive AddManyItemPb err",
				zap.Any("err info", err), zap.Any("role", ctx.RoleId), zap.Any("itemid", item.ItemId), zap.Any("item cnt", item.Count))
			return nil, err
		}
		rewardList = append(rewardList, item)
	}

	cardInfo.CanReceive = false
	cardInfo.ReceiveTimes++

	mDao.SaveSevenDaysInfo(ctx, data)
	return &servicepb.MoonthlyCard_MoonthlyCardReceiveResponse{
		Info:  cardInfo,
		Items: rewardList,
	}, nil
}

func (this_ *Service) MoonthlyCardCnf(ctx *ctx.Context, req *servicepb.MoonthlyCard_MoonthlyCardCnfRequest) (*servicepb.MoonthlyCard_MoonthlyCardCnfResponse, *errmsg.ErrMsg) {
	ret := &servicepb.MoonthlyCard_MoonthlyCardCnfResponse{}
	mConfigs := rule.MustGetReader(ctx).ActivityMonthlycard.List()
	for _, mCnf := range mConfigs {
		ret.Cnf = append(ret.Cnf, &models.MoonthlyCardCnf{
			Id:              mCnf.Id,
			Language1:       mCnf.Language1,
			ActivityId:      mCnf.ActivityId,
			UnlockCondition: mCnf.UnlockCondition,
			Banner:          mCnf.Banner,
			Timeliness:      mCnf.Timeliness,
			Duration:        mCnf.Duration,
			ActivateType:    mCnf.ActivateType,
			PurchaseOptions: mCnf.PurchaseOptions,
			Price:           mCnf.Price,
			FirstPrice:      mCnf.FirstPrice,
			ShowPrice:       mCnf.ShowPrice,
			AveragePrice:    mCnf.AveragePrice,
			Discount:        mCnf.Discount,
			DailyReward:     mCnf.DailyReward,
			PurchaseReward:  mCnf.PurchaseReward,
			RenewalReward:   mCnf.RenewalReward,
			Value:           mCnf.Value,
		})
	}
	return ret, nil
}

func (this_ *Service) HandleTaskChange(ctx *ctx.Context, d *event.TargetUpdate) *errmsg.ErrMsg {
	data := mDao.GetXDayGoalInfo(ctx, ctx.RoleId)
	this_.UpdateActivity(ctx, data)

	msg := &servicepb.MoonthlyCard_MoonthlyCardProgressPush{
		Progress: make(map[int64]*models.MoonthlyCardProgress),
	}

	haveChange := false
	mConfigs := rule.MustGetReader(ctx).ActivityMonthlycard.List()
	for _, mCnf := range mConfigs {
		if !CheckMoonthlyTable(ctx, &mCnf) {
			continue
		}

		aData, ok := data.Infos[mCnf.ActivityId]
		if !ok {
			continue
		}

		if aData.IsInvalid {
			continue
		}

		if !this_.CheckCardNeedCondition(ctx, aData, mCnf.Id) {
			continue
		}

		if d.Typ == models.TaskType(mCnf.UnlockCondition[0]) && d.Id == mCnf.UnlockCondition[1] {
			if d.Typ == models.TaskType_TaskLogin {
				DefaultRefreshTime, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("DefaultRefreshTime")
				if !ok {
					ctx.Error("HandleTaskChange DefaultRefreshTime error")
					panic("HandleTaskChange DefaultRefreshTime error")
				}

				tBeginRefreshTime := timer.BeginOfDay(timer.Now()).Unix() + DefaultRefreshTime

				_, ok1 := aData.LastLoginTime[mCnf.Id]
				if ok1 {
					if aData.LastLoginTime[mCnf.Id] > tBeginRefreshTime {
						continue
					}
				}
				aData.LastLoginTime[mCnf.Id] = tBeginRefreshTime + 86400
			}

			_, ok = msg.Progress[mCnf.ActivityId]
			if !ok {
				msg.Progress[mCnf.ActivityId] = &models.MoonthlyCardProgress{
					Progress: make(map[int64]int64),
				}
			}
			progress := msg.Progress[aData.ActivityId]

			cfg, ok := rule.MustGetReader(nil).TaskType.GetTaskTypeById(int64(d.Typ))
			if !ok {
				panic(errmsg.NewInternalErr("task_type not found: " + strconv.Itoa(int(d.Typ))))
			}

			if cfg.IsAccumulate {
				aData.Progress[mCnf.Id] = d.Count
			} else {
				aData.Progress[mCnf.Id] += d.Incr
			}

			progress.Progress[mCnf.Id] = aData.Progress[mCnf.Id]
			if aData.Progress[mCnf.Id] >= mCnf.UnlockCondition[2] {
				aData.Progress[mCnf.Id] = mCnf.UnlockCondition[2]
				aData.CanActivationCards = append(aData.CanActivationCards, mCnf.Id)
				progress.CanActivationCards = append(progress.CanActivationCards, mCnf.Id)
			}
			haveChange = true
		}
	}
	if haveChange {
		mDao.SaveSevenDaysInfo(ctx, data)
		ctx.PushMessage(msg)
	}
	return nil
}

func (this_ *Service) HandleRoleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	//this_.Update(ctx)
	return nil
}

func (this_ *Service) HandleDailyPaySuccess(ctx *ctx.Context, d *event.PaySuccess) *errmsg.ErrMsg {
	v, ok := rule.MustGetReader(ctx).Charge.GetChargeById(d.PcId)
	if !ok {
		return errmsg.NewErrActivityGiftEmpty()
	}
	if (v.FunctionType == 1 && v.TargetId == enum.MonthlyCard) || (v.FunctionType == 2 && v.TargetId == values.Integer(models.SystemType_SystemMonthlyCard)) {
		if err := this_.monthlyCardBuyByPayId(ctx, d.PcId, d.ExpireTime); err != nil {
			return err
		}
	}
	return nil
}

type CheckActivityInfo struct {
	isOver  bool
	isStart bool
}

func (this_ *Service) Update(ctx *ctx.Context) *dao.MoonthlyCardData {
	data := mDao.GetXDayGoalInfo(ctx, ctx.RoleId)

	uChange := this_.UpdateActivity(ctx, data)
	rCHange := this_.RefreshMoothlyCard(ctx, data)
	if uChange || rCHange {
		mDao.SaveSevenDaysInfo(ctx, data)
	}
	return data
}

func (this_ *Service) UpdateActivity(ctx *ctx.Context, data *dao.MoonthlyCardData) bool {
	uChange := false
	timeNow := timer.Now().UTC().Unix()
	mConfigs := rule.MustGetReader(ctx).ActivityMonthlycard.List()
	activityMap := make(map[int64]*CheckActivityInfo)
	for _, mCnf := range mConfigs {
		if !CheckMoonthlyTable(ctx, &mCnf) {
			continue
		}

		aData, aDataOk := data.Infos[mCnf.ActivityId]
		_, ok := activityMap[mCnf.ActivityId]
		if !ok {
			checkData := &CheckActivityInfo{}
			aCnf, ok := rule.MustGetReader(ctx).Activity.GetActivityById(mCnf.ActivityId)
			if !ok {
				panic(fmt.Sprintf("ActivityMonthlycard id %d activity id %d is not in Activity", mCnf.Id, mCnf.ActivityId))
			}

			if aCnf.TimeType == 1 {
				checkData.isStart = true
			}

			if aCnf.TimeType == 2 {
				startTime, err := GetTime(aCnf.ActivityOpenTime)
				if err != nil {
					panic(fmt.Sprintf("GetActivityRewardById ActivityId %d GetTime ActivityOpenTime error %s", mCnf.ActivityId, err))
				}
				if timeNow >= startTime {
					checkData.isStart = true
				}
			}

			if checkData.isStart {
				if aDataOk {
					if aCnf.TimeType == 1 {
						EndTime, err := strconv.ParseInt(aCnf.DurationTime, 10, 64)
						if err != nil {
							panic(fmt.Sprintf("GetActivityRewardById ActivityId %d GetTime DurationTime error %s", mCnf.ActivityId, err))
						}
						if EndTime != -1 {
							if timeNow > aData.StartTime+EndTime {
								checkData.isOver = true
							}
						}
					}
				}

				if aCnf.TimeType == 2 {
					EndTime, err := GetTime(aCnf.DurationTime)
					if err != nil {
						panic(fmt.Sprintf("GetActivityRewardById ActivityId %d GetTime DurationTime error %s", mCnf.ActivityId, err))
					}
					if timeNow > EndTime {
						checkData.isOver = true
					}
				}
			}

			activityMap[mCnf.ActivityId] = checkData
		}

		activityCheck := activityMap[mCnf.ActivityId]

		if activityCheck.isOver && aDataOk {
			this_.ProcOverTimeActivity(ctx, aData)
			uChange = true
		}

		if !activityCheck.isStart || activityCheck.isOver {
			continue
		}

		if !aDataOk {
			uChange = true
			aData = &models.MoonthlyCardActivityInfo{
				ActivityId:      mCnf.ActivityId,
				StartTime:       timeNow,
				Progress:        make(map[int64]int64),
				ActivationCards: make(map[int64]*models.MoonthlyCardInfo),
			}
			data.Infos[mCnf.ActivityId] = aData
		}

		aData = data.Infos[mCnf.ActivityId]

		if !this_.CheckCardNeedCondition(ctx, aData, mCnf.Id) {
			continue
		}

		if len(mCnf.UnlockCondition) != 0 {
			isComplete := false
			cfg, ok := rule.MustGetReader(nil).TaskType.GetTaskTypeById(mCnf.UnlockCondition[0])
			if !ok {
				panic(errmsg.NewInternalErr("task_type not found: " + strconv.Itoa(int(mCnf.UnlockCondition[0]))))
			}

			if cfg.IsAccumulate {
				counter, err := this_.TaskService.GetCounterByType(ctx, models.TaskType(mCnf.UnlockCondition[0]))
				if err != nil {
					panic(err)
				}
				count := counter[mCnf.UnlockCondition[1]]
				isComplete = count >= mCnf.UnlockCondition[2]
				aData.Progress[mCnf.Id] = count
				uChange = true
			}

			if isComplete {
				aData.Progress[mCnf.Id] = mCnf.UnlockCondition[2]
				aData.CanActivationCards = append(aData.CanActivationCards, mCnf.Id)
			}
		} else {
			aData.CanActivationCards = append(aData.CanActivationCards, mCnf.Id)
		}

	}
	return uChange
}

func (this_ *Service) RefreshMoothlyCard(ctx *ctx.Context, data *dao.MoonthlyCardData) bool {
	haveChange := false
	for _, info := range data.Infos {
		for key, cardInfo := range info.ActivationCards {
			days := GetDays(ctx, cardInfo.StartTime)
			if days > cardInfo.RemainingTimes {
				this_.ProcOverTimeMoonthlyCard(ctx, cardInfo)
				delete(info.ActivationCards, key)
				haveChange = true
				continue
			}

			if days <= cardInfo.ReceiveTimes {
				continue
			}

			haveChange = true
			cardInfo.CanReceive = true
			sendDays := days - cardInfo.ReceiveTimes - 1
			if sendDays <= 0 {
				continue
			}

			mCnf, ok := GetMoonthlyCardCnf(ctx, cardInfo.Id)
			if !ok {
				ctx.Error("ActivityMonthlycard GetActivityMonthlycardById error", zap.Any("card id", cardInfo.Id))
				continue
			}

			for i := int64(1); i <= sendDays; i++ {
				ret := this_.SendRewardMail(ctx, mCnf.Mail, mCnf.DailyReward, 1, nil)
				if ret {
					cardInfo.ReceiveTimes++
				}
			}
		}
	}
	return haveChange
}

func (this_ *Service) ProcOverTimeMoonthlyCard(ctx *ctx.Context, cardInfo *models.MoonthlyCardInfo) {
	sendDays := cardInfo.RemainingTimes - cardInfo.ReceiveTimes
	if sendDays <= 0 {
		return
	}

	mCnf, ok := GetMoonthlyCardCnf(ctx, cardInfo.Id)
	if !ok {
		return
	}
	for i := int64(1); i <= sendDays; i++ {
		ret := this_.SendRewardMail(ctx, mCnf.Mail, mCnf.DailyReward, 1, nil)
		if ret {
			cardInfo.ReceiveTimes++
		}
	}
}

func (this_ *Service) SendRewardMail(ctx *ctx.Context, mailId int64, reward []int64, times int64, Args []string) bool {
	if len(reward) == 0 {
		return true
	}
	var rewardList []*models.Item
	for i := 0; i < len(reward); i += 2 {
		itemId := reward[i]
		Count := reward[i+1]
		rewardList = append(rewardList, &models.Item{
			ItemId: itemId,
			Count:  Count * times,
		})
	}

	if err := this_.MailService.Add(ctx, ctx.RoleId, &models.Mail{
		Type:       models.MailType_MailTypeSystem,
		TextId:     mailId,
		Attachment: rewardList,
		Args:       Args,
	}); err != nil {
		ctx.Error("send maill error", zap.Any("msg", err), zap.Any("reward", reward), zap.Any("roleid", ctx.RoleId))
		return false
	}
	return true
}

func (this_ *Service) CheckCardNeedCondition(ctx *ctx.Context, aData *models.MoonthlyCardActivityInfo, cardId int64) bool {
	for _, id := range aData.InvalidCards {
		if id == cardId {
			return false
		}
	}

	for _, id := range aData.CanActivationCards {
		if id == cardId {
			return false
		}
	}

	for _, cardInfo := range aData.ActivationCards {
		if cardInfo.Id == cardId {
			return false
		}
	}

	return true
}

func (this_ *Service) ProcOverTimeActivity(ctx *ctx.Context, aData *models.MoonthlyCardActivityInfo) {
	aData.IsInvalid = true
}

func GetMoonthlyCardCnf(ctx *ctx.Context, id int64) (*rule_model.ActivityMonthlycard, bool) {
	mCnf, ok := rule.MustGetReader(ctx).ActivityMonthlycard.GetActivityMonthlycardById(id)
	if !ok {
		ctx.Error("GetMoonthlyCardCnf error", zap.Any("card id", id))
		return nil, false
	}
	if !CheckMoonthlyTable(ctx, mCnf) {
		return nil, false
	}
	return mCnf, true
}

func CheckMoonthlyTable(ctx *ctx.Context, mCnf *rule_model.ActivityMonthlycard) bool {
	if len(mCnf.UnlockCondition) != 3 && len(mCnf.UnlockCondition) != 0 {
		ctx.Error("MoonthlyCard config error", zap.Any("id", mCnf.Id), zap.Any("UnlockCondition len", len(mCnf.UnlockCondition)))
		return false
	}
	if len(mCnf.DailyReward)%2 != 0 {
		ctx.Error("MoonthlyCard config error", zap.Any("id", mCnf.Id), zap.Any("DailyReward len", len(mCnf.DailyReward)))
		return false
	}
	if len(mCnf.PurchaseReward)%2 != 0 {
		ctx.Error("MoonthlyCard config error", zap.Any("id", mCnf.Id), zap.Any("PurchaseReward len", len(mCnf.PurchaseReward)))
		return false
	}
	if len(mCnf.RenewalReward)%2 != 0 {
		ctx.Error("MoonthlyCard config error", zap.Any("id", mCnf.Id), zap.Any("RenewalReward len", len(mCnf.RenewalReward)))
		return false
	}
	return true
}

func GetDays(ctx *ctx.Context, startTime int64) int64 {
	DefaultRefreshTime, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("DefaultRefreshTime")
	if !ok {
		ctx.Error("DefaultRefreshTime error")
		panic("DefaultRefreshTime error")
	}

	tBeginRefreshTime := timer.BeginOfDay(timer.Now()).Unix() + DefaultRefreshTime
	if timer.Now().Unix() > tBeginRefreshTime {
		tBeginRefreshTime += 86400
	}
	return (tBeginRefreshTime-startTime)/86400 + 1
}

func GetTime(value string) (int64, error) {
	tv, err := time.Parse("2006-01-02 15:04:05", value)
	if err != nil {
		return 0, err
	}
	return time.Date(
		time.Time(tv).Year(),
		time.Time(tv).Month(),
		time.Time(tv).Day(),
		time.Time(tv).Hour(),
		time.Time(tv).Minute(),
		time.Time(tv).Second(),
		0, time.Local,
	).UTC().Unix(), nil
}
