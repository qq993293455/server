package activity_weekly

import (
	"fmt"
	"strconv"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	daopb "coin-server/common/proto/dao"
	modelspb "coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum/ActivityId"
	"coin-server/game-server/module"
	"coin-server/game-server/service/activity-weekly/dao"
	rule2 "coin-server/game-server/service/activity-weekly/rule"
	"coin-server/game-server/util"
	"coin-server/rule"

	"go.uber.org/zap"
)

type ChangeInfo struct {
	info []int64
	typ  modelspb.ActivityWeeklyType
}

type Service struct {
	serverId   values.ServerId
	serverType modelspb.ServerType
	svc        *service.Service
	*module.Module
	log *logger.Logger
}

func NewActivityWeeklyService(
	serverId values.ServerId,
	serverType modelspb.ServerType,
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
	module.ActivityWeeklyService = s
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取活动情况", svc.GetActivityInfo)
	svc.svc.RegisterFunc("参与铸剑", svc.CastSwordActivation)
	svc.svc.RegisterFunc("获取挑战奖励", svc.DrawWeeklyChallengeReward)
	svc.svc.RegisterFunc("兑换物品", svc.WeeklyExchange)
	svc.svc.RegisterFunc("购买礼包", svc.WeeklyGift)
	svc.svc.RegisterFunc("获取自己排行", svc.WeeklySelfRankingId)
	svc.svc.RegisterFunc("获取排行列表", svc.WeeklyRankingList)
	svc.svc.RegisterFunc("获取位面（公会）挑战积分", svc.GuildChallengeScore)
	svc.svc.RegisterFunc("获取铸剑配置", svc.WeeklyCastswordCnf)
	svc.svc.RegisterFunc("获取挑战配置", svc.WeeklyChallengeCnf)
	svc.svc.RegisterFunc("获取兑换配置", svc.WeeklyExchangeCnf)
	svc.svc.RegisterFunc("获取礼包配置", svc.WeeklyGiftCnf)
	svc.svc.RegisterFunc("获取排行榜配置", svc.WeeklyRankCnf)

	svc.svc.RegisterFunc("GM刷新周活动", svc.CheatWeeklyRefresh)
	svc.svc.RegisterFunc("GM购买礼包", svc.CheatBuyWeeklyGift)
	eventlocal.SubscribeEventLocal(svc.HandleRoleLoginEvent)
	eventlocal.SubscribeEventLocal(svc.HandlePaySuccess)
	eventlocal.SubscribeEventLocal(svc.HandleGuildEvent)
}

// msg proc ============================================================================================================================================================================================================

func (svc *Service) getActivityWeeklyData(c *ctx.Context, roleId values.RoleId) (*daopb.ActivityWeeklyData, *errmsg.ErrMsg) {
	data, err := dao.GetActivityWeeklyData(c, roleId)
	if err != nil {
		return nil, err
	}

	if data.ActivityWeeklyData == nil {
		data.ActivityWeeklyData = make(map[int64]*modelspb.ActivityWeekly)
	}

	// 1、检查并初始化新活动
	needSave1, err := svc.checkStart(c, data)
	if err != nil {
		return nil, err
	}
	// 2、刷新活动
	needSave2, err := svc.refreshAll(c, data)
	if err != nil {
		return nil, err
	}
	// 3、检查并处理已结束活动
	needSave3, err := svc.checkEnd(c, data)
	if err != nil {
		return nil, err
	}
	if needSave1 || needSave2 || needSave3 {
		dao.SaveActivityWeeklyData(c, data)
	}

	return data, nil
}

func (svc *Service) checkStart(c *ctx.Context, data *daopb.ActivityWeeklyData) (bool, *errmsg.ErrMsg) {
	var needSave bool
	activityIds := map[ActivityId.Enum]struct{}{
		ActivityId.WeeklyCastSword: {},
	}

	var activityId int64
	list := rule.MustGetReader(c).ActivityCircular.List()
	for _, cfg := range list {
		activityId = cfg.ActivityId
		if _, ok := activityIds[activityId]; !ok {
			continue
		}

		if cfg.TimeType != 2 { // 如果不是绝对时间，配置错误不处理
			c.Error("activity config error", zap.Int64("activity_id", activityId))
			continue
		}

		start, err := ParseTime(cfg.ActivityOpenTime)
		if err != nil {
			c.Error("activity config ActivityOpenTime error", zap.Int64("activity_id", activityId), zap.Error(err))
			continue
		}
		if timer.Now().Before(start) { // 活动未开始
			c.Info("activity not end", zap.Int64("activity_id", activityId))
			continue
		}
		end, err := ParseTime(cfg.DurationTime)
		if err != nil {
			c.Error("activity config DurationTime error", zap.Int64("activity_id", activityId), zap.Error(err))
			continue
		}
		if timer.Now().After(end) { // 活动已结束
			c.Info("activity end", zap.Int64("activity_id", activityId))
			continue
		}

		if aw, ok := data.ActivityWeeklyData[activityId]; ok {
			if aw.Version != cfg.ActivityVersion { // 开新活动了
				err2 := svc.endActivity(c, aw)
				if err2 != nil {
					return false, err2
				}
				delete(data.ActivityWeeklyData, activityId)
			} else { // 如果已存在，读最新的结束时间
				if aw.IsFinished {
					continue
				}
				aw.EndTime = end.UnixMilli()
				needSave = true
				continue
			}
		}

		aw := &modelspb.ActivityWeekly{
			Version:         cfg.ActivityVersion,
			ActivityId:      activityId,
			StartTime:       start.UnixMilli(),
			EndTime:         end.UnixMilli(),
			NextRefreshTime: util.DefaultNextRefreshTime().UnixMilli(),
			Type:            modelspb.ActivityWeeklyType_AWT_CastSword,
			Score:           0,
			FreeTimes:       rule2.MustGetCastswordFreeTimes(c),
			ChallengeInfo: &modelspb.WeeklyChallenge{
				Rewards:      map[int64]modelspb.RewardStatus{},
				GuildRewards: map[int64]modelspb.RewardStatus{},
			},
			ExchangeInfo: &modelspb.WeeklyExchange{ExchangeTimes: map[int64]int64{}},
			GiftInfo:     &modelspb.WeeklyGift{BuyTimes: map[int64]int64{}},
			RankingInfo:  &modelspb.WeeklyRanking{RankingIndex: ""},
		}
		err2 := svc.refreshChallengeInfo(c, c.RoleId, aw)
		if err2 != nil {
			return needSave, err2
		}
		svc.startRanking(c, aw)
		data.ActivityWeeklyData[activityId] = aw
		needSave = true
	}
	return needSave, nil
}

func (svc *Service) refreshAll(c *ctx.Context, data *daopb.ActivityWeeklyData) (needSave bool, err *errmsg.ErrMsg) {
	for _, aw := range data.ActivityWeeklyData {
		if aw.IsFinished {
			continue
		}
		// 初始化
		if aw.ChallengeInfo.Rewards == nil {
			aw.ChallengeInfo.Rewards = map[int64]modelspb.RewardStatus{}
		}
		if aw.ChallengeInfo.GuildRewards == nil {
			aw.ChallengeInfo.GuildRewards = map[int64]modelspb.RewardStatus{}
		}
		if aw.ExchangeInfo.ExchangeTimes == nil {
			aw.ExchangeInfo.ExchangeTimes = map[int64]int64{}
		}
		if aw.GiftInfo.BuyTimes == nil {
			aw.GiftInfo.BuyTimes = map[int64]int64{}
		}

		if aw.NextRefreshTime > timer.Now().UnixMilli() {
			continue
		}
		err = svc.refresh(c, aw)
		if err != nil {
			return
		}
		needSave = true
	}
	return
}

func (svc *Service) refresh(c *ctx.Context, aw *modelspb.ActivityWeekly) *errmsg.ErrMsg {
	switch aw.Type {
	case modelspb.ActivityWeeklyType_AWT_CastSword:
		svc.refreshCastSwordInfo(c, aw)
	}

	err := svc.refreshChallengeInfo(c, c.RoleId, aw)
	if err != nil {
		return err
	}
	svc.refreshExchangeInfo(c, aw)
	svc.refreshGiftInfo(c, aw)

	aw.NextRefreshTime = util.DefaultNextRefreshTime().UnixMilli()
	return nil
}

func (svc *Service) checkEnd(c *ctx.Context, data *daopb.ActivityWeeklyData) (needSave bool, err *errmsg.ErrMsg) {
	for _, aw := range data.ActivityWeeklyData {
		if aw.EndTime > timer.Now().UnixMilli() {
			continue
		}
		err = svc.endActivity(c, aw)
		if err != nil {
			return
		}
		needSave = true
	}
	return
}

func (svc *Service) endActivity(c *ctx.Context, aw *modelspb.ActivityWeekly) (err *errmsg.ErrMsg) {
	if aw.IsFinished {
		return nil
	}
	err = svc.EndCastSword(c)
	if err != nil {
		return
	}
	err = svc.EndChallenge(c, aw)
	if err != nil {
		return
	}
	err = svc.EndRanking(c, aw)
	if err != nil {
		return
	}
	aw.IsFinished = true
	return nil
}

func (svc *Service) GetActivityInfo(c *ctx.Context, req *servicepb.ActivityWeekly_GetActivityInfoRequest) (*servicepb.ActivityWeekly_GetActivityInfoResponse, *errmsg.ErrMsg) {
	data, err := svc.getActivityWeeklyData(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	info, ok := data.ActivityWeeklyData[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	return &servicepb.ActivityWeekly_GetActivityInfoResponse{
		Info: info,
	}, nil
}

func (svc *Service) CastSwordActivation(c *ctx.Context, req *servicepb.ActivityWeekly_CastSwordActivationRequest) (*servicepb.ActivityWeekly_CastSwordActivationResponse, *errmsg.ErrMsg) {
	if req.Times < 0 || req.Times > 10 {
		return nil, errmsg.NewErrActivityWeeklyTimes()
	}

	data, err := svc.getActivityWeeklyData(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	info, ok := data.ActivityWeeklyData[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	rewards, err := svc.CastSword(c, info, req.Times)
	if err != nil {
		return nil, err
	}

	_, err = svc.updateChallenge(c, info, req.Times)
	if err != nil {
		return nil, err
	}
	_, err = svc.updateRanking(c, info)
	if err != nil {
		return nil, err
	}

	dao.SaveActivityWeeklyData(c, data)

	return &servicepb.ActivityWeekly_CastSwordActivationResponse{
		Info:    info,
		Rewards: rewards,
	}, nil
}

//DrawWeeklyChallengeReward 领取挑战奖励 一键领取所有
func (svc *Service) DrawWeeklyChallengeReward(c *ctx.Context, req *servicepb.ActivityWeekly_WeeklyChallengeReceiveRequest) (*servicepb.ActivityWeekly_WeeklyChallengeReceiveResponse, *errmsg.ErrMsg) {
	data, err := svc.getActivityWeeklyData(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	aw, ok := data.ActivityWeeklyData[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	items, err := svc.DrawChallengeReward(c, aw, req.Id)
	if err != nil {
		return nil, err
	}

	dao.SaveActivityWeeklyData(c, data)
	return &servicepb.ActivityWeekly_WeeklyChallengeReceiveResponse{
		Info:  aw,
		Items: items,
	}, nil
}

func (svc *Service) WeeklyExchange(c *ctx.Context, req *servicepb.ActivityWeekly_WeeklyExchangeRequest) (*servicepb.ActivityWeekly_WeeklyExchangeResponse, *errmsg.ErrMsg) {
	data, err := svc.getActivityWeeklyData(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	aw, ok := data.ActivityWeeklyData[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	items, err := svc.ExchangeItems(c, aw, req.Id)
	if err != nil {
		return nil, err
	}

	dao.SaveActivityWeeklyData(c, data)
	return &servicepb.ActivityWeekly_WeeklyExchangeResponse{
		Info:  aw,
		Items: items,
	}, nil
}

func (svc *Service) WeeklyGift(c *ctx.Context, req *servicepb.ActivityWeekly_WeeklyGiftRequest) (*servicepb.ActivityWeekly_WeeklyGiftResponse, *errmsg.ErrMsg) {
	data, err := svc.getActivityWeeklyData(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	aw, ok := data.ActivityWeeklyData[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	err = svc.BuyGiftItems(c, aw, req.Id)
	if err != nil {
		return nil, err
	}

	dao.SaveActivityWeeklyData(c, data)
	return &servicepb.ActivityWeekly_WeeklyGiftResponse{
		Info: aw,
	}, nil
}

func (svc *Service) BuyGiftItemsByCash(c *ctx.Context, pcId int64) *errmsg.ErrMsg {
	data, err := svc.getActivityWeeklyData(c, c.RoleId)
	if err != nil {
		return err
	}

	for _, aw := range data.ActivityWeeklyData {
		for _, cfg := range rule.MustGetReader(c).ActivityWeeklyGift.List() {
			if cfg.ActivityId != aw.ActivityId || cfg.BuyType != 2 {
				continue
			}
			if cfg.PayId[0] != pcId {
				continue
			}
			err = svc.buyGiftItemsByCash(c, aw, cfg.Id)
			if err != nil {
				return err
			}
			dao.SaveActivityWeeklyData(c, data)
			return nil
		}
	}

	return nil
}

func (svc *Service) WeeklySelfRankingId(c *ctx.Context, req *servicepb.ActivityWeekly_WeeklySelfRankingIdRequest) (*servicepb.ActivityWeekly_WeeklySelfRankingIdResponse, *errmsg.ErrMsg) {
	data, err := svc.getActivityWeeklyData(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	aw, ok := data.ActivityWeeklyData[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	rankingId, err := svc.GetRankingSelf(c, aw)
	if err != nil {
		return nil, err
	}

	return &servicepb.ActivityWeekly_WeeklySelfRankingIdResponse{
		SelfRankingId: rankingId,
	}, nil
}

func (svc *Service) WeeklyRankingList(c *ctx.Context, req *servicepb.ActivityWeekly_WeeklyRankingListRequest) (*servicepb.ActivityWeekly_WeeklyRankingListResponse, *errmsg.ErrMsg) {
	if req.StartIndex < 0 {
		return nil, errmsg.NewErrActivityWeeklyParam()
	}

	if req.Count < 0 {
		return nil, errmsg.NewErrActivityWeeklyParam()
	}

	// 防止性能问题，限制一次性最多获取10个
	if req.Count > 10 {
		return nil, errmsg.NewErrActivityWeeklyParam()
	}

	data, err := svc.getActivityWeeklyData(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	aw, ok := data.ActivityWeeklyData[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	rankingId, roleInfos, nextRefreshTime, err := svc.GetRankingList(c, aw, req.StartIndex, req.Count)
	if err != nil {
		return nil, err
	}

	return &servicepb.ActivityWeekly_WeeklyRankingListResponse{
		SelfRankingId:   rankingId,
		RankingInfos:    roleInfos,
		NextRefreshTime: nextRefreshTime,
	}, nil
}

func (svc *Service) GuildChallengeScore(c *ctx.Context, req *servicepb.ActivityWeekly_GetGuildChallengeScoreRequest) (*servicepb.ActivityWeekly_GetGuildChallengeScoreResponse, *errmsg.ErrMsg) {
	gci, err := dao.GetGuildChallengeInfo(c, req.ActivityId, req.Version, req.GuildId)
	if err != nil {
		return nil, err
	}

	return &servicepb.ActivityWeekly_GetGuildChallengeScoreResponse{
		Score: gci.Score,
	}, nil
}

func (svc *Service) WeeklyCastswordCnf(c *ctx.Context, _ *servicepb.ActivityWeekly_WeeklyCastswordCnfRequest) (*servicepb.ActivityWeekly_WeeklyCastswordCnfResponse, *errmsg.ErrMsg) {
	return &servicepb.ActivityWeekly_WeeklyCastswordCnfResponse{
		Cnf:           svc.GetCastSwordConfigs(c),
		CastswordDraw: rule2.MustGetCastswordDraw(c),
	}, nil
}

func (svc *Service) WeeklyChallengeCnf(c *ctx.Context, req *servicepb.ActivityWeekly_WeeklyChallengeCnfRequest) (*servicepb.ActivityWeekly_WeeklyChallengeCnfResponse, *errmsg.ErrMsg) {
	return &servicepb.ActivityWeekly_WeeklyChallengeCnfResponse{
		Cnf: svc.GetChallengeConfigs(c, req.ActivityId),
	}, nil
}

func (svc *Service) WeeklyExchangeCnf(c *ctx.Context, req *servicepb.ActivityWeekly_WeeklyExchangeCnfRequest) (*servicepb.ActivityWeekly_WeeklyExchangeCnfResponse, *errmsg.ErrMsg) {
	return &servicepb.ActivityWeekly_WeeklyExchangeCnfResponse{
		Cnf: svc.GetExchangeConfigs(c, req.ActivityId),
	}, nil
}

func (svc *Service) WeeklyGiftCnf(c *ctx.Context, req *servicepb.ActivityWeekly_WeeklyGiftCnfRequest) (*servicepb.ActivityWeekly_WeeklyGiftCnfResponse, *errmsg.ErrMsg) {
	return &servicepb.ActivityWeekly_WeeklyGiftCnfResponse{
		Cnf: svc.GetGiftConfigs(c, req.ActivityId),
	}, nil
}

func (svc *Service) WeeklyRankCnf(c *ctx.Context, req *servicepb.ActivityWeekly_WeeklyRankCnfRequest) (*servicepb.ActivityWeekly_WeeklyRankCnfResponse, *errmsg.ErrMsg) {
	return &servicepb.ActivityWeekly_WeeklyRankCnfResponse{
		Cnf: svc.GetRankingConfigs(c, req.ActivityId),
	}, nil
}

//ConvertActivityItem 活动结束时 积分自动兑换奖励
func (svc *Service) ConvertActivityItem(ctx *ctx.Context, key string) *errmsg.ErrMsg {
	convertParam, ok := rule.MustGetReader(ctx).KeyValue.GetIntegerArray(key)
	if !ok {
		ctx.Error("ConvertActivityItem load error", zap.String("key", key))
		panic(fmt.Sprintf("ConvertActivityItem error load error  key %s", key))
	}
	if len(convertParam) < 4 {
		ctx.Error("ConvertActivityItem len error", zap.String("key", key), zap.Int("len", len(convertParam)))
		panic(fmt.Sprintf("ConvertActivityItem error len error  key %s  len %d", key, len(convertParam)))
	}
	mailId := convertParam[0]
	itemIdA := convertParam[1]
	itemIdB := convertParam[2]
	scale := convertParam[3]

	count, err := svc.BagService.GetItem(ctx, ctx.RoleId, itemIdA)
	if count <= 0 {
		return nil
	}

	cost := make(map[int64]int64)
	cost[itemIdA] = count
	err = svc.BagService.SubManyItem(ctx, ctx.RoleId, cost)
	if err != nil {
		ctx.Error("ConvertActivityItem SubManyItem error", zap.Any("err", err), zap.Any("cost", cost))
		return err
	}
	var convertItems []*modelspb.Item
	convertItems = append(convertItems, &modelspb.Item{
		ItemId: itemIdB,
		Count:  scale * count,
	})

	if err = svc.MailService.Add(ctx, ctx.RoleId, &modelspb.Mail{
		Type:       modelspb.MailType_MailTypeSystem,
		TextId:     mailId,
		Attachment: convertItems,
		Args:       []string{strconv.Itoa(int(itemIdA)), strconv.Itoa(int(itemIdB)), strconv.Itoa(int(scale))},
	}); err != nil {
		ctx.Error("send mail error", zap.Error(err), zap.Any("convert_items", convertItems), zap.String("role_id", ctx.RoleId), zap.String("key", key))
		return err
	}
	return nil
}

func GetActivityWeeklyRefreshTime(ctx *ctx.Context) int64 {
	ActivityWeeklyRankRefreshTime, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("ActivityWeeklyRankRefreshTime")
	if !ok {
		ctx.Error("GetActivityWeeklyRefreshTime ActivityWeeklyRankRefreshTime error")
		panic("GetActivityWeeklyRefreshTime ActivityWeeklyRankRefreshTime error")
	}
	return ActivityWeeklyRankRefreshTime
}

func ParseTime(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", value, time.UTC)
}

// ----------- 作弊器 -----------

// CheatWeeklyRefresh 重置活动
func (svc *Service) CheatWeeklyRefresh(c *ctx.Context, req *servicepb.ActivityWeekly_CheatWeeklyRefreshRequest) (*servicepb.ActivityWeekly_CheatWeeklyRefreshResponse, *errmsg.ErrMsg) {
	data, err := dao.GetActivityWeeklyData(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	info, ok := data.ActivityWeeklyData[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	info.NextRefreshTime = util.DefaultNextRefreshTime().AddDate(0, 0, -1).UnixMilli()
	err = svc.refresh(c, info)
	if err != nil {
		return nil, err
	}

	dao.SaveActivityWeeklyData(c, data)
	return &servicepb.ActivityWeekly_CheatWeeklyRefreshResponse{Info: info}, nil
}

// CheatBuyWeeklyGift 作弊购买礼包
func (svc *Service) CheatBuyWeeklyGift(c *ctx.Context, req *servicepb.ActivityWeekly_CheatBuyWeeklyGiftRequest) (*servicepb.ActivityWeekly_CheatBuyWeeklyGiftResponse, *errmsg.ErrMsg) {
	data, err := dao.GetActivityWeeklyData(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	aw, ok := data.ActivityWeeklyData[req.ActivityId]
	if !ok {
		return nil, errmsg.NewErrActivityWeeklyNoActivity()
	}

	err = svc.buyGiftItemsByCash(c, aw, req.Id)
	if err != nil {
		return nil, err
	}
	dao.SaveActivityWeeklyData(c, data)

	return &servicepb.ActivityWeekly_CheatBuyWeeklyGiftResponse{}, nil
}
