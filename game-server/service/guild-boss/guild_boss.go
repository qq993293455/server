package guild_boss

import (
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"coin-server/common/iggsdk"
	"coin-server/common/values/env"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/handler"
	"coin-server/common/logger"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	newcenterpb "coin-server/common/proto/newcenter"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/guild-boss/dao"
	"coin-server/rule"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	module     *module.Module
	log        *logger.Logger
}

func NewGuildBossService(
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
	s.module.GuildBossService = s
	return s
}

const GuildIDKey = "GuildID"

// IsUnlock 中间件，判断是否解锁编队
func (this_ *Service) IsUnlock(next handler.HandleFunc) handler.HandleFunc {
	return func(c *ctx.Context) *errmsg.ErrMsg {
		err := this_.module.GuildService.IsUnlockGuildBoss(c)
		if err != nil {
			return err
		}
		c.StartTime = timer.Now().UnixNano()
		return next(c)
	}
}

func (this_ *Service) Router() {
	h := this_.svc.Group(this_.IsUnlock)
	h.RegisterFunc("获取奖励信息", this_.GetCurrDayRewards)
	h.RegisterFunc("领取当天奖励", this_.RewardCurrDay)
	h.RegisterFunc("领取其他天的奖励", this_.RewardOtherDay)
	h.RegisterFunc("检查是否可进入Boss", this_.CanEnter)
	h.RegisterFunc("获取工会Boss的Battle信息", this_.GetBattleServerInfo)
	h.RegisterEvent("保存伤害数据", this_.SaveDamageInfo)
	h.RegisterEvent("触发工会boss次数事件", this_.HandleGuildBossFinishPush)
	h.RegisterFunc("查询排行榜数据", this_.QueryRank)
	h.RegisterFunc("查询战斗在线人数", this_.QueryOnlineInfo)
	h.RegisterFunc("查询自己造成的伤害和排名", this_.GetSelfDamageAndRank)

}

func (this_ *Service) HandleGuildBossFinishPush(c *ctx.Context, msg *servicepb.GuildBoss_GuildBossFinishPush) {
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskGuildBossNumAcc: {
			Typ: values.Integer(models.TaskType_TaskGuildBossNumAcc),
			Cnt: 1,
		},
		models.TaskType_TaskGuildBossNum: {
			Typ: values.Integer(models.TaskType_TaskGuildBossNum),
			Cnt: 1,
		},
	}
	this_.module.UpdateTargets(c, c.RoleId, tasks)
}

func (this_ *Service) IsGuildBossFighting(c *ctx.Context, roleId string) (bool, *errmsg.ErrMsg) {
	r, err := this_.module.UserService.GetRoleByRoleId(c, roleId)
	if err != nil {
		return false, err
	}
	u, err := this_.module.UserService.GetUserById(c, r.UserId)
	if err != nil {
		return false, err
	}
	if u.MapId == int64(models.BattleType_UnionBoss) {
		return true, nil
	}
	return false, nil
}

func (this_ *Service) CurrDayRefreshUnix(c *ctx.Context) int64 {
	return this_.module.RefreshService.GetCurrDayFreshTime(c).Unix()
}

func (this_ *Service) SaveDamageInfo(c *ctx.Context, msg *servicepb.GuildBoss_SyncDamageInfoPush) {
	err := this_.saveDamageInfo(c, msg)
	if err != nil {
		db := c.GetOrmForMiddleWare()
		if db != nil {
			db.Reset()
		}
		this_.log.Error("SaveDamageInfo", zap.Error(err))
	}
}

func (this_ *Service) saveDamageInfo(c *ctx.Context, msg *servicepb.GuildBoss_SyncDamageInfoPush) *errmsg.ErrMsg {
	gb, err := dao.GetGuildBoss(c, msg.GuildDayId)
	if err != nil {
		this_.log.Error("SaveDamageInfo GetGuildBoss", zap.Any("msg", msg), zap.Error(err))
		return err
	}

	gb.TotalDamage = msg.TotalDamages
	var newGbfiIDS []string
	for k, v := range msg.PlayerDamages {
		d, ok := gb.Damages[k]
		if !ok {
			d = &daopb.GuildBossDamageAndFightCount{}
			gb.Damages[k] = d
			newGbfiIDS = append(newGbfiIDS, k)
		}
		d.Damage = v
	}

	if !gb.IsCheck {
		gb.IsCheck = true
		gbl, err := dao.GetGuildBossList(c, strings.Split(msg.GuildDayId, ":")[0])
		if err != nil {
			this_.log.Error("SaveDamageInfo GetGuildBossList", zap.Any("msg", msg), zap.Error(err))
			return err
		}
		foundToday := false

		for _, v := range gbl.GuildDayIds {
			if v == msg.GuildDayId {
				foundToday = true
			}
		}
		change := false
		if !foundToday {
			gbl.GuildDayIds = append(gbl.GuildDayIds, msg.GuildDayId)
			change = true
		}
		gblLen := len(gbl.GuildDayIds)
		const maxSaveSize = 30
		if gblLen > maxSaveSize {
			moreIDS := make([]string, 0, gblLen-maxSaveSize)
			moreIDS = append(moreIDS, gbl.GuildDayIds[:gblLen-maxSaveSize]...)
			copy(gbl.GuildDayIds, gbl.GuildDayIds[:maxSaveSize])
			gbl.GuildDayIds = gbl.GuildDayIds[:maxSaveSize]
			dao.DeleteGBS(c, moreIDS)
			change = true
		}
		if change {
			dao.SaveGuildBossList(c, gbl)
		}
	}

	if len(newGbfiIDS) > 0 {
		fis, err := dao.GetManyGuildBossUserFightInfo(c, newGbfiIDS)
		if err != nil {
			this_.log.Error("SaveDamageInfo GetManyGuildBossUserFightInfo", zap.Any("msg", msg), zap.Error(err))
			return err
		}
		for _, v := range fis {
			v.GuildDayIds = append(v.GuildDayIds, msg.GuildDayId)
			if len(v.GuildDayIds) > 7 {
				v.GuildDayIds = v.GuildDayIds[1:]
			}
		}
		dao.SaveManyGuildBossUserFightInfo(c, fis)
	}

	if msg.OverRole != "" {
		d, ok := gb.Damages[msg.OverRole]
		if !ok {
			d = &daopb.GuildBossDamageAndFightCount{
				Damage: 0,
				Count:  1,
			}
			gb.Damages[msg.OverRole] = d
		} else {
			d.Count++
		}
		err := this_.svc.GetNatsClient().Publish(msg.OverRoleServerId, &models.ServerHeader{
			StartTime:      c.StartTime,
			RoleId:         msg.OverRole,
			ServerId:       this_.serverId,
			ServerType:     this_.serverType,
			TraceId:        c.TraceId,
			InServerId:     msg.OverRoleServerId,
			BattleServerId: c.BattleServerId,
			BattleMapId:    c.BattleMapId,
		}, &servicepb.GuildBoss_GuildBossFinishPush{Count: 1})
		if err != nil {
			this_.log.Error("saveDamageInfo GuildBoss_GuildBossFinishPush failed", zap.Error(err), zap.String("role", msg.OverRole))
		}
	}

	dao.SaveGuildBoss(c, gb)
	return nil
}

func (this_ *Service) GetCurrDayRewards(c *ctx.Context, _ *servicepb.GuildBoss_GetCurrDayRewardsRequest) (*servicepb.GuildBoss_GetCurrDayRewardsResponse, *errmsg.ErrMsg) {

	now := this_.CurrDayRefreshUnix(c)
	guildId, err := this_.module.GuildService.GetGuildIdByRole(c)
	if err != nil {
		return nil, err
	}
	guildDayId := guildId + ":" + strconv.Itoa(int(now))

	gbfi, err := dao.GetGuildBossUserFightInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	gbui, err := dao.GetRewardsUserInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	guilds := append(gbfi.GuildDayIds, guildDayId)

	gbs, err := dao.GetManyGuildBoss(c, guilds)
	if err != nil {
		return nil, err
	}

	currRewards, otherRewards := this_.GetRewardInfo(c, guildDayId, gbfi, gbui, gbs)
	return &servicepb.GuildBoss_GetCurrDayRewardsResponse{
		CurrDayRewards:  currRewards,
		OtherDayRewards: otherRewards,
	}, nil
}

func (this_ *Service) GetRewardInfo(
	c *ctx.Context,
	guildDayId string,
	gbfi *daopb.GuildBossUserFightInfo,
	gbui *daopb.GuildBossUserInfo,
	gbs map[string]*daopb.GuildBoss,
) ([]*servicepb.GuildBoss_CurrDayRewards, []*servicepb.GuildBoss_OtherRewards) {
	gbList := rule.MustGetReader(c).GuildBoss.List()
	currGB := gbs[guildDayId]
	currRewards := make([]*servicepb.GuildBoss_CurrDayRewards, 0, len(gbList))
	var df *daopb.GuildBossDamageAndFightCount
	if currGB != nil {
		df = currGB.Damages[c.RoleId]
	}

	if currGB == nil || df == nil || df.Damage == 0 {
		for _, v := range gbList {
			if len(v.HpReward) != 0 {
				currRewards = append(currRewards, &servicepb.GuildBoss_CurrDayRewards{
					Damage: v.Hp,
					Status: 0,
					Items:  MapToItems(v.HpReward),
				})
			}
		}
	} else {
		totalDamage := currGB.TotalDamage
		rewardsId := gbui.AllRewards[guildDayId]
		for _, v := range gbList {
			if len(v.HpReward) > 0 {
				cdr := &servicepb.GuildBoss_CurrDayRewards{
					Damage: v.Hp,
					Items:  MapToItems(v.HpReward),
				}
				if totalDamage >= v.Hp { // 如果伤害达到
					if rewardsId < v.Id { // 如果已领取的小于当前ID
						cdr.Status = 1 // 未领取
					} else {
						cdr.Status = 2 // 已领取
					}
				}
				currRewards = append(currRewards, cdr)
			}
		}
	}

	otherRewards := make([]*servicepb.GuildBoss_OtherRewards, 0, len(gbfi.GuildDayIds))

	for _, v := range gbfi.GuildDayIds {
		if v == guildDayId {
			continue
		}
		vGB := gbs[v]

		if vGB == nil {
			continue
		}
		df := vGB.Damages[c.RoleId]
		if df == nil || df.Damage == 0 {
			continue
		}
		itemMap := map[values.Integer]values.Integer{}
		totalDamage := vGB.TotalDamage
		rewardsId := gbui.AllRewards[v]
		for _, v := range gbList {
			if totalDamage >= v.Hp {
				if v.Id > rewardsId {
					for k, v := range v.HpReward {
						itemMap[k] = itemMap[k] + v
					}
				}
			} else {
				break
			}
		}

		if len(itemMap) > 0 {
			if vGB.Day == 0 {
				d, err := strconv.Atoi(strings.Split(vGB.GuildDayId, ":")[1])
				if err != nil {
					panic(err)
				}
				vGB.Day = int64(d)
			}
			otherRewards = append(otherRewards, &servicepb.GuildBoss_OtherRewards{
				Timestamp:   vGB.Day,
				SelfDamage:  df.Damage,
				TotalDamage: vGB.TotalDamage,
				Items:       MapToItems(itemMap),
			})
		}
	}
	return currRewards, otherRewards
}

func (this_ *Service) RewardCurrDay(c *ctx.Context, _ *servicepb.GuildBoss_RewardCurrDayRequest) (*servicepb.GuildBoss_RewardCurrDayResponse, *errmsg.ErrMsg) {
	now := this_.CurrDayRefreshUnix(c)
	guildId, err := this_.module.GuildService.GetGuildIdByRole(c)
	if err != nil {
		return nil, err
	}
	guildDayId := guildId + ":" + strconv.Itoa(int(now))

	gbfi, err := dao.GetGuildBossUserFightInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	gbui, err := dao.GetRewardsUserInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	guilds := append(gbfi.GuildDayIds, guildDayId)

	gbs, err := dao.GetManyGuildBoss(c, guilds)
	if err != nil {
		return nil, err
	}

	rewardsId := gbui.AllRewards[guildDayId]
	rewardsMap := map[values.Integer]values.Integer{}
	hasChange := this_.ClearExpireData(gbfi, gbui)
	currGB := gbs[guildDayId]
	if currGB != nil {
		df := currGB.Damages[c.RoleId]
		if df != nil && df.Damage > 0 {
			gbList := rule.MustGetReader(c).GuildBoss.List()
			for _, v := range gbList {
				if currGB.TotalDamage >= v.Hp {
					if rewardsId < v.Id {
						for itemId, count := range v.HpReward {
							rewardsMap[itemId] += count
						}
						rewardsId = v.Id
						hasChange = true
					}
				} else {
					break
				}
			}
		}
	}

	if hasChange {
		gbui.AllRewards[guildDayId] = rewardsId
		dao.SaveRewardsUserInfo(c, gbui)
	}

	if len(rewardsMap) > 0 {
		_, err := this_.module.BagService.AddManyItem(c, c.RoleId, rewardsMap)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errmsg.NewErrorGuildBossAlreadyRewards()
	}

	currRewards, otherRewards := this_.GetRewardInfo(c, guildDayId, gbfi, gbui, gbs)
	return &servicepb.GuildBoss_RewardCurrDayResponse{
		CurrDayRewards:  currRewards,
		OtherDayRewards: otherRewards,
	}, nil

}

func (this_ *Service) ClearExpireData(gbfi *daopb.GuildBossUserFightInfo, gbui *daopb.GuildBossUserInfo) bool {
	mapFI := make(map[string]struct{}, len(gbfi.GuildDayIds))
	for _, v := range gbfi.GuildDayIds {
		mapFI[v] = struct{}{}
	}
	change := false
	for k := range gbui.AllRewards {
		_, ok := mapFI[k]
		if !ok {
			delete(gbui.AllRewards, k)
			change = true
		}
	}
	return change
}

func (this_ *Service) RewardOtherDay(c *ctx.Context, _ *servicepb.GuildBoss_RewardOtherDayRequest) (*servicepb.GuildBoss_RewardOtherDayResponse, *errmsg.ErrMsg) {
	now := this_.CurrDayRefreshUnix(c)
	guildId, err := this_.module.GuildService.GetGuildIdByRole(c)
	if err != nil {
		return nil, err
	}
	guildDayId := guildId + ":" + strconv.Itoa(int(now))

	gbfi, err := dao.GetGuildBossUserFightInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	gbui, err := dao.GetRewardsUserInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}

	guilds := append(gbfi.GuildDayIds, guildDayId)

	gbs, err := dao.GetManyGuildBoss(c, guilds)
	if err != nil {
		return nil, err
	}
	gbList := rule.MustGetReader(c).GuildBoss.List()
	rewardsMap := map[values.Integer]values.Integer{}
	hasOtherChange := this_.ClearExpireData(gbfi, gbui)

	for _, v := range gbfi.GuildDayIds {
		if v == guildDayId {
			continue
		}
		vGB := gbs[v]
		if vGB == nil {
			continue
		}
		df := vGB.Damages[c.RoleId]
		if df == nil || df.Damage == 0 {
			continue
		}
		totalDamage := vGB.TotalDamage
		rewardsId := gbui.AllRewards[v]
		for _, v := range gbList {
			if totalDamage >= v.Hp {
				if v.Id > rewardsId {
					for k, v := range v.HpReward {
						rewardsMap[k] = rewardsMap[k] + v
					}
					rewardsId = v.Id
				}
			} else {
				break
			}
		}
		if gbui.AllRewards[v] != rewardsId {
			gbui.AllRewards[v] = rewardsId
			hasOtherChange = true
		}
	}
	if hasOtherChange {
		dao.SaveRewardsUserInfo(c, gbui)
	}

	if len(rewardsMap) > 0 {
		_, err := this_.module.BagService.AddManyItem(c, c.RoleId, rewardsMap)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errmsg.NewErrorGuildBossAlreadyRewards()
	}

	currRewards, otherRewards := this_.GetRewardInfo(c, guildDayId, gbfi, gbui, gbs)
	return &servicepb.GuildBoss_RewardOtherDayResponse{
		CurrDayRewards:  currRewards,
		OtherDayRewards: otherRewards,
	}, nil

}

func MapToItems(m map[values.Integer]values.Integer) []*models.Item {
	items := make([]*models.Item, 0, len(m))
	for k, v := range m {
		items = append(items, &models.Item{
			ItemId: k,
			Count:  v,
		})
	}
	return items
}

func (this_ *Service) CanEnter(c *ctx.Context, _ *servicepb.GuildBoss_CanEnterRequest) (*servicepb.GuildBoss_CanEnterResponse, *errmsg.ErrMsg) {
	today := this_.CurrDayRefreshUnix(c)
	guildId, err := this_.module.GuildService.GetGuildIdByRole(c)
	if err != nil {
		return nil, err
	}
	todayStr := strconv.Itoa(int(today))
	guildDayId := guildId + ":" + todayStr
	fi, err := dao.GetGuildBossUserFightInfo(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	otherGuildUsedCount := int64(0)
	if len(fi.GuildDayIds) > 0 {
		lastGuildDayIds := fi.GuildDayIds[len(fi.GuildDayIds)-1]
		strS := strings.Split(lastGuildDayIds, ":")
		if strS[1] == todayStr && strS[0] != guildId {
			lastGB, err := dao.GetGuildBoss(c, lastGuildDayIds)
			if err != nil {
				return nil, err
			}
			damage, ok := lastGB.Damages[c.RoleId]
			if ok {
				otherGuildUsedCount = damage.Count
			}
		}
	}

	gb, err := dao.GetGuildBoss(c, guildDayId)
	if err != nil {
		return nil, err
	}
	totalDamage := int64(0)
	if gb != nil {
		totalDamage = gb.TotalDamage
	}
	guildMaxPlayer, err := this_.module.GuildService.GetGuildMaxMemberCount(c, guildId)
	if err != nil {
		return nil, err
	}
	r := rule.MustGetReader(c)
	maxEnter, _ := r.KeyValue.GetInt64("GuildBossTotal")
	maxEnter += guildMaxPlayer
	cannotEnterSeconds, ok := r.KeyValue.GetInt("GuildBossNo")
	if !ok {
		cannotEnterSeconds = 15 * 60
	}

	nextDayFlush := timer.NextDay(this_.module.RefreshService.GetCurrDayFreshTime(c))
	now := time.Unix(0, c.StartTime).UTC()
	remainTime := nextDayFlush.Sub(now)
	if remainTime < time.Duration(cannotEnterSeconds)*time.Second || maxEnter <= int64(len(gb.Damages)) {
		return &servicepb.GuildBoss_CanEnterResponse{
			Can:           false,
			RemainSeconds: int64(remainTime.Seconds()),
			TotalDamage:   totalDamage,
		}, nil
	}
	//remainTime -= time.Duration(cannotEnterSeconds) * time.Second

	d := gb.Damages[c.RoleId]
	can := false
	onePlayerMaxCount, ok := r.KeyValue.GetInt64("GuildBossNum")
	if !ok {
		onePlayerMaxCount = 1
	}
	rec := onePlayerMaxCount
	usedCount := otherGuildUsedCount
	if d != nil {
		usedCount += d.Count
	}
	if usedCount < onePlayerMaxCount {
		can = true
	} else {
		rec = onePlayerMaxCount - usedCount
	}

	return &servicepb.GuildBoss_CanEnterResponse{
		Can:           can,
		RemainSeconds: int64(remainTime.Seconds()),
		//RefreshRemain: int64(remainTime.Seconds()),
		TotalDamage:      totalDamage,
		RemainEnterCount: rec,
	}, nil
}

func (this_ *Service) GetBattleServerInfo(c *ctx.Context, _ *servicepb.GuildBoss_GetGBBattleServerInfoRequest) (*servicepb.GuildBoss_GetGBBattleServerInfoResponse, *errmsg.ErrMsg) {
	canEnter, err := this_.CanEnter(c, &servicepb.GuildBoss_CanEnterRequest{})
	if err != nil {
		return nil, err
	}
	resp := &servicepb.GuildBoss_GetGBBattleServerInfoResponse{
		Can:           canEnter.Can,
		RemainSeconds: canEnter.RemainSeconds,
		TotalDamage:   canEnter.TotalDamage,
		RefreshRemain: canEnter.RefreshRemain,
		MapSceneId:    0,
		BattleId:      0,
	}
	if !canEnter.Can {
		return resp, nil
	}
	today := this_.CurrDayRefreshUnix(c)
	guildId, err := this_.module.GuildService.GetGuildIdByRole(c)
	if err != nil {
		return nil, err
	}
	guildDayId := guildId + ":" + strconv.Itoa(int(today))
	gb, err := dao.GetGuildBoss(c, guildDayId)
	if err != nil {
		return nil, err
	}
	out := &newcenterpb.NewCenter_CurGuildBossInfoResponse{}
	damages := make(map[string]int64, len(gb.Damages))
	for k, v := range gb.Damages {
		damages[k] = v.Damage
	}
	err = this_.svc.GetNatsClient().RequestWithOut(c, env.GetCenterServerId(), &newcenterpb.NewCenter_CurGuildBossInfoRequest{
		MapId:        5,
		UnionId:      guildId,
		UnionBossId:  guildDayId,
		TotalDamages: gb.TotalDamage,
		Damages:      damages,
	}, out)
	if err != nil {
		return nil, err
	}
	resp.BattleId = out.BattleServerId
	resp.MapSceneId = 5
	if out.IsNew {
		role, err := this_.module.GetRoleByRoleId(c, c.RoleId)
		if err != nil {
			return nil, err
		}
		if role.GameId > 0 {
			this_.log.Info("iggsdk push msg", zap.String("igg_id", role.UserId), zap.Int64("game_id", role.GameId))
			language, ok := rule.MustGetReader(c).LanguageBackend.GetLanguageBackendById(1)
			if ok {
				context := language.GetContext(iggsdk.GetPushIns().GetLag(role.GameId))
				if len(context) > 0 {
					iggsdk.GetPushIns().SendMsg(role.GameId, role.UserId, context)
				}
			}
		}
	}
	return resp, nil
}

func (this_ *Service) QueryRank(c *ctx.Context, _ *servicepb.GuildBoss_QueryRankRequest) (*servicepb.GuildBoss_QueryRankResponse, *errmsg.ErrMsg) {
	today := this_.CurrDayRefreshUnix(c)
	guildId, err := this_.module.GuildService.GetGuildIdByRole(c)
	if err != nil {
		return nil, err
	}
	guildDayId := guildId + ":" + strconv.Itoa(int(today))
	gb, err := dao.GetGuildBoss(c, guildDayId)
	if err != nil {
		return nil, err
	}

	nextDayFlush := timer.NextDay(this_.module.RefreshService.GetCurrDayFreshTime(c))
	now := time.Unix(0, c.StartTime).UTC()
	remainTime := nextDayFlush.Sub(now)
	if len(gb.Damages) == 0 {
		return &servicepb.GuildBoss_QueryRankResponse{
			Ranks:      nil,
			RemainTime: int64(remainTime.Seconds()),
		}, nil
	}
	roleIds := make([]string, 0, len(gb.Damages))
	for k := range gb.Damages {
		roleIds = append(roleIds, k)
	}

	roles, err := this_.module.UserService.GetRole(c, roleIds)
	if err != nil {
		return nil, err
	}
	rankInfos := make([]*servicepb.GuildBoss_RankInfo, 0, len(gb.Damages))
	for k, v := range gb.Damages {
		if v.Damage <= 0 {
			continue
		}
		ri := &servicepb.GuildBoss_RankInfo{RoleId: k, SelfDamage: v.Damage}
		role, ok := roles[k]
		if ok {
			ri.AvatarId = role.AvatarId
			ri.AvatarFrame = role.AvatarFrame
			ri.Lv = role.Level
			ri.NickName = role.Nickname
			ri.Combat = role.Power
		}
		rankInfos = append(rankInfos, ri)
	}
	return &servicepb.GuildBoss_QueryRankResponse{
		Ranks:      rankInfos,
		RemainTime: int64(remainTime.Seconds()),
	}, nil
}

func (this_ *Service) QueryOnlineInfo(c *ctx.Context, _ *servicepb.GuildBoss_OnlineCountRequest) (*servicepb.GuildBoss_OnlineCountResponse, *errmsg.ErrMsg) {
	today := this_.CurrDayRefreshUnix(c)
	guildId, err := this_.module.GuildService.GetGuildIdByRole(c)
	if err != nil {
		return nil, err
	}
	guildDayId := guildId + ":" + strconv.Itoa(int(today))
	req := &newcenterpb.NewCenter_UnionBossOnlineCountRequest{
		UnionBossId: guildDayId,
	}
	out := &newcenterpb.NewCenter_UnionBossOnlineCountResponse{}
	err = this_.svc.GetNatsClient().RequestWithOut(c, env.GetCenterServerId(), req, out)
	if err != nil {
		return nil, err
	}
	return &servicepb.GuildBoss_OnlineCountResponse{Count: out.Count}, nil
}

func (this_ *Service) GetSelfDamageAndRank(c *ctx.Context, _ *servicepb.GuildBoss_GetSelfDamageAndRankRequest) (*servicepb.GuildBoss_GetSelfDamageAndRankResponse, *errmsg.ErrMsg) {
	today := this_.CurrDayRefreshUnix(c)
	guildId, err := this_.module.GuildService.GetGuildIdByRole(c)
	if err != nil {
		return nil, err
	}
	guildDayId := guildId + ":" + strconv.Itoa(int(today))
	gb, err := dao.GetGuildBoss(c, guildDayId)
	if err != nil {
		return nil, err
	}

	if gb.Damages == nil {
		return &servicepb.GuildBoss_GetSelfDamageAndRankResponse{}, nil
	}
	selfDamage, ok := gb.Damages[c.RoleId]
	if !ok || selfDamage.Damage <= 0 {
		return &servicepb.GuildBoss_GetSelfDamageAndRankResponse{}, nil
	}
	selfDamageAddr := unsafe.Pointer(selfDamage)
	damages := make([]*daopb.GuildBossDamageAndFightCount, 0, len(gb.Damages))
	for _, v := range gb.Damages {
		damages = append(damages, v)
	}
	sort.Slice(damages, func(i, j int) bool {
		return damages[i].Damage > damages[j].Damage
	})
	for i, v := range damages {
		if unsafe.Pointer(v) == selfDamageAddr {
			return &servicepb.GuildBoss_GetSelfDamageAndRankResponse{
				SelfDamage: v.Damage,
				Rank:       int64(i + 1),
			}, nil
		}
	}
	return &servicepb.GuildBoss_GetSelfDamageAndRankResponse{}, nil
}
