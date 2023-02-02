package statistic

import (
	"strconv"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/iggsdk"
	"coin-server/common/proto/models"
	protosvc "coin-server/common/proto/service"
	"coin-server/common/service"
	"coin-server/common/statistical"
	st_models "coin-server/common/statistical/models"
	"coin-server/common/statistical2"
	st2_models "coin-server/common/statistical2/models"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/game-server/module"

	"github.com/rs/xid"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	funcMap    map[string]typModel
	*module.Module
}

type typModel func(*ctx.Context, *protosvc.Statistic_TrackingRequest) st_models.Model

func NewStatisticService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		Module:     module,
	}

	s.funcMap = map[string]typModel{
		"Game_Start":       gameStatistic,
		"Story":            gameStatistic,
		"Story_Skip":       gameStatistic,
		"Tutorial":         gameStatistic,
		"Tutorial_Skip":    gameStatistic,
		"Battle_Start":     battleStatistic,
		"Battle_Win":       battleStatistic,
		"Battle_Lose":      battleStatistic,
		"Battle_Exit":      battleStatistic,
		"PVP_Start":        pvpStatistic,
		"PVP_Win":          pvpStatistic,
		"PVP_Lose":         pvpStatistic,
		"PVP_Exit":         pvpStatistic,
		"Multiplayer_copy": roguelikeStatistic,
	}
	return s
}

// ---------------------------------------------------proto------------------------------------------------------------//

func (svc *Service) Router() {
	svc.svc.RegisterFunc("打点", svc.Tracking)
	svc.svc.RegisterEvent("打点推送", svc.TrackingPush)
}

func (svc *Service) Tracking(c *ctx.Context, req *protosvc.Statistic_TrackingRequest) (*protosvc.Statistic_TrackingResponse, *errmsg.ErrMsg) {
	if _, exist := svc.funcMap[req.EventStr]; !exist {
		return nil, errmsg.NewInternalErr("EventStr not exist")
	}
	statistical.Save(c.NewLogServer(), svc.funcMap[req.EventStr](c, req))
	game2 := gameStatistic2(c, req)
	if game2 != nil {
		statistical2.Save(c.NewLogServer2(), game2)
	}
	return &protosvc.Statistic_TrackingResponse{}, nil
}

func (svc *Service) TrackingPush(c *ctx.Context, req *protosvc.Statistic_TrackingPush) {
	if _, exist := svc.funcMap[req.EventStr]; !exist {
		return
	}
	statistical.Save(c.NewLogServer(), svc.funcMap[req.EventStr](c, &protosvc.Statistic_TrackingRequest{
		EventStr: req.EventStr,
		Data:     req.Data,
	}))
	return
}

func gameStatistic(c *ctx.Context, req *protosvc.Statistic_TrackingRequest) st_models.Model {
	m := &st_models.Game{
		IggId:     iggsdk.ConvertToIGGId(c.UserId),
		EventTime: timer.Now(),
		GwId:      statistical.GwId(),
		RoleId:    c.RoleId,
	}
	switch req.EventStr {
	case "Game_Start":
		m.Memory = req.Data["memory"]
	case "Story":
		m.StoryChapter = req.Data["chapter"]
	case "Story_Skip":
		m.SkipChapter = req.Data["chapter"]
	case "Tutorial":
		m.TutorialStep = req.Data["step"]
	case "Tutorial_Skip":
		m.TutorialSkipStep = req.Data["step"]
	}
	return m
}

func gameStatistic2(c *ctx.Context, req *protosvc.Statistic_TrackingRequest) st2_models.Model {
	m := &st2_models.Game{
		Time:     time.Now(),
		IggId:    c.UserId,
		ServerId: c.ServerId,
		Xid:      xid.New().String(),
		RoleId:   c.RoleId,
	}
	switch req.EventStr {
	case "Game_Start":
		m.Memory = req.Data["memory"]
	case "Story":
		m.StoryChapter = req.Data["chapter"]
	case "Story_Skip":
		m.SkipChapter = req.Data["chapter"]
	case "Tutorial":
		m.TutorialStep = req.Data["step"]
	case "Tutorial_Skip":
		m.TutorialSkipStep = req.Data["step"]
	}
	return m
}

func battleStatistic(c *ctx.Context, req *protosvc.Statistic_TrackingRequest) st_models.Model {
	i, _ := strconv.ParseInt(req.Data["mission"], 10, 64)
	m := &st_models.Battle{
		IggId:     iggsdk.ConvertToIGGId(c.UserId),
		EventTime: timer.Now(),
		GwId:      statistical.GwId(),
		RoleId:    c.RoleId,
		Mission:   i,
	}
	switch req.EventStr {
	case "Battle_Start":
		m.EventTyp = st_models.BattleEventTypStart
	case "Battle_Win":
		m.EventTyp = st_models.BattleEventTypWin
	case "Battle_Lose":
		m.EventTyp = st_models.BattleEventTypLose
	case "Battle_Exit":
		m.EventTyp = st_models.BattleEventTypExit
	}
	return m
}

func roguelikeStatistic(c *ctx.Context, req *protosvc.Statistic_TrackingRequest) st_models.Model {
	id, _ := strconv.ParseInt(req.Data["copy"], 10, 64)
	duration, _ := strconv.ParseInt(req.Data["duration"], 10, 64)
	var isSucc int64
	if req.Data["is_succ"] == "true" {
		isSucc = 1
	}
	m := &st_models.Roguelike{
		IggId:       iggsdk.ConvertToIGGId(c.UserId),
		EventTime:   timer.Now(),
		GwId:        statistical.GwId(),
		RoleId:      c.RoleId,
		RoguelikeId: id,
		Duration:    duration,
		IsSucc:      isSucc,
	}
	return m
}

func pvpStatistic(c *ctx.Context, req *protosvc.Statistic_TrackingRequest) st_models.Model {
	rank, _ := strconv.ParseInt(req.Data["rank"], 10, 64)
	point, _ := strconv.ParseInt(req.Data["point"], 10, 64)
	m := &st_models.Pvp{
		IggId:     iggsdk.ConvertToIGGId(c.UserId),
		EventTime: timer.Now(),
		GwId:      statistical.GwId(),
		RoleId:    c.RoleId,
		Opponent:  req.Data["opponent"],
		Rank:      rank,
		Point:     point,
	}
	switch req.EventStr {
	case "PVP_Start":
		m.EventTyp = st_models.PvpEventTypStart
	case "PVP_Win":
		m.EventTyp = st_models.PvpEventTypWin
	case "PVP_Lose":
		m.EventTyp = st_models.PvpEventTypLose
	case "PVP_Exit":
		m.EventTyp = st_models.PvpEventTypExit
	}
	return m
}
