package battle

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/cppbattle"
	"coin-server/common/proto/models"
	"coin-server/common/values/enum/AdditionType"
	"coin-server/game-server/event"
	dao2 "coin-server/game-server/service/battle/dao"
	"coin-server/rule"
)

// ---------------------------------------- event handler --------------------------------------------//

func (this_ *Service) HandleRoleLvUpEvent(c *ctx.Context, e *event.UserLevelChange) *errmsg.ErrMsg {
	battleServerId, err1 := this_.GetCurBattleSrvId(c)
	if err1 != nil {
		return err1
	}
	if battleServerId > 0 {
		return this_.svc.GetNatsClient().Publish(battleServerId, c.ServerHeader, &cppbattle.CPPBattle_UserLevelChangePush{
			ObjId: c.RoleId,
			Level: e.Level,
		})
	}

	for _, cfg := range rule.MustGetReader(c).TempBag.List() {
		if e.Level == cfg.RoleLevel {
			bag, err := dao2.GetTempBag(c, c.RoleId)
			if err != nil {
				return err
			}
			bag.ProfitUpper = cfg.ProfitUpper * 3600
			dao2.SaveTempBag(c, bag)
			return nil
		}
	}
	return nil
}

// HandleMainTaskFinished 任务更新
func (this_ *Service) HandleMainTaskFinished(c *ctx.Context, d *event.MainTaskFinished) *errmsg.ErrMsg {
	if d.ExpProfit <= 0 {
		return nil
	}
	bag, err := dao2.GetTempBag(c, c.RoleId)
	if err != nil {
		return err
	}
	bag.ExpProfitBase += d.ExpProfit
	storeTempBagExpProfit(bag)
	dao2.SaveTempBag(c, bag)
	return nil
}

func (this_ *Service) HandleRoleTitleEvent(c *ctx.Context, e *event.UserTitleChange) *errmsg.ErrMsg {
	battleServerId, err1 := this_.GetCurBattleSrvId(c)
	if err1 != nil {
		return err1
	}
	if battleServerId > 0 {
		return this_.svc.GetNatsClient().Publish(battleServerId, c.ServerHeader, &cppbattle.CPPBattle_UserTitleChangePush{
			ObjId: c.RoleId,
			Title: e.CurrentTitle,
		})
	}
	return nil
}

// HandleExtraSkillTypTotal 监听特殊词条加成
func (this_ *Service) HandleExtraSkillTypTotal(c *ctx.Context, e *event.ExtraSkillTypTotal) *errmsg.ErrMsg {
	if e.TypId != models.EntrySkillType_ESTIdleExpAdd {
		return nil
	}
	bag, err := dao2.GetTempBag(c, c.RoleId)
	if err != nil {
		return err
	}

	switch e.ValueTyp {
	case AdditionType.Fixed:
		bag.ExpProfitAdd += e.ThisAdd
	case AdditionType.Percent:
		bag.ExpProfitPercent += e.ThisAdd
	default:
		return nil
	}

	storeTempBagExpProfit(bag)
	dao2.SaveTempBag(c, bag)
	return nil
}
