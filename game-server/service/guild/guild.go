package guild

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/gopool"
	"coin-server/common/idgenerate"
	"coin-server/common/im"
	"coin-server/common/logger"
	"coin-server/common/orm"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/guild_filter_service"
	"coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	"coin-server/common/proto/rank_service"
	"coin-server/common/redisclient"
	"coin-server/common/sensitive"
	"coin-server/common/service"
	"coin-server/common/timer"
	"coin-server/common/utils/imutil"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/guild/dao"
	"coin-server/game-server/service/guild/rule"
	rule2 "coin-server/rule"
	rulemodel "coin-server/rule/rule-model"

	"github.com/golang-jwt/jwt"
	json "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type Service struct {
	rankId     values.RankId
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewGuildService(
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
	module.GuildService = s
	return s
}

func (svc *Service) Router() *Service {
	svc.svc.RegisterFunc("查找公会", svc.Find)
	svc.svc.RegisterFunc("一键加入", svc.JoinNow)
	svc.svc.RegisterFunc("通过id获取公会详细信息", svc.FindDetails)
	svc.svc.RegisterFunc("进入公会场景", svc.Enter)
	svc.svc.RegisterFunc("创建公会", svc.Create)
	svc.svc.RegisterFunc("解散公会", svc.Dissolve)
	svc.svc.RegisterFunc("修改公会信息", svc.Modify)
	svc.svc.RegisterFunc("获取公会所有成员", svc.Members)
	// svc.svc.RegisterFunc("邀请入会", svc.InviteJoin)
	svc.svc.RegisterFunc("获取可邀请的玩家列表", svc.InviteList)
	svc.svc.RegisterFunc("处理入会邀请", svc.HandleInvite)
	svc.svc.RegisterFunc("申请加入公会", svc.JoinApply)
	svc.svc.RegisterFunc("取消入会申请", svc.CancelApply)
	svc.svc.RegisterFunc("获取入会申请列表", svc.ApplyList)
	svc.svc.RegisterFunc("处理入会申请", svc.HandleApply)
	svc.svc.RegisterFunc("拒绝所有", svc.RejectAll)
	svc.svc.RegisterFunc("将玩家踢出公会", svc.Remove)
	svc.svc.RegisterFunc("退出公会", svc.Exit)
	svc.svc.RegisterFunc("职位变更", svc.PositionChange)
	// svc.svc.RegisterFunc("降职", svc.Demotion)
	// svc.svc.RegisterFunc("移交会长", svc.LeaderChange)
	svc.svc.RegisterFunc("获取公会信息", svc.GuildInfo)
	svc.svc.RegisterFunc("自定义职位名称", svc.ModifyPositionName)
	svc.svc.RegisterFunc("获取玩家公会建设信息", svc.BuildInfo)
	svc.svc.RegisterFunc("建设公会", svc.Build)
	svc.svc.RegisterFunc("世界邀请", svc.WorldInvite)
	svc.svc.RegisterFunc("个人邀请", svc.PrivateInvite)
	svc.svc.RegisterFunc("获取公会排行榜", svc.Rank)
	svc.svc.RegisterFunc("公会名是否存在", svc.NameExist)

	// 公会祝福
	bless := svc.svc.Group(svc.isJoinedGuild)
	bless.RegisterFunc("获取公会祝福信息", svc.BlessingInfo)
	bless.RegisterFunc("开始祝福", svc.BlessingStart)
	bless.RegisterFunc("激活祝福", svc.BlessingActivate)
	bless.RegisterFunc("切换到下一阶", svc.BlessingNextStage)

	// 作弊器
	svc.svc.RegisterFunc("重置建设次数", svc.CheatResetBuildTimes)
	svc.svc.RegisterFunc("增加公会经验", svc.CheatAddGuildExp)
	svc.svc.RegisterFunc("清除退会cd", svc.CheatRestExitGuildCD)
	svc.svc.RegisterFunc("设置公会祝福stage和page", svc.CheatSetBlessStage)

	eventlocal.SubscribeEventLocal(svc.HandlerUserDailyActiveUpdate)
	eventlocal.SubscribeEventLocal(svc.UserCombatValueChange)

	// 仅GVGGuildServer调用
	svc.svc.RegisterFunc("更新公会祝福加成", svc.UpdateBlessingEffic)

	return svc
}

func (svc *Service) InitGuildRankId() {
	rankId, err := dao.GetRankId()
	if err != nil {
		panic(fmt.Errorf("InitGuildRankId err: %s", err.Error()))
	}
	if rankId == "" {
		out := &rank_service.RankService_CreateRankResponse{}
		if err := svc.svc.GetNatsClient().RequestWithHeaderOut(ctx.GetContext(), 0, &models.ServerHeader{
			ServerType: models.ServerType_GameServer,
		}, &rank_service.RankService_CreateRankRequest{
			RankType: values.Integer(enum.RankGuild),
		}, out); err != nil {
			panic(fmt.Errorf("InitGuildRankId create rank err: %s", err.Error()))
		}
		rankId = out.RankId
		if err := dao.SaveRankId(rankId); err != nil {
			panic(fmt.Errorf("InitGuildRankId save rankId err: %s", err.Error()))
		}
	}
	svc.rankId = rankId
}

func (svc *Service) GetGuildIdByRole(ctx *ctx.Context) (values.GuildId, *errmsg.ErrMsg) {
	user, err := dao.NewGuildUser(ctx.RoleId).Get(ctx)
	if err != nil {
		return "", err
	}
	return user.GuildId, nil
}

func (svc *Service) IsUnlockGuildBoss(ctx *ctx.Context) *errmsg.ErrMsg {
	guildId, err := svc.GetGuildIdByRole(ctx)
	if err != nil {
		return err
	}
	guild, err := dao.NewGuild(guildId).Get(ctx)
	if err != nil {
		return err
	}
	if guild == nil {
		return errmsg.NewErrGuildNotExist()
	}
	guildCfg, ok := rule2.MustGetReader(ctx).Guild.GetGuildById(guild.Level)
	if !ok {
		return errmsg.NewInternalErr(fmt.Sprintf("guild config not found:%d", guild.Level))
	}
	for _, v := range guildCfg.FunctionOpen {
		if v == 4 {
			return nil
		}
	}
	return errmsg.NewErrSystemLock()
}

func (svc *Service) GetUserGuildPositionByGuildId(ctx *ctx.Context, guildId values.GuildId) (values.GuildPosition, *errmsg.ErrMsg) {
	if guildId == "" {
		return 0, nil
	}
	member, ok, err := dao.NewGuildMember(guildId).GetOne(ctx, ctx.RoleId)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}
	return member.Position, nil
}

func (svc *Service) GetUserGuildInfo(ctx *ctx.Context, roleId values.RoleId) (*pbdao.Guild, *errmsg.ErrMsg) {
	if roleId == "" {
		roleId = ctx.RoleId
	}
	user, err := dao.NewGuildUser(roleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if user.GuildId == "" {
		return nil, nil
	}
	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	return guild, nil
}

func (svc *Service) GetMultiGuildByRoleId(ctx *ctx.Context, roleIds []values.RoleId) (map[values.RoleId]*pbdao.Guild, *errmsg.ErrMsg) {
	guildUsers, err := dao.NewGuildUser("").GetMulti(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	roleGuildMap := make(map[values.RoleId]values.GuildId, 0)
	guildIds := make([]values.GuildId, 0)
	for _, user := range guildUsers {
		if user.GuildId != "" {
			guildIds = append(guildIds, user.GuildId)
			roleGuildMap[user.RoleId] = user.GuildId
		}
	}
	guildMap, err := dao.NewGuild("").GetMulti(ctx, guildIds)
	if err != nil {
		return nil, err
	}
	ret := make(map[values.RoleId]*pbdao.Guild, 0)
	for roleId, guildId := range roleGuildMap {
		ret[roleId] = guildMap[guildId]
	}
	return ret, nil
}

func (svc *Service) GetGuildMaxMemberCount(ctx *ctx.Context, guildId values.GuildId) (values.Integer, *errmsg.ErrMsg) {
	guild, err := dao.NewGuild(guildId).Get(ctx)
	if err != nil {
		return 0, err
	}
	if guild == nil {
		return 0, errmsg.NewErrGuildNotExist()
	}
	return rule.GetMaxMemberCount(ctx, guild.Level), nil
}

func (svc *Service) Find(ctx *ctx.Context, req *less_service.Guild_GuildFindRequest) (*less_service.Guild_GuildFindResponse, *errmsg.ErrMsg) {
	out := &guild_filter_service.Guild_GuildFilterFindResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, 0, &guild_filter_service.Guild_GuildFilterFindRequest{
		Name:        req.Name,
		Lang:        req.Lang,
		NoAudit:     req.NoAudit,
		CombatValue: req.CombatValue,
	}, out); err != nil {
		return nil, err
	}
	list, err := svc.getGuildDetails(ctx, out.Id)
	if err != nil {
		return nil, err
	}
	return &less_service.Guild_GuildFindResponse{
		List: list,
	}, nil
}

func (svc *Service) JoinNow(ctx *ctx.Context, req *less_service.Guild_GuildJoinNowRequest) (*less_service.Guild_GuildJoinNowResponse, *errmsg.ErrMsg) {
	user, err := dao.NewGuildUser(ctx.RoleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if user.GuildId != "" {
		return nil, errmsg.NewErrGuildHasJoined()
	}
	if svc.isInCooling(ctx, user) {
		return nil, errmsg.NewErrGuildJoinCD()
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if role.Level < rule.GetGuildUserLevelLimit(ctx) {
		return nil, errmsg.NewErrGuildUserLevelNotEnough()
	}

	out := &guild_filter_service.Guild_GuildFilterCanJoinGuildResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, 0, &guild_filter_service.Guild_GuildFilterCanJoinGuildRequest{
		Lang:        req.Lang,
		CombatValue: req.CombatValue,
	}, out); err != nil {
		return nil, err
	}
	if out.Id == "" {
		return nil, errmsg.NewErrGuildNoAvailableGuild()
	}

	err = svc.getLock(ctx, guildLock+out.Id)
	if err != nil {
		return nil, err
	}

	guild, err := dao.NewGuild(out.Id).Get(ctx)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, errmsg.NewErrGuildNoAvailableGuild()
	}
	if guild.DeletedAt > 0 {
		return nil, errmsg.NewErrGuildNoAvailableGuild()
	}

	memberModel := dao.NewGuildMember(guild.Id)
	members, err := memberModel.Get(ctx)
	if err != nil {
		return nil, err
	}
	if values.Integer(len(members)) >= rule.GetMaxMemberCount(ctx, guild.Level) {
		return nil, errmsg.NewErrGuildNoAvailableGuild()
	}
	memberPosition := rule.GetMemberPosition(ctx)
	// 更新guild user
	user.GuildId = guild.Id
	rewards, err := svc.handlerFirstJoinRewards(ctx, user)
	if err != nil {
		return nil, err
	}
	if err := dao.NewGuildUser(user.RoleId).Save(ctx, user); err != nil {
		return nil, err
	}
	// 更新member
	member := &pbdao.GuildMember{
		RoleId:           user.RoleId,
		ServerId:         ctx.ServerId,
		Position:         memberPosition,
		JoinAt:           timer.StartTime(ctx.StartTime).Unix(),
		LastPosition:     memberPosition,
		CombatValue:      role.Power,
		ActiveValue:      user.ActiveValue,
		TotalActiveValue: user.TotalActiveValue,
	}
	members = append(members, member)
	if err := memberModel.SaveOne(ctx, member); err != nil {
		return nil, err
	}
	// 清除玩家已经申请过的公会记录
	if err := svc.clearUserApplyList(ctx, ctx.RoleId); err != nil {
		return nil, err
	}
	guild.Count = values.Integer(len(members))
	if err := dao.NewGuild(out.Id).Save(ctx, guild); err != nil {
		return nil, err
	}
	svc.joinChatRoom(ctx, ctx.RoleId, guild.Id)
	svc.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskJoinGuild, 0, 1)
	return &less_service.Guild_GuildJoinNowResponse{
		Info:    svc.dao2model(ctx, guild, members),
		Rewards: rewards,
	}, nil
}

func (svc *Service) FindDetails(ctx *ctx.Context, req *less_service.Guild_GuildFindDetailsRequest) (*less_service.Guild_GuildFindDetailsResponse, *errmsg.ErrMsg) {
	if len(req.Id) == 0 || len(req.Id) > 20 {
		return nil, errmsg.NewErrGuildInvalidInput()
	}
	allEmpty := true
	for _, v := range req.Id {
		if v != "" {
			allEmpty = false
			break
		}
	}
	if allEmpty {
		return nil, errmsg.NewErrGuildInvalidInput()
	}
	list, err := svc.getGuildDetails(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &less_service.Guild_GuildFindDetailsResponse{
		List: list,
	}, nil
}

func (svc *Service) Enter(ctx *ctx.Context, _ *less_service.Guild_GuildEnterRequest) (*less_service.Guild_GuildEnterResponse, *errmsg.ErrMsg) {
	user, err := dao.NewGuildUser(ctx.RoleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if user.GuildId == "" {
		return &less_service.Guild_GuildEnterResponse{}, nil
	}

	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	members, err := dao.NewGuildMember(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if err := svc.leaderAutoChange(ctx, guild, members); err != nil {
		return nil, err
	}
	rewards, err := svc.handlerFirstJoinRewards(ctx, user)
	if err != nil {
		return nil, err
	}
	if len(rewards) > 0 {
		if err := dao.NewGuildUser(ctx.RoleId).Save(ctx, user); err != nil {
			return nil, err
		}
	}
	var find *pbdao.GuildMember
	for _, member := range members {
		if member.RoleId == user.RoleId {
			find = member
			break
		}
	}
	if find == nil {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	lastPosition := find.LastPosition
	curPosition := find.Position
	if lastPosition != curPosition {
		find.LastPosition = curPosition
		if err := dao.NewGuildMember(user.GuildId).SaveOne(ctx, find); err != nil {
			return nil, err
		}
	}
	return &less_service.Guild_GuildEnterResponse{
		Info:         svc.dao2model(ctx, guild, members),
		Rewards:      rewards,
		LastPosition: lastPosition,
		CurPosition:  curPosition,
	}, nil
}

func (svc *Service) Create(ctx *ctx.Context, req *less_service.Guild_GuildCreateRequest) (*less_service.Guild_GuildCreateResponse, *errmsg.ErrMsg) {
	ok, err := svc.levelCheck(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrGuildUserLevelNotEnough()
	}
	user, err := dao.NewGuildUser(ctx.RoleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if user.GuildId != "" {
		return nil, errmsg.NewErrGuildHasJoined()
	}

	if err := svc.introCheck(ctx, req.Intro); err != nil {
		return nil, err
	}
	if err := svc.noticeCheck(ctx, req.Notice); err != nil {
		return nil, err
	}
	if !rule.FlagCheck(ctx, req.Flag) {
		return nil, errmsg.NewErrGuildInvalidFlag()
	}
	if !rule.LanguageCheck(ctx, req.Lang) {
		return nil, errmsg.NewErrGuildInvalidLang()
	}

	cost, ok := rule.GetGuildCreateCost(ctx)
	if ok {
		if err := svc.SubManyItemPb(ctx, ctx.RoleId, cost); err != nil {
			return nil, err
		}
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	name, err := svc.nameCheck(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	id, err := idgenerate.GenerateID(context.Background(), idgenerate.GuideIDKey)
	if err != nil {
		return nil, err
	}

	_now := timer.StartTime(ctx.StartTime).Unix()
	guildId := strconv.Itoa(int(id))

	if err := svc.saveGuildId(ctx, guildId); err != nil {
		return nil, err
	}
	guild := &pbdao.Guild{
		Id:        guildId,
		LeaderId:  ctx.RoleId,
		Name:      req.Name,
		Flag:      req.Flag,
		Lang:      req.Lang,
		Intro:     req.Intro,
		Notice:    req.Notice,
		AutoJoin:  req.AutoJoin,
		Level:     1,
		Exp:       0,
		Resources: make(map[int64]int64, 0),
		CreatedAt: _now,
		DeletedAt: 0,
		Count:     1,
	}
	// 创建公会
	if err := dao.NewGuild(guildId).Save(ctx, guild); err != nil {
		return nil, err
	}
	leaderPosition := rule.GetLeaderPosition(ctx)
	all := []*pbdao.GuildMember{{
		RoleId:           ctx.RoleId,
		ServerId:         ctx.ServerId,
		Position:         leaderPosition,
		JoinAt:           _now,
		LastPosition:     leaderPosition,
		CombatValue:      role.Power,
		ActiveValue:      user.ActiveValue,
		TotalActiveValue: user.TotalActiveValue,
	}}
	// 创建公会成员
	if err := dao.NewGuildMember(guildId).Save(ctx, all); err != nil {
		return nil, err
	}

	rewards, err := svc.handlerFirstJoinRewards(ctx, user)
	if err != nil {
		return nil, err
	}
	user.GuildId = guildId
	// 创建公会用户
	if err := dao.NewGuildUser(ctx.RoleId).Save(ctx, user); err != nil {
		return nil, err
	}
	svc.joinChatRoom(ctx, ctx.RoleId, guildId)

	guildModel := svc.dao2model(ctx, guild, all)
	if err := svc.updateToGuildFilterServer(ctx, guildModel); err != nil {
		return nil, err
	}

	if err := svc.updateGuildRankValue(ctx, guildModel.Id, guildModel.CombatValue); err != nil {
		return nil, err
	}
	svc.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskJoinGuild, 0, 1)

	if err := dao.SaveName(&dao.Model{
		Id:   guildId,
		Name: name,
	}); err != nil {
		return nil, err
	}
	svc.joinLocalEvent(ctx, ctx.RoleId, guild.Id)
	return &less_service.Guild_GuildCreateResponse{
		Info:    guildModel,
		Rewards: rewards,
	}, nil
}

func (svc *Service) Dissolve(ctx *ctx.Context, _ *less_service.Guild_GuildDissolveRequest) (*less_service.Guild_GuildDissolveResponse, *errmsg.ErrMsg) {
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	ok, err := svc.IsGuildBossFighting(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, errmsg.NewErrGuildBossFighting()
	}
	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	if guild.LeaderId != ctx.RoleId {
		return nil, errmsg.NewErrGuildOnlyLeader()
	}
	list, err := dao.NewGuildMember(guild.Id).Get(ctx)
	if err != nil {
		return nil, err
	}
	// 只剩会长一人的时候才能解散公会
	if len(list) > 1 {
		return nil, errmsg.NewErrGuildRemoveMemberFirst()
	}

	err = svc.getLock(ctx, guildLock+guild.Id)
	if err != nil {
		return nil, err
	}

	guildApplyDao := dao.NewGuildApply(guild.Id)
	applyList, err := guildApplyDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	if len(applyList) > 0 {
		guildApplyDao.DeleteKey(ctx)
	}
	guild.DeletedAt = timer.StartTime(ctx.StartTime).Unix()

	// 解散公会
	if err := dao.NewGuild(guild.Id).Save(ctx, guild, true); err != nil {
		return nil, err
	}
	// 从member里移除
	member, ok, err := dao.NewGuildMember(user.GuildId).GetOne(ctx, user.RoleId)
	if err != nil {
		return nil, err
	}
	if ok {
		if err := dao.NewGuildMember(user.GuildId).Delete(ctx, member, true); err != nil {
			return nil, err
		}
		svc.syncActive2User(user, member)
	}
	// 更新guild user
	user.GuildId = ""
	if err := dao.NewGuildUser(user.RoleId).Save(ctx, user); err != nil {
		return nil, err
	}

	// 解散公会的时候从公会id集合里删除对应的公会id
	svc.delGuildId(ctx, guild.Id)

	svc.leaveChatRoom(ctx, ctx.RoleId, guild.Id)

	if _, err := svc.svc.GetNatsClient().Request(ctx, 0, &guild_filter_service.Guild_GuildFilterDeleteRequest{
		Id: guild.Id,
	}); err != nil {
		return nil, err
	}
	// 从排行榜里删除
	if err := svc.deleteGuildRankValue(ctx, guild.Id); err != nil {
		return nil, err
	}

	if err := dao.DeleteName(guild.Id); err != nil {
		return nil, err
	}

	svc.exitLocalEvent(ctx, ctx.RoleId, guild.Id)

	return &less_service.Guild_GuildDissolveResponse{}, nil
}

func (svc *Service) Modify(ctx *ctx.Context, req *less_service.Guild_GuildModifyRequest) (*less_service.Guild_GuildModifyResponse, *errmsg.ErrMsg) {
	if req.CombatValueLimit > 9223372036854775807 {
		return nil, errmsg.NewErrGuildInvalidInput()
	}
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	err = svc.getLock(ctx, guildLock+user.GuildId)
	if err != nil {
		return nil, err
	}

	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	member, ok, err := dao.NewGuildMember(guild.Id).GetOne(ctx, user.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	item, ok := rule.GetPermissions(ctx, member.Position)
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}

	if req.Flag != 0 && guild.Flag != req.Flag {
		if !item.ModifyFlag {
			return nil, errmsg.NewErrGuildPermissionDenied()
		}
		if !rule.FlagCheck(ctx, req.Flag) {
			return nil, errmsg.NewErrGuildInvalidFlag()
		}
		guild.Flag = req.Flag
	}
	if req.Lang != 0 && guild.Lang != req.Lang {
		if !item.ModifyFlag {
			return nil, errmsg.NewErrGuildPermissionDenied()
		}
		if !rule.LanguageCheck(ctx, req.Lang) {
			return nil, errmsg.NewErrGuildInvalidLang()
		}
		guild.Lang = req.Lang
	}
	var modifyIntroSendToChat bool
	if req.Intro != "" && guild.Intro != req.Intro {
		if !item.ModifyIntro {
			return nil, errmsg.NewErrGuildPermissionDenied()
		}
		if err := svc.introCheck(ctx, req.Intro); err != nil {
			return nil, err
		}
		guild.Intro = req.Intro
		modifyIntroSendToChat = req.Intro != ""
	}

	if guild.Notice != req.Notice {
		if !item.ModifyNotice {
			return nil, errmsg.NewErrGuildPermissionDenied()
		}
		if err := svc.noticeCheck(ctx, req.Notice); err != nil {
			return nil, err
		}
		guild.Notice = req.Notice
	}
	if req.Greeting != "" && guild.Greeting != req.Greeting {
		if !item.ModifyGreeting {
			return nil, errmsg.NewErrGuildPermissionDenied()
		}
		if err := svc.greetingCheck(ctx, req.Greeting); err != nil {
			return nil, err
		}
		guild.Greeting = req.Greeting
	}
	if req.CombatValueLimit >= 0 && guild.CombatValueLimit != req.CombatValueLimit {
		// 策划说了用HandleApply这个权限
		if !item.HandleApply {
			return nil, errmsg.NewErrGuildPermissionDenied()
		}
		guild.CombatValueLimit = req.CombatValueLimit
	}
	var autoJoin bool
	if req.AutoJoin == 1 {
		autoJoin = true
	}
	if guild.AutoJoin != autoJoin {
		if !item.ModifyAutoJoin {
			return nil, errmsg.NewErrGuildPermissionDenied()
		}
		guild.AutoJoin = autoJoin
	}
	var name string
	// 名称会向布隆过滤器添加，放在最后检查
	if req.Name != "" && guild.Name != req.Name {
		if !item.ModifyName {
			return nil, errmsg.NewErrGuildPermissionDenied()
		}
		name, err = svc.nameCheck(ctx, req.Name)
		if err != nil {
			return nil, err
		}
		guild.Name = name
		var costItemSuccess bool
		// 改名优先使用改名道具
		cost, ok := rule.GetGuildModifyNameItemCost(ctx)
		if ok {
			count, err := svc.GetItem(ctx, ctx.RoleId, cost.ItemId)
			if err != nil {
				return nil, err
			}
			if count >= cost.Count {
				if err := svc.SubItem(ctx, ctx.RoleId, cost.ItemId, cost.Count); err != nil {
					return nil, err
				}
				costItemSuccess = true
			}
		}
		if !costItemSuccess {
			cost, ok = rule.GetGuildModifyNameCost(ctx)
			if ok {
				if err := svc.SubManyItem(ctx, ctx.RoleId, map[values.ItemId]values.Integer{cost.ItemId: cost.Count}); err != nil {
					return nil, err
				}
			}
		}
	}
	members, err := dao.NewGuildMember(guild.Id).Get(ctx)
	if err != nil {
		return nil, err
	}
	if err := dao.NewGuild(guild.Id).Save(ctx, guild); err != nil {
		return nil, err
	}
	guildModel := svc.dao2model(ctx, guild, members)
	if err := svc.updateToGuildFilterServer(ctx, guildModel); err != nil {
		return nil, err
	}

	if modifyIntroSendToChat {
		svc.modifyIntroSysMsg(ctx, guildModel.Id, req.Intro)
	}
	if name != "" {
		if err := dao.SaveName(&dao.Model{
			Id:   guild.Id,
			Name: name,
		}); err != nil {
			return nil, err
		}
	}
	return &less_service.Guild_GuildModifyResponse{
		Info: guildModel,
	}, nil
}

func (svc *Service) Members(ctx *ctx.Context, req *less_service.Guild_GuildMembersRequest) (*less_service.Guild_GuildMembersResponse, *errmsg.ErrMsg) {
	var guildId values.GuildId
	if req.GuildId != "" {
		guildId = req.GuildId
	} else {
		user, err := svc.getGuidUser(ctx, ctx.RoleId)
		if err != nil {
			return nil, err
		}
		guildId = user.GuildId
	}
	members, err := dao.NewGuildMember(guildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	roleIds := make([]values.RoleId, 0, len(members))
	for _, member := range members {
		roleIds = append(roleIds, member.RoleId)
	}
	list := make([]*models.GuildMember, 0, len(members))
	userMap, err := svc.GetRole(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	sevenDaysKey := svc.get7DaysKey(ctx)
	for _, member := range members {
		temp, ok := userMap[member.RoleId]
		if !ok {
			temp = &pbdao.Role{}
		}
		list = append(list, &models.GuildMember{
			RoleId:           member.RoleId,
			ServerId:         member.ServerId,
			Position:         member.Position,
			JoinAt:           member.JoinAt,
			LoginAt:          temp.Login,
			LogoutAt:         temp.Logout,
			Name:             temp.Nickname,
			Level:            temp.Level,
			CombatValue:      temp.Power,
			ActiveValue:      svc.get7DayActive(member, sevenDaysKey),
			TotalActiveValue: member.TotalActiveValue,
			Lang:             temp.Language,
			AvatarId:         temp.AvatarId,
			AvatarFrame:      temp.AvatarFrame,
		})
	}
	return &less_service.Guild_GuildMembersResponse{
		Members: list,
	}, nil
}

// func (svc *Service) InviteJoin(ctx *ctx.Context, req *less_service.Guild_GuildInviteJoinRequest) (*less_service.Guild_GuildInviteJoinResponse, *errmsg.ErrMsg) {
// 	operatorUser, err := svc.getGuidUser(ctx, ctx.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	operatorMember, ok, err := dao.NewGuildMember(operatorUser.GuildId).GetOne(ctx, ctx.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !ok {
// 		return nil, errmsg.NewErrGuildPermissionDenied()
// 	}
// 	permissions, ok := rule.GetPermissions(ctx, operatorMember.Position)
// 	if !ok {
// 		return nil, errmsg.NewErrGuildPermissionDenied()
// 	}
// 	if !permissions.Invite {
// 		return nil, errmsg.NewErrGuildPermissionDenied()
// 	}
// 	ok, err = svc.levelCheck(ctx, req.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !ok {
// 		return nil, errmsg.NewErrGuildUserLevelNotEnough()
// 	}
// 	targetUser, err := dao.NewGuildUser(req.RoleId).Get(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if targetUser.GuildId != "" {
// 		return nil, errmsg.NewErrGuildHasJoined()
// 	}
//
// 	operatorUserInfo, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	guild, err := dao.NewGuild(operatorUser.GuildId).Get(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if guild == nil {
// 		return nil, errmsg.NewErrGuildNotExist()
// 	}
//
// 	err = svc.getLock(ctx, guildUserApplyLock+req.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	inviteDao := dao.NewGuildInvite(req.RoleId)
// 	inviteList, err := svc.getInviteList(ctx, req.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if values.Integer(len(inviteList)) >= rule.GetInviteLimit(ctx) {
// 		return nil, errmsg.NewErrGuildInviteListFull()
// 	}
// 	invite := &pbdao.GuildInvite{
// 		Uuid:       xid.New().String(),
// 		InviteId:   ctx.RoleId,
// 		InviteName: operatorUserInfo.Nickname,
// 		GuildId:    operatorUser.GuildId,
// 		GuildName:  guild.Name,
// 		ExpiredAt:  rule.GetInviteExpired(ctx),
// 	}
// 	if err := inviteDao.SaveOne(ctx, invite); err != nil {
// 		return nil, err
// 	}
// 	return &less_service.Guild_GuildInviteJoinResponse{}, nil
// }

func (svc *Service) InviteList(ctx *ctx.Context, req *less_service.Guild_GuildInviteListRequest) (*less_service.Guild_GuildInviteListResponse, *errmsg.ErrMsg) {
	count := 10
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	list, err := dao.GetInvite(ctx, user.GuildId, req.Refresh, req.Lang, req.CombatValueLimit)
	if err != nil {
		return nil, err
	}

	roles, err := svc.GetRole(ctx, list)
	if err != nil {
		return nil, err
	}
	// TODO 这里是否要从user那边获取
	guildUser, err := dao.NewGuildUser("").GetMulti(ctx, list)
	if err != nil {
		return nil, err
	}
	guildUnlock, err := svc.GetMultiSysUnlockBySys(ctx, list, enum.SystemGuild)
	if err != nil {
		return nil, err
	}
	sevenDaysKey := svc.get7DaysKey(ctx)
	ret := make([]*models.GuildInvite, 0)
	temp := make(map[values.RoleId]struct{})
	// 优先获取在线的
	for _, role := range roles {
		if role.Logout <= role.Login {
			gu, ok := guildUser[role.RoleId]
			if !ok {
				continue
			}
			// 有公会的不推荐
			if gu.GuildId != "" {
				continue
			}
			// 公会未解锁的不推荐
			if !guildUnlock[role.RoleId] {
				continue
			}
			temp[role.RoleId] = struct{}{}
			ret = append(ret, &models.GuildInvite{
				RoleId:           role.RoleId,
				LoginAt:          role.Login,
				LogoutAt:         role.Logout,
				Name:             role.Nickname,
				Level:            role.Level,
				CombatValue:      role.Power,
				ActiveValue:      svc.get7DayActiveByGuildUser(gu, sevenDaysKey),
				TotalActiveValue: gu.TotalActiveValue,
				Lang:             role.Language,
				AvatarId:         role.AvatarId,
				AvatarFrame:      role.AvatarFrame,
			})
			if len(ret) >= count {
				break
			}
		}
	}
	// 一个满足条件的都没有的时候直接随意填补
	if len(ret) == 0 {
		for _, role := range roles {
			if _, ok := temp[role.RoleId]; !ok {
				temp[role.RoleId] = struct{}{}
				ret = append(ret, &models.GuildInvite{
					RoleId:           role.RoleId,
					LoginAt:          role.Login,
					LogoutAt:         role.Logout,
					Name:             role.Nickname,
					Level:            role.Level,
					CombatValue:      role.Power,
					ActiveValue:      svc.get7DayActiveByGuildUser(guildUser[role.RoleId], sevenDaysKey),
					TotalActiveValue: guildUser[role.RoleId].TotalActiveValue,
					Lang:             role.Language,
					AvatarId:         role.AvatarId,
					AvatarFrame:      role.AvatarFrame,
				})
				if len(ret) >= count {
					break
				}
			}
		}
	}
	return &less_service.Guild_GuildInviteListResponse{
		List: ret,
	}, nil
}

func (svc *Service) HandleInvite(ctx *ctx.Context, req *less_service.Guild_GuildHandleInviteRequest) (*less_service.Guild_GuildHandleInviteResponse, *errmsg.ErrMsg) {
	guild, rewards, err := svc.handleInvite(ctx, req)
	if err != nil {
		return nil, err
	}
	svc.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskJoinGuild, 0, 1)
	// var (
	// 	guild *models.Guild
	// 	err   *errmsg.ErrMsg
	// )
	// if req.FromChat {
	// 	guild, err = svc.handleWorldInvite(ctx, req)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// } else {
	// 	guild, err = svc.handleNormalInvite(ctx, req)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return &less_service.Guild_GuildHandleInviteResponse{
		Info:    guild,
		Rewards: rewards,
	}, nil
}

func (svc *Service) JoinApply(ctx *ctx.Context, req *less_service.Guild_GuildJoinApplyRequest) (*less_service.Guild_GuildJoinApplyResponse, *errmsg.ErrMsg) {
	var (
		guild   *models.Guild
		rewards map[values.ItemId]values.Integer
		err     *errmsg.ErrMsg
	)
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if role.Level < rule.GetGuildUserLevelLimit(ctx) {
		return nil, errmsg.NewErrGuildUserLevelNotEnough()
	}
	for _, id := range req.Id {
		guild, rewards, err = svc.joinApplyOne(ctx, id, role.Power)
		if err != nil {
			return nil, err
		}
		if guild != nil {
			break
		}
	}
	if guild != nil {
		if err := svc.updateGuildRankValue(ctx, guild.Id, guild.CombatValue); err != nil {
			return nil, err
		}
		if err := svc.updateToGuildFilterServer(ctx, guild); err != nil {
			return nil, err
		}
		svc.joinLocalEvent(ctx, ctx.RoleId, guild.Id)
	}
	svc.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskJoinGuild, 0, 1)
	return &less_service.Guild_GuildJoinApplyResponse{
		Info:    guild,
		Rewards: rewards,
	}, nil
}

func (svc *Service) joinApplyOne(ctx *ctx.Context, id values.GuildId, combatValue values.Integer) (*models.Guild, map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	user, err := dao.NewGuildUser(ctx.RoleId).Get(ctx)
	if err != nil {
		return nil, nil, err
	}

	if user.GuildId != "" {
		return nil, nil, errmsg.NewErrGuildHasJoined()
	}
	if svc.isInCooling(ctx, user) {
		return nil, nil, errmsg.NewErrGuildJoinCD()
	}
	err = svc.getLock(ctx, guildLock+id)
	if err != nil {
		return nil, nil, err
	}

	userApplyList, err := svc.getUserApplyList(ctx, ctx.RoleId)
	if err != nil {
		return nil, nil, err
	}
	if values.Integer(len(userApplyList)) >= rule.GetGuildUserApplyListLimit(ctx) {
		return nil, nil, errmsg.NewErrGuildApplyFull()
	}
	guild, err := dao.NewGuild(id).Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	if guild == nil {
		return nil, nil, errmsg.NewErrGuildNotExist()
	}

	if guild.DeletedAt > 0 {
		return nil, nil, errmsg.NewErrGuildNotExist()
	}
	if combatValue < guild.CombatValueLimit {
		return nil, nil, errmsg.NewErrGuildCombatValueNotEnough()
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, nil, err
	}
	if guild.AutoJoin {
		memberModel := dao.NewGuildMember(guild.Id)
		members, err := memberModel.Get(ctx)
		if err != nil {
			return nil, nil, err
		}
		if values.Integer(len(members)) >= rule.GetMaxMemberCount(ctx, guild.Level) {
			return nil, nil, errmsg.NewErrGuildMemberFull()
		}

		memberPosition := rule.GetMemberPosition(ctx)
		// 更新guild user
		user.GuildId = guild.Id
		rewards, err := svc.handlerFirstJoinRewards(ctx, user)
		if err != nil {
			return nil, nil, err
		}
		if err := dao.NewGuildUser(user.RoleId).Save(ctx, user); err != nil {
			return nil, nil, err
		}
		// 更新member
		member := &pbdao.GuildMember{
			RoleId:           user.RoleId,
			ServerId:         ctx.ServerId,
			Position:         memberPosition,
			JoinAt:           timer.StartTime(ctx.StartTime).Unix(),
			LastPosition:     memberPosition,
			CombatValue:      role.Power,
			ActiveValue:      user.ActiveValue,
			TotalActiveValue: user.TotalActiveValue,
		}
		members = append(members, member)
		if err := memberModel.SaveOne(ctx, member); err != nil {
			return nil, nil, err
		}
		// 清除玩家已经申请过的公会记录
		if err := svc.clearUserApplyList(ctx, ctx.RoleId); err != nil {
			return nil, nil, err
		}
		guild.Count = values.Integer(len(members))
		if err := dao.NewGuild(id).Save(ctx, guild); err != nil {
			return nil, nil, err
		}
		svc.joinChatRoom(ctx, ctx.RoleId, guild.Id)

		return svc.dao2model(ctx, guild, members), rewards, nil
	}

	applyDao := dao.NewGuildApply(guild.Id)
	applyList, exist, err := svc.getGuildApplyList(ctx, guild.Id)
	if err != nil {
		return nil, nil, err
	}
	// 如果已经申请过了，则不再重复申请
	if exist {
		return nil, nil, nil
	}
	// 判断列表是否已满
	if values.Integer(len(applyList)) >= rule.GetGuildApplyListLimit(ctx) {
		return nil, nil, errmsg.NewErrGuildApplyListFull()
	}
	apply := &pbdao.GuildApply{
		RoleId:  user.RoleId,
		Name:    role.Nickname,
		ApplyAt: timer.StartTime(ctx.StartTime).Unix(),
	}
	if err := applyDao.SaveOne(ctx, apply); err != nil {
		return nil, nil, err
	}
	userApply := &pbdao.GuildUserApply{
		RoleId:    ctx.RoleId,
		GuildId:   id,
		ExpiredAt: rule.GetApplyExpired(ctx),
	}
	if err := dao.NewGuildUserApply(ctx.RoleId).SaveOne(ctx, userApply); err != nil {
		return nil, nil, err
	}
	return nil, nil, nil
}

func (svc *Service) CancelApply(ctx *ctx.Context, req *less_service.Guild_GuildCancelApplyRequest) (*less_service.Guild_GuildCancelApplyResponse, *errmsg.ErrMsg) {
	guildApplyDao := dao.NewGuildApply(req.Id)
	userApplyDao := dao.NewGuildUserApply(ctx.RoleId)
	userApplyList, err := svc.getUserApplyList(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	guildApplyList, err := guildApplyDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	userApply := &pbdao.GuildUserApply{}
	for _, item := range userApplyList {
		if item.GuildId == req.Id {
			userApply = item
			break
		}
	}
	guildApply := &pbdao.GuildApply{}
	for _, item := range guildApplyList {
		if item.RoleId == ctx.RoleId {
			guildApply = item
			break
		}
	}
	if err := guildApplyDao.SaveOne(ctx, guildApply); err != nil {
		return nil, err
	}
	if err := userApplyDao.SaveOne(ctx, userApply); err != nil {
		return nil, err
	}
	return &less_service.Guild_GuildCancelApplyResponse{}, nil
}

func (svc *Service) ApplyList(ctx *ctx.Context, _ *less_service.Guild_GuildApplyListRequest) (*less_service.Guild_GuildApplyListResponse, *errmsg.ErrMsg) {
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	list, _, err := svc.getGuildApplyList(ctx, user.GuildId)
	if err != nil {
		return nil, err
	}
	roleIds := make([]values.RoleId, 0, len(list))
	for _, apply := range list {
		roleIds = append(roleIds, apply.RoleId)
	}
	roleMap, err := svc.GetRole(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	applyList := make([]*models.GuildApply, 0, len(list))
	for _, apply := range list {
		role, ok := roleMap[apply.RoleId]
		if !ok {
			continue
		}
		applyList = append(applyList, &models.GuildApply{
			RoleId:      apply.RoleId,
			ServerId:    ctx.ServerId,
			Name:        apply.Name,
			ApplyAt:     apply.ApplyAt,
			LoginAt:     role.Login,
			LogoutAt:    role.Logout,
			CombatValue: role.Power,
			Lang:        role.Language,
			Level:       role.Level,
			AvatarId:    role.AvatarId,
			AvatarFrame: role.AvatarFrame,
		})
	}
	return &less_service.Guild_GuildApplyListResponse{
		List: applyList,
	}, nil
}

func (svc *Service) HandleApply(ctx *ctx.Context, req *less_service.Guild_GuildHandleApplyRequest) (*less_service.Guild_GuildHandleApplyResponse, *errmsg.ErrMsg) {
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	err = svc.getLock(ctx, guildLock+user.GuildId)
	if err != nil {
		return nil, err
	}
	err = svc.getLock(ctx, guildUserLock+req.RoleId)
	if err != nil {
		return nil, err
	}

	memberDao := dao.NewGuildMember(user.GuildId)
	curMember, ok, err := memberDao.GetOne(ctx, user.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	permissions, ok := rule.GetPermissions(ctx, curMember.Position)
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	if !permissions.HandleApply {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}

	guildDao := dao.NewGuild(user.GuildId)
	guild, err := guildDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	guildApplyDao := dao.NewGuildApply(guild.Id)
	applyTarget, ok, err := guildApplyDao.GetOne(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrGuildApplyExpired()
	}

	// applyList, _, err := svc.getGuildApplyList(ctx, guild.Id)
	// if err != nil {
	//	return nil, err
	// }
	// applyMap := make(map[values.RoleId]*pbdao.GuildApply)
	// for _, apply := range applyList {
	//	applyMap[apply.RoleId] = apply
	// }
	applyUser, err := dao.NewGuildUser(req.RoleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	role, err := svc.GetRoleByRoleId(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	if req.Agree {
		if applyUser.GuildId != "" {
			return nil, errmsg.NewErrGuildJoinedAnotherGuild()
		}
		members, err := memberDao.Get(ctx)
		if err != nil {
			return nil, err
		}
		count := values.Integer(len(members))
		max := rule.GetMaxMemberCount(ctx, guild.Level)
		// 满员检查
		if count >= max {
			return nil, errmsg.NewErrGuildMemberFull()
		}
		memberPosition := rule.GetMemberPosition(ctx)
		// newMember := make([]*pbdao.GuildMember, 0)
		// updateApplyUser := make([]*pbdao.GuildUser, 0)
		// deleteApply := make([]*pbdao.GuildApply, 0)

		applyUser.GuildId = guild.Id

		newMember := &pbdao.GuildMember{
			RoleId:       applyUser.RoleId,
			ServerId:     applyTarget.ServerId,
			Position:     memberPosition,
			JoinAt:       timer.StartTime(ctx.StartTime).Unix(),
			LastPosition: memberPosition,
			CombatValue:  role.Power,
		}
		guild.Count = values.Integer(len(members))
		if err := guildDao.Save(ctx, guild); err != nil {
			return nil, err
		}
		// 更新当前公会的成员列表
		if err := memberDao.SaveOne(ctx, newMember); err != nil {
			return nil, err
		}
		// 更新当前玩家的guildUser存档
		if err := dao.NewGuildUser(req.RoleId).Save(ctx, applyUser); err != nil {
			return nil, err
		}
		// 加入公会聊天频道
		svc.joinChatRoom(ctx, req.RoleId, guild.Id)
		// 将指定公会从玩家的申请列表里移除
		if err := svc.delGuildFromUserApplyList(ctx, req.RoleId, guild.Id); err != nil {
			return nil, err
		}
		// 任务打点
		svc.UpdateTarget(ctx, req.RoleId, models.TaskType_TaskJoinGuild, 0, 1)
		// 玩家职位变化推送
		svc.usersGuildInfoChangePush(ctx, []*pbdao.GuildApply{applyTarget}, guild.Id, memberPosition)
		// 更新GuildFilterServer
		if err := svc.updateToGuildFilterServer(ctx, svc.dao2model(ctx, guild, members)); err != nil {
			return nil, err
		}
		svc.joinLocalEvent(ctx, req.RoleId, guild.Id)
		// for _, roleId := range req.RoleId {
		//	var cv values.Integer
		//	role, roleExist := roleMap[roleId]
		//	if roleExist {
		//		cv = role.Power
		//	}
		//	applyUser := applyUserMap[roleId]
		//	apply, ok := applyMap[roleId]
		//	if !ok {
		//		continue
		//	}
		//	if applyUser.GuildId == "" {
		//		applyUser.GuildId = user.GuildId
		//		updateApplyUser = append(updateApplyUser, applyUser)
		//		newMember = append(newMember, &pbdao.GuildMember{
		//			RoleId:       applyUser.RoleId,
		//			ServerId:     apply.ServerId,
		//			Position:     memberPosition,
		//			JoinAt:       timer.StartTime(ctx.StartTime).Unix(),
		//			LastPosition: memberPosition,
		//			CombatValue:  cv,
		//		})
		//		delete(applyMap, roleId)
		//		deleteApply = append(deleteApply, apply)
		//		count++
		//		if count >= max {
		//			break
		//		}
		//	}
		// }
		// for _, apply := range applyMap {
		//	role, ok := roleMap[apply.RoleId]
		//	if !ok {
		//		role = &pbdao.Role{}
		//	}
		//	respApplyList = append(respApplyList, &models.GuildApply{
		//		RoleId:      apply.RoleId,
		//		ServerId:    apply.ServerId,
		//		Name:        apply.Name,
		//		ApplyAt:     apply.ApplyAt,
		//		LoginAt:     role.Login,
		//		LogoutAt:    role.Logout,
		//		CombatValue: role.Power,
		//		Lang:        role.Language,
		//		Level:       role.Level,
		//	})
		// }
		//
		// if len(newMember) > 0 {
		//	if err := memberDao.Save(ctx, newMember); err != nil {
		//		return nil, err
		//	}
		// }
		// if len(updateApplyUser) > 0 {
		//	if err := dao.NewGuildUser("").BatchSave(ctx, updateApplyUser); err != nil {
		//		return nil, err
		//	}
		// }
		// if len(deleteApply) > 0 {
		//	if err := applyDao.Delete(ctx, deleteApply); err != nil {
		//		return nil, err
		//	}
		//	for _, apply := range deleteApply {
		//		svc.joinChatRoom(ctx, apply.RoleId, guild.Id)
		//		if err := svc.delGuildFromUserApplyList(ctx, apply.RoleId, guild.Id); err != nil {
		//			return nil, err
		//		}
		//
		//		svc.UpdateTarget(ctx, apply.RoleId, models.TaskType_TaskJoinGuild, 0, 1)
		//	}
		//	svc.usersGuildInfoChangePush(ctx, deleteApply, guild.Id, memberPosition)
		// }
		// if err := svc.updateToGuildFilterServer(ctx, svc.dao2model(ctx, guild, members)); err != nil {
		//	return nil, err
		// }
	}
	if err := guildApplyDao.Delete(ctx, []*pbdao.GuildApply{applyTarget}); err != nil {
		return nil, err
	}

	return &less_service.Guild_GuildHandleApplyResponse{}, nil
}

func (svc *Service) RejectAll(ctx *ctx.Context, _ *less_service.Guild_GuildApplyRejectAllRequest) (*less_service.Guild_GuildApplyRejectAllResponse, *errmsg.ErrMsg) {
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	curMember, ok, err := dao.NewGuildMember(user.GuildId).GetOne(ctx, user.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	permissions, ok := rule.GetPermissions(ctx, curMember.Position)
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	if !permissions.HandleApply {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	applyDao := dao.NewGuildApply(user.GuildId)
	list, err := applyDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	if len(list) <= 0 {
		return nil, nil
	}
	for _, item := range list {
		if err := svc.delGuildFromUserApplyList(ctx, item.RoleId, user.GuildId); err != nil {
			return nil, err
		}
	}
	if err := applyDao.DeleteAll(ctx); err != nil {
		return nil, err
	}
	return &less_service.Guild_GuildApplyRejectAllResponse{}, nil
}

func (svc *Service) Remove(ctx *ctx.Context, req *less_service.Guild_GuildRemoveRequest) (*less_service.Guild_GuildRemoveResponse, *errmsg.ErrMsg) {
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	ok, err := svc.IsGuildBossFighting(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, errmsg.NewErrGuildBossFighting()
	}
	targetUser, err := dao.NewGuildUser(req.RoleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	operatorMember, ok, err := dao.NewGuildMember(user.GuildId).GetOne(ctx, user.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	permissions, ok := rule.GetPermissions(ctx, operatorMember.Position)
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	if !permissions.RemoveMember {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	err = svc.getLock(ctx, guildLock+user.GuildId)
	if err != nil {
		return nil, err
	}
	guildDao := dao.NewGuild(user.GuildId)
	guild, err := guildDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	memberDao := dao.NewGuildMember(user.GuildId)
	list, err := memberDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	var member *pbdao.GuildMember
	for _, item := range list {
		if item.RoleId == req.RoleId {
			member = item
			break
		}
	}
	if member == nil {
		return nil, errmsg.NewErrGuildUserNotInGuild()
	}
	if operatorMember.Position >= member.Position {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}

	targetUser.GuildId = ""
	targetUser.Cd = rule.GetRemoveCd(ctx)
	svc.syncActive2User(targetUser, member)

	guild.Count = values.Integer(len(list))
	if err := guildDao.Save(ctx, guild); err != nil {
		return nil, err
	}

	if err := dao.NewGuildUser(targetUser.RoleId).Save(ctx, targetUser); err != nil {
		return nil, err
	}
	if err := memberDao.Delete(ctx, member); err != nil {
		return nil, err
	}
	svc.leaveChatRoom(ctx, req.RoleId, user.GuildId)
	guildModel := svc.dao2model(ctx, guild, list)
	if err := svc.updateGuildRankValue(ctx, guildModel.Id, guildModel.CombatValue); err != nil {
		return nil, err
	}
	if err := svc.updateToGuildFilterServer(ctx, guildModel); err != nil {
		return nil, err
	}
	svc.userGuildInfoChangePush(ctx, targetUser.RoleId, "", 0)
	svc.exitLocalEvent(ctx, targetUser.RoleId, guild.Id)
	return &less_service.Guild_GuildRemoveResponse{
		Info: guildModel,
	}, nil
}

func (svc *Service) Exit(ctx *ctx.Context, _ *less_service.Guild_GuildExitRequest) (*less_service.Guild_GuildExitResponse, *errmsg.ErrMsg) {
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	ok, err := svc.IsGuildBossFighting(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, errmsg.NewErrGuildBossFighting()
	}
	err = svc.getLock(ctx, guildLock+user.GuildId)
	if err != nil {
		return nil, err
	}
	guildDao := dao.NewGuild(user.GuildId)
	guild, err := guildDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	memberDao := dao.NewGuildMember(user.GuildId)
	members, err := memberDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	var member *pbdao.GuildMember
	newMembersList := make([]*pbdao.GuildMember, 0)
	for _, item := range members {
		if item.RoleId == user.RoleId {
			member = item
		} else {
			newMembersList = append(newMembersList, item)
		}
	}

	if member == nil {
		return nil, errmsg.NewErrGuildUserNotInGuild()
	}
	if member.Position == rule.GetLeaderPosition(ctx) {
		return nil, errmsg.NewErrGuildLeaderNotAllowExit()
	}
	if err := memberDao.Delete(ctx, member); err != nil {
		return nil, err
	}
	guildId := user.GuildId
	user.GuildId = ""
	user.Cd = rule.GetExitCd(ctx)
	svc.syncActive2User(user, member)

	guild.Count = values.Integer(len(newMembersList))
	if err := guildDao.Save(ctx, guild); err != nil {
		return nil, err
	}

	if err := dao.NewGuildUser(user.RoleId).Save(ctx, user); err != nil {
		return nil, err
	}

	svc.leaveChatRoom(ctx, ctx.RoleId, guildId)

	guildModel := svc.dao2model(ctx, guild, newMembersList)
	if err := svc.updateGuildRankValue(ctx, guildModel.Id, guildModel.CombatValue); err != nil {
		return nil, err
	}
	if err := svc.updateToGuildFilterServer(ctx, guildModel); err != nil {
		return nil, err
	}

	svc.exitLocalEvent(ctx, ctx.RoleId, guild.Id)

	return &less_service.Guild_GuildExitResponse{}, nil
}

func (svc *Service) PositionChange(ctx *ctx.Context, req *less_service.Guild_GuildPositionChangeRequest) (*less_service.Guild_GuildPositionChangeResponse, *errmsg.ErrMsg) {
	list := rule.GuildPositionList(ctx)
	var find bool
	for _, position := range list {
		if position == req.Position {
			find = true
			break
		}
	}
	if !find {
		return nil, errmsg.NewErrGuildInvalidPosition()
	}
	leaderPosition := rule.GetLeaderPosition(ctx)
	changeLeader := req.Position == leaderPosition
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}

	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	err = svc.getLock(ctx, guildLock+user.GuildId)
	if err != nil {
		return nil, err
	}

	memberDao := dao.NewGuildMember(user.GuildId)
	members, err := memberDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	var target *pbdao.GuildMember
	for _, member := range members {
		if member.RoleId == req.RoleId {
			target = member
		}
	}
	if target == nil {
		return nil, errmsg.NewErrGuildUserNotInGuild()
	}
	var operator *pbdao.GuildMember
	for _, member := range members {
		if member.RoleId == ctx.RoleId {
			operator = member
		}
	}
	if operator == nil {
		return nil, errmsg.NewErrGuildUserNotInGuild()
	}

	update := make([]*pbdao.GuildMember, 0)
	if changeLeader {
		operator.LastPosition = operator.Position
		operator.Position = rule.GetMemberPosition(ctx)
		update = append(update, operator)

		target.LastPosition = target.Position
		target.Position = req.Position
	} else {
		if operator.Position >= target.Position {
			return nil, errmsg.NewErrGuildPermissionDenied()
		}
		curCount := svc.getCountByPosition(members, req.Position)
		maxCount := rule.GetMemberCountByPosition(ctx, guild.Level, req.Position)
		if curCount >= maxCount {
			return nil, errmsg.NewErrGuildPositionFull()
		}
		target.LastPosition = target.Position
		target.Position = req.Position
	}
	update = append(update, target)
	// newPosition, ok := svc.getPosition(ctx, target.Position, true)
	// if !ok {
	// 	return nil, errmsg.NewErrGuildMaxPosition()
	// }

	if err := memberDao.Save(ctx, update); err != nil {
		return nil, err
	}

	svc.userGuildInfoChangePush(ctx, operator.RoleId, guild.Id, operator.Position)
	svc.userGuildInfoChangePush(ctx, target.RoleId, guild.Id, target.Position)

	return &less_service.Guild_GuildPositionChangeResponse{}, nil
}

// func (svc *Service) Demotion(ctx *ctx.Context, req *less_service.Guild_GuildDemotionRequest) (*less_service.Guild_GuildDemotionResponse, *errmsg.ErrMsg) {
// 	user, err := svc.getGuidUser(ctx, ctx.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = svc.getLock(ctx, guildLock+user.GuildId)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	memberDao := dao.NewGuildMember(user.GuildId)
// 	members, err := memberDao.Get(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	target, ok, err := memberDao.GetOne(ctx, req.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !ok {
// 		return nil, errmsg.NewErrGuildUserNotInGuild()
// 	}
// 	operator, ok, err := memberDao.GetOne(ctx, ctx.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !ok {
// 		return nil, errmsg.NewErrGuildUserNotInGuild()
// 	}
// 	if operator.Position >= target.Position {
// 		return nil, errmsg.NewErrGuildPermissionDenied()
// 	}
// 	newPosition, ok := svc.getPosition(ctx, target.Position, false)
// 	if !ok {
// 		return nil, errmsg.NewErrGuildMinPosition()
// 	}
// 	curCount := svc.getCountByPosition(members, newPosition)
// 	maxCount := rule.GetMemberCountByPosition(ctx, guild.Level, newPosition)
// 	if curCount >= maxCount {
// 		return nil, errmsg.NewErrGuildPositionFull()
// 	}
// 	target.LastPosition = target.Position
// 	target.Position = newPosition
//
// 	if err := memberDao.SaveOne(ctx, target); err != nil {
// 		return nil, err
// 	}
//
// 	svc.guildSystemMsg(ctx, guild.Id, &SystemMsg{
// 		TextId: demotion,
// 		Args:   []string{"玩家名字", strconv.Itoa(int(newPosition))}, // TODO
// 	})
//
// 	return &less_service.Guild_GuildDemotionResponse{}, nil
// }

// func (svc *Service) LeaderChange(ctx *ctx.Context, req *less_service.Guild_GuildLeaderChangeRequest) (*less_service.Guild_GuildLeaderChangeResponse, *errmsg.ErrMsg) {
// 	user, err := svc.getGuidUser(ctx, ctx.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if guild.LeaderId != user.RoleId {
// 		return nil, errmsg.NewErrGuildPermissionDenied()
// 	}
// 	err = svc.getLock(ctx, guildLock+user.GuildId)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	err = svc.getLock(ctx, guildLock+user.GuildId)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	memberDao := dao.NewGuildMember(guild.Id)
//
// 	target, ok, err := memberDao.GetOne(ctx, req.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !ok {
// 		return nil, errmsg.NewErrGuildUserNotInGuild()
// 	}
// 	operator, ok, err := memberDao.GetOne(ctx, ctx.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !ok {
// 		return nil, errmsg.NewErrGuildUserNotInGuild()
// 	}
// 	operator.Position = rule.GetMemberPosition(ctx)
// 	target.LastPosition = target.Position
// 	target.Position = rule.GetLeaderPosition(ctx)
// 	guild.LeaderId = req.RoleId
//
// 	if err := dao.NewGuild(guild.Id).Save(ctx, guild); err != nil {
// 		return nil, err
// 	}
// 	if err := memberDao.Save(ctx, []*pbdao.GuildMember{operator, target}); err != nil {
// 		return nil, err
// 	}
//
// 	svc.guildSystemMsg(ctx, guild.Id, &SystemMsg{
// 		TextId: leaderChange,
// 		Args:   []string{"玩家名字", "玩家名字"}, // TODO
// 	})
//
// 	return &less_service.Guild_GuildLeaderChangeResponse{}, nil
// }

func (svc *Service) GuildInfo(ctx *ctx.Context, req *less_service.Guild_GuildGuildInfoRequest) (*less_service.Guild_GuildGuildInfoResponse, *errmsg.ErrMsg) {
	roleId := ctx.RoleId
	if req.RoleId != "" {
		roleId = req.RoleId
	}
	user, err := dao.NewGuildUser(roleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	var (
		id   values.GuildId
		name string
	)
	if user.GuildId != "" {
		id = user.GuildId
		guild, err := dao.NewGuild(id).Get(ctx)
		if err != nil {
			return nil, err
		}
		name = guild.Name
	}
	return &less_service.Guild_GuildGuildInfoResponse{
		Id:   id,
		Name: name,
	}, nil
}

func (svc *Service) ModifyPositionName(ctx *ctx.Context, req *less_service.Guild_GuildModifyPositionNameRequest) (*less_service.Guild_GuildModifyPositionNameResponse, *errmsg.ErrMsg) {
	if len(req.Name) <= 0 {
		return nil, errmsg.NewErrGuildInvalidInput()
	}
	for position, name := range req.Name {
		if err := svc.positionNameCheck(ctx, position, name); err != nil {
			return nil, err
		}
	}
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	err = svc.getLock(ctx, guildLock+user.GuildId)
	if err != nil {
		return nil, err
	}

	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	members, err := dao.NewGuildMember(guild.Id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var member *pbdao.GuildMember
	for _, guildMember := range members {
		if guildMember.RoleId == ctx.RoleId {
			member = guildMember
			break
		}
	}
	if member == nil {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	item, ok := rule.GetPermissions(ctx, member.Position)
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	if !item.ModifyPositionName {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	for _, name := range req.Name {
		if !sensitive.TextValid(name) {
			return nil, errmsg.NewErrSensitive()
		}
	}
	guild.PositionName = req.Name
	if err := dao.NewGuild(guild.Id).Save(ctx, guild); err != nil {
		return nil, err
	}
	if err := svc.updateToGuildFilterServer(ctx, svc.dao2model(ctx, guild, members)); err != nil {
		return nil, err
	}
	return &less_service.Guild_GuildModifyPositionNameResponse{}, nil
}

func (svc *Service) BuildInfo(ctx *ctx.Context, _ *less_service.Guild_GuildBuildInfoRequest) (*less_service.Guild_GuildBuildInfoResponse, *errmsg.ErrMsg) {
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if user.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	cfg, ok := rule.GetGuildConfigByLevel(ctx, guild.Level)
	if !ok {
		return nil, errmsg.NewInternalErr("no guild config")
	}
	reset, info := svc.buildInfo(ctx, user, cfg)
	if reset {
		if err := dao.NewGuildUser(user.GuildId).Save(ctx, user); err != nil {
			return nil, err
		}
	}
	return &less_service.Guild_GuildBuildInfoResponse{
		FreeCount: info.FreeCount,
		PayCount:  info.PayCount,
		PayCost:   info.Cost,
		ResetTime: info.ResetTime,
	}, nil
}

func (svc *Service) Build(ctx *ctx.Context, req *less_service.Guild_GuildBuildRequest) (*less_service.Guild_GuildBuildResponse, *errmsg.ErrMsg) {
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	err = svc.getLock(ctx, guildLock+user.GuildId)
	if err != nil {
		return nil, err
	}

	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	cfg, ok := rule.GetGuildConfigByLevel(ctx, guild.Level)
	if !ok {
		return nil, errmsg.NewInternalErr("no guild config")
	}
	_, info := svc.buildInfo(ctx, user, cfg)
	if req.Free {
		if info.FreeCount <= 0 {
			return nil, errmsg.NewErrGuildBuildTimesNotEnough()
		}
		user.Build.FreeCount++
	} else {
		if info.PayCount <= 0 {
			return nil, errmsg.NewErrGuildBuildTimesNotEnough()
		}
		if err := svc.SubManyItem(ctx, ctx.RoleId, info.Cost); err != nil {
			return nil, err
		}
		user.Build.PayCount++
	}
	surprised := values.Integer(1)
	rewards := make(map[values.Integer]values.Integer)
	if cfg.SurprisedMag[1] < 0 {
		return &less_service.Guild_GuildBuildResponse{}, errmsg.NewInternalErr("invalid SurprisedMag")
	}
	n := rand.Int63n(10001)
	if n <= cfg.SurprisedMag[1] {
		surprised = cfg.SurprisedMag[0]
		for id, count := range cfg.RewardItem {
			rewards[id] = count * cfg.SurprisedMag[0]
		}
	} else {
		rewards = cfg.RewardItem
	}
	members, err := dao.NewGuildMember(guild.Id).Get(ctx)
	if err != nil {
		return nil, err
	}
	if len(rewards) > 0 {
		if _, err := svc.AddManyItem(ctx, ctx.RoleId, rewards); err != nil {
			return nil, err
		}
	}
	exp, levelup := svc.addGuildExp(ctx, guild, cfg, surprised, 0)
	if levelup {
		if err := svc.updateToGuildFilterServer(ctx, svc.dao2model(ctx, guild, members)); err != nil {
			return nil, err
		}
	}
	if err := dao.NewGuildUser(ctx.RoleId).Save(ctx, user); err != nil {
		return nil, err
	}
	if err := dao.NewGuild(guild.Id).Save(ctx, guild); err != nil {
		return nil, err
	}
	_, info = svc.buildInfo(ctx, user, cfg)

	if exp > 0 {
		svc.guildExpChangePush(ctx, guild, members)
	}

	return &less_service.Guild_GuildBuildResponse{
		Level:     guild.Level,
		Exp:       guild.Exp,
		Rewards:   rewards,
		Surprised: surprised,
		FreeCount: info.FreeCount,
		PayCount:  info.PayCount,
		PayCost:   info.Cost,
		ResetTime: info.ResetTime,
	}, nil
}

func (svc *Service) WorldInvite(ctx *ctx.Context, req *less_service.Guild_GuildWorldInviteRequest) (*less_service.Guild_GuildWorldInviteResponse, *errmsg.ErrMsg) {
	operatorUser, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	operatorMember, ok, err := dao.NewGuildMember(operatorUser.GuildId).GetOne(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	permissions, ok := rule.GetPermissions(ctx, operatorMember.Position)
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	if !permissions.Invite {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	guild, err := dao.NewGuild(operatorUser.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, errmsg.NewErrGuildNotExist()
	}
	if req.Save {
		if err := svc.inviteMsgCheck(ctx, req.Msg); err != nil {
			return nil, err
		}
		guild.InviteMsg = req.Msg
	}
	info := svc.worldInviteInfo(ctx, guild)
	if req.Free {
		if info.Free <= 0 {
			return nil, errmsg.NewErrGuildInviteTimesNotEnough()
		}
		guild.WorldInvite.Free++
	} else {
		if info.Pay <= 0 {
			return nil, errmsg.NewErrGuildInviteTimesNotEnough()
		}
		cost, ok := rule.GetGuildWorldInviteCost(ctx)
		if !ok {
			return nil, errmsg.NewInternalErr("no guild world invite cost")
		}
		if err := svc.SubManyItem(ctx, ctx.RoleId, map[values.ItemId]values.Integer{cost.ItemId: cost.Count}); err != nil {
			return nil, err
		}
		guild.WorldInvite.Pay++
	}
	invite := Invite{
		GuildId:   guild.Id,
		GuildName: guild.Name,
		Private:   false,
		Msg:       req.Msg,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Invite: invite,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: rule.GetInviteExpired(ctx),
		},
	})
	_token, err1 := token.SignedString(jwtKey)
	if err1 != nil {
		return nil, errmsg.NewInternalErr("jwt sign failed")
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	invite.Token = _token
	content, err1 := json.MarshalToString(invite)
	if err1 != nil {
		svc.log.Error("marshal guild system msg err", zap.Any("data", invite), zap.Error(err))
		return nil, errmsg.NewInternalErr("marshal world invite msg err")
	}
	if err := dao.NewGuild(guild.Id).Save(ctx, guild); err != nil {
		return nil, err
	}

	if err := im.DefaultClient.SendMessage(ctx, &im.Message{
		Type:      im.MsgTypeBroadcast,
		RoleID:    role.RoleId,
		RoleName:  role.Nickname,
		Content:   content,
		ParseType: im.ParseTypeGuildInvite,
		Extra:     imutil.GetIMRoleInfoExtra(role),
	}); err != nil {
		svc.log.Error("send guild world invite msg err", zap.Error(err))
		return nil, errmsg.NewInternalErr("send guild world invite msg err")
	}

	return &less_service.Guild_GuildWorldInviteResponse{}, nil
}

func (svc *Service) PrivateInvite(ctx *ctx.Context, req *less_service.Guild_GuildPrivateInviteRequest) (*less_service.Guild_GuildPrivateInviteResponse, *errmsg.ErrMsg) {
	operatorUser, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	operatorMember, ok, err := dao.NewGuildMember(operatorUser.GuildId).GetOne(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	permissions, ok := rule.GetPermissions(ctx, operatorMember.Position)
	if !ok {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	if !permissions.Invite {
		return nil, errmsg.NewErrGuildPermissionDenied()
	}
	guild, err := dao.NewGuild(operatorUser.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, errmsg.NewErrGuildNotExist()
	}
	if req.Save {
		if err := svc.inviteMsgCheck(ctx, req.Msg); err != nil {
			return nil, err
		}
		guild.InviteMsg = req.Msg
	}
	invite := Invite{
		GuildId:   guild.Id,
		GuildName: guild.Name,
		Private:   true,
		Msg:       req.Msg,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Invite: invite,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: rule.GetInviteExpired(ctx),
		},
	})
	_token, err1 := token.SignedString(jwtKey)
	if err1 != nil {
		return nil, errmsg.NewInternalErr("jwt sign failed")
	}
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	invite.Token = _token
	content, err1 := json.MarshalToString(invite)
	if err1 != nil {
		svc.log.Error("marshal guild system msg err", zap.Any("data", invite), zap.Error(err))
		return nil, errmsg.NewInternalErr("marshal world invite msg err")
	}
	if err := dao.NewGuild(guild.Id).Save(ctx, guild); err != nil {
		return nil, err
	}

	if err := im.DefaultClient.SendMessage(ctx, &im.Message{
		Type:      im.MsgTypePrivate,
		RoleID:    role.RoleId,
		RoleName:  role.Nickname,
		TargetID:  req.RoleId,
		Content:   content,
		ParseType: im.ParseTypeGuildInvite,
		Extra:     imutil.GetIMRoleInfoExtra(role),
	}); err != nil {
		svc.log.Error("send guild private invite msg err", zap.Error(err))
		return nil, errmsg.NewInternalErr("send guild private invite msg err")
	}

	return &less_service.Guild_GuildPrivateInviteResponse{}, nil
}

func (svc *Service) Rank(ctx *ctx.Context, req *less_service.Guild_GuildRankRequest) (*less_service.Guild_GuildRankResponse, *errmsg.ErrMsg) {
	user, err := dao.NewGuildUser(ctx.RoleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	var start, end = req.Start, req.End
	if start < 1 {
		start = 1
	}
	max := values.Integer(100) // 一次最多取100条
	if start-end > max {
		end = start + max
	}

	return svc.getTopRank(ctx, start, end, user.GuildId)
}

func (svc *Service) getTopRank(ctx *ctx.Context, start, end values.Integer, guildId values.GuildId) (ret *less_service.Guild_GuildRankResponse, err *errmsg.ErrMsg) {
	var (
		selfGuild *models.Guild
		selfRank  values.Integer
	)

	if start == 1 && guildId != "" {
		guild, err := dao.NewGuild(guildId).Get(ctx)
		if err != nil {
			return nil, err
		}
		if guild == nil {
			return nil, errmsg.NewErrGuildNotExist()
		}
		members, err := dao.NewGuildMember(guildId).Get(ctx)
		if err != nil {
			return nil, err
		}
		selfGuild = svc.dao2model(ctx, guild, members)
		selfOut := &rank_service.RankService_GetRankValueByOwnerIdResponse{}
		if err := svc.svc.GetNatsClient().RequestWithOut(ctx, 0, &rank_service.RankService_GetRankValueByOwnerIdRequest{
			RankId:  svc.rankId,
			OwnerId: guildId,
		}, selfOut); err != nil {
			return nil, err
		}
		if selfOut.RankValue != nil {
			selfRank = selfOut.RankValue.Rank
		}
	}
	out := &rank_service.RankService_GetValueByIndexResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, 0, &rank_service.RankService_GetValueByIndexRequest{
		RankId: svc.rankId,
		Start:  start,
		End:    end,
	}, out); err != nil {
		return nil, err
	}
	guildIdList := make([]values.GuildId, 0)
	for _, v := range out.RankValues {
		guildIdList = append(guildIdList, v.OwnerId)
	}
	guildMap, err := dao.NewGuild("").GetMulti(ctx, guildIdList)
	if err != nil {
		return nil, err
	}
	memberMap, err := dao.NewGuildMember("").GetMulti(ctx, guildIdList)
	if err != nil {
		return nil, err
	}
	data := make(map[values.Integer]*models.Guild)
	for _, item := range out.RankValues {
		guild, ok1 := guildMap[item.OwnerId]
		members, ok2 := memberMap[item.OwnerId]
		if !ok1 || !ok2 {
			continue
		}
		data[item.Rank] = svc.dao2model(ctx, guild, members, item.Value1)
	}
	return &less_service.Guild_GuildRankResponse{
		Self:     selfGuild,
		SelfRank: selfRank,
		Data:     data,
		Ending:   out.Ending,
	}, nil
}

// func (svc *Service) getSelfNearbyRank(ctx *ctx.Context, index values.Integer, guildId values.GuildId) (ret *less_service.Guild_GuildRankResponse, err *errmsg.ErrMsg) {
// 	count := values.Integer(10)
// 	out := &rank_service.RankService_GetRankValueByByOwnerIdResponse{}
// 	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, 0, &rank_service.RankService_GetRankValueByOwnerIdRequest{
// 		RankId:  svc.rankId,
// 		OwnerId: guildId,
// 	}, out); err != nil {
// 		return nil, err
// 	}
// 	if out.RankValue == nil || out.RankValue.Rank <= 0 {
// 		return
// 	}
// 	start, end := getRankRange(out.RankValue.Rank)
// 	if index == 1 {
// 		index = start
// 	}
// 	var selfGuild *models.Guild
// 	if index == start {
// 		guild, err := dao.NewGuild(guildId).Get(ctx)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if guild == nil {
// 			return nil, errmsg.NewErrGuildNotExist()
// 		}
// 		members, err := dao.NewGuildMember(guildId).Get(ctx)
// 		if err != nil {
// 			return nil, err
// 		}
// 		selfGuild = svc.dao2model(ctx, guild, members)
// 	}
// 	out1 := &rank_service.RankService_GetValueByIndexResponse{}
// 	if err := svc.svc.GetNatsClient().RequestWithOut(ctx, 0, &rank_service.RankService_GetValueByIndexRequest{
// 		RankId: svc.rankId,
// 		Index:  index,
// 	}, out1); err != nil {
// 		return nil, err
// 	}
// 	guildIdList := make([]values.GuildId, 0)
// 	for _, v := range out1.RankValues {
// 		guildIdList = append(guildIdList, v.OwnerId)
// 	}
// 	guildMap, err := dao.NewGuild("").GetMulti(ctx, guildIdList)
// 	if err != nil {
// 		return nil, err
// 	}
// 	memberMap, err := dao.NewGuildMember("").GetMulti(ctx, guildIdList)
// 	if err != nil {
// 		return nil, err
// 	}
// 	data := make(map[values.Integer]*models.Guild)
// 	for _, item := range out1.RankValues {
// 		guild, ok1 := guildMap[item.OwnerId]
// 		members, ok2 := memberMap[item.OwnerId]
// 		if !ok1 || !ok2 {
// 			continue
// 		}
// 		data[item.Rank] = svc.dao2model(ctx, guild, members)
// 	}
// 	ret = &less_service.Guild_GuildRankResponse{
// 		Index:  index + count,
// 		Self:   selfGuild,
// 		Data:   data,
// 		Ending: out1.Ending,
// 	}
// 	if !ret.Ending && index >= end {
// 		ret.Ending = true
// 	}
// 	return ret, nil
// }

func (svc *Service) getLock(ctx *ctx.Context, key string) *errmsg.ErrMsg {
	return ctx.DRLock(redisclient.GetLocker(), key)
}

func (svc *Service) levelCheck(ctx *ctx.Context, roleId values.RoleId) (bool, *errmsg.ErrMsg) {
	level, err := svc.GetLevel(ctx, roleId)
	if err != nil {
		return false, err
	}
	if level < rule.GetGuildUserLevelLimit(ctx) {
		return false, nil
	}
	return true, nil
}

func (svc *Service) dao2model(ctx *ctx.Context, guild *pbdao.Guild, members []*pbdao.GuildMember, combatValueFromRank ...values.Integer) *models.Guild {
	sevenDaysKey := svc.get7DaysKey(ctx)
	var combatValue, activeValue values.Integer
	for _, member := range members {
		combatValue += member.CombatValue
		activeValue += svc.get7DayActive(member, sevenDaysKey)
	}
	if len(combatValueFromRank) > 0 {
		combatValue = combatValueFromRank[0]
	}
	return &models.Guild{
		Id:               guild.Id,
		Name:             guild.Name,
		Flag:             guild.Flag,
		Lang:             guild.Lang,
		Intro:            guild.Intro,
		Notice:           guild.Notice,
		AutoJoin:         guild.AutoJoin,
		Level:            guild.Level,
		Exp:              guild.Exp,
		Resources:        guild.Resources,
		CombatValue:      combatValue,
		Count:            values.Integer(len(members)),
		Greeting:         guild.Greeting,
		PositionName:     guild.PositionName,
		WorldInvite:      svc.worldInviteInfo(ctx, guild),
		ActiveValue:      activeValue,
		CombatValueLimit: guild.CombatValueLimit,
		InviteMsg:        guild.InviteMsg,
	}
}

func (svc *Service) getGuildApplyList(ctx *ctx.Context, id values.GuildId) ([]*pbdao.GuildApply, bool, *errmsg.ErrMsg) {
	list, err := dao.NewGuildApply(id).Get(ctx)
	if err != nil {
		return nil, false, err
	}
	_now := timer.StartTime(ctx.StartTime).Unix()
	v := rule.GetApplyExpiredVal(ctx)
	newList := make([]*pbdao.GuildApply, 0)
	delList := make([]*pbdao.GuildApply, 0)
	var exist bool
	for _, apply := range list {
		if apply.ApplyAt+v > _now {
			newList = append(newList, apply)
			if !exist && apply.RoleId == ctx.RoleId {
				exist = true
			}
		} else {
			delList = append(delList, apply)
		}
	}
	if len(list) != len(newList) {
		if err := dao.NewGuildApply(id).Delete(ctx, delList); err != nil {
			return nil, exist, err
		}
	}
	return newList, exist, nil
}

func (svc *Service) getUserApplyList(ctx *ctx.Context, roleId values.RoleId) ([]*pbdao.GuildUserApply, *errmsg.ErrMsg) {
	list, err := dao.NewGuildUserApply(roleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	_now := timer.StartTime(ctx.StartTime).Unix()
	newList := make([]*pbdao.GuildUserApply, 0)
	delList := make([]*pbdao.GuildUserApply, 0)
	for _, apply := range list {
		if apply.ExpiredAt > _now {
			newList = append(newList, apply)
		} else {
			delList = append(delList, apply)
		}
	}
	if len(list) != len(newList) {
		if err := dao.NewGuildUserApply(roleId).Delete(ctx, delList); err != nil {
			return nil, err
		}
	}
	return newList, nil
}

// func (svc *Service) getInviteList(ctx *ctx.Context, id values.RoleId) ([]*pbdao.GuildInvite, *errmsg.ErrMsg) {
// 	list, err := dao.NewGuildInvite(id).Get(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	now := time.Now().Unix()
// 	newList := make([]*pbdao.GuildInvite, 0)
// 	delList := make([]*pbdao.GuildInvite, 0)
// 	for _, invite := range list {
// 		if invite.ExpiredAt > now {
// 			newList = append(newList, invite)
// 		} else {
// 			delList = append(delList, invite)
// 		}
// 	}
// 	if len(list) != len(newList) {
// 		if err := dao.NewGuildInvite(id).Delete(ctx, delList); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return newList, nil
// }

func (svc *Service) nameCheck(ctx *ctx.Context, name string) (string, *errmsg.ErrMsg) {
	name = strings.TrimSpace(name)
	if name == "" {
		return name, errmsg.NewErrGuildNameEmpty()
	}
	if !sensitive.TextValid(name) {
		return name, errmsg.NewErrSensitive()
	}
	temp := []rune(name)
	if len(temp) > rule.GetGuildNameLen(ctx) {
		return name, errmsg.NewErrGuildNameTooLong()
	}
	exist, err := dao.NameExist(name)
	if err != nil {
		return name, err
	}
	if exist {
		return name, errmsg.NewErrGuildNameExist()
	}
	return name, nil
	// // 重名检查
	// store := redisclient.GetDefaultRedis()
	// bf := bloomfilter.New(store, guildNameBloomKey, guildNameBloomSize)
	// ok, err := bf.Exists([]byte(name))
	// if err != nil {
	// 	return err
	// }
	// if ok {
	// 	return errmsg.NewErrGuildNameExist()
	// }
	// return bf.Add([]byte(name))
}

func (svc *Service) introCheck(ctx *ctx.Context, intro string) *errmsg.ErrMsg {
	intro = strings.TrimSpace(intro)
	if intro == "" {
		return nil
	}
	if !sensitive.TextValid(intro) {
		return errmsg.NewErrSensitive()
	}
	temp := []rune(intro)
	if len(temp) > rule.GetGuildIntroLen(ctx) {
		return errmsg.NewErrGuildIntroTooLong()
	}
	return nil
}

func (svc *Service) noticeCheck(ctx *ctx.Context, notice string) *errmsg.ErrMsg {
	// 公告可以为空
	notice = strings.TrimSpace(notice)
	if notice == "" {
		return nil
	}
	temp := []rune(notice)
	if !sensitive.TextValid(notice) {
		return errmsg.NewErrSensitive()
	}
	if len(temp) > rule.GetGuildNoticeLen(ctx) {
		return errmsg.NewErrGuildNoticeTooLong()
	}
	return nil
}

func (svc *Service) greetingCheck(ctx *ctx.Context, greeting string) *errmsg.ErrMsg {
	greeting = strings.TrimSpace(greeting)
	if greeting == "" {
		return nil
	}
	if !sensitive.TextValid(greeting) {
		return errmsg.NewErrSensitive()
	}
	temp := []rune(greeting)
	if len(temp) > rule.GetGuildGreetingLen(ctx) {
		return errmsg.NewErrGuildGreetingTooLong()
	}
	return nil
}

func (svc *Service) inviteMsgCheck(ctx *ctx.Context, msg string) *errmsg.ErrMsg {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return nil
	}
	if !sensitive.TextValid(msg) {
		return errmsg.NewErrSensitive()
	}
	temp := []rune(msg)
	if len(temp) > rule.GuildInviteMsgLen(ctx) {
		return errmsg.NewErrGuildInviteMsgTooLong()
	}
	return nil
}

func (svc *Service) getGuidUser(ctx *ctx.Context, roleId values.RoleId) (*pbdao.GuildUser, *errmsg.ErrMsg) {
	user, err := dao.NewGuildUser(roleId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if user.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	return user, nil
}

func (svc *Service) isInCooling(ctx *ctx.Context, user *pbdao.GuildUser) bool {
	return user.Cd > timer.StartTime(ctx.StartTime).Unix()
}

// 将指定公会从玩家的申请列表里移除
func (svc *Service) delGuildFromUserApplyList(ctx *ctx.Context, roleId values.RoleId, guildId values.GuildId) *errmsg.ErrMsg {
	userApplyDao := dao.NewGuildUserApply(roleId)
	list, err := userApplyDao.Get(ctx)
	if err != nil {
		return err
	}
	apply := &pbdao.GuildUserApply{}
	for _, item := range list {
		if item.GuildId == guildId {
			apply = item
			break
		}
	}
	return userApplyDao.DeleteOne(ctx, apply)
}

// 清空玩家的申请记录
func (svc *Service) clearUserApplyList(ctx *ctx.Context, roleId values.RoleId) *errmsg.ErrMsg {
	return dao.NewGuildUserApply(roleId).DeleteAll(ctx)
}

func (svc *Service) getCountByPosition(list []*pbdao.GuildMember, p values.GuildPosition) values.Integer {
	var count values.Integer
	for _, member := range list {
		if member.Position == p {
			count++
		}
	}
	return count
}

// 根据当前职位获取它的上级或者下级职位
func (svc *Service) getPosition(ctx *ctx.Context, p values.GuildPosition, superior bool) (values.GuildPosition, bool) {
	list := rule.GuildPositionList(ctx)
	var index int
	for i, pos := range list {
		if pos == p {
			index = i
			break
		}
	}
	if superior {
		index--
		if index < 0 {
			return 0, false
		}
		return list[index], true
	}
	index++
	if index >= len(list) {
		return 0, false
	}
	return list[index], true
}

func (svc *Service) modifyIntroSysMsg(ctx *ctx.Context, id values.GuildId, notice string) {
	gopool.Submit(func() {
		if err := im.DefaultClient.SendMessage(ctx, &im.Message{
			Type:      im.MsgTypeRoom,
			RoleID:    "system",
			RoleName:  "system",
			TargetID:  id,
			Content:   notice,
			ParseType: im.ParseTypeSys,
		}); err != nil {
			svc.log.Error("send guild system msg err", zap.Error(err))
		}
	})
}

func (svc *Service) leaderAutoChange(ctx *ctx.Context, guild *pbdao.Guild, members []*pbdao.GuildMember) *errmsg.ErrMsg {
	leaderPosition := rule.GetLeaderPosition(ctx)
	memberPosition := rule.GetMemberPosition(ctx)
	leader := &pbdao.GuildMember{}
	for _, member := range members {
		if member.Position == leaderPosition {
			leader = member
			break
		}
	}
	if leader.RoleId == "" {
		return nil
	}
	leaderUser, err := svc.GetRoleByRoleId(ctx, leader.RoleId)
	if err != nil {
		return err
	}
	v := rule.GetGuildLeaderAutoHandOverDay(ctx)
	if timer.StartTime(ctx.StartTime).Unix()-leaderUser.Logout < v {
		return nil
	}
	newLeader := svc.getNewLeader(ctx, members, v, leaderPosition)
	if newLeader == nil {
		return nil
	}
	leader.Position = memberPosition
	newLeader.Position = leaderPosition
	guild.LeaderId = newLeader.RoleId
	if err := dao.NewGuild(guild.Id).Save(ctx, guild); err != nil {
		return err
	}

	if err := dao.NewGuildMember(guild.Id).Save(ctx, []*pbdao.GuildMember{leader, newLeader}); err != nil {
		return err
	}

	return nil
}

func (svc *Service) getNewLeader(ctx *ctx.Context, members []*pbdao.GuildMember, v values.Integer, leaderPosition values.GuildPosition) *pbdao.GuildMember {
	// 当公会会长XX日没上线，会长一职会自动转让到XX日内有上限的下一级职位，依次类推
	today := timer.BeginOfDay(timer.StartTime(ctx.StartTime)).Unix()
	target := today - v
	positionList := rule.GuildPositionList(ctx)
	var find *pbdao.GuildMember
	// positionList是按职位从高到底排序的
	for _, position := range positionList {
		if position == leaderPosition {
			continue
		}
		for _, member := range members {
			if member.Position == position && len(member.ActiveValue) > 0 {
				for day, val := range member.ActiveValue {
					if day >= target && val > 0 {
						find = member
						break
					}
				}
			}
		}
	}
	return find
}

func (svc *Service) saveGuildId(ctx *ctx.Context, id values.GuildId) *errmsg.ErrMsg {
	redis := redisclient.GetGuildRedis()
	k := getGuildIdKey(id)
	if err := redis.HSet(ctx, k, id, 0).Err(); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}

func (svc *Service) delGuildId(ctx *ctx.Context, id values.GuildId) *errmsg.ErrMsg {
	redis := redisclient.GetGuildRedis()
	if err := redis.HDel(ctx, getGuildIdKey(id), id).Err(); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}

func (svc *Service) joinChatRoom(ctx *ctx.Context, roleId values.RoleId, gid values.GuildId) {
	gopool.Submit(func() {
		if err := im.DefaultClient.JoinRoom(ctx, &im.RoomRole{
			RoomID:  gid,
			RoleIDs: []string{roleId},
		}); err != nil {
			svc.log.Error("join chat room err", zap.Error(err), zap.String("role_id", roleId), zap.String("gid", gid))
		}
	})
}

func (svc *Service) leaveChatRoom(ctx *ctx.Context, roleId values.RoleId, gid values.GuildId) {
	gopool.Submit(func() {
		if err := im.DefaultClient.LeaveRoom(ctx, &im.RoomRole{
			RoomID:  gid,
			RoleIDs: []string{roleId},
		}); err != nil {
			svc.log.Error("leave chat room err", zap.Error(err), zap.String("role_id", roleId), zap.String("gid", gid))
		}
	})
}

func (svc *Service) positionNameCheck(ctx *ctx.Context, position values.GuildPosition, name string) *errmsg.ErrMsg {
	if !rule.PositionCheck(ctx, position) {
		return errmsg.NewErrGuildInvalidPosition()
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errmsg.NewErrGuildNoticeRequired()
	}
	if !sensitive.TextValid(name) {
		return errmsg.NewErrSensitive()
	}
	temp := []rune(name)
	if len(temp) > rule.GetGuildPositionNameLen(ctx) {
		return errmsg.NewErrGuildPositionNameTooLong()
	}
	return nil
}

func (svc *Service) buildInfo(ctx *ctx.Context, user *pbdao.GuildUser, cfg *rulemodel.Guild) (bool, *BuildInfo) {
	userBuildInfo := user.Build
	if userBuildInfo == nil {
		userBuildInfo = &models.GuildBuild{
			FreeCount: 0,
			PayCount:  0,
			ResetTime: 0,
		}
	}
	var reset bool
	if userBuildInfo.ResetTime < timer.StartTime(ctx.StartTime).Unix() {
		reset = true
		userBuildInfo.FreeCount = 0
		userBuildInfo.PayCount = 0
		userBuildInfo.ResetTime = timer.NextDay(svc.GetCurrDayFreshTime(ctx)).Unix()
		user.Build = userBuildInfo
	}
	freeCount := cfg.BuildFree[0] - userBuildInfo.FreeCount
	if freeCount < 0 {
		freeCount = 0
	}
	payCount := cfg.BuildFree[1] - userBuildInfo.PayCount
	if payCount < 0 {
		payCount = 0
	}
	cost := make(map[values.ItemId]values.Integer)
	for id, count := range cfg.Cost {
		cost[id] = count + values.Integer(math.Floor(values.Float(count*userBuildInfo.PayCount*cfg.CostMultiple/10000)))
	}
	return reset, &BuildInfo{
		FreeCount: freeCount,
		PayCount:  payCount,
		Cost:      cost,
		ResetTime: userBuildInfo.ResetTime,
	}
}

func (svc *Service) addGuildExp(ctx *ctx.Context, guild *pbdao.Guild, cfg *rulemodel.Guild, surprised, cheatExp values.Integer) (values.Integer, bool) {
	var retExp values.Integer
	exp := cfg.GuildExp
	maxLevel := rule.GetMaxGuildLevel(ctx)
	if (exp <= 0 && cheatExp <= 0) || guild.Level >= maxLevel {
		return retExp, false
	}

	if cheatExp > 0 {
		guild.Exp += cheatExp
		retExp = cheatExp
	} else {
		exp *= surprised
		guild.Exp += exp
		retExp = exp
	}
	var levelup bool
	for guild.Level < maxLevel && cfg != nil && guild.Exp >= cfg.Exp {
		guild.Exp -= cfg.Exp
		guild.Level++
		levelup = true
		var ok bool
		cfg, ok = rule.GetGuildConfigByLevel(ctx, guild.Level)
		if !ok {
			break
		}
	}
	return retExp, levelup
}

func (svc *Service) worldInviteInfo(ctx *ctx.Context, guild *pbdao.Guild) *models.GuildWorldInvite {
	if guild.WorldInvite == nil {
		guild.WorldInvite = &models.GuildWorldInvite{
			Free:      0,
			Pay:       0,
			ResetTime: 0,
		}
	}
	free := rule.GetGuildWorldFreeInvite(ctx)
	pay := rule.GetGuildWorldPayInvite(ctx)
	if guild.WorldInvite.ResetTime <= timer.StartTime(ctx.StartTime).Unix() {
		guild.WorldInvite.Free = 0
		guild.WorldInvite.Pay = 0
		guild.WorldInvite.ResetTime = timer.BeginOfDay(timer.StartTime(ctx.StartTime)).Add(time.Hour * 24).Unix()
	} else {
		free -= guild.WorldInvite.Free
		if free <= 0 {
			free = 0
		}
		pay -= guild.WorldInvite.Pay
		if pay <= 0 {
			pay = 0
		}
	}
	return &models.GuildWorldInvite{
		Free:      free,
		Pay:       pay,
		ResetTime: guild.WorldInvite.ResetTime,
	}
}

// func (svc *Service) handleNormalInvite(ctx *ctx.Context, req *less_service.Guild_GuildHandleInviteRequest) (*models.Guild, *errmsg.ErrMsg) {
// 	guildUserDao := dao.NewGuildUser(ctx.RoleId)
// 	user, ok, err := guildUserDao.Get(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !ok {
// 		user = &pbdao.GuildUser{
// 			RoleId:  ctx.RoleId,
// 			GuildId: "",
// 			Cd:      0,
// 		}
// 	}
//
// 	err = svc.getLock(ctx, guildUserApplyLock+ctx.RoleId)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	inviteDao := dao.NewGuildInvite(ctx.RoleId)
// 	inviteList, err := inviteDao.Get(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	inviteItem := &pbdao.GuildInvite{}
// 	for _, invite := range inviteList {
// 		if invite.Uuid == req.Uuid {
// 			inviteItem = invite
// 			break
// 		}
// 	}
// 	var guildInfo *models.Guild
// 	if req.Agree {
// 		if user.GuildId != "" {
// 			return nil, errmsg.NewErrGuildHasJoined()
// 		}
// 		if inviteItem.Uuid == "" {
// 			return nil, errmsg.NewErrGuildInviteNotExist()
// 		}
// 		err = svc.getLock(ctx, guildLock+inviteItem.GuildId)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		guild, err := dao.NewGuild(inviteItem.GuildId).Get(ctx)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if guild == nil {
// 			return nil, errmsg.NewErrGuildNotExist()
// 		}
// 		memberDao := dao.NewGuildMember(inviteItem.GuildId)
// 		members, err := memberDao.Get(ctx)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if values.Integer(len(members)) >= rule.GetMaxMemberCount(ctx, guild.Level) {
// 			return nil, errmsg.NewErrGuildMemberFull()
// 		}
// 		memberPosition := rule.GetMemberPosition(ctx)
// 		user.GuildId = guild.Id
// 		member := &pbdao.GuildMember{
// 			RoleId:       ctx.RoleId,
// 			ServerId:     ctx.ServerId,
// 			Position:     memberPosition,
// 			JoinAt:       time.Now().Unix(),
// 			LastPosition: memberPosition,
// 		}
// 		members = append(members, member)
// 		if err := guildUserDao.Save(ctx, user); err != nil {
// 			return nil, err
// 		}
// 		if err := memberDao.SaveOne(ctx, member); err != nil {
// 			return nil, err
// 		}
// 		guildInfo = svc.dao2model(guild, members)
// 	}
// 	if err := inviteDao.DeleteOne(ctx, inviteItem); err != nil {
// 		return nil, err
// 	}
// 	if req.Agree {
// 		svc.guildSystemMsg(ctx, inviteItem.GuildId, &SystemMsg{
// 			TextId: join,
// 			Args:   []string{"玩家名字"}, // TODO
// 		})
// 		svc.joinChatRoom(ctx, ctx.RoleId, inviteItem.GuildId)
// 	}
// 	return guildInfo, nil
// }

func (svc *Service) handleInvite(ctx *ctx.Context, req *less_service.Guild_GuildHandleInviteRequest) (*models.Guild, map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	role, err := svc.GetRoleByRoleId(ctx, ctx.RoleId)
	if err != nil {
		return nil, nil, err
	}
	if role.Level < rule.GetGuildUserLevelLimit(ctx) {
		return nil, nil, errmsg.NewErrGuildUserLevelNotEnough()
	}
	token, err := svc.parseToken(req.Token)
	if err != nil {
		return nil, nil, err
	}
	guildUserDao := dao.NewGuildUser(ctx.RoleId)
	user, err := guildUserDao.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	if svc.isInCooling(ctx, user) {
		return nil, nil, errmsg.NewErrGuildJoinCD()
	}
	var guildInfo *models.Guild
	if user.GuildId != "" {
		return nil, nil, errmsg.NewErrGuildHasJoined()
	}

	guildId := token.GuildId
	err = svc.getLock(ctx, guildLock+guildId)
	if err != nil {
		return nil, nil, err
	}

	guildDao := dao.NewGuild(guildId)
	guild, err := guildDao.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	if guild == nil {
		return nil, nil, errmsg.NewErrGuildNotExist()
	}
	if guild.DeletedAt > 0 {
		return nil, nil, errmsg.NewErrGuildNotExist()
	}
	if !token.Private && role.Power < guild.CombatValueLimit {
		return nil, nil, errmsg.NewErrGuildCombatValueNotEnough()
	}
	memberDao := dao.NewGuildMember(guildId)
	members, err := memberDao.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	if values.Integer(len(members)) >= rule.GetMaxMemberCount(ctx, guild.Level) {
		return nil, nil, errmsg.NewErrGuildMemberFull()
	}
	user.GuildId = guild.Id
	p := rule.GetMemberPosition(ctx)
	member := &pbdao.GuildMember{
		RoleId:           ctx.RoleId,
		ServerId:         ctx.ServerId,
		Position:         p,
		JoinAt:           timer.StartTime(ctx.StartTime).Unix(),
		LastPosition:     p,
		CombatValue:      role.Power,
		ActiveValue:      user.ActiveValue,
		TotalActiveValue: user.TotalActiveValue,
	}
	rewards, err := svc.handlerFirstJoinRewards(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	guild.Count++
	members = append(members, member)

	if err := guildDao.Save(ctx, guild); err != nil {
		return nil, nil, err
	}
	if err := guildUserDao.Save(ctx, user); err != nil {
		return nil, nil, err
	}
	if err := memberDao.SaveOne(ctx, member); err != nil {
		return nil, nil, err
	}
	guildInfo = svc.dao2model(ctx, guild, members)

	if err := svc.updateGuildRankValue(ctx, guildInfo.Id, guildInfo.CombatValue); err != nil {
		return guildInfo, nil, err
	}

	svc.joinChatRoom(ctx, ctx.RoleId, guildId)

	if err := svc.updateToGuildFilterServer(ctx, guildInfo); err != nil {
		return nil, nil, err
	}
	svc.joinLocalEvent(ctx, ctx.RoleId, guild.Id)
	return guildInfo, rewards, nil
}

func (svc *Service) parseToken(tokenStr string) (*Claims, *errmsg.ErrMsg) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, errmsg.NewErrGuildInviteExpired()
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errmsg.NewErrGuildInviteExpired()
}

func (svc *Service) syncActive2User(user *pbdao.GuildUser, member *pbdao.GuildMember) {
	if member.ActiveValue == nil {
		member.ActiveValue = make(map[int64]int64)
	}
	user.ActiveValue = member.ActiveValue
	user.TotalActiveValue = member.TotalActiveValue
}

// 获取最近7天的key（含今天）
func (svc *Service) get7DaysKey(ctx *ctx.Context) []values.Integer {
	list := make([]values.Integer, 0, 7)
	today := timer.BeginOfDay(timer.StartTime(ctx.StartTime))
	list = append(list, today.Unix())
	for i := 1; i < 7; i++ {
		today = today.Add(-24 * time.Hour * time.Duration(i))
		list = append(list, today.Unix())
	}
	return list
}

func (svc *Service) get7DaysLastKey(ctx *ctx.Context) values.Integer {
	today := timer.BeginOfDay(timer.StartTime(ctx.StartTime))
	return today.Add(-24 * time.Hour * time.Duration(6)).Unix()
}

func (svc *Service) handlerFirstJoinRewards(ctx *ctx.Context, user *pbdao.GuildUser) (map[values.ItemId]values.Integer, *errmsg.ErrMsg) {
	rewards := make(map[values.ItemId]values.Integer)
	if !user.FirstJoinReward {
		cfg, ok := rule.GetGuildFirstJoinRewards(ctx)
		if ok {
			if err := svc.AddItem(ctx, ctx.RoleId, cfg.ItemId, cfg.Count); err != nil {
				return nil, err
			}
		}
		rewards[cfg.ItemId] = cfg.Count
		user.FirstJoinReward = true
	}
	return rewards, nil
}

func (svc *Service) get7DayActive(member *pbdao.GuildMember, sevenDaysKey []values.Integer) values.Integer {
	var total values.Integer
	for _, key := range sevenDaysKey {
		total += member.ActiveValue[key]
	}
	return total
}

func (svc *Service) get7DayActiveByGuildUser(user *pbdao.GuildUser, sevenDaysKey []values.Integer) values.Integer {
	var total values.Integer
	for _, key := range sevenDaysKey {
		total += user.ActiveValue[key]
	}
	return total
}

func (svc *Service) updateGuildRankValue(ctx *ctx.Context, guildId values.GuildId, combatValue values.Integer) *errmsg.ErrMsg {
	if _, err := svc.svc.GetNatsClient().Request(ctx, 0, &rank_service.RankService_UpdateRankValueRequest{
		RankValue: &models.RankValue{
			RankId:   svc.rankId,
			RankType: values.Integer(enum.RankGuild),
			OwnerId:  guildId,
			Value1:   combatValue,
		},
	}); err != nil {
		return err
	}
	return nil
}

func (svc *Service) deleteGuildRankValue(ctx *ctx.Context, guildId values.GuildId) *errmsg.ErrMsg {
	if _, err := svc.svc.GetNatsClient().Request(ctx, 0, &rank_service.RankService_DeleteRankValueRequest{
		RankId:  svc.rankId,
		OwnerId: guildId,
	}); err != nil {
		return err
	}
	return nil
}

func (svc *Service) updateToGuildFilterServer(ctx *ctx.Context, guild *models.Guild) *errmsg.ErrMsg {
	full := guild.Count >= rule.GetMaxMemberCount(ctx, guild.Level)
	_, err := svc.svc.GetNatsClient().Request(ctx, 0, &guild_filter_service.Guild_GuildFilterUpdateRequest{
		Id:               guild.Id,
		Name:             guild.Name,
		Level:            guild.Level,
		Lang:             guild.Lang,
		CombatValueLimit: guild.CombatValueLimit,
		AutoJoin:         guild.AutoJoin,
		Active:           guild.ActiveValue,
		Full:             full,
		Count:            guild.Count,
	})
	return err
}

func (svc *Service) getGuildDetails(ctx *ctx.Context, idList []values.GuildId) ([]*models.Guild, *errmsg.ErrMsg) {
	guildKeys := make([]orm.RedisInterface, 0, len(idList))
	for i := 0; i < len(idList); i++ {
		guildKeys = append(guildKeys, &pbdao.Guild{Id: idList[i]})
	}
	guildNotFound, err := orm.GetOrm(ctx).MGetPB(redisclient.GetGuildRedis(), guildKeys...)
	if err != nil {
		return nil, err
	}
	notFoundMap := make(map[int]bool, len(guildNotFound))
	for _, v := range guildNotFound {
		notFoundMap[v] = true
	}

	list := make([]*models.Guild, 0)
	for i := 0; i < len(guildKeys); i++ {
		if _, ok := notFoundMap[i]; ok {
			continue
		}
		model, ok := guildKeys[i].(*pbdao.Guild)
		if !ok {
			continue
		}
		// DeletedAt > 0  表示公会已解散
		if model == nil || model.DeletedAt > 0 {
			continue
		}
		members, err := dao.NewGuildMember(model.Id).Get(ctx)
		if err != nil {
			return nil, err
		}
		list = append(list, svc.dao2model(ctx, model, members))
	}
	return list, nil
}

func (svc *Service) guildExpChangePush(ctx *ctx.Context, guild *pbdao.Guild, members []*pbdao.GuildMember) {
	roleIds := make([]values.RoleId, 0, len(members))
	for _, member := range members {
		roleIds = append(roleIds, member.RoleId)
	}
	ctx.PushMessageToRoles(roleIds, &less_service.Guild_GuildExpChange{
		Level: guild.Level,
		Exp:   guild.Exp,
	})
}

func (svc *Service) userGuildInfoChangePush(ctx *ctx.Context, roleId values.RoleId, guildId values.GuildId, position values.GuildPosition) {
	ctx.PushMessageToRole(roleId, &less_service.Guild_UserGuildInfoChange{
		GuildId:  guildId,
		Position: position,
	})
}

func (svc *Service) usersGuildInfoChangePush(ctx *ctx.Context, list []*pbdao.GuildApply, guildId values.GuildId, position values.GuildPosition) {
	roleIds := make([]values.RoleId, 0, len(list))
	for _, apply := range list {
		roleIds = append(roleIds, apply.RoleId)
	}
	ctx.PushMessageToRoles(roleIds, &less_service.Guild_UserGuildInfoChange{
		GuildId:  guildId,
		Position: position,
	})
}

func (svc *Service) NameExist(_ *ctx.Context, req *less_service.Guild_GuildNameExistRequest) (*less_service.Guild_GuildNameExistResponse, *errmsg.ErrMsg) {
	if req.Name == "" {
		return nil, errmsg.NewErrInvalidRequestParam()
	}
	exist, err := dao.NameExist(req.Name)
	if err != nil {
		return nil, err
	}
	return &less_service.Guild_GuildNameExistResponse{
		Exist: exist,
	}, nil
}

func (svc *Service) joinLocalEvent(ctx *ctx.Context, roleId values.RoleId, guildId values.GuildId) {
	ctx.PublishEventLocal(&event.GuildEvent{
		RoleId:  roleId,
		GuildId: guildId,
		Action:  event.GuildActionJoin,
	})
}

func (svc *Service) exitLocalEvent(ctx *ctx.Context, roleId values.RoleId, guildId values.GuildId) {
	ctx.PublishEventLocal(&event.GuildEvent{
		RoleId:  roleId,
		GuildId: guildId,
		Action:  event.GuildActionLeave,
	})
}

func (svc *Service) CheatResetBuildTimes(ctx *ctx.Context, _ *less_service.Guild_CheatGuildResetBuildRequest) (*less_service.Guild_CheatGuildResetBuildResponse, *errmsg.ErrMsg) {
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if user.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	build := user.Build
	if build == nil {
		build = &models.GuildBuild{
			FreeCount: 0,
			PayCount:  0,
			ResetTime: timer.BeginOfDay(timer.StartTime(ctx.StartTime)).Add(time.Hour * 24).Unix(),
		}
	}
	build.FreeCount = 0
	build.PayCount = 0
	user.Build = build
	if err := dao.NewGuildUser(ctx.RoleId).Save(ctx, user); err != nil {
		return nil, err
	}
	return &less_service.Guild_CheatGuildResetBuildResponse{}, nil
}

func (svc *Service) CheatAddGuildExp(ctx *ctx.Context, req *less_service.Guild_CheatGuildAddExpRequest) (*less_service.Guild_CheatGuildAddExpResponse, *errmsg.ErrMsg) {
	if req.Exp <= 0 {
		return nil, nil
	}
	user, err := svc.getGuidUser(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if user.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	guild, err := dao.NewGuild(user.GuildId).Get(ctx)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	cfg, ok := rule.GetGuildConfigByLevel(ctx, guild.Level)
	if !ok {
		return nil, errmsg.NewInternalErr("no guild config")
	}
	members, err := dao.NewGuildMember(guild.Id).Get(ctx)
	if err != nil {
		return nil, err
	}
	exp, levelup := svc.addGuildExp(ctx, guild, cfg, 1, req.Exp)
	if levelup {
		if err := svc.updateToGuildFilterServer(ctx, svc.dao2model(ctx, guild, members)); err != nil {
			return nil, err
		}
	}
	if exp > 0 {
		svc.guildExpChangePush(ctx, guild, members)
	}
	if err := dao.NewGuild(guild.Id).Save(ctx, guild); err != nil {
		return nil, err
	}
	return &less_service.Guild_CheatGuildAddExpResponse{}, nil
}

func (svc *Service) CheatRestExitGuildCD(ctx *ctx.Context, _ *less_service.Guild_CheatRestExitGuildCDRequest) (*less_service.Guild_CheatRestExitGuildCDResponse, *errmsg.ErrMsg) {
	userDao := dao.NewGuildUser(ctx.RoleId)
	user, err := userDao.Get(ctx)
	if err != nil {
		return nil, err
	}
	user.Cd = 0
	if err := userDao.Save(ctx, user); err != nil {
		return nil, err
	}
	return &less_service.Guild_CheatRestExitGuildCDResponse{}, nil
}
