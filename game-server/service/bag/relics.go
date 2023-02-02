package bag

import (
	"fmt"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/service/bag/dao"
	"coin-server/rule"
)

func (s *Service) addManyRelics(ctx *ctx.Context, roleId values.RoleId, items map[values.ItemId]values.Integer) *errmsg.ErrMsg {
	var add []*pbdao.Relics
	relicsList := make([]*pbdao.Relics, 0)
	for k := range items {
		relicsList = append(relicsList, &pbdao.Relics{RelicsId: k})
	}
	// 没找到的添加遗物
	r := rule.MustGetReader(ctx)
	nfi, err := dao.GetMultiRelics(ctx, roleId, relicsList)
	if err != nil {
		return err
	}
	for _, n := range nfi {
		items[relicsList[n].RelicsId]--
		if items[relicsList[n].RelicsId] == 0 {
			delete(items, relicsList[n].RelicsId)
		}
		if add == nil {
			add = make([]*pbdao.Relics, 0)
		}
		add = append(add, newRelics(ctx, relicsList[n].RelicsId))
		cfg, ok := r.Item.GetItemById(relicsList[n].RelicsId)
		if !ok {
			continue
		}
		s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskRelicsQualityGetAcc, cfg.Quality, 1)
	}
	if len(add) != 0 {
		dao.SaveMultiRelics(ctx, roleId, add)
		ctx.PublishEventLocal(&event.RelicsUpdate{
			IsNewRelics: true,
			Relics:      RelicsDao2Models(add),
		})
		// 遗物收集打点
		s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGetRelicsAcc, 0, values.Integer(len(add)))
	}

	// 已有的转化为碎片
	if len(items) == 0 {
		return nil
	}
	pieceMap := map[values.ItemId]values.Integer{}
	for k, v := range items {
		cfg, ok := r.Relics.GetRelicsById(k)
		if !ok {
			continue
		}
		pieceKV := cfg.FragmentRelics
		if len(pieceKV) != 2 {
			panic(fmt.Sprintf("Relics config err: %d FragmentRelics length not equals 2", k))
		}
		pieceMap[pieceKV[0]] = v * pieceKV[1]
	}
	_, err = s.BagService.AddManyItem(ctx, roleId, pieceMap)
	if err != nil {
		return err
	}
	return nil
}

func newRelics(ctx *ctx.Context, relicsId values.ItemId) *pbdao.Relics {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.Relics.GetRelicsById(relicsId)
	if !ok {
		panic(fmt.Sprintf("Relics config err: %d not exist", relicsId))
	}
	if len(cfg.AttrId) == 0 || len(cfg.AttrId) != len(cfg.AttrValue) {
		panic(fmt.Sprintf("Relics config err: %d AttrId's length not equal AttrValue", relicsId))
	}
	attr := map[values.Integer]values.Integer{}
	for idx, k := range cfg.AttrId {
		attr[k] = cfg.AttrValue[idx]
	}
	relics := &pbdao.Relics{
		RelicsId: relicsId,
		Level:    DefaultRelicsLevel,
		Star:     DefaultRelicsStar,
		Attr:     attr,
		IsNew:    true,
	}
	return relics
}
