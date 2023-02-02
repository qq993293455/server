package XDayGoal

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
	xDayGoalDao "coin-server/game-server/service/xdaygoal/dao"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"

	"go.uber.org/zap"
)

const (
	activityMailId7  = 100013
	scoreMailId7     = 100012
	activityMailId14 = 100037
	scoreMailId14    = 100036
)

type Cnf struct {
	ActivityId       int64
	StartTime        int64
	EndTime          int64
	ActivityTimeType int64
	IsForever        bool
	ActivityXDayCnf  map[int64]rulemodel.ActivityXdaygoal
}

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	cnf        map[int64]*Cnf
	*module.Module
}

func NewXDayGoalService(
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
		cnf:        make(map[int64]*Cnf),
	}
	module.XDayGoalService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("查询x天所有活动情况", this_.GetXDayGoalAllInfo)
	this_.svc.RegisterFunc("获取x天活动奖励", this_.GetReceiveActivity)
	this_.svc.RegisterFunc("获取x天活动积分奖励", this_.GetReceiveScoreReward)
	this_.svc.RegisterFunc("获取x天活动配置", this_.GetActivityConf)
	this_.svc.RegisterFunc("获取x天活动积分配置", this_.GetScoreConf)

	this_.svc.RegisterFunc("GM", this_.XDayGmChangeEndTime)
	this_.svc.RegisterFunc("GM", this_.XDayGmChangeDays)

	eventlocal.SubscribeEventLocal(this_.HandleTaskChange)
	eventlocal.SubscribeEventLocal(this_.HandleRoleLoginEvent)
	this_.Init()
}

func (this_ *Service) Init() {
	context := ctx.GetContext()
	xDayCnf := rule.MustGetReader(context).ActivityXdaygoal.List()

	for _, cnf := range xDayCnf {
		_, ok1 := this_.cnf[cnf.ActivityId]
		if !ok1 {
			activityRewardCnf, ok2 := rule.MustGetReader(context).Activity.GetActivityById(cnf.ActivityId)
			if !ok2 {
				panic(fmt.Sprintf("GetActivityRewardById error not find ActivityId %d", cnf.ActivityId))
			}

			aCnf := &Cnf{
				ActivityId:       cnf.ActivityId,
				ActivityTimeType: activityRewardCnf.TimeType,
				IsForever:        false,
				ActivityXDayCnf:  make(map[int64]rulemodel.ActivityXdaygoal),
			}

			if aCnf.ActivityTimeType == 1 {
				activityOpenTime, err := strconv.Atoi(activityRewardCnf.ActivityOpenTime)
				if err != nil {
					panic(fmt.Errorf("activity表ActivityOpenTime配置错误，id=%d，错误信息：%s", activityRewardCnf.Id, err.Error()))
				}
				// 这个时间是需要包含当天的
				if activityOpenTime >= 86400 {
					activityOpenTime -= 86400
				}
				aCnf.StartTime = values.Integer(activityOpenTime)
				aCnf.EndTime, err = strconv.ParseInt(activityRewardCnf.DurationTime, 10, 64)
				if err != nil {
					panic(err)
				}
				if aCnf.EndTime < 0 {
					aCnf.IsForever = true
				}
			} else if aCnf.ActivityTimeType == 2 {
				var err error
				aCnf.StartTime, err = GetTime(activityRewardCnf.ActivityOpenTime)
				if err != nil {
					panic(fmt.Sprintf("GetActivityRewardById ActivityId %d GetTime Start error %s", cnf.ActivityId, err))
				}
				aCnf.EndTime, err = GetTime(activityRewardCnf.DurationTime)
				if err != nil {
					panic(fmt.Sprintf("GetActivityRewardById ActivityId %d GetTime End error %s", cnf.ActivityId, err))
				}
			} else {
				panic(fmt.Sprintf("GetActivityRewardById ActivityId %d Time Type error", cnf.ActivityId))
			}

			this_.cnf[cnf.ActivityId] = aCnf
		}
		xCnf := this_.cnf[cnf.ActivityId]
		xCnf.ActivityXDayCnf[cnf.Id] = cnf
	}
}

// msg proc ============================================================================================================================================================================================================
func (this_ *Service) GetXDayGoalAllInfo(ctx *ctx.Context, req *servicepb.XDayGoal_XDayGoalAllInfoRequest) (*servicepb.XDayGoal_XDayGoalAllInfoResponse, *errmsg.ErrMsg) {
	data, err := this_.GetData(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	registerTime, err := this_.GetRegisterTime(ctx)
	if err != nil {
		return nil, err
	}
	this_.UpdateActivity(ctx, data, registerTime)

	var infos map[int64]*models.XDayGoalInfoDetail = make(map[int64]*models.XDayGoalInfoDetail)
	for _, infoPb := range data.Infos {
		ctx.Info("xday infoPb", zap.Any("infoPb", infoPb))
		info, ok := this_.GetInfo(ctx, infoPb)
		if !ok {
			continue
		}
		infos[infoPb.ActivityId] = info
	}

	return &servicepb.XDayGoal_XDayGoalAllInfoResponse{
		Infos: infos,
	}, nil
}

func (this_ *Service) GetInfo(ctx *ctx.Context, infoPb *models.XDayGoalInfo) (*models.XDayGoalInfoDetail, bool) {
	cnf, ok := this_.cnf[infoPb.ActivityId]
	if !ok {
		ctx.Error("data error", zap.Int64("activityId", infoPb.ActivityId))
		return nil, false
	}

	maxDay := int64(0)
	for _, cfg := range cnf.ActivityXDayCnf {
		if cfg.Days > maxDay {
			maxDay = cfg.Days
		}
	}

	signDays := infoPb.Days
	if infoPb.Days > maxDay {
		signDays = maxDay
	}

	info := &models.XDayGoalInfoDetail{
		ActivityId: infoPb.ActivityId,
		Score:      infoPb.Score,
		Days:       signDays,
		OverTime:   infoPb.OverTime,
		Progress:   make(map[int64]int64),
	}
	for id := range infoPb.ReceivedIds {
		info.ReceivedIds = append(info.ReceivedIds, id)
	}
	for id := range infoPb.CanReceiveIds {
		_, ok := cnf.ActivityXDayCnf[id]
		if !ok {
			continue
		}
		temp, ok := cnf.ActivityXDayCnf[id]
		if !ok {
			continue
		}
		// 还未到指定的天数，不下发给客户端
		if temp.Days > info.Days {
			continue
		}
		info.CanReceiveIds = append(info.CanReceiveIds, id)
	}
	for id := range infoPb.ScoreReceivedIds {
		info.ScoreReceivedIds = append(info.ScoreReceivedIds, id)
	}

	for _, cfg := range cnf.ActivityXDayCnf {
		if _, ok := infoPb.ReceivedIds[cfg.Id]; ok {
			continue
		}
		// if _, ok := infoPb.CanReceiveIds[cfg.Id]; ok {
		// 	continue
		// }

		rCfg, ok := rule.MustGetReader(nil).TaskType.GetTaskTypeById(cfg.TaskType[0])
		if !ok {
			panic(errmsg.NewInternalErr("task_type not found: " + strconv.Itoa(int(cfg.TaskType[0]))))
		}

		if rCfg.IsAccumulate {
			counter, err := this_.TaskService.GetCounterByType(ctx, models.TaskType(cfg.TaskType[0]))
			if err != nil {
				panic(err)
			}
			count := counter[cfg.TaskType[1]]
			info.Progress[cfg.Id] = count
		} else {
			info.Progress[cfg.Id] = infoPb.Progress[cfg.Id]
		}
	}
	return info, true
}

func (this_ *Service) GetReceiveActivity(ctx *ctx.Context, req *servicepb.XDayGoal_XDayGoalGetReceiveActivityRequest) (*servicepb.XDayGoal_XDayGoalGetReceiveActivityResponse, *errmsg.ErrMsg) {
	data, err := this_.GetData(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	cnf, ok := this_.cnf[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrXDayGoalNotData()
	}

	info, ok := data.Infos[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrXDayGoalNotData()
	}

	_, ok = info.ReceivedIds[req.Id]
	if ok {
		return nil, errmsg.NewErrXDayGoalNoReward()
	}

	_, ok = info.CanReceiveIds[req.Id]
	if !ok {
		return nil, errmsg.NewErrXDayGoalNotComplete()
	}

	dCnf, ok := cnf.ActivityXDayCnf[req.Id]
	if !ok {
		return nil, errmsg.NewErrXDayGoalConfig()
	}
	if dCnf.Days > info.Days {
		return nil, errmsg.NewErrXDayGoalNotDay()
	}

	score, rewardItems, err := this_.GetActivityReward(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	info.Score += score
	info.ReceivedIds[req.Id] = rewardItems

	delete(info.CanReceiveIds, req.Id)
	xDayGoalDao.SaveSevenDaysInfo(ctx, data)

	outInfo, _ := this_.GetInfo(ctx, info)

	return &servicepb.XDayGoal_XDayGoalGetReceiveActivityResponse{
		RewardItem: rewardItems,
		Info:       outInfo,
	}, nil
}

func (this_ *Service) GetReceiveScoreReward(ctx *ctx.Context, req *servicepb.XDayGoal_XDayGoalGetReceiveScoreRewardRequest) (*servicepb.XDayGoal_XDayGoalGetReceiveScoreRewardResponse, *errmsg.ErrMsg) {
	data, err := this_.GetData(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	info, ok := data.Infos[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrXDayGoalNotData()
	}

	_, ok = info.ScoreReceivedIds[req.Id]
	if ok {
		return nil, errmsg.NewErrXDayGoalScoreReceived()
	}

	rewardItems, err := this_.GetScoreReward(ctx, info.Score, req.Id)
	if err != nil {
		return nil, err
	}

	info.ScoreReceivedIds[req.Id] = rewardItems
	xDayGoalDao.SaveSevenDaysInfo(ctx, data)

	outInfo, _ := this_.GetInfo(ctx, info)
	return &servicepb.XDayGoal_XDayGoalGetReceiveScoreRewardResponse{
		RewardItem: rewardItems,
		Info:       outInfo,
	}, nil
}

func (this_ *Service) GetActivityConf(ctx *ctx.Context, req *servicepb.XDayGoal_XDayGoalGetActivityConfRequest) (*servicepb.XDayGoal_XDayGoalGetActivityConfResponse, *errmsg.ErrMsg) {
	activityInfo, ok := this_.cnf[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrXDayGoalNotData()
	}

	var tabCnf []*models.XDayCnfInfo
	for _, cfg := range activityInfo.ActivityXDayCnf {
		tabCnf = append(tabCnf, &models.XDayCnfInfo{
			Id:           cfg.Id,
			ActivityId:   cfg.ActivityId,
			Kind:         cfg.Kind,
			Language1:    cfg.Language1,
			Days:         cfg.Days,
			TaskType:     cfg.TaskType,
			Language2:    cfg.Language2,
			Reward:       cfg.Reward,
			Points:       cfg.Points,
			GuideProcess: cfg.GuideProcess,
			GuideType:    cfg.GuideType,
		})
	}

	return &servicepb.XDayGoal_XDayGoalGetActivityConfResponse{
		Cnf: tabCnf,
	}, nil
}

func (this_ *Service) GetScoreConf(ctx *ctx.Context, req *servicepb.XDayGoal_XDayGoalGetScoreConfRequest) (*servicepb.XDayGoal_XDayGoalGetScoreConfResponse, *errmsg.ErrMsg) {
	scoreCnf := rule.MustGetReader(ctx).ActivityXdaygoalScorereward.List()
	var tabCnf []*models.XDayScoreCnfInfo
	for _, cfg := range scoreCnf {
		if cfg.ActivityId == req.ActivityId {
			tabCnf = append(tabCnf, &models.XDayScoreCnfInfo{
				Id:         cfg.Id,
				ActivityId: cfg.ActivityId,
				Points:     cfg.Points,
				Reward:     cfg.Reward,
			})
		}
	}
	return &servicepb.XDayGoal_XDayGoalGetScoreConfResponse{
		Cnf: tabCnf,
	}, nil
}

func (this_ *Service) XDayGmChangeEndTime(ctx *ctx.Context, req *servicepb.XDayGoal_CheatXDayGmChangeEndTimeRequest) (*servicepb.XDayGoal_CheatXDayGmChangeEndTimeResponse, *errmsg.ErrMsg) {
	activity, ok := this_.cnf[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrXDayGoalNotData()
	}

	activity.EndTime = req.EndTime

	return &servicepb.XDayGoal_CheatXDayGmChangeEndTimeResponse{}, nil
}

func (this_ *Service) XDayGmChangeDays(ctx *ctx.Context, req *servicepb.XDayGoal_CheatXDayGmChangeDaysRequest) (*servicepb.XDayGoal_CheatXDayGmChangeDaysResponse, *errmsg.ErrMsg) {
	data, err := this_.GetData(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	activity, ok := data.Infos[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrXDayGoalNotData()
	}
	activity.Days = req.Days

	xDayGoalDao.SaveSevenDaysInfo(ctx, data)

	info, ok := this_.GetInfo(ctx, activity)
	if !ok {
		return nil, errmsg.NewErrXDayGoalNotData()
	}
	return &servicepb.XDayGoal_CheatXDayGmChangeDaysResponse{
		Info: info,
	}, nil
}

func (this_ *Service) GetActivityReward(ctx *ctx.Context, id int64) (int64, *models.Items, *errmsg.ErrMsg) {
	score := int64(0)
	cnf, ok := rule.MustGetReader(ctx).ActivityXdaygoal.GetActivityXdaygoalById(id)
	if !ok {
		return 0, nil, errmsg.NewErrXDayGoalConfig()
	}
	rewardItems := &models.Items{}
	if len(cnf.Reward)%2 != 0 {
		panic(fmt.Sprintf("ActivityXdaygoal is error id %d len %d", id, len(cnf.Reward)))
	}
	for i := 0; i < len(cnf.Reward); i += 2 {
		itemId := cnf.Reward[i]
		Count := cnf.Reward[i+1]
		if itemId == cnf.Points {
			score += Count
			continue
		}

		item := &models.Item{
			ItemId: itemId,
			Count:  Count,
		}
		err := this_.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
		if err != nil {
			ctx.Error("xDay Activity ReceiveReward AddManyItemPb err",
				zap.Any("err info", err), zap.Any("role", ctx.RoleId), zap.Any("itemid", item.ItemId), zap.Any("item cnt", item.Count))
			break
		}
		rewardItems.Items = append(rewardItems.Items, item)
	}
	return score, rewardItems, nil
}

func (this_ *Service) SendAcitvityReward(ctx *ctx.Context, info *models.XDayGoalInfo, activityId values.Integer) {
	normalTextId := activityMailId7
	scoreTextId := scoreMailId7
	if activityId == enum.XDayGoal14 {
		normalTextId = activityMailId14
		scoreTextId = scoreMailId14
	}
	score := int64(0)
	rewardItems := make(map[int64]int64)
	for id := range info.CanReceiveIds {
		cnf, ok := rule.MustGetReader(ctx).ActivityXdaygoal.GetActivityXdaygoalById(id)
		if !ok {
			ctx.Error("GetActivityXdaygoalById err not find", zap.Any("id", id))
		}

		if len(cnf.Reward)%2 != 0 {
			panic(fmt.Sprintf("ActivityXdaygoal is error id %d len %d", id, len(cnf.Reward)))
		}
		for i := 0; i < len(cnf.Reward); i += 2 {
			itemId := cnf.Reward[i]
			Count := cnf.Reward[i+1]
			if itemId == cnf.Points {
				score += Count
				continue
			}
			rewardItems[itemId] += Count
		}
	}
	if len(rewardItems) > 0 {
		var rewardList []*models.Item
		for itemId, count := range rewardItems {
			rewardList = append(rewardList, &models.Item{
				ItemId: itemId,
				Count:  count,
			})
		}

		if err := this_.MailService.Add(ctx, ctx.RoleId, &models.Mail{
			Type:       models.MailType_MailTypeSystem,
			TextId:     values.Integer(normalTextId),
			Attachment: rewardList,
		}); err != nil {
			ctx.Error("send maill error", zap.Any("msg", err), zap.Any("reward", rewardItems), zap.Any("roleid", ctx.RoleId))
		}
	}

	for k := range rewardItems {
		delete(rewardItems, k)
	}

	info.Score += score
	scoreCnf := rule.MustGetReader(ctx).ActivityXdaygoalScorereward.List()
	for _, cfg := range scoreCnf {
		_, ok := info.ScoreReceivedIds[cfg.Id]
		if ok {
			continue
		}
		if cfg.ActivityId != activityId {
			continue
		}
		if cfg.Points <= info.Score {
			for i := 0; i < len(cfg.Reward); i += 2 {
				itemId := cfg.Reward[i]
				Count := cfg.Reward[i+1]
				rewardItems[itemId] += Count
			}
		}
	}

	if len(rewardItems) == 0 {
		return
	}

	var rewardList []*models.Item
	for itemId, count := range rewardItems {
		rewardList = append(rewardList, &models.Item{
			ItemId: itemId,
			Count:  count,
		})
	}

	if err := this_.MailService.Add(ctx, ctx.RoleId, &models.Mail{
		Type:       models.MailType_MailTypeSystem,
		TextId:     values.Integer(scoreTextId),
		Attachment: rewardList,
	}); err != nil {
		ctx.Error("send maill error", zap.Any("msg", err), zap.Any("reward", rewardItems), zap.Any("roleid", ctx.RoleId))
	}
}

func (this_ *Service) GetScoreReward(ctx *ctx.Context, score int64, id int64) (*models.Items, *errmsg.ErrMsg) {
	cnf, ok := rule.MustGetReader(ctx).ActivityXdaygoalScorereward.GetActivityXdaygoalScorerewardById(id)
	if !ok {
		return nil, errmsg.NewErrXDayGoalConfig()
	}

	if score < cnf.Points {
		return nil, errmsg.NewErrXDayGoalNotScore()
	}

	rewardItems := &models.Items{}
	for i := 0; i < len(cnf.Reward); i += 2 {
		itemId := cnf.Reward[i]
		Count := cnf.Reward[i+1]
		item := &models.Item{
			ItemId: itemId,
			Count:  Count,
		}
		err := this_.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
		if err != nil {
			ctx.Error("xDay Score ReceiveReward AddManyItemPb err",
				zap.Any("err info", err), zap.Any("role", ctx.RoleId), zap.Any("itemid", item.ItemId), zap.Any("item cnt", item.Count))
			break
		}
		rewardItems.Items = append(rewardItems.Items, item)
	}
	return rewardItems, nil
}

func (this_ *Service) UpdateActivity(ctx *ctx.Context, data *dao.XDayGoalData, createdAt values.Integer) {
	haveChange := false
	timeNow := timer.Now().UTC().Unix()
	for _, activity := range this_.cnf {
		startTime := createdAt + activity.StartTime
		if timeNow < startTime {
			continue
		}

		isTimeOut := false
		overTime := int64(0)
		if activity.ActivityTimeType == 1 {
			overTime = -1
			if !activity.IsForever {
				overTime = startTime + activity.EndTime
				if timeNow > overTime {
					isTimeOut = true
				}
			}
		}
		if activity.ActivityTimeType == 2 {
			overTime = activity.EndTime
			if timeNow > activity.EndTime {
				isTimeOut = true
			}
		}

		if isTimeOut {
			aData, ok := data.Infos[activity.ActivityId]
			if ok {
				this_.SendAcitvityReward(ctx, aData, activity.ActivityId)
				delete(data.Infos, activity.ActivityId)
				haveChange = true
			}
			continue
		}

		isNew := false
		_, ok := data.Infos[activity.ActivityId]
		if !ok {
			aData := &models.XDayGoalInfo{
				ActivityId:       activity.ActivityId,
				Score:            0,
				Days:             0,
				ReceivedIds:      make(map[int64]*models.Items),
				ScoreReceivedIds: make(map[int64]*models.Items),
				CanReceiveIds:    make(map[int64]int64),
				Progress:         make(map[int64]int64),
				LastLoginTime:    make(map[int64]int64),
			}
			data.Infos[activity.ActivityId] = aData
			isNew = true
		}
		aData := data.Infos[activity.ActivityId]
		if aData.OverTime != overTime {
			aData.OverTime = overTime
			haveChange = true
		}

		if activity.ActivityTimeType == 1 {
			loginDays := this_.GetDays(ctx, startTime)
			if aData.Days != loginDays {
				aData.Days = loginDays
				haveChange = true
			}
		}
		if activity.ActivityTimeType == 2 {
			days := this_.GetDays(ctx, activity.StartTime)
			if aData.Days != days {
				aData.Days = days
				haveChange = true
			}
		}

		if isNew {
			for _, aCfg := range activity.ActivityXDayCnf {
				if len(aCfg.TaskType) != 3 {
					ctx.Error("xDay config error", zap.Any("id", aCfg.Id), zap.Any("taskCnf len", len(aCfg.TaskType)))
					continue
				}

				isComplete := false

				cfg, ok := rule.MustGetReader(nil).TaskType.GetTaskTypeById(aCfg.TaskType[0])
				if !ok {
					panic(errmsg.NewInternalErr("task_type not found: " + strconv.Itoa(int(aCfg.TaskType[0]))))
				}

				if cfg.IsAccumulate {
					counter, err := this_.TaskService.GetCounterByType(ctx, models.TaskType(aCfg.TaskType[0]))
					if err != nil {
						panic(err)
					}
					count := counter[aCfg.TaskType[1]]
					if cfg.IsReversed {
						isComplete = count <= aCfg.TaskType[2]
					} else {
						isComplete = count >= aCfg.TaskType[2]
					}
					aData.Progress[aCfg.Id] = count
					haveChange = true
				}

				if isComplete {
					aData.Progress[aCfg.Id] = aCfg.TaskType[2]
					aData.CanReceiveIds[aCfg.Id] = timeNow
				}
			}
		}
	}
	if haveChange {
		xDayGoalDao.SaveSevenDaysInfo(ctx, data)
	}
}

func (this_ *Service) HandleTaskChange(ctx *ctx.Context, d *event.TargetUpdate) *errmsg.ErrMsg {
	dataPb, err := this_.GetData(ctx, ctx.RoleId)
	if err != nil {
		ctx.Error("handleTaskChange Error", zap.Any("error", err))
		return err
	}

	msg := &servicepb.XDayGoal_XDayGoalActivityProgressPush{
		Progress: make(map[int64]*models.ActivityProgress),
	}
	timeNow := timer.Now().Unix()

	haveChange := false
	for _, info := range dataPb.Infos {
		activityCnf, ok := this_.cnf[info.ActivityId]
		if !ok {
			ctx.Debug("this_.cnf not find activity", zap.Int64("ActivityId", info.ActivityId))
			continue
		}

		isTimeOut := false
		if activityCnf.ActivityTimeType == 1 {
			if !activityCnf.IsForever {
				if timeNow > dataPb.RegisterTime+activityCnf.StartTime+activityCnf.EndTime {
					isTimeOut = true
				}
			}
		}
		if activityCnf.ActivityTimeType == 2 {
			if timeNow > activityCnf.EndTime {
				isTimeOut = true
			}
		}

		if isTimeOut {
			continue
		}

		for _, cnf := range activityCnf.ActivityXDayCnf {
			_, ok := info.ReceivedIds[cnf.Id]
			if ok {
				continue
			}
			_, ok = info.CanReceiveIds[cnf.Id]
			if ok {
				continue
			}

			if len(cnf.TaskType) != 3 {
				ctx.Error("xday TargetUpdate fail task type error", zap.Int64("activity id", cnf.Id), zap.Int("task type len", len(cnf.TaskType)))
				continue
			}
			if d.Typ == models.TaskType(cnf.TaskType[0]) && d.Id == cnf.TaskType[1] {
				_, ok = msg.Progress[info.ActivityId]
				if !ok {
					msg.Progress[info.ActivityId] = &models.ActivityProgress{
						Progress: make(map[int64]int64),
					}
				}
				if d.Typ == models.TaskType_TaskLogin {
					DefaultRefreshTime, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("DefaultRefreshTime")
					if !ok {
						ctx.Error("HandleTaskChange DefaultRefreshTime error")
						panic("HandleTaskChange DefaultRefreshTime error")
					}

					tBeginRefreshTime := timer.BeginOfDay(timer.Now()).Unix() + DefaultRefreshTime

					_, ok1 := info.LastLoginTime[cnf.Id]
					if ok1 {
						if info.LastLoginTime[cnf.Id] > tBeginRefreshTime {
							ctx.Info("xday_login", zap.Any("lastLoginTime", info.LastLoginTime[cnf.Id]), zap.Any("Progress", info.Progress[cnf.Id]))
							continue
						}
					}
					info.LastLoginTime[cnf.Id] = tBeginRefreshTime + 86400
				}

				progress := msg.Progress[info.ActivityId]

				cfg, ok := rule.MustGetReader(ctx).TaskType.GetTaskTypeById(int64(d.Typ))
				if !ok {
					ctx.Warn("task_type not found", zap.Int64("type", int64(d.Typ)))
					continue
				}

				if cfg.IsAccumulate {
					info.Progress[cnf.Id] = d.Count
				} else {
					if d.IsReplace {
						info.Progress[cnf.Id] = d.Incr
					} else {
						info.Progress[cnf.Id] += d.Incr
					}
				}
				progress.Progress[cnf.Id] = info.Progress[cnf.Id]
				if cfg.IsReversed {
					if info.Progress[cnf.Id] <= cnf.TaskType[2] {
						info.Progress[cnf.Id] = cnf.TaskType[2]
						info.CanReceiveIds[cnf.Id] = timeNow
						// 还未到指定的天数，不下发给客户端
						if cnf.Days <= info.Days {
							progress.CanReceiveIds = append(progress.CanReceiveIds, cnf.Id)
						}
					}
				} else {
					if info.Progress[cnf.Id] >= cnf.TaskType[2] {
						info.Progress[cnf.Id] = cnf.TaskType[2]
						info.CanReceiveIds[cnf.Id] = timeNow
						// 还未到指定的天数，不下发给客户端
						if cnf.Days <= info.Days {
							progress.CanReceiveIds = append(progress.CanReceiveIds, cnf.Id)
						}
					}
				}
				haveChange = true
			}
		}
	}
	if haveChange {
		xDayGoalDao.SaveSevenDaysInfo(ctx, dataPb)
		ctx.PushMessage(msg)
	}
	return nil
}

func (this_ *Service) HandleRoleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	data, err := this_.GetData(ctx, ctx.RoleId)
	if err != nil {
		ctx.Error("HandleRoleLoginEvent SevenDays Error", zap.Any("errmsg", err))
		return nil
	}
	registerTime, err := this_.GetRegisterTime(ctx)
	if err != nil {
		return err
	}

	this_.UpdateActivity(ctx, data, registerTime)
	return nil
}

func (this_ *Service) GetData(ctx *ctx.Context, roleId string) (*dao.XDayGoalData, *errmsg.ErrMsg) {
	data := xDayGoalDao.GetXDayGoalInfo(ctx, roleId)
	if data == nil {
		registerTime, err := this_.GetRegisterTime(ctx)
		if err != nil {
			return nil, err
		}
		data = &dao.XDayGoalData{
			RoleId:       roleId,
			RegisterTime: registerTime,
			Infos:        make(map[int64]*models.XDayGoalInfo),
		}
	}

	if data.Infos == nil {
		data.Infos = make(map[int64]*models.XDayGoalInfo)
	}

	for _, info := range data.Infos {
		if info.ReceivedIds == nil {
			info.ReceivedIds = make(map[int64]*models.Items)
		}
		if info.ScoreReceivedIds == nil {
			info.ScoreReceivedIds = make(map[int64]*models.Items)
		}
		if info.CanReceiveIds == nil {
			info.CanReceiveIds = make(map[int64]int64)
		}
		if info.Progress == nil {
			info.Progress = make(map[int64]int64)
		}
		if info.LastLoginTime == nil {
			info.LastLoginTime = make(map[int64]int64)
		}
	}
	return data, nil
}

func (this_ *Service) GetDays(ctx *ctx.Context, startTime int64) int64 {
	DefaultRefreshTime, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("DefaultRefreshTime")
	if !ok {
		ctx.Error("DefaultRefreshTime error")
		panic("DefaultRefreshTime error")
	}

	tBeginRefreshTime := timer.BeginOfDay(timer.Now()).Unix() + DefaultRefreshTime
	if timer.Now().Unix() > tBeginRefreshTime {
		tBeginRefreshTime += 86400
	}
	return (tBeginRefreshTime - startTime) / 86400
}

// func ============================================================================================================================================================================================================
func GetTime(value string) (int64, error) {
	tv, err := time.Parse("2006-01-02 15:04:05", value)
	if err != nil {
		return 0, err
	}
	return tv.Unix(), nil
	// return time.Date(
	// 	time.Time(tv).Year(),
	// 	time.Time(tv).Month(),
	// 	time.Time(tv).Day(),
	// 	time.Time(tv).Hour(),
	// 	time.Time(tv).Minute(),
	// 	time.Time(tv).Second(),
	// 	0, time.Local,
	// ).UTC().Unix(), nil
}

func (this_ *Service) GetRegisterTime(ctx *ctx.Context) (values.Integer, *errmsg.ErrMsg) {
	role, err := this_.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return 0, err
	}
	registerTime := time.Unix(role.CreateTime/1000, 0).UTC()
	refresh := this_.RefreshService.GetCurrDayFreshTimeWith(ctx, registerTime)

	return refresh.Unix(), nil
}
