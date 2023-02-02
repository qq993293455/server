package activity_weekly

import (
	"math/rand"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	modelspb "coin-server/common/proto/models"
	"coin-server/common/timer"
	"coin-server/common/utils/generic/maps"
	"coin-server/common/values"
	rule2 "coin-server/game-server/service/activity-weekly/rule"
	"coin-server/game-server/util/trans"
	"coin-server/rule"
)

func (svc *Service) GetCastSwordConfigs(ctx *ctx.Context) []*modelspb.WeeklyCastswordCnf {
	var ret []*modelspb.WeeklyCastswordCnf
	cfgs := rule.MustGetReader(ctx).ActivityWeeklyCastsword.List()
	for _, cfg := range cfgs {
		ret = append(ret, &modelspb.WeeklyCastswordCnf{
			Id:           cfg.Id,
			ActivityId:   cfg.ActivityId,
			Sequence:     cfg.Sequence,
			ActivityItem: cfg.ActivityItem,
			Reward:       cfg.Reward,
		})
	}
	return ret
}

func (svc *Service) refreshCastSwordInfo(c *ctx.Context, aw *modelspb.ActivityWeekly) {
	if aw.NextRefreshTime > timer.Now().UnixMilli() {
		return
	}
	aw.FreeTimes = rule2.MustGetCastswordFreeTimes(c)
}

//CastSword 拔剑
func (svc *Service) CastSword(c *ctx.Context, aw *modelspb.ActivityWeekly, times int64) ([]*modelspb.WeeklyItems, *errmsg.ErrMsg) {
	ret := make([]*modelspb.WeeklyItems, 0, times)
	rewards := make(map[values.Integer]values.Integer, times)
	costs := make(map[values.Integer]values.Integer, times)

	for i := int64(0); i < times; i++ {
		reward, cost := svc.CastSwordOnce(c, aw, times == 1)
		maps.Merge(rewards, trans.ItemProtoToMap(reward.Items))
		maps.Merge(costs, cost)
		ret = append(ret, reward)
	}

	err := svc.BagService.SubManyItem(c, c.RoleId, costs)
	if err != nil {
		return nil, err
	}
	_, err = svc.BagService.AddManyItem(c, c.RoleId, rewards)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

//CastSwordOnce 拔剑一次
func (svc *Service) CastSwordOnce(c *ctx.Context, aw *modelspb.ActivityWeekly, isSingle bool) (ret *modelspb.WeeklyItems, cost map[values.Integer]values.Integer) {
	cfgs := rule.MustGetReader(c).ActivityWeeklyCastsword.List()
	idx := int(aw.Score) % len(cfgs)
	cfg := cfgs[idx]
	ret = new(modelspb.WeeklyItems)

	if isSingle && aw.FreeTimes > 0 { // 如果是单抽 才优先用免费次数
		aw.FreeTimes--
	} else {
		cost = trans.ItemSliceToMap(cfg.ActivityItem)
	}

	aw.Score++
	prob, doubleTimes, guaranteeDouble := rule2.MustGetCastswordMultiplier(c)
	if aw.Score != 0 && aw.Score%guaranteeDouble == 0 {
		ret.IsDouble = true
	}
	if !ret.IsDouble {
		rn := rand.Int63n(1000)
		if rn < prob {
			ret.IsDouble = true
		}
	}
	if ret.IsDouble {
		ret.Items = trans.ItemSliceToPbMulti(cfg.Reward, doubleTimes)
	} else {
		ret.Items = trans.ItemSliceToPb(cfg.Reward)
	}

	return ret, cost
}

func (svc *Service) EndCastSword(c *ctx.Context) *errmsg.ErrMsg {
	return svc.ConvertActivityItem(c, "CastswordRecycling")
}
