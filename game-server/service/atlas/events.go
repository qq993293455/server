package atlas

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/common/values/enum/ItemType"
	"coin-server/game-server/event"
	"coin-server/rule"
)

func (s *Service) HandleEquipUpdate(ctx *ctx.Context, d *event.EquipUpdate) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	var equipIds []values.Integer
	for itemId := range d.Items {
		cfg, ok := r.Item.GetItemById(itemId)
		if !ok {
			continue
		}
		if cfg.Typ == ItemType.Equipment {
			if equipIds == nil {
				equipIds = make([]values.Integer, 0)
			}
			equipIds = append(equipIds, itemId)
		}
	}
	if len(equipIds) != 0 {
		err := s.addMultiToAtlas(ctx, models.AtlasType_EquipAtlas, equipIds)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) HandleRelicsUpdate(ctx *ctx.Context, d *event.RelicsUpdate) *errmsg.ErrMsg {
	if !d.IsNewRelics {
		return nil
	}
	r := rule.MustGetReader(ctx)
	var relicsIds []values.Integer
	for _, relics := range d.Relics {
		cfg, ok := r.Item.GetItemById(relics.RelicsId)
		if !ok {
			continue
		}
		if cfg.Typ == ItemType.Relics {
			if relicsIds == nil {
				relicsIds = make([]values.Integer, 0)
			}
			relicsIds = append(relicsIds, relics.RelicsId)
		}
	}
	if len(relicsIds) != 0 {
		err := s.addMultiToAtlas(ctx, models.AtlasType_RelicsAtlas, relicsIds)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) HandleMainTaskUpdate(ctx *ctx.Context, d *event.MainTaskFinished) *errmsg.ErrMsg {
	if d.Illustration == 0 {
		return nil
	}
	r := rule.MustGetReader(ctx)
	_, ok := r.StoryIllustration.GetStoryIllustrationById(d.Illustration)
	if !ok {
		return nil
	}
	err := s.addToAtlas(ctx, models.AtlasType_PictureAtlas, d.Illustration)
	if err != nil {
		return err
	}
	return nil
}
