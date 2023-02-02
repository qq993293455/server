package rule

import (
	"fmt"
	"strconv"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/proto/models"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"

	"go.uber.org/zap"
)

func GetEventById(ctx *ctx.Context, id values.Integer) (*rulemodel.Activity, bool) {
	cfg, ok := rule.MustGetReader(ctx).Activity.GetActivityById(id)
	return cfg, ok
}

func GetLevelGrowthFund(ctx *ctx.Context) []rulemodel.ActivityGrowthfund {
	return rule.MustGetReader(ctx).ActivityGrowthfund.List()
}

func GetLevelGrowthFundById(ctx *ctx.Context, id values.Integer) (*rulemodel.ActivityGrowthfund, bool) {
	cfg, ok := rule.MustGetReader(ctx).ActivityGrowthfund.GetActivityGrowthfundById(id)
	return cfg, ok
}

func GetLevelGrowthFundLen(ctx *ctx.Context) int {
	return rule.MustGetReader(ctx).ActivityGrowthfund.Len()
}

func GetAllAvailableActivity(ctx *ctx.Context, createdAt values.TimeStamp) ([]*models.Activity, []*models.Activity) {
	list := rule.MustGetReader(ctx).Activity.List()
	now := timer.StartTime(ctx.StartTime).UnixMilli()
	activated := make([]*models.Activity, 0)
	inactive := make([]*models.Activity, 0)
	registerTime := time.Unix(createdAt, 0).UTC()
	for i := 0; i < len(list); i++ {
		item := &list[i]
		if item.TimeType == enum.RoleCreated {
			openTime, err := strconv.Atoi(item.ActivityOpenTime)
			if err != nil {
				ctx.Error("activity ActivityOpenTime strconv.Atoi err", zap.Int64("id", item.Id), zap.Error(err))
			}
			// 这个时间是需要包含当天的
			if openTime >= 86400 {
				openTime -= 86400
			}
			start := registerTime.Add(time.Duration(openTime) * time.Second)
			duration, err := strconv.Atoi(item.DurationTime)
			if err != nil {
				ctx.Error("activity DurationTime strconv.Atoi err", zap.Int64("id", item.Id), zap.Error(err))
				continue
			}
			end := values.Integer(-1)
			if duration > 0 {
				// end = time.Unix(0, values.Integer(time.Duration(createdAt)*time.Millisecond)).Add(time.Duration(duration) * time.Second).UnixMilli()
				end = start.Add(time.Duration(duration) * time.Second).UnixMilli()
			}
			if item.Id == enum.ZeroPurchase {
				activated = append(activated, &models.Activity{
					Id:                         item.Id,
					Sort:                       item.ActivitySort,
					ActivityDescribeLanguageId: item.ActivityDescribeLanguageId,
					Begin:                      start.UnixMilli(),
					End:                        end,
					SystemId:                   item.SystemId,
					ChargeId:                   item.ChargeId,
				})
				continue
			}
			if start.UnixMilli() > now && (end == -1 || end >= now) {
				inactive = append(inactive, &models.Activity{
					Id:                         item.Id,
					Sort:                       item.ActivitySort,
					ActivityDescribeLanguageId: item.ActivityDescribeLanguageId,
					Begin:                      start.UnixMilli(),
					End:                        end,
					SystemId:                   item.SystemId,
					ChargeId:                   item.ChargeId,
				})
				continue
			}
			if end == -1 || end >= now {
				activated = append(activated, &models.Activity{
					Id:                         item.Id,
					Sort:                       item.ActivitySort,
					ActivityDescribeLanguageId: item.ActivityDescribeLanguageId,
					Begin:                      start.UnixMilli(),
					End:                        end,
					SystemId:                   item.SystemId,
					ChargeId:                   item.ChargeId,
				})
			}
		} else if item.TimeType == enum.AbsoluteTime {
			begin, err := time.ParseInLocation("2006-01-02 15:04:05", item.ActivityOpenTime, time.UTC)
			if err != nil {
				ctx.Error("activity ActivityOpenTime time.ParseInLocation err", zap.Int64("id", item.Id), zap.Error(err))
				continue
			}
			end, err := time.ParseInLocation("2006-01-02 15:04:05", item.DurationTime, time.UTC)
			if err != nil {
				ctx.Error("activity DurationTime time.ParseInLocation err", zap.Int64("id", item.Id), zap.Error(err))
				continue
			}
			if end.UnixMilli() <= now {
				continue
			}
			if begin.UnixMilli() <= now {
				activated = append(activated, &models.Activity{
					Id:                         item.Id,
					Sort:                       item.ActivitySort,
					ActivityDescribeLanguageId: item.ActivityDescribeLanguageId,
					Begin:                      begin.UnixMilli(),
					End:                        end.UnixMilli(),
					SystemId:                   item.SystemId,
					ChargeId:                   item.ChargeId,
				})
			}
			if begin.UnixMilli() > now {
				inactive = append(inactive, &models.Activity{
					Id:                         item.Id,
					Sort:                       item.ActivitySort,
					ActivityDescribeLanguageId: item.ActivityDescribeLanguageId,
					Begin:                      begin.UnixMilli(),
					End:                        end.UnixMilli(),
					SystemId:                   item.SystemId,
					ChargeId:                   item.ChargeId,
				})
			}
		}
	}

	list2 := rule.MustGetReader(ctx).ActivityCircular.List()
	for _, item := range list2 {
		if item.TimeType != enum.AbsoluteTime {
			continue
		}
		begin, err := time.ParseInLocation("2006-01-02 15:04:05", item.ActivityOpenTime, time.UTC)
		if err != nil {
			panic(err)
		}
		end, err := time.ParseInLocation("2006-01-02 15:04:05", item.DurationTime, time.UTC)
		if err != nil {
			panic(err)
		}
		if end.UnixMilli() <= now {
			continue
		}
		if begin.UnixMilli() <= now {
			activated = append(activated, &models.Activity{
				Id:                         item.ActivityId,
				Sort:                       item.ActivitySort,
				ActivityDescribeLanguageId: item.ActivityDescribeLanguageId,
				Begin:                      begin.UnixMilli(),
				End:                        end.UnixMilli(),
				SystemId:                   item.SystemId,
				ChargeId:                   item.ChargeId,
			})
		}
		if begin.UnixMilli() > now {
			inactive = append(inactive, &models.Activity{
				Id:                         item.ActivityId,
				Sort:                       item.ActivitySort,
				ActivityDescribeLanguageId: item.ActivityDescribeLanguageId,
				Begin:                      begin.UnixMilli(),
				End:                        end.UnixMilli(),
				SystemId:                   item.SystemId,
				ChargeId:                   item.ChargeId,
			})
		}
	}
	return activated, inactive
}

func GetFirstPay(ctx *ctx.Context) []rulemodel.ActivityFirstpay {
	return rule.MustGetReader(ctx).ActivityFirstpay.List()
}

func GetFirstPayById(ctx *ctx.Context, id values.Integer) (*rulemodel.ActivityFirstpay, bool) {
	cfg, ok := rule.MustGetReader(ctx).ActivityFirstpay.GetActivityFirstpayById(id)
	return cfg, ok
}

func MustGetActivityPassesAwards(c *ctx.Context, id values.Integer) *rulemodel.ActivityPassesAwards {
	cfg, ok := rule.MustGetReader(c).ActivityPassesAwards.GetActivityPassesAwardsById(id)
	if !ok {
		panic(fmt.Sprintf("ActivityPassesAwards config not found: %d", id))
	}
	return cfg
}

func GetRebateReward(ctx *ctx.Context) (map[values.ItemId]values.Integer, bool) {
	cfg, ok := rule.MustGetReader(ctx).KeyValue.GetMapInt64Int64("ActivityGrowthfundRebateReward")
	return cfg, ok
}

func GetMailConfigTextId(ctx *ctx.Context, id values.Integer) (*rulemodel.Mail, bool) {
	item, ok := rule.MustGetReader(ctx).Mail.GetMailById(id)
	return item, ok
}

func GetStellargemShopmallByPurchasePrice(ctx *ctx.Context, id values.Integer) (rulemodel.ActivityStellargemShopmall, bool) {
	list := rule.MustGetReader(ctx).ActivityStellargemShopmall.List()
	for _, shopmall := range list {
		if shopmall.PurchasePrice == id {
			return shopmall, true
		}
	}
	return rulemodel.ActivityStellargemShopmall{}, false
}

func GetChargeById(ctx *ctx.Context, id values.Integer) (*rulemodel.Charge, bool) {
	return rule.MustGetReader(ctx).Charge.GetChargeById(id)
}
