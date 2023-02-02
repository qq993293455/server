package service

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/sensitive"
	"coin-server/common/service"
	"coin-server/common/values"
	"coin-server/game-server/module"
	"coin-server/game-server/service/achievement"
	"coin-server/game-server/service/activity"
	activityWeekly "coin-server/game-server/service/activity-weekly"
	"coin-server/game-server/service/arena"
	"coin-server/game-server/service/atlas"
	"coin-server/game-server/service/bag"
	"coin-server/game-server/service/battle"
	bossHall "coin-server/game-server/service/boss-hall"
	discord "coin-server/game-server/service/discord"
	"coin-server/game-server/service/divination"
	equip_forge "coin-server/game-server/service/equip-forge"
	"coin-server/game-server/service/expedition"
	"coin-server/game-server/service/formation"
	"coin-server/game-server/service/friend"
	"coin-server/game-server/service/gacha"
	googleQuest "coin-server/game-server/service/google_quest"
	"coin-server/game-server/service/guide"
	"coin-server/game-server/service/guild"
	guild_boss "coin-server/game-server/service/guild-boss"
	guild_gvg "coin-server/game-server/service/guild-gvg"
	"coin-server/game-server/service/hero"
	"coin-server/game-server/service/im"
	"coin-server/game-server/service/journey"
	"coin-server/game-server/service/mail"
	mapevent "coin-server/game-server/service/map-event"
	MoonthlyCard "coin-server/game-server/service/moonthly_card"
	"coin-server/game-server/service/notice"
	"coin-server/game-server/service/npc"
	personalboss "coin-server/game-server/service/personal-boss"
	"coin-server/game-server/service/plane-dungeon"
	predownload "coin-server/game-server/service/pre_download"
	questReward "coin-server/game-server/service/quest-reward"
	racing_rank "coin-server/game-server/service/racing-rank"
	"coin-server/game-server/service/rank"
	"coin-server/game-server/service/refresh"
	"coin-server/game-server/service/relics"
	roguelike "coin-server/game-server/service/roguelike"
	sevenDays "coin-server/game-server/service/sevendays"
	"coin-server/game-server/service/shop"
	"coin-server/game-server/service/stage"
	"coin-server/game-server/service/statistic"
	system_unlock "coin-server/game-server/service/system-unlock"
	"coin-server/game-server/service/talent"
	loop_task "coin-server/game-server/service/task/loop-task"
	maintask "coin-server/game-server/service/task/main-task"
	npc_task "coin-server/game-server/service/task/npc-task"
	ritualtask "coin-server/game-server/service/task/ritual-task"
	"coin-server/game-server/service/task/task"
	"coin-server/game-server/service/tower"
	"coin-server/game-server/service/user"
	xDayGoal "coin-server/game-server/service/xdaygoal"

	"go.uber.org/zap"
)

type Service struct {
	*module.Module
	svc        *service.Service
	log        *logger.Logger
	serverId   values.ServerId
	serverType models.ServerType
}

func NewService(
	urls []string,
	log *logger.Logger,
	serverId values.ServerId,
	serverType models.ServerType,
) *Service {
	svc := service.NewService(urls, log, serverId, serverType, true, true, eventlocal.CreateEventLocal(true))
	svc.AddBussMid(system_unlock.SystemChecker)
	s := &Service{
		svc:        svc,
		log:        log,
		serverId:   serverId,
		serverType: serverType,
		Module:     &module.Module{},
	}
	return s
}

func (this_ *Service) Serve() {

	this_.Router()
	this_.svc.Start(func(event interface{}) {
		this_.log.Warn("unknown event", zap.Any("event", event))
	}, true)
}

func (this_ *Service) Stop() {
	this_.svc.Close()
}

func (this_ *Service) Router() {
	refresh.NewRefreshService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log)
	bag.NewBagService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	user.NewUserService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	task.NewTaskService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router() // 优先加载
	mail.NewMailService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	im.NewImService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	rank.NewRankService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	achievement.NewAchievementService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	loop_task.NewLoopTaskService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	battle.NewService(this_.serverId, this_.serverType, this_.svc, this_.log, this_.Module).Router()
	shop.NewShopService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	friend.NewFriendService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	guild.NewGuildService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router().InitGuildRankId()
	stage.NewStageService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	maintask.NewMainTaskService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	hero.NewHeroService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	divination.NewDivinationService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	mapevent.NewMapEventService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	npc.NewNpcService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	npc_task.NewNpcTaskService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	atlas.NewAtlasService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	talent.NewTalentService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	relics.NewRelicsService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	equip_forge.NewEquipForgeService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	gacha.NewGachaService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	ritualtask.NewRitualTaskService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	tower.NewTowerService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	system_unlock.NewSysUnlockService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	formation.NewFormationService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	arena.NewArenaService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	activityWeekly.NewActivityWeeklyService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	guild_boss.NewGuildBossService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	racing_rank.NewRacingRank(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	guide.NewGuideService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	plane.NewPlaneService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	expedition.NewExpeditionService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	sevenDays.NewSevenDaysService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	xDayGoal.NewXDayGoalService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	MoonthlyCard.NewMoonthlyCardService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	discord.NewDiscordService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	questReward.NewQuestRewardService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	activity.NewActivityService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	statistic.NewStatisticService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	googleQuest.NewGoogleQuestService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	predownload.NewPreDownloadService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	bossHall.BossHallService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	journey.NewJourneyService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	notice.NewNoticeService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	personalboss.NewPersonalBossService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	roguelike.NewRoguelikeService(this_.serverId, this_.serverType, this_.svc, this_.Module).Router()
	guild_gvg.NewGuildGVGService(this_.serverId, this_.serverType, this_.svc, this_.Module, this_.log).Router()
	this_.svc.RegisterFunc("检查敏感字", this_.CheckSensitiveText)
}

func (this_ *Service) CheckSensitiveText(c *ctx.Context, req *servicepb.Common_CheckSensitiveTextRequest) (*servicepb.Common_CheckSensitiveTextResponse, *errmsg.ErrMsg) {
	if req.Txt == "" {
		return &servicepb.Common_CheckSensitiveTextResponse{IsPass: true}, nil
	}
	isPass := true
	if !sensitive.TextValid(req.Txt) {
		isPass = false
	}
	return &servicepb.Common_CheckSensitiveTextResponse{IsPass: isPass}, nil
}
