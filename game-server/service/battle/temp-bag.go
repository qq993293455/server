package battle

import (
	"errors"
	"fmt"
	"math"
	"math/rand"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	mapdata "coin-server/common/map-data"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/utils/generic/gmath"
	"coin-server/common/utils/percent"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/battle/dao"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"

	"go.uber.org/zap"
)

func lockKey(roleId values.RoleId) string {
	return "temp_bag_lock:role:" + roleId
}

const syncInterval = 10

func (this_ *Service) HandleLoginEvent(c *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	tb, err := dao.GetTempBag(c, d.RoleId)
	if err != nil {
		return err
	}
	// TODO 动态调节间隔时间后需要改这里
	if timer.Now().Unix()-tb.KickOutTime <= syncInterval {
		return nil
	}
	exp := calTempBagProfit(tb, true)
	if exp > 0 { // 挂机经验
		err = this_.UserService.AddExp(c, c.RoleId, exp, true)
		if err != nil {
			this_.log.Error("call UserService.AddExp failed.", zap.Error(err))
			return err
		}
	}
	dao.SaveTempBag(c, tb)
	return nil
}

func (this_ *Service) HandleTempBagSyncEvent(c *ctx.Context, req *servicepb.BattleEvent_TempBagSyncEvent) {
	//ids := make([]string, 0, len(req.TempBags))
	//keys := make([]string, 0, len(req.TempBags))
	//tbMap := make(map[string]*models.RoleTempBag, len(req.TempBags))
	//for _, tb := range req.TempBags {
	//	ids = append(ids, tb.RoleId)
	//	keys = append(keys, lockKey(tb.RoleId))
	//	tbMap[tb.RoleId] = tb
	//}
	//
	//err := c.DRLock(redisclient.GetLocker(), keys...)
	//if err != nil {
	//	return
	//}
	//
	//tbs, err := dao.BatchGetTempBag(c, ids)
	//if err != nil {
	//	this_.log.Warn("HandleTempBagSyncEvent: BatchGetTempBag failed!", zap.String("error: ", err.Error()))
	//	return
	//}
	ids := make([]string, 0, len(req.TempBags))
	for _, tb := range req.TempBags {
		ids = append(ids, tb.RoleId)
	}
	roleMap, err := this_.UserService.GetRole(c, ids)
	if err != nil {
		panic(err)
	}

	for _, tb := range req.TempBags {
		role, ok := roleMap[tb.RoleId]
		if !ok {
			this_.log.Error("role not found", zap.String("role_id", tb.RoleId))
			continue
		}
		header := new(models.ServerHeader)
		utils.DeepCopy(header, c.ServerHeader)
		header.RoleId = tb.RoleId
		header.UserId = role.UserId
		newCtx := ctx.NewContext(header, &servicepb.BattleEvent_SyncTempBagEvent{TempBag: tb}, nil)
		this_.svc.GetEventLoop().PostEventQueue(newCtx)
	}

	//dao.BatchSaveTempBag(c, tbs)
}

func (this_ *Service) SyncTempBag(c *ctx.Context, req *servicepb.BattleEvent_SyncTempBagEvent) {
	tb, err := dao.GetTempBag(c, c.RoleId)
	if err != nil {
		panic(err)
	}
	this_.mergeTempBag(c, tb, req.TempBag)
	dao.SaveTempBag(c, tb)
}

func (this_ *Service) GetTempBag(c *ctx.Context, _ *servicepb.GameBattle_GameGetTempBagRequest) (*servicepb.GameBattle_GameGetTempBagResponse, *errmsg.ErrMsg) {
	tb, err := dao.GetTempBag(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	//calTempBagProfit(tb, false)
	return &servicepb.GameBattle_GameGetTempBagResponse{TempBag: tb.TempBag, MapId: tb.MapId}, nil
}

func (this_ *Service) DrawTempBag(c *ctx.Context, _ *servicepb.GameBattle_GameDrawTempBagRequest) (*servicepb.GameBattle_GameDrawTempBagResponse, *errmsg.ErrMsg) {
	//err := c.DRLock(redisclient.GetLocker(), lockKey(c.RoleId))
	//if err != nil {
	//	return nil, err
	//}

	tb, err := dao.GetTempBag(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	//calTempBagProfit(tb, false)
	items := tb.TempBag.Items
	exp := items[enum.RoleExp]
	if exp > 0 {
		err = this_.UserService.AddExp(c, c.RoleId, exp, false)
		if err != nil {
			return nil, err
		}
		delete(items, enum.RoleExp)
	}
	_, err = this_.BagService.AddManyItem(c, c.RoleId, items)
	if err != nil {
		return nil, err
	}
	resetTempBag(tb)
	dao.SaveTempBag(c, tb)
	return &servicepb.GameBattle_GameDrawTempBagResponse{Items: items}, nil
}

func (this_ *Service) mergeTempBag(c *ctx.Context, tb1 *daopb.RoleTempBag, tb2 *models.RoleTempBag) {
	tb1.MapId = tb2.MapId
	exp := calTempBagProfit(tb1, false)
	if exp > 0 {
		err := this_.UserService.AddExp(c, tb1.RoleId, exp, true)
		if err != nil {
			this_.log.Error("call UserService.AddExp failed.", zap.Error(err))
			return
		}
	}

	for itemId, count := range tb2.TempBag.Items {
		addTempBagItem(tb1, itemId, count)
	}
	tb1.TempBag.KillMonster += tb2.TempBag.KillMonster
	tb1.TempBag.TakeMedicine += tb2.TempBag.TakeMedicine
	tb1.TempBag.KillBoss += tb2.TempBag.KillBoss
	tb1.TempBag.DeadCount += tb2.TempBag.DeadCount
	tb1.KickOutTime = timer.Now().Unix()
}

// calTempBagProfit 计算收益 返回挂机经验收益
func calTempBagProfit(bag *daopb.RoleTempBag, isLogin bool) int64 {
	if bag.LastCalcTime-bag.TempBag.StartTime >= bag.ProfitUpper {
		return 0
	}

	now := timer.Now().Unix()
	if now-bag.TempBag.StartTime > bag.ProfitUpper {
		now = bag.TempBag.StartTime + bag.ProfitUpper
	}
	sec := now - bag.LastCalcTime
	if sec < 1 {
		return 0
	}

	//mapScene, ok := rule.MustGetReader(nil).MapScene.GetMapSceneById(bag.MapId)
	//if !ok {
	//	panic(errors.New("map_scene not found. id:" + strconv.Itoa(int(bag.MapId))))
	//}

	if bag.TempBag.Items == nil {
		bag.TempBag.Items = map[int64]int64{}
	}
	// 现在不能通过挂机获得金币了
	//gold := sec * mapScene.GoldProfit
	//if gold > 0 {
	//	bag.TempBag.Items[enum.Gold] += gold
	//}
	exp := gmath.CeilTo[int64](float64(sec) * bag.TempBag.ExpProfit)

	if isLogin {
		calKickOutDrop(bag, now)
	}

	bag.LastCalcTime = now
	return exp
}

// 计算离线掉落
func calKickOutDrop(bag *daopb.RoleTempBag, now int64) {
	if bag.KickOutTime < now {
		if bag.KickOutTime < bag.TempBag.StartTime {
			bag.KickOutTime = bag.TempBag.StartTime
		}
		reader := rule.MustGetReader(nil)

		onlineTime := bag.KickOutTime - bag.TempBag.StartTime
		if onlineTime < syncInterval {
			onlineTime = syncInterval
		}
		kmRate := float64(bag.TempBag.KillMonster) / float64(onlineTime)
		count := int64(math.Ceil(kmRate * float64(now-bag.KickOutTime)))
		kt, ok := reader.KeyValue.GetInt64("TempBagKillTime")
		if !ok {
			panic("KeyValue TempBagKillTime not found")
		}
		// 计算保底杀怪
		minCount := int64(math.Ceil(float64(now-bag.KickOutTime) / float64(kt)))
		if count < minCount {
			count = minCount
		}
		bag.TempBag.KillMonster += count

		lmap := mapdata.MustGetMapData(mapdata.GetLogicMapId(nil, bag.MapId))
		monsterMap := make(map[values.Integer]struct{})
		for _, m := range lmap.Monsters {
			monsterMap[m.MonsterId] = struct{}{}
		}
		monsters := make([]*rulemodel.Monster, 0, len(monsterMap))
		for id := range monsterMap {
			mcfg, ok := reader.Monster.GetMonsterById(id)
			if !ok || mcfg.MonsterType != 1 {
				continue
			}
			monsters = append(monsters, mcfg)
		}

		if len(monsters) > 0 {
			for i := 0; i < int(count); i++ {
				mcfg := monsters[rand.Intn(len(monsters))]
				for k, v := range reader.GetDropList(mcfg.DropListId) {
					addTempBagItem(bag, k, v)
					//bag.TempBag.Items[k] += v
				}
			}
		}

		bag.KickOutTime = now
	}
}

// storeTempBagExpProfit 每秒经验收益有变动时算好存下来
func storeTempBagExpProfit(bag *daopb.RoleTempBag) {
	expProfit := bag.ExpProfitBase // 每秒经验
	expProfit += bag.ExpProfitAdd
	bag.TempBag.ExpProfit = float64(expProfit) + percent.AdditionFloat(expProfit, bag.ExpProfitPercent)
}

func resetTempBag(bag *daopb.RoleTempBag) {
	now := timer.Now().Unix()
	bag.LastCalcTime = now
	bag.KickOutTime = now
	bag.TempBag.StartTime = now
	bag.TempBag.Items = map[int64]int64{}
	bag.TempBag.KillMonster = 0
	bag.TempBag.TakeMedicine = 0
	bag.TempBag.KillBoss = 0
	bag.TempBag.DeadCount = 0
	cap_, ok := rule.MustGetReader(nil).KeyValue.GetInt64("TempBagCap")
	if !ok {
		panic(errors.New("KeyValue TempBagCap not found"))
	}
	bag.CapLimit = cap_
	bag.BagSize = 0
	storeTempBagExpProfit(bag)
}

func addTempBagItem(bag *daopb.RoleTempBag, id, count values.Integer) {
	if bag.TempBag.Items == nil {
		bag.TempBag.Items = map[int64]int64{}
	}
	cfg, ok := rule.MustGetReader(nil).Item.GetItemById(id)
	if !ok {
		panic(fmt.Sprintf("item config not found: %d", id))
	}
	isStack := cfg.Stack != -1 // 是否可堆叠
	_, isOwn := bag.TempBag.Items[id]

	if bag.BagSize < bag.CapLimit { // 如果临时背包没满
		bag.TempBag.Items[id] += count
		if !isOwn || !isStack { // 如果未拥有 或者 不可堆叠
			bag.BagSize++ // 占用容量+1
		}
	} else if isOwn && isStack { // 如果已拥有且可堆叠
		bag.TempBag.Items[id] += count
	}
}

// TempBagExpProfit 获取临时背包每秒经验收益
func (this_ *Service) TempBagExpProfit(c *ctx.Context, roleId string) (float64, *errmsg.ErrMsg) {
	bag, err := dao.GetTempBag(c, roleId)
	if err != nil {
		return 0, err
	}
	return bag.TempBag.ExpProfit, nil
}
