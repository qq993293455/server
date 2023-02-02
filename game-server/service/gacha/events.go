package gacha

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/game-server/event"
	"coin-server/game-server/service/gacha/dao"
)

func (s *Service) HandleRoleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	if d.IsRegister {
		unlock, err := s.checkAndUnlock(ctx, 1, 1)
		if err != nil {
			return err
		}

		if len(unlock) != 0 {
			dao.SaveMultiGacha(ctx, ctx.RoleId, GachaModels2Dao(unlock))
		}
	}

	unLockGachaSlice, err := s.checkAndUnlockGacha(ctx)
	if err != nil {
		return err
	}

	gachas, err := dao.GetGacha(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	for _, gacha := range gachas {
		if s.refresh(ctx, GachaDao2Model(gacha)) {
			dao.SaveGacha(ctx, ctx.RoleId, gacha)
		}
	}

	if len(unLockGachaSlice) != 0 {
		for _, ulock := range unLockGachaSlice {
			hasGachas := false
			for _, gacha := range gachas {
				if ulock.GachaId == gacha.GachaId {
					hasGachas = true
					break
				}
			}

			if hasGachas {
				continue
			}

			dao.SaveGacha(ctx, ctx.RoleId, GachaModel2Dao(ulock))
		}
	}
	return nil
}

func (s *Service) HandleLevelChange(ctx *ctx.Context, d *event.TargetUpdate, args any) *errmsg.ErrMsg {
	unlock, err := s.checkAndUnlock(ctx, 1, d.Count)
	if err != nil {
		return err
	}
	if len(unlock) != 0 {
		dao.SaveMultiGacha(ctx, ctx.RoleId, GachaModels2Dao(unlock))
	}
	return nil
}

func (s *Service) HandleMainTaskFinish(ctx *ctx.Context, d *event.MainTaskFinished) *errmsg.ErrMsg {
	unlock, err := s.checkAndUnlock(ctx, 2, d.TaskNo)
	if err != nil {
		return err
	}
	if len(unlock) != 0 {
		dao.SaveMultiGacha(ctx, ctx.RoleId, GachaModels2Dao(unlock))
	}
	return nil
}
