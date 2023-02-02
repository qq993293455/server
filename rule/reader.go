package rule

import (
	"coin-server/common/ctx"
	rule_model "coin-server/rule/rule-model"
)

func GetReader(ctx *ctx.Context) (*rule_model.Reader, error) {
	reader := rule_model.GetReader()
	// TODO 待ServerHeader完善
	//var version string
	//if ctx.ServerHeader != nil {
	//	version = ctx.ServerHeader.RuleVersion
	//}
	//if version == "" || version != r.version {
	//	return nil, fmt.Errorf("版本号错误")
	//}
	return reader, nil
}

func MustGetReader(ctx *ctx.Context) *rule_model.Reader {
	reader := rule_model.GetReader()
	// TODO 待ServerHeader完善
	//var version string
	//if ctx.ServerHeader != nil {
	//	version = ctx.ServerHeader.RuleVersion
	//}
	//if version == "" || version != reader.Version {
	//	panic(fmt.Errorf("版本号错误"))
	//}
	return reader
}
