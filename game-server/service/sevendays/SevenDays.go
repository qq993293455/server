package sevenDays

import (
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
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	sevenDaysDao "coin-server/game-server/service/sevendays/dao"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"

	"go.uber.org/zap"
)

const (
	CheckError         = -1
	CheckIsNoStart     = 0
	CheckIsOver        = 1
	CheckIsActivaction = 2
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	*module.Module
	log *logger.Logger
}

func NewSevenDaysService(
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
	module.SevenDaysService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("查询7天情况", this_.GetSevenDaysInfo)
	this_.svc.RegisterFunc("领取7天奖品", this_.ReceiveReward)
	this_.svc.RegisterFunc("获取配置", this_.SevenDaysGetCnf)
	this_.svc.RegisterFunc("GM", this_.SevenDayGm)
	eventlocal.SubscribeEventLocal(this_.HandleRoleLoginEvent)
}

// msg proc ============================================================================================================================================================================================================
func (this_ *Service) GetSevenDaysInfo(ctx *ctx.Context, req *servicepb.SevenDays_SevenDaysInfoRequest) (*servicepb.SevenDays_SevenDaysInfoResponse, *errmsg.ErrMsg) {
	this_.update(ctx)
	data := sevenDaysDao.GetSevenDaysInfo(ctx, ctx.RoleId)

	ret := &servicepb.SevenDays_SevenDaysInfoResponse{}

	for _, aData := range data.Infos {
		ret.Infos = append(ret.Infos, &models.SevenDaysInfo{
			SignedInDay:        aData.SignedInDay,
			ReceivedRewardDays: aData.ReceivedRewardDays,
			ActivityId:         aData.ActivitId,
		})

	}

	return ret, nil
}

func (this_ *Service) IsSevenDaysReceiveAll(ctx *ctx.Context) bool {
	data := sevenDaysDao.GetSevenDaysInfo(ctx, ctx.RoleId)

	aConfigs := rule.MustGetReader(ctx).ActivityLoginReward.List()

	for _, config := range aConfigs {
		rLen := len(config.ActivityReward)
		if rLen < 3 || rLen%2 != 0 {
			this_.log.Error("GetActivityRewardData err", zap.Any("err msg", errmsg.NewErrSevenDaysConfig()))
			return false
		}

		day := config.Day

		isReceived := false
		for _, aData := range data.Infos {
			if aData.ActivitId != config.TypId {
				continue
			}

			for _, tempDay := range aData.ReceivedRewardDays {
				if tempDay == day {
					isReceived = true
					break
				}
			}
			break
		}

		if isReceived {
			continue
		}

		return false
	}

	return true
}

func (this_ *Service) SevenDaysGetCnf(ctx *ctx.Context, req *servicepb.SevenDays_SevenDaysGetCnfRequest) (*servicepb.SevenDays_SevenDaysGetCnfResponse, *errmsg.ErrMsg) {
	cnfs := rule.MustGetReader(ctx).ActivityLoginReward.List()
	ret := &servicepb.SevenDays_SevenDaysGetCnfResponse{}
	for _, cnf := range cnfs {
		ret.Cnfs = append(ret.Cnfs, &models.SevenDayCnf{
			Id:             cnf.Id,
			TypId:          cnf.TypId,
			Day:            cnf.Day,
			ActivityReward: cnf.ActivityReward,
		})
	}
	return ret, nil
}

func (this_ *Service) ReceiveReward(ctx *ctx.Context, req *servicepb.SevenDays_SevenDaysReceiveRequest) (*servicepb.SevenDays_SevenDaysReceiveResponse, *errmsg.ErrMsg) {
	this_.update(ctx)
	data := sevenDaysDao.GetSevenDaysInfo(ctx, ctx.RoleId)

	var aData *models.SevenDayInfo
	for _, aD := range data.Infos {
		if aD.ActivitId != req.ActivityId {
			continue
		}
		aData = aD
		break
	}

	if aData == nil {
		return nil, errmsg.NewErrSevenDaysNoActiviy()
	}

	if !req.ReceivedAll {
		for _, day := range aData.ReceivedRewardDays {
			if day == req.ReceivedDay {
				return nil, errmsg.NewErrSevenDaysReceived()
			}
		}

		if req.ReceivedDay > aData.SignedInDay {
			return nil, errmsg.NewErrSevenDaysNotTime()
		}
		// return nil, errmsg.NewErrSevenDaysNoActiviy()
	}

	hasChange := false
	var thisTimeReceived []int64

	aConfigs := rule.MustGetReader(ctx).ActivityLoginReward.List()

	for _, config := range aConfigs {
		rLen := len(config.ActivityReward)
		if rLen < 3 || rLen%2 != 0 {
			return nil, errmsg.NewErrSevenDaysConfig()
		}
		day := config.Day
		if !req.ReceivedAll && day != int64(req.ReceivedDay) {
			continue
		}

		isReceived := false
		for _, tempDay := range aData.ReceivedRewardDays {
			if int64(tempDay) == day {
				isReceived = true
				break
			}
		}

		if isReceived {
			continue
		}

		for i := 0; i < rLen; i += 2 {
			item := &models.Item{
				ItemId: config.ActivityReward[i],
				Count:  config.ActivityReward[i+1],
			}
			err := this_.BagService.AddManyItemPb(ctx, ctx.RoleId, item)
			if err != nil {
				this_.log.Error("sevenday ReceiveReward AddManyItemPb err",
					zap.Any("err info", err), zap.Any("role", ctx.RoleId), zap.Any("itemid", item.ItemId), zap.Any("item cnt", item.Count))
				break
			}
		}
		hasChange = true
		aData.ReceivedRewardDays = append(aData.ReceivedRewardDays, day)
		thisTimeReceived = append(thisTimeReceived, day)
	}

	if hasChange {
		sevenDaysDao.SaveSevenDaysInfo(ctx, data)
	}

	return &servicepb.SevenDays_SevenDaysReceiveResponse{
		SignedInDay:                aData.SignedInDay,
		ReceivedRewardDays:         aData.ReceivedRewardDays,
		ThisTimeReceivedRewardDays: thisTimeReceived,
	}, nil
}

func (this_ *Service) SevenDayGm(ctx *ctx.Context, req *servicepb.SevenDays_CheatSevenDayGmRequest) (*servicepb.SevenDays_CheatSevenDayGmResponse, *errmsg.ErrMsg) {
	data := sevenDaysDao.GetSevenDaysInfo(ctx, ctx.RoleId)

	// var aData *models.SevenDayInfo
	// for _, aD := range data.Infos {
	// 	if aD.ActivitId != req.ActivityId {
	// 		continue
	// 	}
	// 	aData = aD
	// 	break
	// }
	if req.StartTime > 7 {
		req.StartTime = 7
	}
	var lastDay values.Integer
	for i := 0; i < len(data.Infos); i++ {
		if data.Infos[i].ActivitId == req.ActivityId {
			lastDay = data.Infos[i].SignedInDay
			data.Infos[i].SignedInDay = req.StartTime
		}
	}

	if lastDay < req.StartTime {
		data.RegisterTime = data.RegisterTime - ((req.StartTime - lastDay) * 24 * 60 * 60)
	}
	// if aData == nil {
	// 	return nil, errmsg.NewErrSevenDaysNoActiviy()
	// }
	//
	// if req.StartTime > 0 {
	// 	aData.StartTime = req.StartTime
	// }

	// if req.RegisterTime > 0 {
	// 	data.RegisterTime = req.RegisterTime
	// }

	sevenDaysDao.SaveSevenDaysInfo(ctx, data)

	return &servicepb.SevenDays_CheatSevenDayGmResponse{}, nil
}

func (this_ *Service) HandleRoleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	this_.update(ctx)
	return nil
}

func (this_ *Service) update(ctx *ctx.Context) {
	haveChange := false
	data := sevenDaysDao.GetSevenDaysInfo(ctx, ctx.RoleId)
	aConfigs := rule.MustGetReader(ctx).ActivityLoginReward.List()
	nowTime := timer.Now().Unix()
	if data.RegisterTime == 0 {
		data.RegisterTime = nowTime
	}
	procA := make(map[int64]int64)
	for _, aRConfig := range aConfigs {
		checkData, ok := procA[aRConfig.TypId]
		if !ok {
			procA[aRConfig.TypId] = this_.CheckActivity(ctx, data, aRConfig.TypId, nowTime, "seven days")
		}

		isFind := false
		for _, aData := range data.Infos {
			if aData.ActivitId == aRConfig.TypId {
				isFind = true
				break
			}
		}

		switch checkData {
		case CheckIsActivaction:
			{
				if !isFind {
					haveChange = true
					data.Infos = append(data.Infos, &models.SevenDayInfo{
						ActivitId: aRConfig.TypId,
						StartTime: nowTime,
					})
				}
			}
		case CheckIsOver:
			{
				if isFind {
					// 暂时不处理 策划说后面有需求再开发
				}
			}
		}
	}

	for _, info := range data.Infos {
		aCnf, ok := rule.MustGetReader(ctx).Activity.GetActivityById(info.ActivitId)
		if !ok {
			ctx.Error("not fund activity", zap.Int64("ActivityId", info.ActivitId))
			continue
		}
		loginDays := int64(0)
		if aCnf.TimeType == 1 {
			loginDays = this_.GetDays(ctx, data.RegisterTime)
		}
		if aCnf.TimeType == 2 {
			loginDays = this_.GetDays(ctx, info.StartTime)
		}
		if info.SignedInDay != loginDays {
			info.SignedInDay = loginDays
			haveChange = true
		}
	}
	if haveChange {
		sevenDaysDao.SaveSevenDaysInfo(ctx, data)
	}
}

func (this_ *Service) CheckActivity(ctx *ctx.Context, data *dao.SevenDayData, activityId int64, nowTime int64, activityName string) int64 {
	aCnf, ok := rule.MustGetReader(ctx).Activity.GetActivityById(activityId)
	if !ok {
		ctx.Error("not fund activity", zap.Int64("ActivityId", activityId), zap.String("activityName", activityName))
		return CheckError
	}

	if aCnf.TimeType == 1 {
		endTime, err := strconv.ParseInt(aCnf.DurationTime, 10, 64)
		if err != nil {
			ctx.Error("DurationTime is error", zap.Int64("ActivityId", activityId), zap.String("DurationTime", aCnf.DurationTime), zap.String("activityName", activityName))
			return CheckError
		}
		if endTime > 0 {
			if nowTime > endTime+data.RegisterTime {
				return CheckIsOver
			}
		}
	} else if aCnf.TimeType == 2 {
		startTime, err := GetTime(aCnf.ActivityOpenTime)
		if err != nil {
			ctx.Error("activity ActivityOpenTime error", zap.Int64("ActivityId", activityId), zap.String("ActivityOpenTime", aCnf.ActivityOpenTime), zap.Any("error", err), zap.String("activityName", activityName))
			return CheckError
		}

		if nowTime < startTime {
			return CheckIsNoStart
		}

		endTime, err := GetTime(aCnf.DurationTime)
		if err != nil {
			ctx.Error("activity DurationTime error", zap.Int64("ActivityId", activityId), zap.String("DurationTime", aCnf.ActivityOpenTime), zap.Any("error", err), zap.String("activityName", activityName))
			return CheckError
		}

		if nowTime > endTime {
			return CheckIsOver
		}
	}
	return CheckIsActivaction
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
	return (tBeginRefreshTime-startTime)/86400 + 1
}

// func ============================================================================================================================================================================================================

func GetHMS(timeUnix int64) (int, int, int) {
	local_time := time.Unix(timeUnix, 0).UTC()
	return local_time.Hour(), local_time.Minute(), local_time.Second()
}

func GetStartAndEndDay(config *rulemodel.Activity, log *logger.Logger, registerDay int64) (int64, int64, *errmsg.ErrMsg) {
	if config.TimeType == 1 {
		endTime, err := strconv.ParseInt(config.DurationTime, 10, 64)
		if err != nil {
			return 0, 0, errmsg.NewErrSevenDaysConfig()
		}
		return registerDay, registerDay + endTime, nil
	}
	if config.TimeType == 2 {
		startTime, err := strconv.ParseInt(config.ActivityOpenTime, 10, 64)
		if err != nil {
			return 0, 0, errmsg.NewErrSevenDaysConfig()
		}
		endTime, err := strconv.ParseInt(config.DurationTime, 10, 64)
		if err != nil {
			return 0, 0, errmsg.NewErrSevenDaysConfig()
		}
		return startTime, endTime, nil
	}
	return 0, 0, errmsg.NewErrSevenDaysConfig()
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
