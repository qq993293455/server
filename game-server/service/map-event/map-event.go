package map_event

import (
	"fmt"
	"math/rand"
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/iggsdk"
	"coin-server/common/logger"
	mapdata "coin-server/common/map-data"
	"coin-server/common/proto/cppbattle"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/statistical"
	models2 "coin-server/common/statistical/models"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/map-event/dao"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"

	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewMapEventService(
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
	module.MapEventService = s
	return s
}

func (s *Service) Router() {
	s.svc.RegisterFunc("获取逸闻", s.GetMapStoryRequest)
	s.svc.RegisterFunc("获取地图事件", s.GetMapEventRequest)
	s.svc.RegisterFunc("获取逸闻奖励", s.GetStoryRewardRequest)
	s.svc.RegisterFunc("刷新事件", s.RefreshEventRequest)
	s.svc.RegisterFunc("开始事件", s.AppointEventStart)
	s.svc.RegisterFunc("结束事件", s.ComplateEventRequest)
	s.svc.RegisterFunc("查询事件", s.GetAppointEventInfo)

	s.svc.RegisterFunc("开宝箱游戏", s.TreasureChestRequest)
	s.svc.RegisterFunc("采矿游戏", s.CollectMineRequest)
	s.svc.RegisterFunc("黑市游戏", s.BlackMarketRequest)
	s.svc.RegisterFunc("拼图游戏", s.JigsawPuzzleRequest)
	s.svc.RegisterFunc("翻牌游戏", s.CardMatchRequest)
	s.svc.RegisterFunc("通缉游戏", s.WantingRequest)
	s.svc.RegisterFunc("偶遇游戏", s.MeetRequest)
	s.svc.RegisterFunc("偶遇战斗开始", s.MeetBattleStart)
	s.svc.RegisterFunc("偶遇战斗结束", s.MeetBattleFinish)
	s.svc.RegisterFunc("武器制作游戏", s.BuildArmsRequest)
	s.svc.RegisterFunc("开启地图事件", s.UnlockEventRequest)
	s.svc.RegisterFunc("重发 完成消息", s.ReSendFinishRequset)

	s.svc.RegisterFunc("作弊器开启地图事件", s.CheatUnlockEventRequest)
	s.svc.RegisterFunc("作弊器完成地图事件", s.CheatFinishEventRequest)
	s.svc.RegisterFunc("作弊器添加地图事件", s.CheatAddEventRequest)
	s.svc.RegisterFunc("作弊器更改地图", s.CheatChangeMapRequest)
	s.svc.RegisterFunc("作弊器获取逸闻碎片", s.CheatGetStoryRequest)

	eventlocal.SubscribeEventLocal(s.HandleRoleLoginEvent)

	eventlocal.SubscribeEventLocal(s.HandleEventFinished)
}

func (s *Service) AddEvent(ctx *ctx.Context, roleId values.RoleId, eventId values.EventId, mapId values.MapId) *errmsg.ErrMsg {
	e, err := dao.GetMapEvent(ctx, roleId)
	if err != nil {
		return err
	}
	newEvent := s.randPointEvent(ctx, eventId, mapId, e.Triggered)
	e.Triggered = append(e.Triggered, newEvent)
	dao.SaveMapEvent(ctx, e)
	return nil
}

func (s *Service) GetMapStoryRequest(ctx *ctx.Context, _ *servicepb.MapEvent_GetMapStoryRequest) (*servicepb.MapEvent_GetMapStoryResponse, *errmsg.ErrMsg) {
	story, err := dao.GetMapStory(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.MapEvent_GetMapStoryResponse{
		Story: StoryDao2Models(story),
	}, nil
}

func (s *Service) GetMapEventRequest(ctx *ctx.Context, _ *servicepb.MapEvent_GetMapEventRequest) (*servicepb.MapEvent_GetMapEventResponse, *errmsg.ErrMsg) {
	e, err := dao.GetMapEvent(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, errmsg.NewErrMapEventLock()
	}
	refresh, err := dao.GetEventRefresh(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	ret := &servicepb.MapEvent_GetMapEventResponse{
		Events:      e.Triggered,
		Count:       0,
		NextRefresh: 0,
		Curr:        refresh.Curr,
	}
	limit := s.getRefreshLimit(ctx)
	// 有几次可刷新
	pass := timer.StartTime(ctx.StartTime).UnixMilli() - refresh.Refreshed
	gap := s.getRefreshGap(ctx) * 1000
	times := pass / gap
	if times >= limit {
		ret.Count = limit
		ret.NextRefresh = 0
	} else {
		ret.Count = times
		ret.NextRefresh = refresh.Refreshed + (times+1)*gap
	}
	return ret, nil
}

func (s *Service) GetStoryRewardRequest(ctx *ctx.Context, req *servicepb.MapEvent_GetStoryRewardRequest) (*servicepb.MapEvent_GetStoryRewardResponse, *errmsg.ErrMsg) {
	story, err := dao.GetMapStoryById(ctx, ctx.RoleId, req.StoryId)
	if err != nil {
		return nil, err
	}
	if story == nil {
		return nil, errmsg.NewErrMapStoryNotExist()
	}
	if story.Status != models.RewardStatus_Unlocked {
		return nil, nil
	}
	reader := rule.MustGetReader(ctx)
	anecdote, ok := reader.Anecdotes.GetAnecdotesById(req.StoryId)
	if !ok {
		return nil, nil
	}
	_, err = s.BagService.AddManyItem(ctx, ctx.RoleId, anecdote.CollectItem)
	if err != nil {
		return nil, err
	}

	story.Status = models.RewardStatus_Received
	err = dao.SaveMapStory(ctx, ctx.RoleId, story)
	if err != nil {
		return nil, err
	}
	return &servicepb.MapEvent_GetStoryRewardResponse{
		Reward: anecdote.CollectItem,
	}, nil
}

// TODO 支持指定刷新到某地图某个点

func (s *Service) RefreshEventRequest(ctx *ctx.Context, _ *servicepb.MapEvent_RefreshEventRequest) (*servicepb.MapEvent_RefreshEventResponse, *errmsg.ErrMsg) {
	events, err := dao.GetMapEvent(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if events == nil {
		return nil, nil
	}
	refresh, err := dao.GetEventRefresh(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}
	if events.BattleMapId == curRes.SceneId {
		// 若当前事件未完成，返回当前事件
		if refresh.Curr != 0 {
			var em *models.MapEvent
			for _, e := range events.Triggered {
				if refresh.Curr == e.EventId {
					em = e
				}
			}
			if em == nil {
				return nil, errmsg.NewErrMapEventNotFinish()
			}
			return &servicepb.MapEvent_RefreshEventResponse{Event: em}, nil
		}
	}

	t := s.getRefreshGap(ctx) * 1000

	// 没到刷新时间
	gap := timer.StartTime(ctx.StartTime).UnixMilli() - refresh.Refreshed
	if gap < t {
		return nil, nil
	}

	newEvent, err := s.refreshEvent(ctx, refresh, events)
	if err != nil {
		return nil, err
	}
	dao.SaveMapEvent(ctx, events)
	dao.SaveEventRefresh(ctx, refresh)
	// 逸闻探索打点
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskTaskMapEventNum: {
			Typ:     values.Integer(models.TaskType_TaskTaskMapEventNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskTaskMapEventNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskTaskMapEventNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}
	s.TaskService.UpdateTargets(ctx, ctx.RoleId, tasks)
	return &servicepb.MapEvent_RefreshEventResponse{Event: newEvent}, nil
}

func (s *Service) ComplateEventRequest(ctx *ctx.Context, req *servicepb.MapEvent_ComplateEventRequest) (*servicepb.MapEvent_ComplateEventResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(req.EventId)
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	if cfg.Typ != 2 {
		return nil, errmsg.NewErrMapGameNotExist()
	}
	err := s.finishEvent(ctx, 0, req.EventId, true, nil, 0, 1)
	if err != nil {
		return nil, err
	}
	return &servicepb.MapEvent_ComplateEventResponse{}, nil
}

func (s *Service) AppointEventStart(ctx *ctx.Context, req *servicepb.MapEvent_AppointEventStartRequest) (*servicepb.MapEvent_AppointEventStartResponse, *errmsg.ErrMsg) {
	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}
	datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
	if err != nil {
		return nil, err
	}

	for _, data := range datas {
		if data.EventId == req.EventId && data.MapId == curRes.SceneId {
			if data.IsOver {
				return nil, errmsg.NewErrAppointMapEventComplate()
			}
			return &servicepb.MapEvent_AppointEventStartResponse{}, nil
		}
	}

	datas = append(datas, &models.AppointMapEvent{
		EventId: req.EventId,
		MapId:   curRes.SceneId,
		GameId:  req.GameId,
	})

	// 逸闻探索打点
	tasks := map[models.TaskType]*models.TaskUpdate{
		models.TaskType_TaskTaskMapEventNum: {
			Typ:     values.Integer(models.TaskType_TaskTaskMapEventNum),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
		models.TaskType_TaskTaskMapEventNumAcc: {
			Typ:     values.Integer(models.TaskType_TaskTaskMapEventNumAcc),
			Id:      0,
			Cnt:     1,
			Replace: false,
		},
	}

	s.TaskService.UpdateTargets(ctx, ctx.RoleId, tasks)

	dao.SaveAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId, datas)
	return &servicepb.MapEvent_AppointEventStartResponse{}, nil
}

func (s *Service) GetAppointEventInfo(ctx *ctx.Context, req *servicepb.MapEvent_GetAppointEventInfoRequest) (*servicepb.MapEvent_GetAppointEventInfoResponse, *errmsg.ErrMsg) {
	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}
	data, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
	if err != nil {
		return nil, err
	}
	return &servicepb.MapEvent_GetAppointEventInfoResponse{
		EventInfos: data,
	}, nil
}

// ---------------------------------------------------cheat-----------------------------------------------------------//
func (s *Service) ReSendFinishRequset(ctx *ctx.Context, req *servicepb.MapEvent_FixAppointEventFinishStatusRequest) (*servicepb.MapEvent_FixAppointEventFinishStatusResponse, *errmsg.ErrMsg) {
	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}

	datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
	if err != nil {
		return nil, err
	}

	hasEvent := false
	for _, data := range datas {
		if data.EventId == req.EventId {
			if !data.IsOver {
				return nil, errmsg.NewErrMapEventNotFinish()
			}
			hasEvent = true
			s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGivenMapEvent, req.EventId, 1)
			break
		}
	}

	if !hasEvent {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	return &servicepb.MapEvent_FixAppointEventFinishStatusResponse{}, nil
}

func (s *Service) UnlockEventRequest(ctx *ctx.Context, _ *servicepb.MapEvent_UnlockEventRequest) (*servicepb.MapEvent_UnlockEventResponse, *errmsg.ErrMsg) {
	err := s.unlockMapEvent(ctx)
	if err != nil {
		return nil, err
	}
	return &servicepb.MapEvent_UnlockEventResponse{}, nil
}

func (s *Service) CheatUnlockEventRequest(ctx *ctx.Context, _ *servicepb.MapEvent_CheatUnlockEventRequest) (*servicepb.MapEvent_CheatUnlockEventResponse, *errmsg.ErrMsg) {
	err := s.unlockMapEvent(ctx)
	if err != nil {
		return nil, err
	}
	return &servicepb.MapEvent_CheatUnlockEventResponse{}, nil
}

func (s *Service) CheatAddEventRequest(ctx *ctx.Context, req *servicepb.MapEvent_CheatAddEventRequest) (*servicepb.MapEvent_CheatAddEventResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(ctx)
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(req.EventId)
	if !ok {
		return nil, errmsg.NewErrMapGameNotExist()
	}

	if cfg.Typ == 1 {
		e, err := dao.GetMapEvent(ctx, ctx.RoleId)
		if err != nil {
			return nil, err
		}
		newEvent := &models.MapEvent{
			EventId: req.EventId,
			MapId:   req.MapId,
			GameId:  req.GameId,
			Point:   req.Point,
		}
		if req.EventId == 1 {
			gameCfg, ok := r.AnecdotesGame1.GetAnecdotesGame1ById(req.GameId)
			if !ok {
				return nil, errmsg.NewErrMapEventAddFailed()
			}
			newEvent.Extra = []string{strconv.Itoa(rand.Intn(len(gameCfg.StoryText)))}
		}
		if req.EventId == 7 {
			gameCfg, ok := r.AnecdotesGame7.GetAnecdotesGame7ById(req.GameId)
			if !ok {
				return nil, errmsg.NewErrMapEventAddFailed()
			}
			newEvent.Extra = []string{strconv.Itoa(rand.Intn(len(gameCfg.StoryText)))}
		}
		e.Triggered = append(e.Triggered, newEvent)
		dao.SaveMapEvent(ctx, e)
		return &servicepb.MapEvent_CheatAddEventResponse{}, nil
	}

	if cfg.Typ == 2 {
		curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
		if err1 != nil {
			return nil, err1
		}

		datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
		if err != nil {
			return nil, err
		}

		for _, data := range datas {
			if data.EventId == req.EventId && data.MapId == curRes.SceneId {
				if data.IsOver {
					return nil, errmsg.NewErrAppointMapEventComplate()
				}
				return &servicepb.MapEvent_CheatAddEventResponse{}, nil
			}
		}

		datas = append(datas, &models.AppointMapEvent{
			EventId: req.EventId,
			MapId:   curRes.SceneId,
			GameId:  req.GameId,
		})

		dao.SaveAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId, datas)
		return &servicepb.MapEvent_CheatAddEventResponse{}, nil
	}
	return nil, errmsg.NewErrMapGameNotExist()
}

func (s *Service) CheatFinishEventRequest(ctx *ctx.Context, req *servicepb.MapEvent_CheatFinishEventRequest) (*servicepb.MapEvent_CheatFinishEventResponse, *errmsg.ErrMsg) {
	err := s.finishEvent(ctx, 0, req.EventId, true, nil, 0, 1)
	if err != nil {
		return nil, err
	}
	return &servicepb.MapEvent_CheatFinishEventResponse{}, nil
}

func (s *Service) CheatChangeMapRequest(ctx *ctx.Context, req *servicepb.MapEvent_CheatChangeMapRequest) (*servicepb.MapEvent_CheatChangeMapResponse, *errmsg.ErrMsg) {
	// e, err := dao.GetMapEvent(ctx, ctx.RoleId)
	// if err != nil {
	//	return nil, err
	// }
	// if e == nil {
	//	return nil, nil
	// }
	// s.changeBattleMap(ctx, e, req.MapId)
	return &servicepb.MapEvent_CheatChangeMapResponse{}, nil
}

func (s *Service) CheatGetStoryRequest(ctx *ctx.Context, req *servicepb.MapEvent_CheatGetStoryRequest) (*servicepb.MapEvent_CheatGetStoryResponse, *errmsg.ErrMsg) {
	story, err := dao.GetMapStoryById(ctx, ctx.RoleId, req.StoryId)
	if err != nil {
		return nil, err
	}
	if story == nil {
		return nil, nil
	}
	story.Piece = EncodeBit1(story.Piece, req.Piece)
	err = dao.SaveMapStory(ctx, ctx.RoleId, story)
	if err != nil {
		return nil, err
	}
	piece, err := dao.GetStoryPiece(ctx, ctx.RoleId)
	piece.Drop[req.StoryId]--
	if piece.Drop[req.StoryId] == 0 {
		delete(piece.Drop, req.StoryId)
		story.Status = models.RewardStatus_Unlocked
		err = dao.SaveMapStory(ctx, ctx.RoleId, story)
		if err != nil {
			return nil, err
		}
	}
	err = dao.SaveStoryPiece(ctx, piece)
	if err != nil {
		return nil, err
	}
	return &servicepb.MapEvent_CheatGetStoryResponse{}, nil
}

// ------------------------------------------------------service-------------------------------------------------------//

func (s *Service) randPointEvent(ctx *ctx.Context, eventId values.EventId, mapId values.MapId, triggered []*models.MapEvent) *models.MapEvent {
	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		panic("GetCurrBattleInfo")
	}
	// 当前地图的刷新点
	points := s.getEventPoints(ctx, curRes.SceneId)
	// 找可用刷新点
	for _, tri := range triggered {
		if tri.MapId == curRes.SceneId {
			delete(points, tri.Point)
		}
	}
	if len(points) == 0 {
		panic("无可用刷新点")
	}
	newEvent := &models.MapEvent{
		EventId: eventId,
		MapId:   mapId,
		GameId:  0,
		Point:   0,
	}
	// 随机一个点刷事件，由于map遍历无序
	for k := range points {
		newEvent.Point = k
		break
	}
	return newEvent
}

func (s *Service) unlockMapEvent(ctx *ctx.Context) *errmsg.ErrMsg {
	e, err := dao.GetMapEvent(ctx, ctx.RoleId)
	if err != nil {
		return nil
	}
	if e != nil {
		return nil
	}
	r := rule.MustGetReader(ctx)

	e = &pbdao.MapEvent{
		RoleId:    ctx.RoleId,
		Triggered: make([]*models.MapEvent, 0),
	}

	refresh := &pbdao.MapEventRefresh{
		RoleId:    ctx.RoleId,
		Status:    map[values.Integer]models.EventStatus{},
		Refreshed: 0,
		Curr:      0,
	}

	cfg := r.Anecdotes.GetAutoRefreshEvent()
	for _, games := range cfg {
		refresh.Status[games.Id] = models.EventStatus_EventNotEffect
	}

	stories := make([]*pbdao.MapStory, 0)
	drop := make(map[values.StoryId]values.Integer, 0)
	for k, cnfs := range r.Anecdotes.CustomAnecdotes() {
		anecdote := &pbdao.MapStory{
			StoryId: k,
			Piece:   0,
			Status:  models.RewardStatus_Locked,
		}
		drop[k] = int64(len(cnfs))

		for _, cnf := range cnfs {
			if cnf.DefaultHave == 1 {
				anecdote.Piece = anecdote.Piece | 1<<(cnf.Id-1)
				drop[k]--
				if drop[k] == 0 {
					delete(drop, k)
					anecdote.Status = models.RewardStatus_Unlocked
				}
			}
		}

		stories = append(stories, anecdote)
	}
	dao.SaveManyMapStory(ctx, ctx.RoleId, stories)

	piece := &pbdao.StoryPiece{
		RoleId: ctx.RoleId,
		Drop:   drop,
	}
	err = dao.SaveStoryPiece(ctx, piece)
	if err != nil {
		return err
	}
	dao.SaveMapEvent(ctx, e)
	dao.SaveEventRefresh(ctx, refresh)
	return nil
}

// 完成任务后推送事件奖励、掉落碎片
func (s *Service) finishEvent(ctx *ctx.Context, gameId int64, eventId values.EventId, success bool, rewards map[values.ItemId]values.Integer, dropListId int64, ratio values.Integer) *errmsg.ErrMsg {
	r := rule.MustGetReader(ctx)
	// 更新事件
	events, err := dao.GetMapEvent(ctx, ctx.RoleId)
	if err != nil {
		return err
	}
	cfg, ok := r.AnecdotesGames.GetAnecdotesGamesById(eventId)
	if !ok {
		panic(fmt.Sprintf("没找到id=%d的逸闻配置", eventId))
	}

	isWin := int64(0)
	if success {
		isWin = 1
	}

	// 埋点
	statistical.Save(ctx.NewLogServer(), &models2.EventGames{
		IggId:     iggsdk.ConvertToIGGId(ctx.UserId),
		EventTime: timer.Now(),
		GwId:      statistical.GwId(),
		RoleId:    ctx.RoleId,
		GameId:    gameId,
		IsWin:     isWin,
	})

	mapId := int64(0)
	if cfg.Typ == 1 {
		refresh, err1 := dao.GetEventRefresh(ctx, ctx.RoleId)
		if err1 != nil {
			return err1
		}
		// 如果之前累计是满的，设置刷新时间为比上限少一次
		t := s.getRefreshGap(ctx) * 1000
		gap := timer.StartTime(ctx.StartTime).UnixMilli() - refresh.Refreshed
		if gap/t >= s.getRefreshLimit(ctx) {
			refresh.Refreshed = timer.StartTime(ctx.StartTime).UnixMilli() - (s.getRefreshLimit(ctx)-1)*t
		}

		_, ok1 := refresh.Status[eventId]
		if ok1 {
			refresh.Status[eventId] = models.EventStatus_EventCompleted
		}
		refresh.Curr = 0
		dao.SaveEventRefresh(ctx, refresh)

		var e *models.MapEvent
		for idx := range events.Triggered {
			if events.Triggered[idx].EventId == eventId {
				e = events.Triggered[idx]
				events.Triggered = append(events.Triggered[:idx], events.Triggered[idx+1:]...)
				break
			}
		}
		if e == nil {
			return errmsg.NewErrMapGameNotExist()
		}
		mapId = e.MapId
		dao.SaveMapEvent(ctx, events)
		s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskFinishMapEvent, 0, 1)
	}

	if cfg.Typ == 2 {
		curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
		if err1 != nil {
			return err1
		}

		datas, err := dao.LoadAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId)
		if err != nil {
			return err
		}
		mapId = curRes.SceneId
		hasEvent := false
		for _, data := range datas {
			if data.EventId == eventId && !data.IsOver {
				hasEvent = true
				data.IsOver = true
				break
			}
		}

		if !hasEvent {
			return errmsg.NewErrMapGameNotExist()
		}

		dao.SaveAppointMapEvent(ctx, ctx.RoleId, curRes.SceneId, datas)
		s.TaskService.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskGivenMapEvent, eventId, 1)
		s.log.Info("UpdateTarget", zap.Any("roleId", ctx.RoleId), zap.Any("TaskType", models.TaskType_TaskGivenMapEvent), zap.Any("eventId", eventId))
	}

	eve := &event.MapEventFinished{
		EventId:   eventId,
		IsSuccess: success,
		Rewards:   rewards,
		StoryId:   0,
		Piece:     0,
		MapId:     mapId,
		Ratio:     ratio,
	}
	// 掉落卡牌
	if success {
		rn := rand.Int63n(10000)
		var (
			storyId values.StoryId
			piece   values.Integer
		)
		if rn < cfg.CardPro {
			if len(cfg.DropId) > 0 {
				story, err1 := dao.GetStoryPiece(ctx, ctx.RoleId)
				if err1 != nil {
					return err1
				}

				storyNum := len(story.Drop)
				if storyNum != 0 {
					total := int64(0)
					weight := make([]values.Integer, 0)
					ids := make([]values.Integer, 0)

					for id, cnt := range story.Drop {
						isFind := false
						for _, dropId := range cfg.DropId {
							if dropId == id {
								isFind = true
								break
							}
						}
						if !isFind {
							continue
						}
						total += cnt
						weight = append(weight, total)
						ids = append(ids, id)
					}

					rn = rand.Int63n(total)
					var offset values.Integer
					for idx := range ids {
						if rn < weight[idx] {
							storyId = ids[idx]
							if idx == 0 {
								offset = rn
							} else {
								offset = rn - weight[idx-1]
							}
							break
						}
					}

					anecdote, err2 := dao.GetMapStoryById(ctx, ctx.RoleId, storyId)
					if err2 != nil {
						return err2
					}

					anecdoteCfg, ok1 := r.Anecdotes.GetAnecdotesById(storyId)
					if !ok1 {
						panic(fmt.Sprintf("没找到id=%d的故事配置", storyId))
					}

					zeroBit := DecodeBit0(anecdote.Piece, int64(len(r.Anecdotes.CustomAnecdotes()[storyId])))
					if anecdoteCfg.Typ == 5 {
						piece = zeroBit[0]
					} else {
						piece = zeroBit[offset]
					}

					anecdote.Piece = EncodeBit1(anecdote.Piece, piece)
					story.Drop[storyId]--
					if story.Drop[storyId] == 0 {
						delete(story.Drop, storyId)
						anecdote.Status = models.RewardStatus_Unlocked
					}
					err = dao.SaveStoryPiece(ctx, story)
					if err != nil {
						return err
					}
					err = dao.SaveMapStory(ctx, ctx.RoleId, anecdote)
					if err != nil {
						return err
					}
				}
			}
		}
		eve.StoryId = storyId
		eve.Piece = piece

		if dropListId > 0 {
			s.GetDropList(ctx, dropListId, rewards)
		}
	}

	// 获取道具
	_, err = s.BagService.AddManyItem(ctx, ctx.RoleId, rewards)
	if err != nil {
		return err
	}

	if len(rewards) > 0 {
		curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
		if err1 != nil {
			ctx.Error("GetCurrBattleInfo error", zap.Any("roleid", ctx.RoleId), zap.Any("err msg", err1))
			return err1
		}

		notice := &cppbattle.CPPBattle_DropItemPush{}
		for itemId, count := range rewards {
			notice.Items = append(notice.Items, &models.Item{
				ItemId: itemId,
				Count:  count,
			})
		}
		s.log.Info("CPPBattle_DropItemPush", zap.Any("battle id", curRes.BattleId))
		err = s.svc.GetNatsClient().PublishCtx(ctx, curRes.BattleId, notice)
		if err != nil {
			s.log.Error("push DropItemPush error", zap.Any("Error Msg", err))
		}
	}

	// 发送结束事件
	ctx.PublishEventLocal(eve)

	return nil
}

func (s *Service) GetDropList(ctx *ctx.Context, dropId int64, rewards map[int64]int64) {
	r := rule.MustGetReader(ctx)
	dropListsMini := r.DropListsMini.List()

	var dropList []int64
	prob := rand.Int63n(10001)
	for _, dInfo := range dropListsMini {
		if dInfo.DropListsId != dropId {
			continue
		}
		if prob <= dInfo.DropProb {
			dropList = append(dropList, dInfo.DropId)
		}
	}

	dropMiniCnfs := r.DropMini.List()
	var dropMList map[int64][]rule_model.DropMini = make(map[int64][]rule_model.DropMini)
	for _, dropListId := range dropList {
		itemTotalWeight := int64(0)
		_, ok := dropMList[dropListId]
		if !ok {
			for _, dropMiniCnf := range dropMiniCnfs {
				if dropMiniCnf.DropId == dropListId {
					var dCnf rule_model.DropMini = dropMiniCnf
					dCnf.ItemWeight += itemTotalWeight
					itemTotalWeight = dCnf.ItemWeight
					dropMList[dropListId] = append(dropMList[dropListId], dCnf)
				}
			}
		}
		dProb := rand.Int63n(itemTotalWeight + 1)
		for _, dItem := range dropMList[dropListId] {
			if dProb > dItem.ItemWeight {
				continue
			}
			for itemId, itemCnt := range dItem.ItemId {
				rewards[itemId] += itemCnt
			}
			break
		}
	}
}

// 刷自动刷新的事件
func (s *Service) refreshEvent(ctx *ctx.Context, refresh *pbdao.MapEventRefresh, events *pbdao.MapEvent) (*models.MapEvent, *errmsg.ErrMsg) {
	notEffect := make([]values.EventId, 0)
	triggered := values.EventId(0)

	curRes, err1 := s.Module.GetCurrBattleInfo(ctx, &servicepb.GameBattle_GetCurrBattleInfoRequest{})
	if err1 != nil {
		return nil, err1
	}
	mapCnf := mapdata.GetLogicMapCnf(ctx, curRes.SceneId)

	for k, v := range refresh.Status {
		if v == models.EventStatus_EventNotEffect {
			for _, id := range mapCnf.ContainedEvent {
				if k == id {
					notEffect = append(notEffect, k)
					break
				}
			}
		}
	}

	// 当前轮次还有事件未刷新，刷事件
	if len(notEffect) > 0 {
		r := rand.Intn(len(notEffect))
		refresh.Status[notEffect[r]] = models.EventStatus_EventTriggered
		triggered = notEffect[r]
		notEffect = append(notEffect[:r], notEffect[r+1:]...)
	}
	// 本轮刷新完毕
	if len(notEffect) == 0 {
		// 把已完成的状态改为未触发
		for k, v := range refresh.Status {
			if v == models.EventStatus_EventCompleted {
				refresh.Status[k] = models.EventStatus_EventNotEffect
			}
		}
	}

	newEvent := s.randPointEvent(ctx, triggered, curRes.SceneId, events.Triggered)
	// 根据事件ID刷游戏
	err := s.genEventGame(ctx, newEvent)
	if err != nil {
		return nil, err
	}
	events.Triggered = append(events.Triggered, newEvent)
	refresh.Refreshed += s.getRefreshGap(ctx) * 1000
	refresh.Curr = newEvent.EventId
	return newEvent, nil
}

func (s *Service) genEventGame(ctx *ctx.Context, e *models.MapEvent) *errmsg.ErrMsg {
	reader := rule.MustGetReader(ctx)
	switch e.EventId {
	case 1:
		role, err := s.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
		if err != nil {
			return err
		}
		total := int64(0)
		weight := make([]values.Integer, 0)
		gameId := make([]values.Integer, 0)
		for idx := range reader.AnecdotesGame1.List() {
			if role.Level >= reader.AnecdotesGame1.List()[idx].GameLv {
				total += reader.AnecdotesGame1.List()[idx].Weight
				weight = append(weight, total)
				gameId = append(gameId, reader.AnecdotesGame1.List()[idx].Id)
			}
		}

		r := rand.Int63n(total)
		for idx := range weight {
			if r < weight[idx] {
				e.GameId = gameId[idx]
				cfg, _ := reader.AnecdotesGame1.GetAnecdotesGame1ById(e.GameId)
				storyIdx := rand.Intn(len(cfg.StoryText))
				if len(e.Extra) == 0 {
					e.Extra = make([]string, 1)
				}
				e.Extra[0] = strconv.Itoa(storyIdx)
				break
			}
		}
	case 2:
		role, err := s.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
		if err != nil {
			return err
		}
		for _, v := range reader.AnecdotesGame2.List() {
			if role.Level >= v.GameLv {
				e.GameId = v.Id
				continue
			}
			break
		}
	case 3:
		role, err := s.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
		if err != nil {
			return err
		}
		total := int64(0)
		weight := make([]values.Integer, 0)
		gameId := make([]values.Integer, 0)
		for idx := range reader.AnecdotesGame3.List() {
			if role.Level >= reader.AnecdotesGame3.List()[idx].GameLv {
				total += reader.AnecdotesGame3.List()[idx].Weight
				weight = append(weight, total)
				gameId = append(gameId, reader.AnecdotesGame3.List()[idx].Id)
			}
		}

		r := rand.Int63n(total)
		for idx := range weight {
			if r < weight[idx] {
				e.GameId = gameId[idx]
				break
			}
		}
	case 4:
		role, err := s.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
		if err != nil {
			return err
		}
		total := int64(0)
		weight := make([]values.Integer, 0)
		gameId := make([]values.Integer, 0)
		for idx := range reader.AnecdotesGame4.List() {
			if role.Level >= reader.AnecdotesGame4.List()[idx].GameLv {
				total += reader.AnecdotesGame4.List()[idx].Weight
				weight = append(weight, total)
				gameId = append(gameId, reader.AnecdotesGame4.List()[idx].Id)
			}
		}

		r := rand.Int63n(total)
		for idx := range weight {
			if r < weight[idx] {
				e.GameId = gameId[idx]
				break
			}
		}
	case 5:
		role, err := s.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
		if err != nil {
			return err
		}
		total := int64(0)
		weight := make([]values.Integer, 0)
		gameId := make([]values.Integer, 0)
		for idx := range reader.AnecdotesGame5.List() {
			if role.Level >= reader.AnecdotesGame5.List()[idx].GameLv {
				total += reader.AnecdotesGame5.List()[idx].Weight
				weight = append(weight, total)
				gameId = append(gameId, reader.AnecdotesGame5.List()[idx].Id)
			}
		}

		r := rand.Int63n(total)
		for idx := range weight {
			if r < weight[idx] {
				e.GameId = gameId[idx]
				break
			}
		}
	case 7:
		role, err := s.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
		if err != nil {
			return err
		}
		total := int64(0)
		weight := make([]values.Integer, 0)
		gameId := make([]values.Integer, 0)
		for idx := range reader.AnecdotesGame7.List() {
			if role.Level >= reader.AnecdotesGame7.List()[idx].GameLv {
				total += reader.AnecdotesGame7.List()[idx].Weight
				weight = append(weight, total)
				gameId = append(gameId, reader.AnecdotesGame7.List()[idx].Id)
			}
		}

		r := rand.Int63n(total)
		for idx := range weight {
			if r < weight[idx] {
				e.GameId = gameId[idx]
				cfg, _ := reader.AnecdotesGame1.GetAnecdotesGame1ById(e.GameId)
				storyIdx := rand.Intn(len(cfg.StoryText))
				if len(e.Extra) == 0 {
					e.Extra = make([]string, 1)
				}
				e.Extra[0] = strconv.Itoa(storyIdx)
				break
			}
		}
	case 8:
		role, err := s.UserService.GetRoleByRoleId(ctx, ctx.RoleId)
		if err != nil {
			return err
		}
		total := int64(0)
		weight := make([]values.Integer, 0)
		gameId := make([]values.Integer, 0)
		for idx := range reader.AnecdotesGame8.List() {
			if role.Level >= reader.AnecdotesGame8.List()[idx].GameLv {
				total += reader.AnecdotesGame8.List()[idx].Weight
				weight = append(weight, total)
				gameId = append(gameId, reader.AnecdotesGame8.List()[idx].Id)
			}
		}

		r := rand.Int63n(total)
		for idx := range weight {
			if r < weight[idx] {
				e.GameId = gameId[idx]
				break
			}
		}
	default:
	}
	return nil
}

func (s *Service) getEventPoints(ctx *ctx.Context, mapId values.MapId) map[values.EventId]struct{} {
	// 可用的刷新点
	data := mapdata.MustGetMapData(mapdata.GetLogicMapId(ctx, mapId))
	points := map[values.EventId]struct{}{}
	for _, p := range data.Events {
		points[p.Id] = struct{}{}
	}
	return points
}

func (s *Service) getRefreshLimit(ctx *ctx.Context) values.Integer {
	r := rule.MustGetReader(ctx)
	limit, ok := r.KeyValue.GetInt64(AnecdotesNum)
	if !ok {
		panic(fmt.Sprintf("AnecdotesNum Key not found"))
	}
	return limit
}

func (s *Service) getRefreshGap(ctx *ctx.Context) values.Integer {
	r := rule.MustGetReader(ctx)
	t, ok := r.KeyValue.GetInt64(AnecdotesAddTime)
	if !ok {
		panic(fmt.Sprintf("AnecdotesAddTime Key not found"))
	}
	return t
}
