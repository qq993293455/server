package bag

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/iggsdk"
	"coin-server/common/logger"
	"coin-server/common/proto/cppbattle"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/statistical"
	models2 "coin-server/common/statistical/models"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/common/values/enum/ItemType"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/bag/dao"
	rule2 "coin-server/game-server/service/bag/rule"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"
	"github.com/gogo/protobuf/proto"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type Updater func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg
type Querier func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) (count values.Integer, err *errmsg.ErrMsg)
type Exchanger func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg
type Events struct {
	// itemUpdate *event.ItemUpdate
	stoneUpdate *event.SkillStoneUpdate
	runeUpdate  *event.TalentRuneUpdate
}

type Service struct {
	serverId       values.ServerId
	serverType     models.ServerType
	svc            *service.Service
	updatersById   map[values.ItemId]Updater
	updatersByType map[values.ItemType]Updater
	querierById    map[values.ItemId]Querier
	exchangerById  map[values.ItemId]Exchanger
	*module.Module
	log *logger.Logger
}

func NewBagService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:       serverId,
		serverType:     serverType,
		svc:            svc,
		updatersById:   map[values.ItemId]Updater{},
		updatersByType: map[values.ItemType]Updater{},
		querierById:    map[values.ItemId]Querier{},
		exchangerById:  map[values.ItemId]Exchanger{},
		Module:         module,
		log:            log,
	}
	module.BagService = s
	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("获取背包", s.GetBagRequest)
	s.svc.RegisterFunc("获取单个道具", s.GetItemRequest)
	s.svc.RegisterFunc("使用道具", s.UseItemRequest)
	s.svc.RegisterFunc("卖出道具或装备", s.SellThings)
	s.svc.RegisterFunc("增加道具", s.AddItemRequest)
	s.svc.RegisterFunc("增加多个装备", s.AddItemsRequest)
	s.svc.RegisterFunc("获取自动出售配置", s.GetBagConfigRequest)
	s.svc.RegisterFunc("保存自动出售配置", s.SaveBagConfigRequest)
	s.svc.RegisterFunc("扩充背包", s.ExpandCapacity)
	s.svc.RegisterFunc("花费钻石立即完成背包扩充", s.ExpandCapacityImmediately)
	s.svc.RegisterFunc("钻石兑换绑钻", s.DiamondExchangeBound)
	s.svc.RegisterFunc("道具合成", s.SynthesisItem)
	s.svc.RegisterFunc("获取装备详情", s.GetEquipDetails)

	s.svc.RegisterEvent("增加多个装备通知", s.AddItemsEvent)
	s.svc.RegisterEvent("增加道具通知", s.AddItemEvent)

	s.svc.RegisterFunc("作弊器增加道具", s.CheatAddItem)
	s.svc.RegisterFunc("作弊器获取所有道具", s.CheatAddAll)
	s.svc.RegisterFunc("作弊器设置道具", s.CheatSetItem)
	s.svc.RegisterFunc("作弊器清空背包", s.CheatDeleteAll)

	eventlocal.SubscribeEventLocal(s.HandleItemUpdateEvent)
	eventlocal.SubscribeEventLocal(s.HandleMedicineChange)
	eventlocal.SubscribeEventLocal(s.HandlerUserLevelUp)
	eventlocal.SubscribeEventLocal(s.HandleEquipUpdateEvent)
	eventlocal.SubscribeEventLocal(s.HandleEquipDestroyedEvent)
	eventlocal.SubscribeEventLocal(s.HandleRelicsUpdate)
	eventlocal.SubscribeEventLocal(s.HandleSkillStoneUpdate)
	eventlocal.SubscribeEventLocal(s.HandleTalentRuneUpdate)
	eventlocal.SubscribeEventLocal(s.HandleTalentRuneDel)
	eventlocal.SubscribeEventLocal(s.HandleBattleSettingChange)
}

func (s *Service) InitBagConfig(ctx *ctx.Context) {
	dao.InitBagConfig(ctx, rule2.GetBagInitCap(ctx))
}

func (s *Service) GetBagConfig(ctx *ctx.Context) (*pbdao.BagConfig, *errmsg.ErrMsg) {
	return dao.GetBagConfig(ctx, ctx.RoleId)
}

func (s *Service) SaveBagConfig(ctx *ctx.Context, data *pbdao.BagConfig) {
	dao.SaveBagConfig(ctx, data)
}

func (s *Service) IsBagEnough(ctx *ctx.Context, items map[values.ItemId]values.Integer) (bool, *errmsg.ErrMsg) {
	_, add2mail, err := s.beforeAdd2Bag(ctx, ctx.RoleId, items, false)
	if err != nil {
		return false, err
	}
	return len(add2mail) <= 0, nil
}

func (s *Service) GetItem(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) (values.Integer, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(ctx)
	if _, ok := reader.Item.GetItemById(itemId); !ok {
		return 0, errmsg.NewErrBagNoSuchItem()
	}
	if querier, ok := s.querierById[itemId]; ok {
		return querier(ctx, roleId, itemId)
	}
	item, err := dao.GetItem(ctx, roleId, itemId)
	if err != nil {
		return 0, err
	}
	return item.Count, nil
}

func (s *Service) GetItemPb(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) (*models.Item, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(ctx)
	if _, ok := reader.Item.GetItemById(itemId); !ok {
		return nil, errmsg.NewErrBagNoSuchItem()
	}
	if querier, ok := s.querierById[itemId]; ok {
		count, err := querier(ctx, roleId, itemId)
		if err != nil {
			return nil, err
		}
		return &models.Item{ItemId: itemId, Count: count}, nil
	}
	item, err := dao.GetItem(ctx, roleId, itemId)
	if err != nil {
		return nil, err
	}
	return ItemDao2Model(item), nil
}

func (s *Service) GetManyItem(ctx *ctx.Context, roleId values.RoleId, itemsId []values.ItemId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	itemList := make([]*pbdao.Item, 0)
	reader := rule.MustGetReader(ctx)
	ret := make(map[values.ItemId]values.Integer)
	for _, id := range itemsId {
		if _, ok := reader.Item.GetItemById(id); !ok {
			return nil, errmsg.NewErrBagNoSuchItem()
		}
		if querier, ok := s.querierById[id]; ok {
			count, err := querier(ctx, roleId, id)
			if err != nil {
				return nil, err
			}
			ret[id] = count
			continue
		}
		itemList = append(itemList, &pbdao.Item{ItemId: id})
	}
	err := dao.GetItems(ctx, roleId, itemList)
	if err != nil {
		return nil, err
	}
	for _, item := range itemList {
		ret[item.ItemId] = item.Count
	}
	return ret, nil
}

func (s *Service) AddItem(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg {
	// if count < 0 {
	//	return errmsg.NewErrBagItemBelowZero()
	// }
	// reader := rule.MustGetReader(ctx)
	// if _, ok := reader.Item.GetItemById(itemId); !ok {
	//	return errmsg.NewErrBagNoSuchItem()
	// }
	// if updater, ok := s.getUpdater(itemId); ok {
	//	return updater(ctx, roleId, itemId, count)
	// }
	// item, err := dao.GetItem(ctx, roleId, itemId)
	// if err != nil {
	//	return err
	// }
	// if item.Count == 0 && isShow(ctx, itemId) {
	//	err = checkAndUpdate(ctx, 1)
	//	if err != nil {
	//		return err
	//	}
	// }
	// item.Count += count
	// dao.SaveItem(ctx, roleId, item)
	// ctx.PublishEventLocal(&event.ItemUpdate{
	//	RoleId: roleId,
	//	Items:  []*models.Item{ItemDao2Model(item)},
	//	Incr:   []values.Integer{count},
	// })
	// return nil
	_, err := s.AddManyItem(ctx, roleId, map[values.ItemId]values.Integer{itemId: count})
	return err
}

func (s *Service) AddManyItem(ctx *ctx.Context, roleId values.RoleId, items map[values.ItemId]values.Integer) ([]*models.Equipment, *errmsg.ErrMsg) {
	events, equips, err := s.AddWithOutEvent(ctx, roleId, items)
	if err != nil {
		return nil, err
	}
	if events == nil {
		return equips, nil
	}
	if events.stoneUpdate != nil {
		ctx.PublishEventLocal(events.stoneUpdate)
	}
	if events.runeUpdate != nil {
		ctx.PublishEventLocal(events.runeUpdate)
	}
	msg := proto.MessageName(ctx.Req)
	for id, count := range items {
		if id == enum.Diamond || id == enum.BoundDiamond {
			if count >= 50000 {
				iggsdk.GetAlarmIns().SendRes(fmt.Sprintf("玩家 %d 钻石单次获取超过5W", utils.Base34DecodeString(roleId)))
			}
		}
		statistical.Save(ctx.NewLogServer(), &models2.GotItem{
			IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
			EventTime: timer.Now(),
			GwId:      statistical.GwId(),
			RoleId:    roleId,
			ItemId:    id,
			Num:       count,
			Msg:       msg,
		})
	}

	return equips, nil
}

func (s *Service) AddWithOutEvent(ctx *ctx.Context, roleId values.RoleId, items map[values.ItemId]values.Integer) (*Events, []*models.Equipment, *errmsg.ErrMsg) {
	add2bag, add2mail, err := s.beforeAdd2Bag(ctx, roleId, items, true)
	if err != nil {
		return nil, nil, err
	}
	if len(add2mail) > 0 {
		attachment := make([]*models.Item, 0)
		for id, count := range add2mail {
			attachment = append(attachment, &models.Item{
				ItemId: id,
				Count:  count,
			})
		}
		id := values.Integer(enum.ItemReissueId)
		var expiredAt values.Integer
		cfg, ok := rule2.GetMailConfigTextId(ctx, id)
		if ok {
			expiredAt = timer.StartTime(ctx.StartTime).Add(time.Second * time.Duration(cfg.Overdue)).UnixMilli()
		}
		if err := s.Add(ctx, ctx.RoleId, &models.Mail{
			Id:         xid.New().String(),
			Type:       models.MailType_MailTypeSystem,
			TextId:     id,
			ExpiredAt:  expiredAt,
			Attachment: attachment,
		}); err != nil {
			return nil, nil, err
		}
	}
	if len(add2bag) == 0 {
		return nil, nil, nil
	}
	var (
		itemList  []*pbdao.Item
		equipMap  map[values.ItemId]values.Integer
		relicsMap map[values.ItemId]values.Integer
		stoneMap  map[values.ItemId]values.Integer
		runeMap   map[values.ItemId]values.Integer
		events    = &Events{}
	)
	reader := rule.MustGetReader(ctx)
	for id, cnt := range add2bag {
		if cnt < 0 {
			return nil, nil, errmsg.NewErrBagItemBelowZero()
		}
		if cnt == 0 {
			continue
		}
		cfg, ok := reader.Item.GetItemById(id)
		if !ok {
			return nil, nil, errmsg.NewErrBagNoSuchItem()
		}
		if updater, ok := s.updatersById[id]; ok {
			err := updater(ctx, roleId, id, cnt)
			if err != nil {
				return nil, nil, err
			}
			continue
		}
		if updater, ok := s.updatersByType[cfg.Typ]; ok {
			err := updater(ctx, roleId, id, cnt)
			if err != nil {
				return nil, nil, err
			}
			if cfg.Typ != ItemType.Relics {
				continue
			}
		}
		if cfg.Typ == ItemType.Equipment {
			if equipMap == nil {
				equipMap = map[values.ItemId]values.Integer{}
			}
			equipMap[id] = cnt
			continue
		}
		if cfg.Typ == ItemType.Relics {
			if relicsMap == nil {
				relicsMap = map[values.ItemId]values.Integer{}
			}
			relicsMap[id] = cnt
			continue
		}
		if cfg.Typ == ItemType.SkillStone {
			if len(stoneMap) == 0 {
				stoneMap = map[values.ItemId]values.Integer{}
			}
			stoneMap[id] = cnt
			continue
		}
		if cfg.Typ == ItemType.TalentRune {
			if len(runeMap) == 0 {
				runeMap = map[values.ItemId]values.Integer{}
			}
			runeMap[id] = cnt
			continue
		}
		if len(itemList) == 0 {
			itemList = make([]*pbdao.Item, 0)
		}
		itemList = append(itemList, &pbdao.Item{ItemId: id})
		if id == enum.Diamond || id == enum.BoundDiamond {
			if cnt >= 50000 {
				iggsdk.GetAlarmIns().SendRes(fmt.Sprintf("玩家 %d 钻石单次获取超过5W", utils.Base34DecodeString(roleId)))
			}
		}
	}
	equips := make([]*models.Equipment, 0)
	if len(equipMap) != 0 {
		var err *errmsg.ErrMsg
		equips, err = s.addManyEquip(ctx, roleId, equipMap)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(relicsMap) != 0 {
		err := s.addManyRelics(ctx, ctx.RoleId, relicsMap)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(stoneMap) != 0 {
		eStone, err := s.addManyStones(ctx, ctx.RoleId, stoneMap)
		if err != nil {
			return nil, nil, err
		}
		events.stoneUpdate = eStone
	}
	if len(runeMap) != 0 {
		eRune, err := s.addManyRunes(ctx, roleId, runeMap)
		if err != nil {
			return nil, nil, err
		}
		events.runeUpdate = eRune
	}
	if len(itemList) == 0 {
		return events, equips, nil
	}
	err = dao.GetItems(ctx, roleId, itemList)
	if err != nil {
		return nil, nil, err
	}
	i := 0
	itemIncr := make([]values.Integer, len(itemList))
	for _, item := range itemList {
		// if item.Count == 0 && isShow(ctx, item.ItemId) {
		// 	err = addBagLength(ctx, 1)
		// 	if err != nil {
		// 		return nil, nil, err
		// 	}
		// }
		cfg, _ := reader.Item.GetItemById(item.ItemId)
		if cfg.ExpiredTime > 0 {
			item.Expire = cfg.ExpiredTime * 1000
		} else if cfg.ExpiredTime < 0 {
			item.Expire = timer.StartTime(ctx.StartTime).UnixMilli() - cfg.ExpiredTime*1000
		}
		item.Count += add2bag[item.ItemId]
		itemIncr[i] = add2bag[item.ItemId]
		i++
	}

	dao.SaveManyItem(ctx, roleId, itemList)
	ctx.PublishEventLocal(&event.ItemUpdate{
		RoleId: roleId,
		Items:  ItemDao2Models(itemList),
		Incr:   itemIncr,
	})
	// 累计获得道具打点
	for idx := range itemList {
		s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGetItemTaskAcc, itemList[idx].ItemId, itemIncr[idx])
		s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGetItemTask, itemList[idx].ItemId, itemIncr[idx])
		s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskCollect, itemList[idx].ItemId, itemIncr[idx])
	}
	return events, equips, nil
}

func (s *Service) SubItem(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg {
	// if count < 0 {
	//	return errmsg.NewErrBagItemBelowZero()
	// }
	// if count == 0 {
	//	return nil
	// }
	// reader := rule.MustGetReader(ctx)
	// if _, ok := reader.Item.GetItemById(itemId); !ok {
	//	return errmsg.NewErrBagNoSuchItem()
	// }
	// item, err := dao.GetItem(ctx, roleId, itemId)
	// if err != nil {
	//	return err
	// }
	// if item.Count-count < 0 {
	//	return errmsg.NewErrBagNotEnough()
	// }
	// item.Count -= count
	// if item.Count == 0 && isShow(ctx, itemId) {
	//	err = dao.DelItem(ctx, roleId, item)
	//	if err != nil {
	//		return err
	//	}
	// } else {
	//	dao.SaveItem(ctx, roleId, item)
	// }
	// ctx.PublishEventLocal(&event.ItemUpdate{
	//	RoleId: roleId,
	//	Items:  []*models.Item{ItemDao2Model(item)},
	//	Incr:   []values.Integer{-count},
	// })
	// return nil
	return s.SubManyItem(ctx, roleId, map[values.ItemId]values.Integer{itemId: count})
}

func (s *Service) SubManyItem(ctx *ctx.Context, roleId values.RoleId, items map[values.ItemId]values.Integer) *errmsg.ErrMsg {
	events, count, err := s.SubWithoutEvent(ctx, roleId, items)
	if err != nil {
		return err
	}
	if err := s.afterSub(ctx, roleId, count); err != nil {
		return err
	}

	if events == nil {
		return nil
	}
	if events.stoneUpdate != nil {
		ctx.PublishEventLocal(events.stoneUpdate)
	}
	msg := proto.MessageName(ctx.Req)
	for id, count := range items {
		statistical.Save(ctx.NewLogServer(), &models2.SubItem{
			IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
			EventTime: timer.Now(),
			GwId:      statistical.GwId(),
			RoleId:    roleId,
			ItemId:    id,
			Num:       count,
			Msg:       msg,
		})
	}

	return nil
}

func (s *Service) SubWithoutEvent(ctx *ctx.Context, roleId values.RoleId, items map[values.ItemId]values.Integer) (*Events, values.Integer, *errmsg.ErrMsg) {
	if len(items) == 0 {
		return nil, 0, nil
	}
	var (
		save, del, itemList []*pbdao.Item
		events              = &Events{}
	)
	itemList = make([]*pbdao.Item, 0)
	var stoneMap map[values.ItemId]values.Integer
	reader := rule.MustGetReader(ctx)
	for id, cnt := range items {
		if cnt < 0 {
			return nil, 0, errmsg.NewErrBagItemBelowZero()
		}
		if cnt == 0 {
			continue
		}
		cfg, ok := reader.Item.GetItemById(id)
		if !ok {
			return nil, 0, errmsg.NewErrBagNoSuchItem()
		}
		if cfg.Typ == ItemType.Equipment || cfg.Typ == ItemType.Relics || cfg.Typ == ItemType.TalentRune {
			return nil, 0, errmsg.NewErrBagRelicsEquipCantSub()
		}
		if cfg.Typ == ItemType.SkillStone {
			if stoneMap == nil {
				stoneMap = map[values.ItemId]values.Integer{}
			}
			stoneMap[id] = cnt
			continue
		}
		itemList = append(itemList, &pbdao.Item{ItemId: id})
	}
	if len(stoneMap) != 0 {
		eStone, err := s.subManyStones(ctx, ctx.RoleId, stoneMap)
		if err != nil {
			return nil, 0, err
		}
		events.stoneUpdate = eStone
	}
	if len(itemList) == 0 {
		return events, 0, nil
	}
	err := dao.GetItems(ctx, roleId, itemList)
	if err != nil {
		return nil, 0, nil
	}

	itemIncr := make([]values.Integer, 0)
	for _, item := range itemList {
		if items[item.ItemId] == 0 {
			continue
		}
		if item.Count-items[item.ItemId] < 0 {
			return nil, 0, errmsg.NewErrBagNotEnough()
		}
		item.Count -= items[item.ItemId]
		if item.Count == 0 && isShow(ctx, item.ItemId) {
			if del == nil {
				del = []*pbdao.Item{item}
			} else {
				del = append(del, item)
			}
		} else {
			if save == nil {
				save = []*pbdao.Item{item}
			} else {
				save = append(save, item)
			}
		}
		itemIncr = append(itemIncr, -items[item.ItemId])
		// cfg, _ := reader.Item.GetItemById(item.ItemId)
		// statistical.Save(ctx.NewLogServer(), &models2.Item{
		//	Id:          0,
		//	RoleId:      roleId,
		//	Time:        time.Now(),
		//	IggId:       ctx.UserId,
		//	ServerId:    ctx.ServerId,
		//	OperateType: 0,
		//	ItemId:      item.ItemId,
		//	Quality:     cfg.Quality,
		//	OldNum:      item.Count + items[item.ItemId],
		//	NewNum:      item.Count,
		// })
	}

	if len(del) != 0 {
		dao.DelManyItem(ctx, roleId, del)
	}
	if len(save) != 0 {
		dao.SaveManyItem(ctx, roleId, save)
	}
	ctx.PublishEventLocal(&event.ItemUpdate{
		RoleId: roleId,
		Items:  ItemDao2Models(itemList),
		Incr:   itemIncr,
	})
	return events, values.Integer(len(del)), nil
}

func (s *Service) ExchangeManyItem(ctx *ctx.Context, roleId values.RoleId, add map[values.ItemId]values.Integer, sub map[values.ItemId]values.Integer) *errmsg.ErrMsg {
	esub, count, err := s.SubWithoutEvent(ctx, roleId, sub)
	if err != nil {
		return err
	}
	if err := s.afterSub(ctx, roleId, count); err != nil {
		return err
	}

	eadd, _, err := s.AddWithOutEvent(ctx, roleId, add)
	if err != nil {
		return err
	}
	if esub.stoneUpdate != nil && eadd.stoneUpdate != nil {
		// id相同add为最新
		notfind := make([]int, 0)
		for i := range esub.stoneUpdate.Stones {
			nf := true
			for j := range eadd.stoneUpdate.Stones {
				if esub.stoneUpdate.Stones[i].ItemId == eadd.stoneUpdate.Stones[j].ItemId {
					nf = false
					break
				}
			}
			if nf {
				notfind = append(notfind, i)
			}
		}
		for _, idx := range notfind {
			eadd.stoneUpdate.Stones = append(eadd.stoneUpdate.Stones, esub.stoneUpdate.Stones[idx])
			eadd.stoneUpdate.Incr = append(eadd.stoneUpdate.Incr, esub.stoneUpdate.Incr[idx])
		}
		ctx.PublishEventLocal(eadd.stoneUpdate)
	} else {
		if esub.stoneUpdate != nil {
			ctx.PublishEventLocal(esub.stoneUpdate)
		}
		if eadd.stoneUpdate != nil {
			ctx.PublishEventLocal(eadd.stoneUpdate)
		}
	}
	msg := proto.MessageName(ctx.Req)
	for id, count := range add {
		statistical.Save(ctx.NewLogServer(), &models2.GotItem{
			IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
			EventTime: timer.Now(),
			GwId:      statistical.GwId(),
			RoleId:    roleId,
			ItemId:    id,
			Num:       count,
			Msg:       msg,
		})
	}
	for id, count := range sub {
		statistical.Save(ctx.NewLogServer(), &models2.SubItem{
			IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
			EventTime: timer.Now(),
			GwId:      statistical.GwId(),
			RoleId:    roleId,
			ItemId:    id,
			Num:       count,
			Msg:       msg,
		})
	}

	return nil
}

func (s *Service) AddManyItemPb(ctx *ctx.Context, roleId values.RoleId, items ...*models.Item) *errmsg.ErrMsg {
	if len(items) <= 0 {
		return nil
	}
	itemMap := make(map[values.ItemId]values.Integer)
	for _, item := range items {
		itemMap[item.ItemId] += item.Count
	}
	_, err := s.AddManyItem(ctx, roleId, itemMap)
	return err
	// events, err := s.AddPbWithoutEvent(ctx, roleId, items...)
	// if err != nil {
	// 	return err
	// }
	// if events == nil {
	// 	return nil
	// }
	// if events.stoneUpdate != nil {
	// 	ctx.PublishEventLocal(events.stoneUpdate)
	// }
	// if events.runeUpdate != nil {
	// 	ctx.PublishEventLocal(events.runeUpdate)
	// }
	// msg := proto.MessageName(ctx.Req)
	// for _, model := range items {
	// 	statistical.Save(ctx.NewLogServer(), &models2.GotItem{
	// 		IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
	// 		EventTime: timer.Now(),
	// 		GwId:      0,
	// 		RoleId:    roleId,
	// 		ItemId:    model.ItemId,
	// 		Num:       model.Count,
	// 		Msg:       msg,
	// 	})
	// 	if model.ItemId == enum.Diamond || model.ItemId == enum.BoundDiamond {
	// 		if model.Count >= 50000 {
	// 			iggsdk.GetAlarmIns().Send(fmt.Sprintf("玩家 %d 钻石单次获取超过5W", utils.Base34DecodeString(roleId)))
	// 		}
	// 	}
	// }
	//
	// return nil
}

func (s *Service) AddPbWithoutEvent(ctx *ctx.Context, roleId values.RoleId, items ...*models.Item) (*Events, *errmsg.ErrMsg) {
	if len(items) == 0 {
		return nil, nil
	}
	var (
		itemList  []*pbdao.Item
		itemIdx   map[values.ItemId]int
		equipMap  map[values.ItemId]values.Integer
		relicsMap map[values.ItemId]values.Integer
		stoneMap  map[values.ItemId]values.Integer
		runeMap   map[values.ItemId]values.Integer
		events    = &Events{}
	)
	itemList = make([]*pbdao.Item, 0)
	itemIdx = map[values.ItemId]int{}
	reader := rule.MustGetReader(ctx)
	for i, item := range items {
		if item.Count < 0 {
			return nil, errmsg.NewErrBagItemBelowZero()
		}
		if item.Count == 0 {
			continue
		}
		cfg, ok := reader.Item.GetItemById(item.ItemId)
		if !ok {
			return nil, errmsg.NewErrBagNoSuchItem()
		}
		if updater, ok := s.updatersById[item.ItemId]; ok {
			err := updater(ctx, roleId, item.ItemId, item.Count)
			if err != nil {
				return nil, err
			}
			continue
		}
		if updater, ok := s.updatersByType[cfg.Typ]; ok {
			err := updater(ctx, roleId, item.ItemId, item.Count)
			if err != nil {
				return nil, err
			}
			// TODO: 注册会导致下面逻辑无法进行
			if cfg.Typ != ItemType.Relics {
				continue
			}
		}
		if cfg.Typ == ItemType.Equipment {
			if equipMap == nil {
				equipMap = map[values.ItemId]values.Integer{}
			}
			equipMap[item.ItemId] += item.Count
			continue
		}
		if cfg.Typ == ItemType.Relics {
			if relicsMap == nil {
				relicsMap = map[values.ItemId]values.Integer{}
			}
			relicsMap[item.ItemId] = item.Count
			continue
		}
		if cfg.Typ == ItemType.SkillStone {
			if stoneMap == nil {
				stoneMap = map[values.ItemId]values.Integer{}
			}
			stoneMap[item.ItemId] = item.Count
			continue
		}
		if cfg.Typ == ItemType.TalentRune {
			if runeMap == nil {
				runeMap = map[values.ItemId]values.Integer{}
			}
			runeMap[item.ItemId] = item.Count
			continue
		}
		itemList = append(itemList, &pbdao.Item{ItemId: item.ItemId})
		itemIdx[item.ItemId] = i
		if item.ItemId == enum.Diamond || item.ItemId == enum.BoundDiamond {
			if item.Count >= 50000 {
				iggsdk.GetAlarmIns().SendRes(fmt.Sprintf("玩家 %d 钻石单次获取超过5W", utils.Base34DecodeString(roleId)))
			}
		}
	}
	if len(equipMap) != 0 {
		_, err := s.addManyEquip(ctx, roleId, equipMap)
		if err != nil {
			return nil, err
		}
	}
	if len(relicsMap) != 0 {
		err := s.addManyRelics(ctx, ctx.RoleId, relicsMap)
		if err != nil {
			return nil, err
		}
	}
	if len(stoneMap) != 0 {
		eStone, err := s.addManyStones(ctx, ctx.RoleId, stoneMap)
		if err != nil {
			return nil, err
		}
		events.stoneUpdate = eStone
	}
	if len(runeMap) != 0 {
		eRune, err := s.addManyRunes(ctx, ctx.RoleId, runeMap)
		if err != nil {
			return nil, err
		}
		events.runeUpdate = eRune
	}
	if len(itemList) == 0 {
		return events, nil
	}
	err := dao.GetItems(ctx, roleId, itemList)
	if err != nil {
		return nil, err
	}
	var i = 0
	itemIncr := make([]values.Integer, len(itemList))
	for _, item := range itemList {
		// if item.Count == 0 && isShow(ctx, item.ItemId) {
		// 	err = addBagLength(ctx, 1)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// }
		incr := items[itemIdx[item.ItemId]].Count
		item.Count += incr
		itemIncr[i] = incr
		i++
	}
	dao.SaveManyItem(ctx, roleId, itemList)
	ctx.PublishEventLocal(&event.ItemUpdate{
		RoleId: roleId,
		Items:  ItemDao2Models(itemList),
		Incr:   itemIncr,
	})
	return events, nil
}

func (s *Service) SubManyItemPb(ctx *ctx.Context, roleId values.RoleId, items ...*models.Item) *errmsg.ErrMsg {
	itemsMap := make(map[values.ItemId]values.Integer)
	for _, item := range items {
		itemsMap[item.ItemId] += item.Count
	}
	return s.SubManyItem(ctx, roleId, itemsMap)
	// events, err := s.SubPbWithoutEvent(ctx, roleId, items...)
	// if err != nil {
	// 	return err
	// }
	// if events == nil {
	// 	return nil
	// }
	// if events.stoneUpdate != nil {
	// 	ctx.PublishEventLocal(events.stoneUpdate)
	// }
	// msg := proto.MessageName(ctx.Req)
	// for _, model := range items {
	// 	statistical.Save(ctx.NewLogServer(), &models2.SubItem{
	// 		IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
	// 		EventTime: timer.Now(),
	// 		GwId:      0,
	// 		RoleId:    roleId,
	// 		ItemId:    model.ItemId,
	// 		Num:       model.Count,
	// 		Msg:       msg,
	// 	})
	// }
	// return nil
}

func (s *Service) SubPbWithoutEvent(ctx *ctx.Context, roleId values.RoleId, items ...*models.Item) (*Events, *errmsg.ErrMsg) {
	if len(items) == 0 {
		return nil, nil
	}
	var (
		save, del []*pbdao.Item
		events    = &Events{}
	)
	itemList := make([]*pbdao.Item, 0)
	itemIdx := make([]int, 0)
	var stoneMap map[values.ItemId]values.Integer
	reader := rule.MustGetReader(ctx)
	for i, item := range items {
		if item.Count == 0 {
			continue
		}
		cfg, ok := reader.Item.GetItemById(item.ItemId)
		if !ok {
			return nil, errmsg.NewErrBagNoSuchItem()
		}
		if cfg.Typ == ItemType.Equipment || cfg.Typ == ItemType.Relics || cfg.Typ == ItemType.TalentRune {
			return nil, errmsg.NewErrBagRelicsEquipCantSub()
		}
		if cfg.Typ == ItemType.SkillStone {
			if stoneMap == nil {
				stoneMap = map[values.ItemId]values.Integer{}
			}
			stoneMap[item.ItemId] = item.Count
			continue
		}
		itemList = append(itemList, &pbdao.Item{ItemId: item.ItemId})
		itemIdx = append(itemIdx, i)
	}
	if len(stoneMap) > 0 {
		eStone, err := s.subManyStones(ctx, ctx.RoleId, stoneMap)
		if err != nil {
			return nil, err
		}
		events.stoneUpdate = eStone
	}
	if len(itemList) == 0 {
		return events, nil
	}
	err := dao.GetItems(ctx, roleId, itemList)
	if err != nil {
		return nil, err
	}
	itemIncr := make([]values.Integer, len(itemList))
	for i, item := range itemList {
		if item.Count == 0 {
			continue
		}
		if item.Count < items[itemIdx[i]].Count {
			return nil, errmsg.NewErrBagNotEnough()
		}
		item.Count -= items[itemIdx[i]].Count
		if item.Count == 0 && isShow(ctx, item.ItemId) {
			if del == nil {
				del = []*pbdao.Item{item}
			} else {
				del = append(del, item)
			}
		} else {
			if save == nil {
				save = []*pbdao.Item{item}
			} else {
				save = append(save, item)
			}
		}
		itemIncr[i] = -items[itemIdx[i]].Count
	}

	if len(del) != 0 {
		dao.DelManyItem(ctx, roleId, del)
	}
	if len(save) != 0 {
		dao.SaveManyItem(ctx, roleId, save)
	}
	ctx.PublishEventLocal(&event.ItemUpdate{
		RoleId: roleId,
		Items:  ItemDao2Models(itemList),
		Incr:   itemIncr,
	})
	return events, nil
}

// func (s *Service) ExchangeManyItemPb(ctx *ctx.Context, roleId values.RoleId, add []*models.Item, sub []*models.Item) *errmsg.ErrMsg {
// 	esub, err := s.SubPbWithoutEvent(ctx, roleId, sub...)
// 	if err != nil {
// 		return err
// 	}
// 	eadd, err := s.AddPbWithoutEvent(ctx, roleId, add...)
// 	if err != nil {
// 		return err
// 	}
// 	if esub.stoneUpdate != nil && eadd.stoneUpdate != nil {
// 		// id相同add为最新
// 		notfind := make([]int, 0)
// 		for i := range esub.stoneUpdate.Stones {
// 			nf := true
// 			for j := range eadd.stoneUpdate.Stones {
// 				if esub.stoneUpdate.Stones[i].ItemId == eadd.stoneUpdate.Stones[j].ItemId {
// 					nf = false
// 					break
// 				}
// 			}
// 			if nf {
// 				notfind = append(notfind, i)
// 			}
// 		}
// 		for _, idx := range notfind {
// 			eadd.stoneUpdate.Stones = append(eadd.stoneUpdate.Stones, esub.stoneUpdate.Stones[idx])
// 			eadd.stoneUpdate.Incr = append(eadd.stoneUpdate.Incr, esub.stoneUpdate.Incr[idx])
// 		}
// 		ctx.PublishEventLocal(eadd.stoneUpdate)
// 	} else {
// 		if esub.stoneUpdate != nil {
// 			ctx.PublishEventLocal(esub.stoneUpdate)
// 		}
// 		if eadd.stoneUpdate != nil {
// 			ctx.PublishEventLocal(eadd.stoneUpdate)
// 		}
// 	}
// 	return nil
// }

func (s *Service) UseItemCase(ctx *ctx.Context, itemId values.ItemId, count values.Integer, choose map[values.ItemId]values.Integer) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	if count <= 0 {
		return nil, errmsg.NewErrBagInvalidUse()
	}
	r := rule.MustGetReader(ctx)
	ex, ok := r.Exchange.GetExchangeById(itemId)
	if !ok {
		return nil, errmsg.NewErrBagNotExchange()
	}
	exCfg, ok := r.ExchangeItemWeight(itemId)
	if !ok {
		return nil, errmsg.NewErrBagNotExchange()
	}
	if exchanger, has := s.getExchanger(itemId); has {
		err := exchanger(ctx, ctx.RoleId, itemId, count)
		if err != nil {
			return nil, err
		}
	}
	items := map[values.ItemId]values.Integer{}
	exType := ex.Typ[0]
	switch exType {
	case values.Integer(models.ExchangeType_All):
		for _, exItem := range exCfg.ExchangeWeights {
			cnt := exItem.ItemCount * count
			items[exItem.ItemId] = cnt
		}
	case values.Integer(models.ExchangeType_RandomN):
		rn := ex.Typ[1]
		if rn*count > 0 && exCfg.TotalWeight <= 0 {
			return nil, errmsg.NewInternalErr("invalid exCfg.TotalWeight")
		}
		// 放回抽取n个
		for n := 0; n < int(rn*count); n++ {
			if exCfg.TotalWeight <= 0 {
				break
			}
			random := rand.Int63n(exCfg.TotalWeight)
			for _, exItem := range exCfg.ExchangeWeights {
				if random < exItem.ItemWeight {
					items[exItem.ItemId] += exItem.ItemCount
					break
				}
			}
		}
	case values.Integer(models.ExchangeType_ChooseOne):
		if len(choose) == 0 {
			return nil, nil
		}
		var useNum int64
		kv := make(map[values.ItemId]rule_model.ExchangeItem, len(exCfg.ExchangeWeights))
		for _, exItem := range exCfg.ExchangeWeights {
			kv[exItem.ItemId] = exItem
		}
		for cid, ccnt := range choose {
			useNum += ccnt
			lItems, has := kv[cid]
			if !has {
				return nil, errmsg.NewErrBagChooseNotMatch()
			}
			cnt := lItems.ItemCount * ccnt
			items[lItems.ItemId] = cnt
		}
		if count != useNum {
			return nil, errmsg.NewErrBagChooseNotMatch()
		}
	case values.Integer(models.ExchangeType_RandomNoReplace):
		rn := int(ex.Typ[1])
		if rn > len(exCfg.ExchangeWeights) {
			return nil, errmsg.NewErrBagInvalidUse()
		}
		if rn == len(exCfg.ExchangeWeights) {
			for _, exItem := range exCfg.ExchangeWeights {
				cnt := exItem.ItemCount * count
				items[exItem.ItemId] = cnt
			}
			return items, nil
		}
		// 不放回抽取
		for cnt := 0; cnt < int(count); cnt++ {
			// 深拷贝
			weight := exCfg.TotalWeight
			exItems := make([]rule_model.ExchangeItem, len(exCfg.ExchangeWeights))
			for i, v := range exCfg.ExchangeWeights {
				exItems[i] = v
			}
			if weight <= 0 {
				return nil, errmsg.NewInternalErr("invalid weight")
			}
			for n := 0; n < rn; n++ {
				if weight <= 0 {
					break
				}
				random := rand.Int63n(weight)
				index := 0
				var loseWeight values.Integer
				var prevWeight values.Integer
				for idx, lItems := range exItems {
					if random < lItems.ItemWeight {
						items[lItems.ItemId] += lItems.ItemCount
						index = idx
						loseWeight = lItems.ItemWeight - prevWeight
						weight -= loseWeight
						break
					}
					prevWeight = lItems.ItemWeight
				}
				exItems = append(exItems[:index], exItems[index+1:]...)
				for ; index < len(exItems); index++ {
					exItems[index].ItemWeight -= loseWeight
				}

			}
		}
	default:
		break
	}
	return items, nil
}

func (s *Service) RegisterUpdaterById(id values.ItemId, f func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg) {
	s.updatersById[id] = f
}

func (s *Service) RegisterUpdaterByType(typ values.ItemType, f func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg) {
	s.updatersByType[typ] = f
}

func (s *Service) RegisterQuerierById(id values.ItemId, f func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) (count values.Integer, err *errmsg.ErrMsg)) {
	s.querierById[id] = f
}

func (s *Service) RegisterExchangerById(id values.ItemId, f func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.ItemId) *errmsg.ErrMsg) {
	s.exchangerById[id] = f
}

func (s *Service) getQuerier(id values.ItemId) (Querier, bool) {
	querier, ok := s.querierById[id]
	if ok {
		return querier, true
	}
	return nil, false
}

func (s *Service) getExchanger(id values.ItemId) (Exchanger, bool) {
	ex, ok := s.exchangerById[id]
	if ok {
		return ex, true
	}
	return nil, false
}

func (s *Service) GetManyEquipBagMap(ctx *ctx.Context, roleId values.RoleId, equipId ...values.EquipId) (map[values.EquipId]*models.Equipment, *errmsg.ErrMsg) {
	if len(equipId) <= 0 {
		return nil, nil
	}
	equips := make([]*pbdao.Equipment, 0, len(equipId))
	for _, id := range equipId {
		equips = append(equips, &pbdao.Equipment{EquipId: id})
	}
	if err := dao.GetManyEquip(ctx, roleId, equips); err != nil {
		return nil, err
	}
	return EquipDao2ModelMap(equips, true), nil
}

func (s *Service) GetEquipById(ctx *ctx.Context, roleId values.RoleId, equipId values.EquipId) (*models.Equipment, *errmsg.ErrMsg) {
	equips, err := dao.GetEquip(ctx, roleId, equipId)
	if err != nil {
		return nil, err
	}
	if equips == nil {
		return nil, errmsg.NewErrBagEquipNotExist()
	}
	return EquipDao2Model(equips, true), nil
}

func (s *Service) GetRelicsById(ctx *ctx.Context, roleId values.RoleId, relicsId values.ItemId) (*models.Relics, *errmsg.ErrMsg) {
	relics, err := dao.GetRelicsById(ctx, roleId, relicsId)
	if err != nil {
		return nil, err
	}
	if relics == nil {
		return nil, errmsg.NewErrBagRelicsNotExist()
	}
	return RelicsDao2Model(relics), nil
}

func (s *Service) GetManyRelics(ctx *ctx.Context, roleId values.RoleId, relicsIds []values.ItemId) ([]*models.Relics, *errmsg.ErrMsg) {
	relicsList := make([]*pbdao.Relics, 0, len(relicsIds))
	for _, relicsId := range relicsIds {
		relicsList = append(relicsList, &pbdao.Relics{
			RelicsId: relicsId,
		})
	}
	notFindIdx, err := dao.GetMultiRelics(ctx, roleId, relicsList)
	if err != nil {
		return nil, err
	}
	res := make([]*models.Relics, 0, len(relicsList)-len(notFindIdx))
	for idx, relics := range relicsList {
		has := false
		for _, notFind := range notFindIdx {
			if idx == notFind {
				has = true
				break
			}
		}
		if !has {
			res = append(res, RelicsDao2Model(relics))
		}
	}
	return res, nil
}

func (s *Service) GetRelics(ctx *ctx.Context, roleId values.RoleId) ([]*models.Relics, *errmsg.ErrMsg) {
	relics, err := dao.GetRelics(ctx, roleId)
	if err != nil {
		return nil, err
	}
	if relics == nil {
		return nil, errmsg.NewErrBagRelicsNotExist()
	}
	return RelicsDao2Models(relics), nil
}

func (s *Service) UpdateRelics(ctx *ctx.Context, roleId values.RoleId, relics *models.Relics) {
	dao.SaveRelics(ctx, roleId, RelicsModel2Dao(relics))
	ctx.PublishEventLocal(&event.RelicsUpdate{
		IsNewRelics: false,
		Relics:      []*models.Relics{relics},
	})
}

func (s *Service) UpdateMultiRelics(ctx *ctx.Context, roleId values.RoleId, data []*models.Relics) {
	dao.SaveMultiRelics(ctx, roleId, RelicsModels2Dao(data))
	ctx.PublishEventLocal(&event.RelicsUpdate{
		IsNewRelics: false,
		Relics:      data,
	})
}

func (s *Service) SaveEquipment(ctx *ctx.Context, roleId values.RoleId, equip *models.Equipment) *errmsg.ErrMsg {
	dao.SaveEquip(ctx, roleId, EquipModel2Dao(equip))
	return nil
}

// func (s *Service) SaveEquipmentBrief(ctx *ctx.Context, roleId values.RoleId, equips ...*models.Equipment) {
// 	eb := make([]*pbdao.EquipmentBrief, 0, len(equips))
// 	for _, equip := range equips {
// 		eb = append(eb, EquipModel2EquipmentBrief(equip))
// 	}
// 	dao.SaveEquipBrief(ctx, roleId, eb...)
// }

func (s *Service) SaveManyEquipment(ctx *ctx.Context, roleId values.RoleId, equips []*models.Equipment) {
	dao.SaveManyEquip(ctx, roleId, EquipModels2Dao(equips))
}

func (s *Service) AutoTakeMedicine(ctx *ctx.Context, roleId values.RoleId, typ values.Integer, mapId values.MapId) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	list, ok := r.Medicament.GetMedicineByType(typ)
	if !ok {
		return nil
	}
	s.log.Debug("medicines", zap.String("roleId", roleId), zap.Any("list", list))
	level, err := s.UserService.GetLevel(ctx, roleId)
	if err != nil {
		return err
	}
	// 吃药成功，返回
	for _, cfg := range list {
		if level >= cfg.Level {
			mapMedicine, has := r.MapScene.GetMapSceneById(mapId)
			if !has {
				s.log.Warn("BattleMapIdNotFound!", zap.Int64("MapSceneId", mapId))
				continue
			}
			find := false
			for _, mapM := range mapMedicine.MedicamentId {
				if mapM == cfg.Id {
					find = true
				}
			}
			// 判断当前地图是否可以吃这个药
			if !find {
				continue
			}
			cnt, err1 := s.GetItem(ctx, roleId, cfg.Id)
			if err1 != nil {
				return err1
			}
			if cnt == 0 {
				continue
			}
			err2 := s.takeMedicine(ctx, roleId, cfg.Id, mapId)
			if err2 != nil {
				continue
			}
		}
	}

	_, err1 := s.checkMedicine(ctx, 1)
	if err1 != nil {
		return err1
	}
	return nil
}

func (s *Service) GetMedicineInfo(ctx *ctx.Context, roleId values.RoleId) (*pbdao.MedicineInfo, *errmsg.ErrMsg) {
	cfg, err := dao.GetMedicineInfo(ctx, roleId)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Service) GetMedicineMsg(ctx *ctx.Context, roleId values.RoleId, battleMapId values.Integer) (map[values.Integer]*models.MedicineInfo, *errmsg.ErrMsg) {
	cfg, err := dao.GetMedicineInfo(ctx, roleId)
	if err != nil {
		return nil, err
	}
	r := rule.MustGetReader(ctx)
	medicineCfg := r.Medicament.GetMedicine()
	medicines := map[values.Integer]*models.MedicineInfo{}
	for typ, ms := range medicineCfg {
		info := &models.MedicineInfo{
			CdTime:   0,
			Open:     cfg.Open[typ],
			AutoTake: cfg.AutoTake[typ],
		}
		if cfg.NextTake[typ]-timer.StartTime(ctx.StartTime).UnixMilli() <= 0 {
			info.CdTime = 0
		} else {
			info.CdTime = cfg.NextTake[typ] - timer.StartTime(ctx.StartTime).UnixMilli()
		}
		mapMedicine, ok := r.MapScene.GetMapSceneById(battleMapId)
		if !ok {
			s.log.Warn("BattleMapIdNotFound!", zap.Int64("MapSceneId", battleMapId))
			return nil, errmsg.NewErrMapNotExist()
		}
		for _, m := range ms {
			find := false
			for _, mapM := range mapMedicine.MedicamentId {
				if mapM == m.Id {
					find = true
				}
			}
			if !find {
				continue
			}
		}
		medicines[typ] = info
	}
	return medicines, nil
}

func (s *Service) SaveMedicineInfo(ctx *ctx.Context, info *pbdao.MedicineInfo) {
	dao.SaveMedicineInfo(ctx, info)
}

func (s *Service) takeMedicine(ctx *ctx.Context, roleId values.RoleId, medicineId values.ItemId, mapId values.MapId) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.Medicament.GetMedicamentById(medicineId)
	if !ok {
		return errmsg.NewErrBagMedicineNotExist()
	}
	info, err := dao.GetMedicineInfo(ctx, roleId)
	if err != nil {
		return err
	}
	next := info.NextTake[cfg.Typ]
	// 没到可以吃药的时间
	if next > timer.StartTime(ctx.StartTime).UnixMilli() {
		return errmsg.NewErrBagMedicineCD()
	}
	level, err := s.UserService.GetLevel(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	if level < cfg.Level {
		return errmsg.NewErrBagMedicineLevel()
	}
	u, err := s.UserService.GetUserById(ctx, ctx.UserId)
	err = s.SubItem(ctx, ctx.RoleId, medicineId, 1)
	if err != nil {
		return err
	}

	req := &cppbattle.CPPBattle_TakeMedicineRequest{
		Id:     medicineId,
		CdTime: cfg.Cd,
		Flag:   1,
	}
	mapCnf, ok := r.MapScene.GetMapSceneById(mapId)
	if mapCnf.MapType == values.Integer(models.BattleType_HangUp) ||
		mapCnf.MapType == values.Integer(models.BattleType_Roguelike) ||
		mapCnf.MapType == values.Integer(models.BattleType_UnionBoss) ||
		mapCnf.MapType == values.Integer(models.BattleType_BossHall) {
		out := &cppbattle.CPPBattle_TakeMedicineResponse{}
		err = s.svc.GetNatsClient().RequestWithOut(ctx, u.BattleServerId, req, out)
		if err != nil {
			ctx.Warn("call cpp take medicine fail", zap.Error(err))
			return err
		}
	} else {
		ctx.SetValue("CTX_MEDICINE_REQ", req)
	}
	nextTime := timer.StartTime(ctx.StartTime).UnixMilli() + cfg.Cd
	info.NextTake[cfg.Typ] = nextTime
	dao.SaveMedicineInfo(ctx, info)

	ctx.PushMessage(&servicepb.GameBattle_MedicineCdPush{
		Cd: nextTime,
	})
	return err
}

// func addBagLengthOne(ctx *ctx.Context) *errmsg.ErrMsg {
// 	r := rule.MustGetReader(ctx)
// 	capacity, ok := r.KeyValue.GetInt64("BagCapacity")
// 	if !ok {
// 		panic(fmt.Sprintf("BagCapacity Key not found"))
// 	}
// 	bagLen, err := dao.GetBagLen(ctx, ctx.RoleId)
// 	if err != nil {
// 		return err
// 	}
// 	if bagLen.Length >= capacity {
// 		return errmsg.NewErrBagCapLimit()
// 	}
// 	bagLen.Length++
// 	err = dao.UpdateBagLen(ctx, bagLen)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func addBagLength(ctx *ctx.Context, add values.Integer) *errmsg.ErrMsg {
// 	if add == 1 {
// 		return addBagLengthOne(ctx)
// 	}
// 	r := rule.MustGetReader(ctx)
// 	capacity, ok := r.KeyValue.GetInt64("BagCapacity")
// 	if !ok {
// 		panic(fmt.Sprintf("BagCapacity Key not found"))
// 	}
// 	if add > capacity {
// 		return errmsg.NewErrBagCapLimit()
// 	}
// 	bagLen, err := dao.GetBagLen(ctx, ctx.RoleId)
// 	if err != nil {
// 		return err
// 	}
// 	bagLen.Length += add
// 	if bagLen.Length > capacity {
// 		return errmsg.NewErrBagCapLimit()
// 	}
// 	err = dao.UpdateBagLen(ctx, bagLen)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func subBagLength(ctx *ctx.Context, del values.Integer) *errmsg.ErrMsg {
// 	bagLen, err := dao.GetBagLen(ctx, ctx.RoleId)
// 	if err != nil {
// 		return err
// 	}
// 	bagLen.Length -= del
// 	err = dao.UpdateBagLen(ctx, bagLen)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func isShow(ctx *ctx.Context, itemId values.ItemId) bool {
	r := rule.MustGetReader(ctx)
	item, _ := r.Item.GetItemById(itemId)
	if item.Show == 1 {
		return true
	}
	return false
}

func recovery(ctx *ctx.Context, itemMap map[values.ItemId]values.Integer, equipIds map[values.EquipId]values.ItemId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(ctx)
	items := map[values.ItemId]values.Integer{}
	for id, cnt := range itemMap {
		itemCfg, ok := reader.Item.GetItemById(id)
		if !ok {
			return nil, errmsg.NewErrBagNoSuchItem()
		}
		if itemCfg.SellType == 0 {
			return nil, errmsg.NewErrBagItemCantSell()
		}
		price := itemCfg.Price
		for k, v := range price {
			items[k] += v * cnt
		}
	}
	for _, id := range equipIds {
		itemCfg, ok := reader.Item.GetItemById(id)
		if !ok {
			return nil, errmsg.NewErrBagNoSuchItem()
		}
		if itemCfg.SellType == 0 {
			return nil, errmsg.NewErrBagItemCantSell()
		}
		if itemCfg.Typ != ItemType.Equipment {
			return nil, errmsg.NewErrBagItemNotEquip()
		}
		price := itemCfg.Price
		for k, v := range price {
			items[k] += v
		}
	}
	return items, nil
}

func (s *Service) Lock(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) *errmsg.ErrMsg {
	itemRule, ok := rule.MustGetReader(ctx).Item.GetItemById(itemId)
	if !ok {
		return errmsg.NewErrBagNoSuchItem()
	}
	switch itemRule.Typ {
	case ItemType.SkillStone:
		return s.lockSkillStone(ctx, roleId, itemId)
	default:
		return nil
	}
}

func (s *Service) Unlock(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) *errmsg.ErrMsg {
	itemRule, ok := rule.MustGetReader(ctx).Item.GetItemById(itemId)
	if !ok {
		return errmsg.NewErrBagNoSuchItem()
	}
	switch itemRule.Typ {
	case ItemType.SkillStone:
		return s.unlockSkillStone(ctx, roleId, itemId)
	default:
		return nil
	}
}

// ---------------------------------------------------proto------------------------------------------------------------//

func (s *Service) GetBagRequest(ctx *ctx.Context, _ *servicepb.Bag_GetBagInfoRequest) (*servicepb.Bag_GetBagInfoResponse, *errmsg.ErrMsg) {
	items, err := dao.GetItemBag(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	// equips, err := dao.GetEquipBag(ctx, ctx.RoleId)
	// if err != nil {
	// 	return nil, err
	// }
	equips, err := dao.GetEquipBag(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	relics, err := dao.GetRelics(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	stones, err := dao.GetSkillStoneBag(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	runes, err := dao.GetTalentRuneBag(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if err := s.syncBagConfig(ctx, items, equips); err != nil {
		return nil, err
	}
	// length, err := dao.GetBagLen(ctx, ctx.RoleId)
	// if err != nil {
	// 	return nil, err
	// }
	// equipList := make([]*pbdao.Equipment, 0)
	// for _, equip := range equips {
	// 	// 只发给客户端未装备的装备
	// 	if equip.HeroId == 0 {
	// 		equipList = append(equipList, equip)
	// 	}
	// }
	return &servicepb.Bag_GetBagInfoResponse{
		Bag: &models.Bag{
			Item:       ItemDao2Models(items),
			Equipment:  EquipDao2Models(equips, true),
			Relics:     RelicsDao2Models(relics),
			SkillStone: SkillStoneDao2Models(stones),
			TalentRune: TalentRuneDao2Models(runes),
		}}, nil
}

func (s *Service) GetItemRequest(ctx *ctx.Context, req *servicepb.Bag_GetItemRequest) (*servicepb.Bag_GetItemResponse, *errmsg.ErrMsg) {
	item, err := s.GetItemPb(ctx, ctx.RoleId, req.ItemId)
	if err != nil {
		return nil, err
	}
	return &servicepb.Bag_GetItemResponse{Item: item}, nil
}

// 道具换道具：内部处理，道具换其他：注册exchanger使用器

func (s *Service) UseItemRequest(ctx *ctx.Context, req *servicepb.Bag_UseItemRequest) (*servicepb.Bag_UseItemResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.Item.GetItemById(req.ItemId)
	if !ok {
		return nil, errmsg.NewErrBagNoSuchItem()
	}
	if cfg.Typ == ItemType.Medicine {
		mapMedicine, has := r.MapScene.GetMapSceneById(ctx.BattleMapId)
		if !has {
			s.log.Warn("BattleMapIdNotFound!", zap.Int64("MapSceneId", ctx.BattleMapId))
		}
		find := false
		for _, mapM := range mapMedicine.MedicamentId {
			if mapM == cfg.Id {
				find = true
			}
		}
		// 判断当前地图是否可以吃这个药
		if !find {
			return nil, errmsg.NewErrBagMedicineMap()
		}
		err := s.takeMedicine(ctx, ctx.RoleId, req.ItemId, ctx.BattleMapId)
		if err != nil {
			return nil, err
		}
		return &servicepb.Bag_UseItemResponse{Items: nil}, nil
	} else if cfg.Typ == ItemType.FashionActivate {
		// 时装激活道具一次只能使用一个
		if err := s.SubManyItem(ctx, ctx.RoleId, map[values.ItemId]values.Integer{req.ItemId: 1}); err != nil {
			return nil, err
		}
		items, err := s.ActivateFashion(ctx, req.ItemId)
		if err != nil {
			return nil, err
		}
		return &servicepb.Bag_UseItemResponse{
			Items: items,
		}, nil
	}
	ex, ok := r.Exchange.GetExchangeById(req.ItemId)
	if !ok {
		return nil, errmsg.NewErrBagNotExchange()
	}
	item, err := dao.GetItem(ctx, ctx.RoleId, req.ItemId)
	if err != nil {
		return nil, err
	}
	if item.Expire != 0 && item.Expire <= timer.StartTime(ctx.StartTime).UnixMilli() {
		return nil, errmsg.NewErrBagItemHasExpired()
	}
	err = s.SubItem(ctx, ctx.RoleId, req.ItemId, ex.Count*req.Count)
	if err != nil {
		return nil, err
	}
	items, err := s.UseItemCase(ctx, req.ItemId, req.Count, req.Choose)
	if err != nil {
		return nil, err
	}
	equips, err := s.AddManyItem(ctx, ctx.RoleId, items)
	if err != nil {
		return nil, err
	}
	return &servicepb.Bag_UseItemResponse{
		Items:  items,
		Equips: equips,
	}, nil
}

func (s *Service) SellThings(ctx *ctx.Context, req *servicepb.Bag_SellThingsRequest) (*servicepb.Bag_SellThingsResponse, *errmsg.ErrMsg) {
	if len(req.Items) == 0 && len(req.Equips) == 0 {
		return nil, nil
	}
	idList := make([]values.ItemId, 0)
	for id := range req.Items {
		idList = append(idList, id)
	}
	if len(req.Items) != 0 {
		items, err := s.GetManyItem(ctx, ctx.RoleId, idList)
		if err != nil {
			return nil, err
		}

		for id, count := range items {
			if count < req.Items[id] {
				return nil, errmsg.NewErrBagNotEnough()
			}
		}
	}

	if len(req.Equips) != 0 {
		equips := make([]*pbdao.Equipment, 0, len(req.Equips))
		for v := range req.Equips {
			equips = append(equips, &pbdao.Equipment{EquipId: v})
		}
		err := dao.GetManyEquip(ctx, ctx.RoleId, equips)
		if err != nil {
			return nil, err
		}
		s.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskRecycleEquip, 0, int64(len(req.Equips)))
	}

	item, err := recovery(ctx, req.Items, req.Equips)
	if err != nil {
		return nil, err
	}
	err = s.SubManyItem(ctx, ctx.RoleId, req.Items)
	if err != nil {
		return nil, err
	}
	equips := make([]values.EquipId, 0, len(req.Equips))
	for id := range req.Equips {
		equips = append(equips, id)
	}
	err = s.DelEquipment(ctx, equips...)
	if err != nil {
		return nil, err
	}
	_, err = s.AddManyItem(ctx, ctx.RoleId, item)
	if err != nil {
		return nil, err
	}

	return &servicepb.Bag_SellThingsResponse{Items: item}, nil
}

func (s *Service) AddItemRequest(ctx *ctx.Context, req *servicepb.Bag_AddItemRequest) (*servicepb.Bag_AddItemResponse, *errmsg.ErrMsg) {
	if ctx.ServerType != models.ServerType_GatewayStdTcp {
		return nil, nil
	}
	err := s.AddItem(ctx, ctx.RoleId, req.ItemId, req.Count)
	if err != nil {
		return nil, err
	}
	return &servicepb.Bag_AddItemResponse{}, nil
}

func (s *Service) AddItemEvent(ctx *ctx.Context, req *servicepb.Bag_AddItemEvent) {
	err := s.AddItem(ctx, ctx.RoleId, req.ItemId, req.Count)
	if err != nil {
		ctx.Warn("AddItemsEvent error", zap.Error(err), zap.Any("req", req))
	}
}

func (s *Service) AddItemsEvent(ctx *ctx.Context, req *servicepb.BattleEvent_AddItemsEvent) {
	if len(req.Items) == 0 {
		return
	}
	_, err := s.AddManyItem(ctx, ctx.RoleId, req.Items)
	if err != nil {
		ctx.Warn("AddItemsEvent error", zap.Error(err), zap.Any("req", req))
	}
}

func (s *Service) AddItemsRequest(ctx *ctx.Context, req *servicepb.Bag_AddItemsRequest) (*servicepb.Bag_AddItemsResponse, *errmsg.ErrMsg) {
	if ctx.ServerType == models.ServerType_GatewayStdTcp {
		return nil, nil
	}
	if len(req.Items) == 0 {
		return nil, nil
	}
	_, err := s.AddManyItem(ctx, ctx.RoleId, req.Items)
	if err != nil {
		return nil, err
	}
	return &servicepb.Bag_AddItemsResponse{}, nil
}

func (s *Service) GetBagConfigRequest(ctx *ctx.Context, _ *servicepb.Bag_GetBagConfigRequest) (*servicepb.Bag_GetBagConfigResponse, *errmsg.ErrMsg) {
	cfg, err := dao.GetBagConfig(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	s.handleBagLatticeUnlock(ctx, cfg)
	return &servicepb.Bag_GetBagConfigResponse{
		Config: cfg.Config,
	}, nil
}

func (s *Service) SaveBagConfigRequest(ctx *ctx.Context, req *servicepb.Bag_SaveBagConfigRequest) (*servicepb.Bag_SaveBagConfigResponse, *errmsg.ErrMsg) {
	cfg, err := dao.GetBagConfig(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	cfg.Config.Quality = req.Quality
	dao.SaveBagConfig(ctx, cfg)
	return &servicepb.Bag_SaveBagConfigResponse{}, nil
}

func (s *Service) ExpandCapacity(ctx *ctx.Context, _ *servicepb.Bag_ExpandCapacityRequest) (*servicepb.Bag_ExpandCapacityResponse, *errmsg.ErrMsg) {
	cfg, err := dao.GetBagConfig(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	now := timer.StartTime(ctx.StartTime)
	if cfg.Config.UnlockTime > now.Unix() {
		return nil, errmsg.NewErrBagNeedUnlockLattice()
	}
	if cfg.Config.Capacity >= rule2.GetBagMaxCap(ctx) {
		return nil, errmsg.NewErrBagMaxCapacity()
	}
	duration := rule2.GetUnlockTime(ctx, cfg.Config.BuyCount+1)
	cfg.Config.UnlockTime = now.Add(duration).Unix()
	cfg.Config.BuyCount++
	dao.SaveBagConfig(ctx, cfg)
	return &servicepb.Bag_ExpandCapacityResponse{
		LatticeCount: rule2.GetBagOnceUnlockLatticeCount(ctx),
		UnlockTime:   cfg.Config.UnlockTime,
	}, nil
}

func (s *Service) ExpandCapacityImmediately(ctx *ctx.Context, _ *servicepb.Bag_ExpandCapacityImmediatelyRequest) (*servicepb.Bag_ExpandCapacityImmediatelyResponse, *errmsg.ErrMsg) {
	cfg, err := dao.GetBagConfig(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	now := timer.StartTime(ctx.StartTime).Unix()
	if cfg.Config.UnlockTime <= now {
		return &servicepb.Bag_ExpandCapacityImmediatelyResponse{
			Capacity: cfg.Config.Capacity,
		}, nil
	}
	cost := rule2.GetExpandCapacitySpeedUpCost(ctx)
	if len(cost) != 2 {
		return nil, errmsg.NewInternalErr("AddBagTimeCost key not found")
	}
	d := values.Float(cfg.Config.UnlockTime - now)
	count := values.Integer(math.Ceil(d / values.Float(cost[1])))
	if count > 0 {
		costItem := map[values.ItemId]values.Integer{cost[0]: count}
		if err := s.SubManyItem(ctx, ctx.RoleId, costItem); err != nil {
			return nil, err
		}
	}
	cfg.Config.UnlockTime = 0
	cfg.Config.Capacity += rule2.GetBagOnceUnlockLatticeCount(ctx)
	dao.SaveBagConfig(ctx, cfg)
	return &servicepb.Bag_ExpandCapacityImmediatelyResponse{
		Capacity: cfg.Config.Capacity,
	}, nil
}

// DiamondExchangeBound 钻石兑换绑钻
func (s *Service) DiamondExchangeBound(ctx *ctx.Context, req *servicepb.Bag_DiamondExchangeBoundRequest) (*servicepb.Bag_DiamondExchangeBoundResponse, *errmsg.ErrMsg) {
	err := s.SubItem(ctx, ctx.RoleId, enum.Diamond, req.Num)
	if err != nil {
		return nil, err
	}
	err = s.AddItem(ctx, ctx.RoleId, enum.BoundDiamond, req.Num)
	if err != nil {
		return nil, err
	}
	return &servicepb.Bag_DiamondExchangeBoundResponse{}, nil
}

func (s *Service) SynthesisItem(ctx *ctx.Context, req *servicepb.Bag_SynthesisItemRequest) (*servicepb.Bag_SynthesisItemResponse, *errmsg.ErrMsg) {
	cfg, ok := rule2.GetItemById(ctx, req.ItemId)
	if !ok {
		return nil, errmsg.NewInternalErr("item not found")
	}
	if len(cfg.SynthesisItem) <= 0 {
		return nil, errmsg.NewErrBagCanNotSynthesis()
	}
	num := req.Num
	if num <= 0 {
		num = 1
	}
	// 去掉限制，后期背包做上限，合成的时候先判断合成的物品是否可叠加，然后再判断背包剩余格子数，如果格子不够则不让合成
	// if num > 1000 {
	// 	num = 1000
	// }
	items := make(map[values.ItemId]values.Integer)
	for id, count := range cfg.SynthesisItem {
		items[id] = count * num
	}
	if err := s.SubManyItem(ctx, ctx.RoleId, items); err != nil {
		return nil, err
	}
	if err := s.AddItem(ctx, ctx.RoleId, req.ItemId, num); err != nil {
		return nil, err
	}
	return &servicepb.Bag_SynthesisItemResponse{}, nil
}

func (s *Service) GetEquipDetails(ctx *ctx.Context, req *servicepb.Equip_EquipDetailRequest) (*servicepb.Equip_EquipDetailResponse, *errmsg.ErrMsg) {
	if len(req.EquipId) <= 0 || len(req.EquipId) > 10 {
		return nil, errmsg.NewErrInvalidRequestParam()
	}
	roleId := ctx.RoleId
	if req.RoleId != "" {
		roleId = req.RoleId
	}
	data, err := s.GetManyEquipBagMap(ctx, roleId, req.EquipId...)
	if err != nil {
		return nil, err
	}
	equips := make([]*models.Equipment, 0)
	for _, equip := range data {
		equips = append(equips, equip)
	}
	return &servicepb.Equip_EquipDetailResponse{
		Equip: equips,
	}, nil
}

func (s *Service) handleBagLatticeUnlock(ctx *ctx.Context, bagCfg *pbdao.BagConfig) {
	now := timer.StartTime(ctx.StartTime)
	if bagCfg.Config.UnlockTime > 0 && bagCfg.Config.UnlockTime <= now.Unix() {
		bagCfg.Config.UnlockTime = 0
		bagCfg.Config.Capacity += rule2.GetBagOnceUnlockLatticeCount(ctx)
		dao.SaveBagConfig(ctx, bagCfg)
	}
}

func (s *Service) beforeAdd2Bag(ctx *ctx.Context, roleId values.RoleId, items map[values.ItemId]values.Integer, save bool) (map[values.ItemId]values.Integer, map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	bagCfg, err := dao.GetBagConfig(ctx, roleId)
	if err != nil {
		return nil, nil, err
	}
	s.handleBagLatticeUnlock(ctx, bagCfg)
	stackableMap := make(map[values.ItemId]values.Integer)    // 可堆叠的
	notStackableMap := make(map[values.ItemId]values.Integer) // 不可堆叠的
	equipMap := make(map[values.ItemId]values.Integer)        // 装备
	add2bag := make(map[values.ItemId]values.Integer)         // 添加至背包的物品
	add2mail := make(map[values.ItemId]values.Integer)        // 通过邮件发送的物品

	// 按可堆叠、不可堆叠、装备划分
	stackable, err := s.formatItems(ctx, items, add2bag, stackableMap, notStackableMap, equipMap)
	if err != nil {
		return nil, nil, err
	}
	// 判断是否有自动回收装备
	tempItems, recoveryCount, err := s.autoRecoveryEquip(ctx, bagCfg, equipMap)
	if err != nil {
		return nil, nil, err
	}
	// 自动回收也需要统计
	if recoveryCount > 0 {
		s.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskRecycleEquip, 0, recoveryCount)
	}

	// 将回收装备获得的物品再按可堆叠、不可堆叠、装备划分
	tempStackable, err := s.formatItems(ctx, tempItems, add2bag, stackableMap, notStackableMap, equipMap)
	if err != nil {
		return nil, nil, err
	}
	stackable = append(stackable, tempStackable...)

	stackableData, err := s.GetManyItem(ctx, ctx.RoleId, stackable)
	if err != nil {
		return nil, nil, err
	}
	// 可堆叠，但是背包里已经有了，不占空间
	for id, count := range stackableMap {
		if has, ok := stackableData[id]; ok && has > 0 {
			add2bag[id] += count
			delete(stackableMap, id)
		}
	}
	capacityOccupied := bagCfg.Config.CapacityOccupied
	// 背包没空间
	if bagCfg.Config.Capacity-capacityOccupied <= 0 {
		for id, count := range equipMap {
			add2mail[id] += count
		}
		for id, count := range notStackableMap {
			add2mail[id] += count
		}
		for id, count := range stackableMap {
			add2mail[id] += count
		}
		return add2bag, add2mail, nil
	}

	for id, count := range equipMap {
		notStackableMap[id] += count
	}
	for id, count := range stackableMap {
		add2bag[id] += count
		delete(stackableMap, id)
		capacityOccupied++
		if bagCfg.Config.Capacity-capacityOccupied <= 0 {
			break
		}
	}
	if bagCfg.Config.Capacity > capacityOccupied {
		for id, count := range notStackableMap {
			for i := 0; i < int(count); i++ {
				add2bag[id] += 1
				notStackableMap[id]--
				if notStackableMap[id] <= 0 {
					delete(notStackableMap, id)
				}
				capacityOccupied++
				if bagCfg.Config.Capacity-capacityOccupied <= 0 {
					goto TAG1
				}
			}
		}
	}
TAG1:
	for id, count := range stackableMap {
		add2mail[id] += count
	}
	for id, count := range notStackableMap {
		add2mail[id] += count
	}
	if save {
		bagCfg.Config.CapacityOccupied = capacityOccupied
		dao.SaveBagConfig(ctx, bagCfg)
	}
	return add2bag, add2mail, nil
}

func (s *Service) formatItems(
	ctx *ctx.Context,
	items map[values.ItemId]values.Integer,
	add2bag map[values.ItemId]values.Integer,
	stackableMap map[values.ItemId]values.Integer,
	notStackableMap map[values.ItemId]values.Integer,
	equipMap map[values.ItemId]values.Integer,
) ([]values.ItemId, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(ctx)
	stackable := make([]values.ItemId, 0)
	for id, count := range items {
		itemCfg, ok := reader.Item.GetItemById(id)
		if !ok {
			return nil, errmsg.NewErrBagNoSuchItem()
		}
		// 不在背包里显示的，直接添加
		if itemCfg.Show == 0 {
			add2bag[id] += count
			continue
		}
		// 可以堆叠的
		if itemCfg.Stack != -1 {
			stackable = append(stackable, id)
			stackableMap[id] += count
			continue
		}
		// 装备一定是不可以堆叠的
		if itemCfg.Typ == enum.Equip {
			equipMap[id] += count
			continue
		}
		// 除装备外的其他不可堆叠物品
		notStackableMap[id] += count
	}
	return stackable, nil
}

func (s *Service) autoRecoveryEquip(ctx *ctx.Context, bagCfg *pbdao.BagConfig, items map[values.ItemId]values.Integer,
) (map[values.ItemId]values.Integer, values.Integer, *errmsg.ErrMsg) {
	if bagCfg.Config.Quality == 0 {
		return nil, 0, nil
	}
	equipIds := make(map[string]values.ItemId)
	var index int
	reader := rule.MustGetReader(ctx)
	itemsMap := make(map[values.ItemId]values.Integer)
	for id, count := range items {
		cfg, ok := reader.Item.GetItemById(id)
		if !ok {
			continue
		}
		if cfg.Quality <= bagCfg.Config.Quality {
			for i := 0; i < int(count); i++ {
				equipIds[strconv.Itoa(index)] = id
				index++
			}
			delete(items, id)
		}
	}
	data, err := recovery(ctx, nil, equipIds)
	if err != nil {
		return nil, 0, err
	}
	for id, count := range data {
		itemsMap[id] += count
	}
	return itemsMap, values.Integer(len(equipIds)), nil
}

func (s *Service) afterSub(ctx *ctx.Context, roleId values.RoleId, count values.Integer) *errmsg.ErrMsg {
	bagCfg, err := dao.GetBagConfig(ctx, roleId)
	if err != nil {
		return err
	}
	if count <= 0 {
		return nil
	}
	bagCfg.Config.CapacityOccupied -= count
	if bagCfg.Config.CapacityOccupied < 0 {
		bagCfg.Config.CapacityOccupied = 0
	}
	dao.SaveBagConfig(ctx, bagCfg)
	return nil
}

func (s *Service) syncBagConfig(ctx *ctx.Context, items []*pbdao.Item, equips []*pbdao.Equipment) *errmsg.ErrMsg {
	bagCfg, err := dao.GetBagConfig(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	var occupied values.Integer
	tempItems := make(map[values.Integer]struct{})
	for _, item := range items {
		cfg, ok := rule2.GetItemById(ctx, item.ItemId)
		if !ok {
			continue
		}
		// 不在背包里显示
		if cfg.Show == 0 {
			continue
		}
		// -1不可以堆叠
		if cfg.Stack == -1 {
			occupied += item.Count
		} else {
			if _, ok := tempItems[cfg.Id]; !ok {
				occupied++
				tempItems[cfg.Id] = struct{}{}
			}
		}
	}
	for _, equip := range equips {
		if equip.HeroId == 0 {
			occupied++
		}
	}
	if bagCfg.Config.CapacityOccupied != occupied {
		bagCfg.Config.CapacityOccupied = occupied
		dao.SaveBagConfig(ctx, bagCfg)
	}
	return nil
}

// ---------------------------------------------------cheat------------------------------------------------------------//

func (s *Service) CheatAddItem(ctx *ctx.Context, req *servicepb.Bag_CheatAddItemRequest) (*servicepb.Bag_CheatAddItemResponse, *errmsg.ErrMsg) {
	err := s.AddItem(ctx, ctx.RoleId, req.ItemId, req.Count)
	if err != nil {
		return nil, err
	}
	return &servicepb.Bag_CheatAddItemResponse{}, nil
}

func (s *Service) CheatAddAll(ctx *ctx.Context, req *servicepb.Bag_CheatAddAllItemRequest) (*servicepb.Bag_CheatAddAllItemResponse, *errmsg.ErrMsg) {
	if req.Typ != 0 && req.Typ != ItemType.Equipment && req.Typ != ItemType.Relics {
		return nil, nil
	}
	r := rule.MustGetReader(ctx)
	all := r.Item.List()
	m := map[values.ItemId]values.Integer{}
	for i := range all {
		switch req.Typ {
		case 0:
			if all[i].Typ != ItemType.Equipment && all[i].Typ != ItemType.Relics {
				m[all[i].Id] = req.Count
			}
		case ItemType.Equipment:
			if all[i].Typ == ItemType.Equipment {
				m[all[i].Id] = req.Count
			}
		case ItemType.Relics:
			if all[i].Typ == ItemType.Relics {
				m[all[i].Id] = req.Count
			}
		}
	}
	_, err := s.AddManyItem(ctx, ctx.RoleId, m)
	if err != nil {
		return nil, err
	}
	return &servicepb.Bag_CheatAddAllItemResponse{}, nil
}

func (s *Service) CheatSetItem(ctx *ctx.Context, req *servicepb.Bag_CheatSetItemRequest) (*servicepb.Bag_CheatSetItemResponse, *errmsg.ErrMsg) {
	if req.ItemId == enum.RoleExp && req.Count > 0 {
		if err := s.AddExp(ctx, ctx.RoleId, req.Count, false); err != nil {
			return nil, err
		}
		return &servicepb.Bag_CheatSetItemResponse{}, nil
	}
	cnt, err := s.GetItem(ctx, ctx.RoleId, req.ItemId)
	if err != nil {
		return nil, err
	}
	if cnt == req.Count {
		return nil, nil
	}
	if cnt > req.Count {
		err = s.SubItem(ctx, ctx.RoleId, req.ItemId, cnt-req.Count)
		if err != nil {
			return nil, err
		}
	}
	if cnt < req.Count {
		err = s.AddItem(ctx, ctx.RoleId, req.ItemId, req.Count-cnt)
		if err != nil {
			return nil, err
		}
	}
	return &servicepb.Bag_CheatSetItemResponse{}, nil
}

func (s *Service) CheatDeleteAll(ctx *ctx.Context, req *servicepb.Bag_CheatDeleteAllRequest) (*servicepb.Bag_CheatDeleteAllResponse, *errmsg.ErrMsg) {
	// 道具
	items, err := dao.GetItemBag(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	dao.DelManyItem(ctx, ctx.RoleId, items)

	// 装备
	heroes, err1 := s.Module.GetAllHero(ctx, ctx.RoleId)
	if err1 != nil {
		return nil, err1
	}
	equipsMap, err2 := s.GetManyEquipBagMap(ctx, ctx.RoleId, s.GetHeroesEquippedEquipId(heroes)...)
	if err2 != nil {
		return nil, err2
	}
	// equips, err3 := dao.GetEquipBrief(ctx, ctx.RoleId)
	// if err3 != nil {
	// 	return nil, err3
	// }
	equipIds := make([]values.EquipId, 0, len(equipsMap))
	for _, e := range equipsMap {
		equipIds = append(equipIds, e.EquipId)
	}
	err = s.DelEquipment(ctx, equipIds...)

	// 遗物
	relics, err4 := dao.GetRelics(ctx, ctx.RoleId)
	if err4 != nil {
		return nil, err4
	}
	for _, r := range relics {
		dao.DeleteRelic(ctx, ctx.RoleId, r)
	}

	return &servicepb.Bag_CheatDeleteAllResponse{}, nil
}
