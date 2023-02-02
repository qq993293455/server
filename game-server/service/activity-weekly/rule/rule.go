package rule

import (
	"errors"

	"coin-server/common/ctx"
	"coin-server/rule"
)

func MustGetCastswordFreeTimes(c *ctx.Context) int64 {
	freeTimes, ok := rule.MustGetReader(c).KeyValue.GetInt64("CastswordFreeTimes")
	if !ok {
		panic(errors.New("KeyValue CastswordFreeTimes not found"))
	}
	return freeTimes
}

func MustGetCastswordMultiplier(ctx *ctx.Context) (int64, int64, int64) {
	CastswordMultiplierProbability, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("CastswordMultiplierProbability")
	if !ok {
		ctx.Error("KeyValue CastswordMultiplierProbability not found")
		panic("KeyValue CastswordMultiplierProbability not found")
	}
	CastswordMultiplier, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("CastswordMultiplier")
	if !ok {
		ctx.Error("KeyValue CastswordMultiplier not found")
		panic("KeyValue CastswordMultiplier not found")
	}
	CastswordMultiplierGuarantee, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("CastswordMultiplierGuarantee")
	if !ok {
		ctx.Error("KeyValue CastswordMultiplierGuarantee not found")
		panic("KeyValue CastswordMultiplierGuarantee not found")
	}
	return CastswordMultiplierProbability, CastswordMultiplier, CastswordMultiplierGuarantee
}

func MustGetCastswordDraw(ctx *ctx.Context) string {
	CastswordDraw, ok := rule.MustGetReader(ctx).KeyValue.GetString("CastswordDraw")
	if !ok {
		ctx.Error("KeyValue CastswordDraw not found")
		panic("KeyValue CastswordDraw not found")
	}
	return CastswordDraw
}
