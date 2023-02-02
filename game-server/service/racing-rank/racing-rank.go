package racing_rank

import (
	"sort"
	"strconv"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/proto/racingrank_service"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/racing-rank/dao"
	"coin-server/game-server/service/racing-rank/rule"
	rule2 "coin-server/rule"
	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewRacingRank(
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
		log:        log,
		Module:     module,
	}
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("获取竞速赛信息", svc.Info)
	svc.svc.RegisterFunc("报名参赛", svc.Enroll)
	svc.svc.RegisterFunc("领取排名奖励", svc.GetRankReward)

	eventlocal.SubscribeEventLocal(svc.HandleLogin)
}

func (svc *Service) Info(ctx *ctx.Context, _ *servicepb.RacingRank_GetRacingRankInfoRequest) (*servicepb.RacingRank_GetRacingRankInfoResponse, *errmsg.ErrMsg) {
	status, err := dao.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	// saveStatus, err := svc.handleSettlement(ctx, status)
	// if err != nil {
	// 	return nil, err
	// }
	res := &servicepb.RacingRank_GetRacingRankInfoResponse{
		Enrolled:     status.Enrolled,
		Season:       status.Season,
		RewardedRank: status.RewardedRank,
	}
	if status.Enrolled {
		// role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
		// if err != nil {
		// 	return nil, err
		// }
		// res.DailyReward = svc.canGetDailyReward(ctx, status, role)

		// 已领取了全部奖励，且赛季已结束
		if svc.isGetAllReward(ctx, status) && status.EndTime <= timer.StartTime(ctx.StartTime).UnixMilli() {
			res.Enrolled = false
			res.HighestRank = 0

			status.Enrolled = false
			status.HighestRank = 0
			dao.SaveStatus(ctx, status)
		} else {
			res.EndTime = status.EndTime
			list, err := svc.getRankData(ctx, status)
			if err != nil {
				return nil, err
			}
			if err := svc.getFashion(ctx, list); err != nil {
				return nil, err
			}
			res.List = list

			var rank values.Integer
			for i := 0; i < len(list); i++ {
				if list[i].RoleId == ctx.RoleId {
					rank = values.Integer(i)
					break
				}
			}
			rank++
			if status.HighestRank == 0 || rank < status.HighestRank {
				status.HighestRank = rank
				dao.SaveStatus(ctx, status)
			}
			res.HighestRank = status.HighestRank
			svc.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskRacingRecord, 0, rank, true)
		}
	}
	// else if saveStatus {
	// 	if err := dao.SaveStatus(ctx, status); err != nil {
	// 		return nil, err
	// 	}
	// }
	return res, nil
}

func (svc *Service) Enroll(ctx *ctx.Context, _ *servicepb.RacingRank_EnrollRacingRankRequest) (*servicepb.RacingRank_EnrollRacingRankResponse, *errmsg.ErrMsg) {
	status, err := dao.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	// _, err = svc.handleSettlement(ctx, status)
	// if err != nil {
	// 	return nil, err
	// }
	if status.Enrolled {
		return nil, errmsg.NewErrRacingRankDoNotRepeatEnroll()
	}
	duration := rule.GetRacingRankDuration(ctx)
	if duration <= 0 {
		return nil, errmsg.NewInternalErr("CombatBattleDay is zero")
	}
	count := rule.GetRankCount(ctx)
	if count <= 1 {
		return nil, errmsg.NewInternalErr("racing rank CombatBattleNum is invalid value")
	}
	now := timer.StartTime(ctx.StartTime)
	status.Enrolled = true
	status.NextRefresh = rule.GetNextRefresh(ctx)
	status.HighestRank = 0
	status.RewardedRank = nil
	status.EnrollTime = now.UnixMilli()
	status.EndTime = now.Add(duration).UnixMilli()
	// status.Info = &pbdao.DailyRewardGetInfo{
	// 	NextGetTime: timer.NextDay(svc.GetCurrDayFreshTime(ctx)).UnixMilli(),
	// }
	self, highestPower, err := svc.getSelfData(ctx)
	if err != nil {
		return nil, err
	}
	status.Season++

	svc.startMatching(&pbdao.RacingRankMatch{
		RoleId:      ctx.RoleId,
		CombatValue: highestPower,
		Count:       count - 1,
		EndTime:     status.EndTime,
		Self:        self,
	}, status)

	// // 将匹配信息发送至kafka
	// if err := racingrank.Emitting(ctx, &pbdao.RacingRankMatch{
	// 	RoleId:      ctx.RoleId,
	// 	CombatValue: highestPower,
	// 	Count:       count - 1,
	// 	EndTime:     status.EndTime,
	// 	Self:        self,
	// }); err != nil {
	// 	return nil, err
	// }

	dao.SaveStatus(ctx, status)
	dao.SaveData(ctx, &pbdao.RacingRankData{
		RoleId: ctx.RoleId,
		List:   []*models.RankItem{self},
		Locked: false,
	})

	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskRacingNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskRacingNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskRacingNum: {
			Typ:     values.Integer(models.TaskType_TaskRacingNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	svc.UpdateTargets(ctx, ctx.RoleId, tasks)

	return &servicepb.RacingRank_EnrollRacingRankResponse{
		Self:    self,
		EndTime: status.EndTime,
	}, nil
}

func (svc *Service) GetRankReward(ctx *ctx.Context, req *servicepb.RacingRank_GetRacingRankRewardRequest) (*servicepb.RacingRank_GetRacingRankRewardResponse, *errmsg.ErrMsg) {
	status, err := dao.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	// role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	// if err != nil {
	// 	return nil, err
	// }
	for _, id := range status.RewardedRank {
		if id == req.Id {
			return nil, errmsg.NewErrRacingRankDoNotRepeatGetReward()
		}
	}
	cfg, ok := rule.GetRankRewardById(ctx, req.Id)
	if !ok {
		return nil, errmsg.NewErrRacingRankDidNotReachRanking()
	}
	if status.HighestRank > cfg.SubsectionRank[1] {
		return nil, errmsg.NewErrRacingRankDidNotReachRanking()
	}
	if req.Id > rule.GetMaxRanking(ctx) {
		return nil, errmsg.NewErrRacingRankDidNotReachRanking()
	}
	// if !svc.canGetDailyReward(ctx, status, role) {
	// 	return nil, errmsg.NewErrRacingRankDoNotRepeatGetReward()
	// }
	rewards := make(map[values.ItemId]values.Integer)
	// items, ok := rule.GetConfigByTitle(ctx, role.Title)
	// if ok && len(items) > 0 {
	// 	for id, count := range items {
	// 		rewards[id] += count
	// 	}
	// 	if _, err := svc.AddManyItem(ctx, ctx.RoleId, rewards); err != nil {
	// 		return nil, err
	// 	}
	// }
	// now := timer.StartTime(ctx.StartTime)
	// nextGetTime := timer.StartTime(status.Info.NextGetTime * values.Integer(time.Millisecond))
	// for nextGetTime.UnixMilli() <= now.UnixMilli() {
	// 	nextGetTime = timer.NextDay(svc.GetCurrDayFreshTime(ctx))
	// }
	// status.Info = &pbdao.DailyRewardGetInfo{
	// 	NextGetTime: nextGetTime.UnixMilli(),
	// 	Title:       role.Title,
	// }

	// cfg := rule.GetRankReward(ctx, req.Rank)
	for id, val := range cfg.RankReward {
		rewards[id] += val
	}
	if _, err := svc.AddManyItem(ctx, ctx.RoleId, rewards); err != nil {
		return nil, err
	}
	status.RewardedRank = append(status.RewardedRank, req.Id)

	if svc.isGetAllReward(ctx, status) && status.EndTime <= timer.StartTime(ctx.StartTime).UnixMilli() {
		status.Enrolled = false
	}

	dao.SaveStatus(ctx, status)

	return &servicepb.RacingRank_GetRacingRankRewardResponse{
		Rewards:      rewards,
		RewardedRank: status.RewardedRank,
	}, nil
}

func (svc *Service) handleSettlement(ctx *ctx.Context, status *pbdao.RacingRankStatus) (bool, *errmsg.ErrMsg) {
	if status.EndTime <= 0 || status.EndTime > timer.StartTime(ctx.StartTime).UnixMilli() {
		return false, nil
	}
	data, err := dao.GetData(ctx)
	if err != nil {
		return false, err
	}
	if len(data.List) <= 0 {
		return true, nil
	}
	rank := -1
	for i := 0; i < len(data.List); i++ {
		if data.List[i].RoleId == ctx.RoleId {
			rank = i
			break
		}
	}
	rank++
	if rank <= 0 {
		status.Enrolled = false
		status.EnrollTime = 0
		status.EndTime = 0
		return true, nil
	}
	cfg := rule.GetRankReward(ctx, values.Integer(rank))
	attachment := make([]*models.Item, 0)
	for id, count := range cfg.RankReward {
		attachment = append(attachment, &models.Item{
			ItemId: id,
			Count:  count,
		})
	}
	id := rule.GetMailId(ctx)
	var expiredAt values.Integer
	expire := rule.GetMailExpire(ctx, id)
	if expire > 0 {
		expiredAt = timer.StartTime(ctx.StartTime).Add(time.Second * time.Duration(expire)).UnixMilli()
	}
	if err := svc.MailService.Add(ctx, ctx.RoleId, &models.Mail{
		Type:       models.MailType_MailTypeSystem,
		TextId:     id,
		ExpiredAt:  expiredAt,
		Args:       []string{strconv.Itoa(rank)},
		Attachment: attachment,
	}); err != nil {
		return false, err
	}

	status.Enrolled = false
	status.EnrollTime = 0
	status.EndTime = 0
	return true, nil
}

// 检查匹配是否完成，若未完成（消息未成功发送至kafka或者中间丢消息之类的问题）则直接将玩家的状态置为未报名状态，让玩家再重新报名
func (svc *Service) matchCheck(ctx *ctx.Context, status *pbdao.RacingRankStatus) *errmsg.ErrMsg {
	if !status.Enrolled {
		return nil
	}
	data, err := dao.GetData(ctx)
	if err != nil {
		return err
	}
	// 匹配了10分钟还未完成则认为匹配失败（丢消息了）
	max := values.Integer(10 * 60 * 1000)
	// 列表里只有一个人则未匹配成功
	if len(data.List) <= 1 && status.EnrollTime > 0 && timer.StartTime(ctx.StartTime).UnixMilli()-status.EnrollTime >= max {
		status = &pbdao.RacingRankStatus{
			RoleId:       ctx.RoleId,
			Enrolled:     false,
			NextRefresh:  0,
			Season:       0,
			HighestRank:  0,
			RewardedRank: nil,
			EnrollTime:   0,
			EndTime:      0,
		}
		dao.SaveStatus(ctx, status)
	}
	return nil
}

// func (svc *Service) canGetDailyReward(ctx *ctx.Context, status *pbdao.RacingRankStatus, role *pbdao.Role) bool {
// 	if !status.Enrolled {
// 		return false
// 	}
// 	return status.Info == nil || status.Info.NextGetTime <= timer.StartTime(ctx.StartTime).UnixMilli() || status.Info.Title != role.Title
// }

func (svc *Service) getRankData(ctx *ctx.Context, status *pbdao.RacingRankStatus) ([]*models.RankItem, *errmsg.ErrMsg) {
	data, err := dao.GetData(ctx)
	if err != nil {
		return nil, err
	}
	now := timer.StartTime(ctx.StartTime).UnixMilli()
	// 里面只有1个元素的时候表示只有玩家自己，这时候一定还未匹配完成，不刷新排行榜
	if !data.ForceRefresh && (data.Locked || status.EndTime <= now || status.NextRefresh > now || len(data.List) == 1) {
		return data.List, nil
	}
	// 刷新排行榜
	roleIds := make([]values.RoleId, 0, len(data.List))
	for _, item := range data.List {
		roleIds = append(roleIds, item.RoleId)
	}
	roles, err := svc.GetRole(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	guilds, err := svc.GetMultiGuildByRoleId(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	list := make([]*models.RankItem, 0, len(roleIds))
	var selfExist bool
	for roleId, role := range roles {
		if roleId == ctx.RoleId {
			selfExist = true
		}
		item := &models.RankItem{
			RoleId:      roleId,
			Nickname:    role.Nickname,
			Level:       role.Level,
			AvatarId:    role.AvatarId,
			AvatarFrame: role.AvatarFrame,
			Power:       role.Power,
			// GuildId:     "",
			GuildName: "",
		}
		guild, ok := guilds[roleId]
		if ok {
			item.GuildName = guild.Name
		}
		list = append(list, item)
	}
	// TODO 目前可能会出现自己数据丢失的BUG，具体原因未知，这是临时解决方案
	if !selfExist {
		self, _, err := svc.getSelfData(ctx)
		if err != nil {
			return nil, err
		}
		list = append(list, self)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Power > list[j].Power
	})

	data.ForceRefresh = false
	data.List = list
	// 更新排行榜下次刷新时间
	status.NextRefresh = rule.GetNextRefresh(ctx)
	dao.SaveStatus(ctx, status)
	dao.SaveData(ctx, data)

	return list, nil
}

func (svc *Service) taskCheck(ctx *ctx.Context, endTime values.Integer) *errmsg.ErrMsg {
	if _, err := svc.svc.GetNatsClient().Request(ctx, 0, &racingrank_service.RacingRankService_CronTaskCheckRequest{
		EndTime: endTime,
	}); err != nil {
		return err
	}
	return nil
}

func (svc *Service) getSelfData(ctx *ctx.Context) (*models.RankItem, values.Integer, *errmsg.ErrMsg) {
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, 0, err
	}
	guild, err := svc.GetUserGuildInfo(ctx, ctx.RoleId)
	if err != nil {
		return nil, 0, err
	}
	self := &models.RankItem{
		RoleId:      ctx.RoleId,
		Nickname:    role.Nickname,
		Level:       role.Level,
		AvatarId:    role.AvatarId,
		AvatarFrame: role.AvatarFrame,
		Power:       role.Power,
	}
	if guild != nil {
		// self.GuildId = guild.Id 不需要公会id
		self.GuildName = guild.Name
	}

	return self, role.HighestPower, nil
}

func (svc *Service) isGetAllReward(ctx *ctx.Context, status *pbdao.RacingRankStatus) bool {
	if !status.Enrolled || status.HighestRank <= 0 {
		return true
	}
	list := rule2.MustGetReader(ctx).CombatBattle.List()
	tempMap := make(map[values.Integer]struct{})
	for _, battle := range list {
		if status.HighestRank <= battle.SubsectionRank[1] {
			tempMap[battle.Id] = struct{}{}
		}
	}
	for _, rank := range status.RewardedRank {
		delete(tempMap, rank)
	}
	return len(tempMap) == 0
}

func (svc *Service) startMatching(self *pbdao.RacingRankMatch, status *pbdao.RacingRankStatus) {
	svc.svc.AfterFunc(time.Millisecond*5, func(ctx *ctx.Context) {
		if err := svc.matching(self); err != nil {
			svc.log.Error("racing rank matching err:", zap.String("role_id", self.RoleId), zap.Error(err))

			// 匹配失败，将玩家状态更新为未报名状态
			status.Enrolled = false
			status.Season--
			if err := dao.SaveStatusImmediately(status); err != nil {
				svc.log.Error("racing rank matching err save status err:", zap.String("role_id", self.RoleId), zap.Error(err))
			}
			return
		}
		ctx.PushMessageToRole(
			self.RoleId,
			&servicepb.RacingRank_RacingRankMatchSuccessPush{},
		)
	})
}

func (svc *Service) getFashion(ctx *ctx.Context, list []*models.RankItem) *errmsg.ErrMsg {
	count := len(list)
	if count <= 1 {
		return nil
	}
	// 只取前3名的时装
	if count > 3 {
		count = 3
	}
	for i := 0; i < count; i++ {
		fashionId, err := svc.getOneFashion(ctx, list[i].RoleId)
		if err != nil {
			return err
		}
		list[i].Fashion = fashionId
	}
	return nil
}

func (svc *Service) getOneFashion(ctx *ctx.Context, roleId values.RoleId) (values.FashionId, *errmsg.ErrMsg) {
	f, err := svc.GetDefaultHeroes(ctx, roleId)
	if err != nil {
		return 0, err
	}
	h, ok, err := svc.GetHero(ctx, roleId, f.HeroOrigin_0)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, errmsg.NewErrHeroNotFound()
	}
	return h.Fashion.Dressed, nil
}
