package stage

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/game-server/event"
	"coin-server/game-server/service/stage/dao"
)

func (svc *Service) HandleEventFinished(ctx *ctx.Context, d *event.MapEventFinished) *errmsg.ErrMsg {
	return svc.updateExplore(ctx, d.EventId, d.MapId)
}

func (svc *Service) HandleTargetUpdate(ctx *ctx.Context, d *event.TargetUpdate) *errmsg.ErrMsg {
	if d.IsAccumulate {
		return nil
	}

	stage, err := dao.GetStage(ctx)
	if err != nil {
		return err
	}
	nextStage, err := svc.tryLock(ctx, stage)
	if err != nil {
		return err
	}
	isSave := false
	if nextStage != stage.CurrStage {
		stage.CurrStage = nextStage
		stage.TypeCount = map[int64]int64{}
		isSave = true
	}
	next, err := svc.NextStage(ctx, stage.CurrStage)
	if err != nil {
		return err
	}
	if next == nil {
		return nil
	}

	for _, v := range next.UnlockCondition {
		if models.TaskType(v[0]) == d.Typ {
			stage.TypeCount[v[0]] += d.Count
			isSave = true
		}
	}
	if isSave {
		dao.SaveStage(ctx, stage)
	}
	return nil
}
