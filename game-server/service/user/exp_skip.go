package user

import (
	"math"

	modelspb "coin-server/common/proto/models"
	"coin-server/common/values/enum"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/service/user/db"
	"coin-server/game-server/service/user/rule"
)

func (this_ *Service) GetUseHangUpExpItemInfo(ctx *ctx.Context, _ *lessservicepb.User_GetUseHangUpExpItemInfoRequest) (*lessservicepb.User_GetUseHangUpExpItemInfoResponse, *errmsg.ErrMsg) {
	es, ok, err := db.GetExpSkip(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		es = &dao.ExpSkip{
			RoleId:       ctx.RoleId,
			ResetTime:    timer.NextDay(this_.GetCurrDayFreshTime(ctx)).Unix(),
			UseCount:     0,
			RateDuration: map[int64]int64{},
		}
	}
	if this_.handleReset(ctx, es) {
		db.SaveExpSkip(ctx, es)
	}
	var totalDuration values.Integer
	for _, d := range es.RateDuration {
		totalDuration += d
	}
	return &lessservicepb.User_GetUseHangUpExpItemInfoResponse{
		TotalDuration: totalDuration,
		Rate:          values.Integer(rule.GetExpSkipEffcient(ctx, es.UseCount).Rate),
		RateDuration:  es.RateDuration,
		UseCount:      es.UseCount,
	}, nil
}

func (this_ *Service) UseHangUpExpItem(ctx *ctx.Context, req *lessservicepb.User_UseHangUpExpItemRequest) (*lessservicepb.User_UseHangUpExpItemResponse, *errmsg.ErrMsg) {
	cfg, ok := rule.GetExpSkipCfg(ctx, req.ItemId)
	if !ok {
		return nil, errmsg.NewInternalErr("exp skip not found")
	}

	es, ok, err := db.GetExpSkip(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		es = &dao.ExpSkip{
			RoleId:       ctx.RoleId,
			ResetTime:    timer.NextDay(this_.GetCurrDayFreshTime(ctx)).Unix(),
			UseCount:     0,
			RateDuration: map[int64]int64{},
		}
	}
	this_.handleReset(ctx, es)
	if es.UseCount > rule.GetExpSkipMaxCount(ctx) {
		return nil, errmsg.NewErrExpSkipUseItemMax()
	}
	es.UseCount++
	rateCfg := rule.GetExpSkipEffcient(ctx, es.UseCount)
	if rateCfg == nil {
		return nil, errmsg.NewInternalErr("exp skip rate is nil")
	}
	if err := this_.SubItem(ctx, ctx.RoleId, req.ItemId, 1); err != nil {
		return nil, err
	}
	baseExp, err := this_.TempBagExpProfit(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	exp := values.Integer(math.Ceil(baseExp * values.Float(cfg.Duration*60) * rateCfg.Rate / 10000))
	// 经验走自动升级流程不再直接往背包里加
	expData := map[values.ItemId]values.Integer{enum.RoleExp: exp}
	//if exp > 0 {
	//	if _, err := this_.AddManyItem(ctx, ctx.RoleId, expData); err != nil {
	//		return nil, err
	//	}
	//}
	if err := this_.AddExp(ctx, ctx.RoleId, exp, false); err != nil {
		return nil, err
	}
	es.RateDuration[values.Integer(rateCfg.Rate)] += cfg.Duration
	var totalDuration values.Integer
	for _, d := range es.RateDuration {
		totalDuration += d
	}

	rateCfg = rule.GetExpSkipEffcient(ctx, es.UseCount+1)
	if rateCfg == nil {
		return nil, errmsg.NewInternalErr("exp skip rate is nil")
	}
	db.SaveExpSkip(ctx, es)
	this_.UpdateTarget(ctx, ctx.RoleId, modelspb.TaskType_TaskHangUpExp, 0, 1)
	return &lessservicepb.User_UseHangUpExpItemResponse{
		TotalDuration: totalDuration,
		Rate:          values.Integer(rateCfg.Rate),
		RateDuration:  es.RateDuration,
		UseCount:      es.UseCount,
		Exp:           expData,
	}, nil
}

func (this_ *Service) handleReset(ctx *ctx.Context, es *dao.ExpSkip) bool {
	if es.ResetTime > timer.StartTime(ctx.StartTime).Unix() {
		return false
	}

	es.ResetTime = timer.NextDay(this_.GetCurrDayFreshTime(ctx)).Unix()
	es.UseCount = 0
	es.RateDuration = map[int64]int64{}
	return true
}

func (this_ *Service) CheatResetExpSkip(ctx *ctx.Context, _ *lessservicepb.User_CheatResetExpSkipRequest) (*lessservicepb.User_CheatResetExpSkipResponse, *errmsg.ErrMsg) {
	es, _, err := db.GetExpSkip(ctx)
	if err != nil {
		return nil, err
	}
	es = &dao.ExpSkip{
		RoleId:       ctx.RoleId,
		ResetTime:    timer.NextDay(this_.GetCurrDayFreshTime(ctx)).Unix(),
		UseCount:     0,
		RateDuration: map[int64]int64{},
	}
	db.SaveExpSkip(ctx, es)
	return &lessservicepb.User_CheatResetExpSkipResponse{
		TotalDuration: 0,
		Rate:          values.Integer(rule.GetExpSkipEffcient(ctx, es.UseCount).Rate),
		RateDuration:  es.RateDuration,
		UseCount:      es.UseCount,
	}, nil
}
