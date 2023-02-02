package journey

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/journey/dao"
	values2 "coin-server/game-server/service/journey/values"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	module     *module.Module
	log        *logger.Logger
}

func NewJourneyService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		module:     module,
		log:        log,
	}
	s.module.JourneyService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("获取征途信息", this_.GetInfo)
	this_.svc.RegisterFunc("开免费宝箱", this_.OpenFreeChest)
	this_.svc.RegisterFunc("开代币宝箱", this_.OpenTokenChest)
	this_.svc.RegisterEvent("多人副本完成推送", this_.RogueLikeFinishPush)
}

func (this_ *Service) RogueLikeFinishPush(c *ctx.Context, msg *servicepb.Journey_RoguelikeFinishPush) {
	err := this_.AddToken(c, c.RoleId, values2.JourneyRougelike, 1)
	if err != nil {
		this_.log.Error("RogueLikeFinishPush error", zap.Error(err))
	}
	c.PublishEventLocal(&event.RLFinishEvent{
		RoguelikeId: values.RoguelikeId(msg.RoguelikeId),
		IsSucc:      msg.Success,
	})
}

func (this_ *Service) AddToken(c *ctx.Context, roleId string, journeyListId int64, count int64) *errmsg.ErrMsg {
	j, err := this_.GetAndFlush(c, roleId)
	if err != nil {
		return err
	}
	r := rule.MustGetReader(c)
	limit, ok := r.KeyValue.GetInt64("JourneyTokenLimit")
	if !ok {
		panic("not found key value:JourneyTokenLimit ")
	}
	itemId, ok := r.KeyValue.GetInt64("JourneyToken")
	if !ok {
		panic("not found key value:JourneyTokenLimit ")
	}
	if j.TodayAddCoinNum >= limit {
		return nil
	}
	l, ok := r.JourneyList.GetJourneyListById(journeyListId)
	if !ok {
		panic(fmt.Sprintf("invalid journeyListId:%d", journeyListId))
	}
	addCount := l.AddItem[itemId] * count
	total := j.TodayAddCoinNum + addCount
	realAdd := addCount
	addItems := make([]*models.Item, 0, len(l.AddItem))
	if realAdd > 0 {
		if total > limit {
			realAdd = limit - j.TodayAddCoinNum
		}
		addItems = append(addItems, &models.Item{ItemId: itemId, Count: realAdd})
	}
	for k, v := range l.AddItem {
		if k != itemId && v > 0 {
			addItems = append(addItems, &models.Item{ItemId: k, Count: v})
		}
	}

	j.TodayAddCoinNum += realAdd
	if len(addItems) > 0 {
		err = this_.module.AddManyItemPb(c, roleId, addItems...)
		if err != nil {
			return err
		}
	}
	dao.Save(c, j)
	return nil
}

func (this_ *Service) GetAndFlush(c *ctx.Context, roleId string) (*daopb.Journey, *errmsg.ErrMsg) {
	j, err := dao.Get(c, roleId)
	if err != nil {
		return nil, err
	}
	now := this_.module.RefreshService.GetCurrDayFreshTime(c).Unix()
	if j.TodayFlushTime != now {
		j.TodayFlushTime = now
		j.TodayAddCoinNum = 0
		dao.Save(c, j)
	}
	return j, nil
}

func (this_ *Service) GetInfo(c *ctx.Context, msg *servicepb.Journey_GetInfoRequest) (*servicepb.Journey_GetInfoResponse, *errmsg.ErrMsg) {
	j, err := this_.GetAndFlush(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	remainFlushTime := int64(0)
	now := time.Unix(0, c.StartTime).UTC().Unix()
	if j.LastOpenFreeTime > now {
		remainFlushTime = j.LastOpenFreeTime - now
	}

	rougelikeCount, err := this_.module.UserService.GetRoguelikeCnt(c)
	if err != nil {
		return nil, err
	}
	towerCount, err := this_.module.TowerService.GetLevelInfo(c, models.TowerType_TT_Default)
	if err != nil {
		return nil, err
	}
	personalSeconds, err := this_.module.PersonalBossService.GetPersonalBossResetRemainSec(c)
	if err != nil {
		return nil, err
	}
	var bossHallSeconds int64
	var bossHallIsOpen bool
	{
		bossHallSeconds = -1
		ut := time.Unix(0, c.StartTime).UTC()
		bt := timer.BeginOfDay(ut)
		s := int64(ut.Sub(bt).Seconds())
		utils.MustTrue(s <= 86400)
		str, ok := rule.MustGetReader(c).KeyValue.GetString("DevilSecretOpenTime")
		if !ok {
			panic("KeyValue:DevilSecretOpenTime not found")
		}
		var rangeTime [][]int64
		err := json.Unmarshal([]byte(str), &rangeTime)
		if err != nil {
			panic(err)
		}
		sort.Slice(rangeTime, func(i, j int) bool {
			return rangeTime[i][0] < rangeTime[j][0]
		})
		rangeTime = append(rangeTime, []int64{rangeTime[0][0] + 86400, rangeTime[0][1] + 86400})
		for _, v := range rangeTime {
			if s >= v[0] && s <= v[1] {
				bossHallSeconds = v[1] - s
				bossHallIsOpen = true
				break
			}
			if s < v[0] {
				bossHallSeconds = v[0] - s
				break
			}
		}
		utils.MustTrue(bossHallSeconds != -1)
	}
	arenaSeconds := this_.module.ArenaService.GetArenaResetRemainSec(c)
	towerUsed := towerCount.CurrentTowerLevel - 1
	if towerCount.IsAllPass {
		towerUsed = towerCount.CurrentTowerLevel
	}
	return &servicepb.Journey_GetInfoResponse{
		FreeFlushSeconds:    remainFlushTime,
		TodayGotCoin:        j.TodayAddCoinNum,
		RougelikeUsed:       rougelikeCount[0],
		RougelikeTotal:      rougelikeCount[1],
		TowerUsed:           towerUsed,
		TowerTotal:          towerCount.MaxChallengeTimes,
		ArenaSeconds:        arenaSeconds,
		BossHallSeconds:     bossHallSeconds,
		BossHallIsOpen:      bossHallIsOpen,
		PersonalBossSeconds: personalSeconds,
	}, nil
}

func (this_ *Service) OpenFreeChest(c *ctx.Context, msg *servicepb.Journey_OpenFreeChestRequest) (*servicepb.Journey_OpenFreeChestResponse, *errmsg.ErrMsg) {
	j, err := dao.Get(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Unix()
	if now < j.LastOpenFreeTime {
		return nil, errmsg.NewErrJourneyFreeNotFlush()
	}
	r := rule.MustGetReader(c)
	cd, ok := r.KeyValue.GetInt64("JourneyChestCD")
	if !ok {
		cd = 150
	}
	j.LastOpenFreeTime = now + cd
	var list []*rule_model.JourneyChest
	all := r.JourneyChest.List()
	for i, v := range all {
		if v.ChestType == 1 {
			list = append(list, &all[i])
		}
	}
	role, err := this_.module.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	var items []*models.Item
	for _, v := range list {
		if v.MinimumLevel <= role.Level && v.HighestLevel >= role.Level {
			itemMap, err := this_.module.UseItemCase(c, v.BoxId, 1, nil)
			if err != nil {
				return nil, err
			}
			for id, count := range v.FixedReward {
				itemMap[id] += count
			}
			for id, count := range itemMap {
				items = append(items, &models.Item{ItemId: id, Count: count})
			}

			break
		}
	}
	if len(items) > 0 {
		err = this_.module.AddManyItemPb(c, c.RoleId, items...)
		if err != nil {
			return nil, err
		}
	}
	dao.Save(c, j)

	return &servicepb.Journey_OpenFreeChestResponse{Items: items}, nil
}

func (this_ *Service) OpenTokenChest(c *ctx.Context, req *servicepb.Journey_OpenTokenChestRequest) (*servicepb.Journey_OpenTokenChestResponse, *errmsg.ErrMsg) {
	openCount := req.Count

	r := rule.MustGetReader(c)
	itemID, ok := r.KeyValue.GetInt64("JourneyToken")
	if !ok {
		panic("not key value config JourneyToken")
	}
	costCount, ok := r.KeyValue.GetInt64("JourneyTokenCost")
	if !ok {
		panic("not key value config JourneyTokenCost")
	}

	err := this_.module.SubItem(c, c.RoleId, itemID, openCount*costCount)
	if err != nil {
		return nil, err
	}

	var list []*rule_model.JourneyChest
	all := r.JourneyChest.List()
	for i, v := range all {
		if v.ChestType == 2 {
			list = append(list, &all[i])
		}
	}
	role, err := this_.module.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	var items []*models.Item
	for _, v := range list {
		if v.MinimumLevel <= role.Level && v.HighestLevel >= role.Level {
			itemMap, err := this_.module.UseItemCase(c, v.BoxId, openCount, nil)
			if err != nil {
				return nil, err
			}
			for id, count := range v.FixedReward {
				itemMap[id] += count
			}
			for id, count := range itemMap {
				items = append(items, &models.Item{ItemId: id, Count: count})
			}
		}
	}
	if len(items) > 0 {
		err = this_.module.AddManyItemPb(c, c.RoleId, items...)
		if err != nil {
			return nil, err
		}
	}

	return &servicepb.Journey_OpenTokenChestResponse{Items: items}, nil
}
