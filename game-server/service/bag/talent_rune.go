package bag

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/bag/dao"
	"coin-server/rule"
	"github.com/rs/xid"
)

func (s *Service) addManyRunes(ctx *ctx.Context, roleId values.RoleId, runeMap map[values.ItemId]values.Integer) (*event.TalentRuneUpdate, *errmsg.ErrMsg) {
	runeList := make([]*pbdao.TalentRune, 0, len(runeMap))
	for itemId, cnt := range runeMap {
		for i := values.Integer(0); i < cnt; i++ {
			runeList = append(runeList, &pbdao.TalentRune{
				RuneId:   xid.New().String(),
				TalentId: itemId,
				Lvl:      1,
			})
			runeRule, ok := rule.MustGetReader(ctx).Talent.GetTalentById(itemId)
			if !ok {
				continue
			}
			s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskOweLvlUpperRune, values.Integer(len(runeRule.RuneShape)), 1)
		}
	}
	err := dao.GetTalentRunes(ctx, roleId, runeList)
	if err != nil {
		return nil, err
	}
	dao.SaveManyTalentRune(ctx, roleId, runeList)
	e := &event.TalentRuneUpdate{
		RoleId: roleId,
		Runes:  TalentRuneDao2Models(runeList),
	}
	return e, nil
}

func (s *Service) DelManyRunes(ctx *ctx.Context, roleId values.RoleId, runeIds []values.RuneId) (*event.TalentRuneDestroyed, *errmsg.ErrMsg) {
	runeList := make([]*pbdao.TalentRune, 0, len(runeIds))
	for _, runeId := range runeIds {
		runeList = append(runeList, &pbdao.TalentRune{
			RuneId: runeId,
		})
	}
	err := dao.GetTalentRunes(ctx, roleId, runeList)
	if err != nil {
		return nil, err
	}
	if len(runeList) != 0 {
		dao.DelManyTalentRune(ctx, roleId, runeList)
	}
	e := &event.TalentRuneDestroyed{
		RoleId:  roleId,
		RuneIds: runeIds,
	}
	return e, nil
}

func (s *Service) GetRuneById(ctx *ctx.Context, roleId values.RoleId, runeId values.RuneId) (*models.TalentRune, *errmsg.ErrMsg) {
	runeItem, err := dao.GetTalentRune(ctx, roleId, runeId)
	if err != nil {
		return nil, err
	}
	return TalentRuneDao2Model(runeItem), nil
}

func (s *Service) SaveRune(ctx *ctx.Context, roleId values.RoleId, rune *models.TalentRune) (*event.TalentRuneUpdate, *errmsg.ErrMsg) {
	dao.SaveTalentRune(ctx, roleId, TalentRuneModel2Dao(rune))
	e := &event.TalentRuneUpdate{
		RoleId: roleId,
		Runes:  []*models.TalentRune{rune},
	}
	return e, nil
}

func (s *Service) GetManyRunes(ctx *ctx.Context, roleId values.RoleId, runeIds []values.RuneId) (map[values.RuneId]*models.TalentRune, *errmsg.ErrMsg) {
	runeList := make([]*pbdao.TalentRune, 0, len(runeIds))
	for _, runeId := range runeIds {
		runeList = append(runeList, &pbdao.TalentRune{
			RuneId: runeId,
		})
	}
	err := dao.GetTalentRunes(ctx, roleId, runeList)
	if err != nil {
		return nil, err
	}
	res := map[values.RuneId]*models.TalentRune{}
	for _, runeItem := range runeList {
		res[runeItem.RuneId] = TalentRuneDao2Model(runeItem)
	}
	return res, nil
}

func (s *Service) SaveManyRunes(ctx *ctx.Context, roleId values.RoleId, runes []*models.TalentRune) (*event.TalentRuneUpdate, *errmsg.ErrMsg) {
	dao.SaveManyTalentRune(ctx, roleId, TalentRuneModels2Dao(runes))
	e := &event.TalentRuneUpdate{
		RoleId: roleId,
		Runes:  runes,
	}
	return e, nil
}
