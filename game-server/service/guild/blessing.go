package guild

import (
	"math"
	"sort"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"
	pbdao "coin-server/common/proto/dao"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	"coin-server/common/timer"
	"coin-server/common/utils/percent"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/guild/dao"
	"coin-server/game-server/service/guild/rule"
	rulemodel "coin-server/rule/rule-model"
	"go.uber.org/zap"
)

const countKey = "COUNT_KEY"

func (svc *Service) isJoinedGuild(next handler.HandleFunc) handler.HandleFunc {
	return func(ctx *ctx.Context) *errmsg.ErrMsg {
		user, err := dao.NewGuildUser(ctx.RoleId).Get(ctx)
		if err != nil {
			return err
		}
		if user.GuildId == "" {
			return errmsg.NewErrGuildNoGuild()
		}
		guild, err := dao.NewGuild(user.GuildId).Get(ctx)
		if err != nil {
			return err
		}
		if guild == nil {
			return errmsg.NewErrGuildNoGuild()
		}
		ctx.SetValue(countKey, guild.Count)
		return next(ctx)
	}
}

func (svc *Service) BlessingInfo(ctx *ctx.Context, _ *lessservicepb.Guild_BlessingInfoRequest) (*lessservicepb.Guild_BlessingInfoResponse, *errmsg.ErrMsg) {
	bless, err := dao.GetBlessing(ctx)
	if err != nil {
		return nil, err
	}
	effic, err := svc.getBlessingEffic(ctx)
	if err != nil {
		return nil, err
	}
	return &lessservicepb.Guild_BlessingInfoResponse{
		Stage:     bless.Stage,
		Queue:     bless.Queue,
		Activated: bless.Activated,
		Effic:     effic,
	}, nil
}

func (svc *Service) BlessingStart(ctx *ctx.Context, req *lessservicepb.Guild_BlessingStartRequest) (*lessservicepb.Guild_BlessingStartResponse, *errmsg.ErrMsg) {
	cfg, ok := rule.GetBlessById(ctx, req.BlessId)
	if !ok {
		return nil, errmsg.NewErrGuildInvalidBless()
	}
	bless, err := dao.GetBlessing(ctx)
	if err != nil {
		return nil, err
	}
	list, dataMap := svc.getAvailableBless(ctx, bless, cfg)
	if len(list) <= 0 {
		return nil, errmsg.NewErrGuildDoNotRepeatBless()
	}
	availableDuration, err := svc.getQueueAvailableDuration(ctx, bless)
	if err != nil {
		return nil, err
	}
	effic, err := svc.getBlessingEffic(ctx)
	if err != nil {
		return nil, err
	}
	availableDuration = values.Integer(math.Ceil(percent.AdditionFloat(availableDuration, effic)))
	var update bool
	lastDoneTime := svc.getQueueLastItemDoneTime(ctx, bless)
	for _, id := range list {
		duration := svc.getBlessDuration(bless, dataMap[id])
		if duration > availableDuration {
			break
		}
		availableDuration -= duration
		update = true
		lastDoneTime += duration
		bless.Queue = append(bless.Queue, &models.BlessingQueue{
			Id:       id,
			Duration: duration,
			DoneTime: lastDoneTime,
		})
	}
	if update {
		dao.SaveBlessing(ctx, bless)
	}
	return &lessservicepb.Guild_BlessingStartResponse{
		Stage:     bless.Stage,
		Queue:     bless.Queue,
		Activated: bless.Activated,
	}, nil
}

func (svc *Service) BlessingActivate(ctx *ctx.Context, req *lessservicepb.Guild_BlessingActivateRequest) (*lessservicepb.Guild_BlessingActivateResponse, *errmsg.ErrMsg) {
	_, ok := rule.GetBlessById(ctx, req.BlessId)
	if !ok {
		return nil, errmsg.NewErrGuildInvalidBless()
	}
	bless, err := dao.GetBlessing(ctx)
	if err != nil {
		return nil, err
	}
	var item *models.BlessingQueue
	newList := make([]*models.BlessingQueue, 0)
	for _, queue := range bless.Queue {
		if queue.Id == req.BlessId {
			item = queue
		} else {
			newList = append(newList, queue)
		}
	}
	if item == nil || item.DoneTime > timer.StartTime(ctx.StartTime).Unix() {
		return nil, errmsg.NewErrGuildBlessNotDone()
	}
	bless.Activated = append(bless.Activated, req.BlessId)
	bless.Queue = newList
	dao.SaveBlessing(ctx, bless)

	// 刷新英雄属性
	ctx.PublishEventLocal(&event.BlessActivated{
		Stage:     bless.Stage,
		Page:      bless.Page,
		Activated: bless.Activated,
	})
	return &lessservicepb.Guild_BlessingActivateResponse{
		Stage:     bless.Stage,
		Queue:     bless.Queue,
		Activated: bless.Activated,
	}, nil
}

func (svc *Service) BlessingNextStage(ctx *ctx.Context, _ *lessservicepb.Guild_BlessingNextStageRequest) (*lessservicepb.Guild_BlessingNextStageResponse, *errmsg.ErrMsg) {
	bless, err := dao.GetBlessing(ctx)
	if err != nil {
		return nil, err
	}
	if bless.Stage >= rule.GetMaxBlessStage(ctx) {
		return nil, errmsg.NewErrGuildMaxBlessStage()
	}
	bless.Stage++
	svc.nextPage(ctx, bless)
	bless.Activated = []values.Integer{}
	dao.SaveBlessing(ctx, bless)
	return &lessservicepb.Guild_BlessingNextStageResponse{}, nil
}

func (svc *Service) getAvailableBless(ctx *ctx.Context, bless *pbdao.Blessing, target *rulemodel.GuildBlessing) ([]values.Integer, map[values.Integer]*rulemodel.GuildBlessing) {
	_, dataMap := svc.getAvailableBlessFromConfigByTarget(ctx, target, bless.Page)
	list := make([]values.Integer, 0)
	for _, id := range bless.Activated {
		delete(dataMap, id)
	}
	for _, queue := range bless.Queue {
		delete(dataMap, queue.Id)
	}
	temp := make([]*rulemodel.GuildBlessing, 0)
	for _, item := range dataMap {
		temp = append(temp, item)
	}
	sort.Slice(temp, func(i, j int) bool {
		return temp[i].UnlockId < temp[j].UnlockId
	})
	for _, blessing := range temp {
		list = append(list, blessing.Id)
	}
	return list, dataMap
}

func (svc *Service) getAvailableBlessFromConfigByTarget(ctx *ctx.Context, target *rulemodel.GuildBlessing, page values.Integer) ([]values.Integer, map[values.Integer]*rulemodel.GuildBlessing) {
	data := map[values.Integer]*rulemodel.GuildBlessing{target.Id: target}
	cfg, ok := rule.GetBlessById(ctx, target.Id)
	if ok {
		svc.getByParent(ctx, cfg.BlessingId, data, page)
	}
	ret := make([]values.Integer, 0, len(data))
	for blessId := range data {
		ret = append(ret, blessId)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})
	return ret, data
}

func (svc *Service) getByParent(ctx *ctx.Context, parent []values.Integer, data map[values.Integer]*rulemodel.GuildBlessing, pageId values.Integer) {
	tempParent := make([]values.Integer, 0)
	for _, id := range parent {
		cfg, ok := rule.GetBlessById(ctx, id)
		if !ok || cfg.PageId != pageId {
			continue
		}
		data[id] = cfg
		tempParent = append(tempParent, cfg.BlessingId...)
	}
	if len(tempParent) > 0 {
		svc.getByParent(ctx, tempParent, data, pageId)
	}
}

func (svc *Service) getQueueAvailableDuration(ctx *ctx.Context, bless *pbdao.Blessing) (values.Integer, *errmsg.ErrMsg) {
	count := ctx.GetValue(countKey).(values.Integer)
	duration := rule.GetQueueDuration(ctx)
	duration += count * rule.GetOneMemberQueueDuration(ctx)
	if len(bless.Queue) <= 0 {
		return duration, nil
	}

	return timer.StartTime(ctx.StartTime).Add(time.Second*time.Duration(duration)).Unix() - bless.Queue[len(bless.Queue)-1].DoneTime, nil
}

func (svc *Service) getBlessDuration(bless *pbdao.Blessing, cfg *rulemodel.GuildBlessing) values.Integer {
	if bless.Stage <= 1 {
		return cfg.BlessingTimeFir
	}
	// 2层（表里的pageId>=2）及以上直接读配置表里第二层配置的时间
	return cfg.BlessingTimeSec
	// if bless.Stage == 2 {
	// 	return cfg.BlessingTimeSec
	// }
	// return values.Integer(math.Floor(values.Float(cfg.BlessingTimeSec) * (values.Float(cfg.AddTime) / 10000 * values.Float(bless.Stage-2))))
}

func (svc *Service) getQueueLastItemDoneTime(ctx *ctx.Context, bless *pbdao.Blessing) values.Integer {
	now := timer.StartTime(ctx.StartTime)
	if len(bless.Queue) <= 0 {
		return now.Unix()
	}
	doneTime := bless.Queue[len(bless.Queue)-1].DoneTime
	if doneTime < now.Unix() {
		return now.Unix()
	}
	return doneTime
}

func (svc *Service) nextPage(ctx *ctx.Context, bless *pbdao.Blessing) {
	max := rule.GetMaxBlessPage(ctx)
	bless.Page++
	if bless.Page > max {
		bless.Page = 1
	}
}

func (svc *Service) getBlessingEffic(ctx *ctx.Context) (values.Integer, *errmsg.ErrMsg) {
	id, err := svc.GetGuildIdByRole(ctx)
	if err != nil {
		return 0, err
	}
	if id == "" {
		return 0, nil
	}
	if err := svc.getLock(ctx, guildBlessingEffic+id); err != nil {
		return 0, err
	}
	data, err := dao.GetBlessingEffic(ctx, id)
	if err != nil {
		return 0, err
	}
	now := timer.StartTime(ctx.StartTime).Unix()
	var update bool
	newEffic := make([]*models.BlessingEfficItem, 0)
	efficVal := values.Integer(percent.BASE)
	for _, item := range data.Effic {
		if item.ExpiredAt <= now {
			update = true
		} else {
			newEffic = append(newEffic, item)
			efficVal += item.Effic
		}
	}
	if update {
		data.Effic = newEffic
		dao.SaveBlessingEffic(ctx, data)
	}
	return efficVal, nil
}

// ----------------------仅GVGGuildServer调用----------------------//
func (svc *Service) UpdateBlessingEffic(ctx *ctx.Context, req *lessservicepb.Guild_UpdateBlessingEfficRequest) (*lessservicepb.Guild_UpdateBlessingEfficResponse, *errmsg.ErrMsg) {
	if ctx.ServerType != models.ServerType_GVGGuildServer {
		return nil, errmsg.NewProtocolErrorInfo("invalid server type request")
	}
	if req.GuildId == "" || req.Effic <= 0 {
		ctx.Error("UpdateBlessingEffic invalid request data", zap.Any("req", req))
		return &lessservicepb.Guild_UpdateBlessingEfficResponse{}, nil
	}
	if req.ExpiredAt <= timer.StartTime(ctx.StartTime).Unix() {
		ctx.Warn("UpdateBlessingEffic expiredAt is less than now")
		return &lessservicepb.Guild_UpdateBlessingEfficResponse{}, nil
	}
	if err := svc.getLock(ctx, guildBlessingEffic+req.GuildId); err != nil {
		return nil, err
	}
	data, err := dao.GetBlessingEffic(ctx, req.GuildId)
	if err != nil {
		return nil, err
	}
	data.Effic = append(data.Effic, &models.BlessingEfficItem{
		Effic:     req.Effic,
		ExpiredAt: req.ExpiredAt,
	})
	dao.SaveBlessingEffic(ctx, data)
	return &lessservicepb.Guild_UpdateBlessingEfficResponse{}, nil
}

func (svc *Service) CheatSetBlessStage(ctx *ctx.Context, req *lessservicepb.Guild_CheatSetBlessStageRequest) (*lessservicepb.Guild_CheatSetBlessStageResponse, *errmsg.ErrMsg) {
	bless, err := dao.GetBlessing(ctx)
	if err != nil {
		return nil, err
	}
	if req.Stage <= 0 || req.Page < 0 {
		return nil, errmsg.NewInternalErr("invalid request")
	}
	if req.Stage > rule.GetMaxBlessStage(ctx) {
		return nil, errmsg.NewErrGuildMaxBlessStage()
	}
	maxPage := rule.GetMaxBlessPage(ctx)
	if req.Page > maxPage {
		req.Page = maxPage
	}
	bless.Stage = req.Stage
	if req.Page > 0 {
		bless.Page = req.Page
	}
	bless.Activated = []values.Integer{}
	dao.SaveBlessing(ctx, bless)
	return &lessservicepb.Guild_CheatSetBlessStageResponse{}, nil
}
