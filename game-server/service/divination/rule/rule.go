package rule

import (
	"coin-server/common/ctx"
	"coin-server/common/proto/models"
	"coin-server/common/utils"
	wr "coin-server/common/utils/weightedrand"
	"coin-server/common/values"
	"coin-server/rule"
)

// GetFreeTimes 占卜的每日免费次数（0点重置）
func GetFreeTimes(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("DivinationFreeNum")
	return v
}

// GetMaxUnusedBallNum 最大累积能量球数量
func GetMaxUnusedBallNum(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("DivinationUseNum")
	return v
}

// DrawEnergy 占卜提取能量
func DrawEnergy(ctx *ctx.Context, lv values.Level, num int) ([]*models.EnergyBall, values.Integer) {
	cfg, ok := rule.MustGetReader(ctx).RoleLv.GetRoleLvById(lv)
	utils.MustTrue(ok)

	ret := make([]*models.EnergyBall, 0, num)
	choices := make([]*wr.Choice[int64, int64], 0, len(cfg.DivinationWeight))
	for k, v := range cfg.DivinationWeight {
		choices = append(choices, wr.NewChoice(k, v))
	}
	chooser, _ := wr.NewChooser(choices...)

	var exps int64
	for i := 0; i < num; i++ {
		quality := chooser.Pick()
		exp := quality * cfg.DivinationExp
		ret = append(ret, &models.EnergyBall{
			Quality: quality,
			Exp:     exp,
		})
		exps += exp
	}

	return ret, exps
}
