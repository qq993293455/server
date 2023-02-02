package rule

import (
	"fmt"
	"math/rand"

	"coin-server/common/ctx"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

const CureRitualTargetId = 101 // 救治仪式目标id

func GetRitualTaskCfg(ctx *ctx.Context, parentId, id values.Integer) *rulemodel.TargetTaskStage {
	v, _ := rule.MustGetReader(ctx).TargetTaskStage.GetTargetTaskStageById(parentId, id)
	return v
}

func MustGetRitualTaskCfg(ctx *ctx.Context, parentId, id values.Integer) *rulemodel.TargetTaskStage {
	v, ok := rule.MustGetReader(ctx).TargetTaskStage.GetTargetTaskStageById(parentId, id)
	if !ok {
		panic(fmt.Sprintf("TargetTaskStage config not found: parent_id: %d, id: %d", parentId, id))
	}
	return v
}

func MustGetTargetCfg(ctx *ctx.Context, targetId values.Integer) *rulemodel.MainTaskChapterTarget {
	v, ok := rule.MustGetReader(ctx).MainTaskChapterTarget.GetMainTaskChapterTargetById(targetId)
	if !ok {
		panic(fmt.Sprintf("MainTaskChapterTarget config not found: %d", targetId))
	}
	return v
}

func GetRitualUnlockMainTaskId(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("ChaosRitualUnlockByMainTaskId")
	return v
}

// RandRitualUnlockHero 混沌仪式解锁英雄
func RandRitualUnlockHero(ctx *ctx.Context, prob values.Integer) values.Integer {
	vs, ok := rule.MustGetReader(ctx).KeyValue.GetIntegerArray("ChaosRitualUnlockHero")
	utils.MustTrue(ok && len(vs) == 2)
	if rand.Int63n(10000) < prob {
		return vs[0]
	}
	return vs[1]
}

func NextRitualTargetCfg(ctx *ctx.Context, targetId values.Integer) (*rulemodel.MainTaskChapterTarget, bool) {
	list := rule.MustGetReader(ctx).MainTaskChapterTarget.List()
	next := false
	for i, cfg := range list {
		if next {
			return &list[i], true
		}
		if cfg.Id == targetId {
			next = true
		}
	}
	return nil, false
}

func NextRitualTargetTaskId(ctx *ctx.Context, targetId, taskIdx values.Integer) (values.Integer, bool) {
	targetCfg := MustGetTargetCfg(ctx, targetId)
	if int(taskIdx) < len(targetCfg.TaskStage)-1 {
		return targetCfg.TaskStage[int(taskIdx+1)], true
	}
	return 0, false
}
