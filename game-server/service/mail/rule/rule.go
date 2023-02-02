package rule

import (
	"coin-server/common/ctx"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

func GetMailConfigTextId(ctx *ctx.Context, id values.Integer) (*rulemodel.Mail, bool) {
	item, ok := rule.MustGetReader(ctx).Mail.GetMailById(id)
	return item, ok
}

func GetUpdateReward(ctx *ctx.Context) (map[values.Integer]values.Integer, bool) {
	return rule.MustGetReader(ctx).KeyValue.GetMapInt64Int64("MaintenanceCompensationBonus")
}

func GetMailMax(ctx *ctx.Context) int {
	v, ok := rule.MustGetReader(ctx).KeyValue.GetInt("MailMax")
	if !ok {
		ctx.Error("MailMax not found in keyvalue,use default value: 50")
		v = 50
	}
	return v
}
