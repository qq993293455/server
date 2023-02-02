package match

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/im"
	"coin-server/common/logger"
	"coin-server/common/proto/cppbattle"
	pbdao "coin-server/common/proto/dao"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	"coin-server/common/proto/newcenter"
	"coin-server/common/proto/roguelike_match"
	service2 "coin-server/common/proto/service"
	"coin-server/common/routine_limit_service"
	"coin-server/common/timer"
	"coin-server/common/utils/imutil"
	"coin-server/common/values"
	"coin-server/common/values/env"
	"coin-server/roguelike-match-server/service/match/dao"
	"coin-server/roguelike-match-server/service/match/values/group_lock"
	"coin-server/roguelike-match-server/service/match/values/room_mgr"
	"coin-server/rule"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

const (
	maxMatchCnt = 100
)

type Service struct {
	serverId       values.ServerId
	serverType     models.ServerType
	svc            *routine_limit_service.RoutineLimitService
	log            *logger.Logger
	mgr            *room_mgr.RoomMgr
	roleInfoMgr    *room_mgr.RoleInfoMgr
	robotInfoMgr   *room_mgr.RobotInfoMgr
	unqRobot       values.Integer
	dailyBossSkill *pbdao.RoguelikeBossSkill
	refreshTime    values.Integer
}

func NewMatchService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *routine_limit_service.RoutineLimitService,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:     serverId,
		serverType:   serverType,
		svc:          svc,
		log:          log,
		mgr:          room_mgr.NewRoomMgr(),
		roleInfoMgr:  room_mgr.NewRoleInfoMgr(),
		robotInfoMgr: room_mgr.NewRobotInfoMgr(),
	}
	svc.AfterFunc(time.Millisecond, func(c *ctx.Context) {
		r := rule.MustGetReader(c)
		s.refreshTime, _ = r.KeyValue.GetInt64("DefaultRefreshTime")
		data, err := dao.GetBossSkill(c)
		if err != nil {
			panic(err)
		}
		s.dailyBossSkill = data
		if s.needGen(c) {
			s.genBossSkill(c)
			dao.SaveBossSkill(c, s.dailyBossSkill)
		}
		now := timer.Now().UTC()
		end := s.getCurrDayFreshTime(c).Add(24 * time.Hour)
		svc.AfterFunc(time.Duration(end.Unix()-now.Unix()+5)*time.Second, func(c1 *ctx.Context) {
			s.genBossSkill(c1)
			dao.SaveBossSkill(c1, s.dailyBossSkill)
			svc.TickFunc(24*time.Hour, func(c2 *ctx.Context) bool {
				s.genBossSkill(c2)
				dao.SaveBossSkill(c2, s.dailyBossSkill)
				return true
			})
		})
	})
	return s
}

func (svc *Service) Router() {
	svc.svc.RegisterFunc("创建房间", svc.CreateRoom)
	svc.svc.RegisterFunc("获取当前房间", svc.GetCurrRoom)
	svc.svc.RegisterFunc("加入房间", svc.JoinRoom)
	svc.svc.RegisterFunc("离开房间", svc.LeaveRoom)
	svc.svc.RegisterFunc("解散房间", svc.DismissRoom)
	svc.svc.RegisterFunc("获取某玩家的房间", svc.GetRoleRoom)
	svc.svc.RegisterFunc("修改房间公开状态", svc.SetPub)
	svc.svc.RegisterFunc("随机获取某地图的房间", svc.GetRandomRoom)
	svc.svc.RegisterFunc("匹配离线玩家", svc.MatchOffline)
	svc.svc.RegisterFunc("开始副本", svc.StartBattle)
	svc.svc.RegisterFunc("重新进入副本", svc.ReGetBattle)
	svc.svc.RegisterFunc("准备", svc.Ready)
	svc.svc.RegisterFunc("选择英雄", svc.ChooseHero)
	svc.svc.RegisterFunc("获取今日boss和次数", svc.GetTodayBoss)
	svc.svc.RegisterFunc("踢人", svc.Kick)
	svc.svc.RegisterFunc("换队长", svc.ChangeOwner)
	svc.svc.RegisterFunc("设置入队条件", svc.SetCombatNeed)

	svc.svc.RegisterFunc("邀请玩家", svc.Invite)
	svc.svc.RegisterFunc("房间开放到世界频道", svc.PubRoomToChat)

	//----------------event--------------------------
	svc.svc.RegisterEvent("战斗结束回调", svc.BattleFinishAck)
	svc.svc.RegisterEvent("战斗手动退出", svc.BattleExit)
	svc.svc.RegisterEvent("解散房间推送", svc.DismissRoomEvent)

	//----------------cheat--------------------------
	svc.svc.RegisterFunc("作弊改变boss", svc.CheatChangeBoss)
	svc.svc.RegisterFunc("作弊清除今日数量", svc.CheatClearTodayCnt)
}

func (svc *Service) getCurrDayFreshTime(c *ctx.Context) time.Time {
	offset := timer.Timer.GetOffset()
	now := time.Unix(0, c.StartTime).Add(offset).UTC()
	begin := timer.BeginOfDay(now)
	offsetDuration := time.Second * time.Duration(svc.refreshTime)
	currDayFreshTime := begin.Add(offsetDuration)
	if now.After(currDayFreshTime) {
		return currDayFreshTime
	}
	return timer.LastDay(begin).Add(offsetDuration)
}

func (svc *Service) needGen(c *ctx.Context) bool {
	return svc.dailyBossSkill.UpdateAt < svc.getCurrDayFreshTime(c).Unix()
}

func (svc *Service) genBossSkill(c *ctx.Context) {
	if svc.dailyBossSkill == nil {
		return
	}
	begin := svc.getCurrDayFreshTime(c)
	/*if svc.dailyBossSkill.UpdateAt == begin.Unix() {
		begin = begin.Add(24 * time.Hour)
	}*/
	newData := &pbdao.RoguelikeBossSkill{
		Key: svc.dailyBossSkill.Key,
	}
	newData.UpdateAt = begin.Unix()
	reader := rule.MustGetReader(c)
	groupData := reader.GenRoguelikeBossGroup()
	if svc.dailyBossSkill.DungeonDay <= 0 || svc.dailyBossSkill.DungeonDay >= int64(len(groupData)) {
		newData.DungeonDay = 1
	} else {
		newData.DungeonDay = svc.dailyBossSkill.DungeonDay + 1
	}
	/* 删除理由：BOSS没有技能了
	if g, exist := groupData[newData.DungeonDay]; exist {
		for _, rlId := range g {
			roguelikeRule, ok := reader.RoguelikeDungeon.GetRoguelikeDungeonById(rlId)
			if !ok {
				continue
			}
			if len(roguelikeRule.BossSkillNum) == 2 {
				bsn := roguelikeRule.BossSkillNum
				newData.BossSkill[rlId] = &models.BossSkillList{
					Skills: make([]values.Integer, 0, bsn[1]),
				}
				svc.pick(rlId, bsn[0], bsn[1], reader, newData, 0)
			}
			if len(roguelikeRule.MonsterSkillNum) == 2 {
				msn := roguelikeRule.MonsterSkillNum
				newData.MonsterSkill[rlId] = &models.BossSkillList{
					Skills: make([]values.Integer, 0, msn[1]),
				}
				svc.pick(rlId, msn[0], msn[1], reader, newData, 1)
			}
		}
	}*/
	svc.dailyBossSkill = newData
}

/*	删除理由：BOSS没有技能了
func (svc *Service) pick(rlId, entryGroupId, cnt values.Integer, reader *rule_model.Reader, newData *pbdao.RoguelikeBossSkill, typ int) {
	entryGroup := reader.GenRoguelikeEntryGroup()
	if entryIds, has := entryGroup[entryGroupId]; has {
		cho := make([]*weightedrand.Choice[values.Integer, values.Integer], 0, len(entryIds))
		for _, entryId := range entryIds {
			if entryRule, ok := reader.RoguelikeEntry.GetRoguelikeEntryById(entryId); ok {
				cho = append(cho, weightedrand.NewChoice(entryRule.Id, entryRule.Entryweight))
			}
		}
		if ec, err := weightedrand.NewChooser(cho...); err == nil {
			for t := values.Integer(0); t < cnt; t++ {
				chooseId := ec.Pick()
				if entryRule, ok := reader.RoguelikeEntry.GetRoguelikeEntryById(chooseId); ok {
					choQ := make([]*weightedrand.Choice[values.Integer, values.Integer], 0, len(entryRule.Qualitysection))
					for k, v := range entryRule.Qualitysection {
						choQ = append(choQ, weightedrand.NewChoice(k, v))
					}
					if qc, err := weightedrand.NewChooser(choQ...); err == nil {
						if typ == 0 {
							newData.BossSkill[rlId].Skills = append(newData.BossSkill[rlId].Skills, qc.Pick())
						}
						if typ == 1 {
							newData.MonsterSkill[rlId].Skills = append(newData.MonsterSkill[rlId].Skills, qc.Pick())
						}
					}
				}
			}
		}
	}
}*/

func (svc *Service) checkIsToday(c *ctx.Context, roguelikeId values.RoguelikeId) bool {
	bossGroup := rule.MustGetReader(c).GenRoguelikeBossGroup()
	if svc.dailyBossSkill == nil {
		return false
	}
	if g, exist := bossGroup[svc.dailyBossSkill.DungeonDay]; exist {
		for _, rId := range g {
			if roguelikeId == values.RoguelikeId(rId) {
				return true
			}
		}
	}
	return false
}

func (svc *Service) checkClose(r *room_mgr.Room) bool {
	return r.CheckClose(svc.mgr.CloseAfter)
}

func (svc *Service) CreateRoom(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLCreateRoomRequest) (*roguelike_match.RoguelikeMatch_RLCreateRoomResponse, *errmsg.ErrMsg) {
	if !svc.checkIsToday(c, values.RoguelikeId(req.RoguelikeId)) {
		return nil, errmsg.NewErrRLDungeonNotToday()
	}
	group_lock.LockRole(c.RoleId)
	defer group_lock.UnlockRole(c.RoleId)
	room, err := svc.mgr.CreateRoom(c, c.RoleId, values.RoguelikeId(req.RoguelikeId), c.InServerId)
	if err != nil {
		return nil, err
	}
	room.SetCombatNeed(req.CombatNeed)
	svc.svc.AfterFuncCtx(c, time.Microsecond, func(c1 *ctx.Context) {
		im.DefaultClient.JoinRoom(c1, &im.RoomRole{
			RoomID:  strconv.FormatInt(int64(room.RoomId()), 10),
			RoleIDs: []string{c1.RoleId},
		})
	})
	return &roguelike_match.RoguelikeMatch_RLCreateRoomResponse{
		Room: &models.RoguelikeRoom{
			RoomId:      int64(room.RoomId()),
			OwnerId:     room.OwnerId(),
			RoguelikeId: int64(room.RoguelikeId()),
			RoleInfos:   svc.fillParty(c, room.GetPartyInfo()),
			IsOpen:      room.IsOpen(),
			Status:      room.Status(),
			UpdateAt:    room.CloseAt(svc.mgr.CloseAfter),
			CombatNeed:  room.CombatNeed(),
		},
	}, nil
}

func (svc *Service) SetCombatNeed(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLChangeCombatNeedRequest) (*roguelike_match.RoguelikeMatch_RLChangeCombatNeedResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, errmsg.NewErrNotIn()
	}
	if !svc.checkIsToday(c, roguelikeId) {
		return nil, errmsg.NewErrRLDungeonNotToday()
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if r.IsStarted() {
		return nil, errmsg.NewErrRoguelikeAlreadyStart()
	}
	if r.OwnerId() != c.RoleId {
		return nil, errmsg.NewErrOnlyOwnerCanStart()
	}
	r.SetCombatNeed(req.CombatNeed)
	svc.roomChangePushWithRoom(c, r)
	return &roguelike_match.RoguelikeMatch_RLChangeCombatNeedResponse{
		Room: &models.RoguelikeRoom{
			RoomId:      int64(r.RoomId()),
			OwnerId:     r.OwnerId(),
			RoguelikeId: int64(r.RoguelikeId()),
			RoleInfos:   svc.fillParty(c, r.GetPartyInfo()),
			IsOpen:      r.IsOpen(),
			Status:      r.Status(),
			UpdateAt:    r.CloseAt(svc.mgr.CloseAfter),
			CombatNeed:  r.CombatNeed(),
		},
	}, nil
}

func (svc *Service) GetCurrRoom(c *ctx.Context, _ *roguelike_match.RoguelikeMatch_RLGetCurrRoomRequest) (*roguelike_match.RoguelikeMatch_RLGetCurrRoomResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return &roguelike_match.RoguelikeMatch_RLGetCurrRoomResponse{
			Room: nil,
		}, nil
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return &roguelike_match.RoguelikeMatch_RLGetCurrRoomResponse{
			Room: nil,
		}, nil
	}
	roguelikeRule, ok := rule.MustGetReader(c).RoguelikeDungeon.GetRoguelikeDungeonById(values.Integer(r.RoguelikeId()))
	if !ok {
		return nil, errmsg.NewErrRoguelikeNotExist()
	}
	if svc.checkClose(r) || (r.IsStarted() && timer.Unix() > (r.StartTime()+roguelikeRule.Times+30)) {
		// 房间待机时间过长或者副本已经结束但是未关闭
		roleIds := r.GetParty()
		c.PushMessageToRoles(roleIds, &roguelike_match.RoguelikeMatch_RLRoguelikeFinishPush{
			RoomId:    int64(roomId),
			BattleId:  r.BattleId(),
			HasReward: false,
		}, &roguelike_match.RoguelikeMatch_RLRoomClosePush{
			RoomId: int64(roomId),
		})
		svc.mgr.DelRoom(r.RoomId(), r.RoguelikeId())
		svc.svc.AfterFuncCtx(c, time.Microsecond, func(c1 *ctx.Context) {
			im.DefaultClient.LeaveRoom(c1, &im.RoomRole{
				RoomID:  strconv.FormatInt(int64(r.RoomId()), 10),
				RoleIDs: roleIds,
			})
		})
		return &roguelike_match.RoguelikeMatch_RLGetCurrRoomResponse{
			Room: nil,
		}, nil
	}
	if r.IsStarted() {
		copyResp := &cppbattle.NSNB_GetRlBattleRoomsResponse{}
		err := svc.svc.GetNatsClient().RequestWithOut(c, r.BattleId(), &cppbattle.NSNB_GetRlBattleRoomsRequest{}, copyResp)
		if err != nil {
			// 肉鸽战斗服务挂了，关闭战斗
			svc.log.Error("GetCurrRoom rl battle fail", zap.Int64("battleId", r.BattleId()), zap.String("err", err.ErrMsg))
			r.StopFight()
			svc.roomChangePushWithRoom(c, r)
		}
	}
	return &roguelike_match.RoguelikeMatch_RLGetCurrRoomResponse{
		Room: &models.RoguelikeRoom{
			RoomId:      int64(r.RoomId()),
			OwnerId:     r.OwnerId(),
			RoguelikeId: int64(r.RoguelikeId()),
			RoleInfos:   svc.fillParty(c, r.GetPartyInfo()),
			IsOpen:      r.IsOpen(),
			Status:      r.Status(),
			UpdateAt:    r.CloseAt(svc.mgr.CloseAfter),
			CombatNeed:  r.CombatNeed(),
		},
	}, nil
}

func (svc *Service) StartBattle(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLStartBattleRequest) (*roguelike_match.RoguelikeMatch_RLStartBattleResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, errmsg.NewErrNotIn()
	}
	if !svc.checkIsToday(c, roguelikeId) {
		return nil, errmsg.NewErrRLDungeonNotToday()
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if r.IsStarted() {
		return nil, errmsg.NewErrRoguelikeAlreadyStart()
	}
	if r.OwnerId() != c.RoleId {
		return nil, errmsg.NewErrOnlyOwnerCanStart()
	}
	if !r.IsAllReady() {
		return nil, errmsg.NewErrRLNotAllReady()
	}
	isChange, cnt, err := svc.handleCnt(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if cnt.TodayJoin >= svc.mgr.OneDayStartLimit+cnt.ExtraCnt {
		return nil, errmsg.NewErrRLNotEnoughJoinCnt()
	}
	var mapId values.Integer = 0
	maps := rule.MustGetReader(c).MapScene.List()
	for _, ma := range maps {
		if ma.RoguelikeId == values.Integer(r.RoguelikeId()) {
			mapId = ma.Id
		}
	}
	if mapId <= 0 {
		return nil, errmsg.NewErrDungeonNotExist()
	}
	var robots []*models.Bot
	robotsIds := r.GetRobots()
	if len(robotsIds) > 0 {
		robots = make([]*models.Bot, 0, len(robotsIds))
		for _, robotId := range robotsIds {
			robot := svc.robotInfoMgr.Get(robotId)
			if robot != nil {
				robots = append(robots, robot)
			}
		}
	}
	/*var be, me []values.Integer
	if svc.dailyBossSkill.BossSkill[int64(roguelikeId)] != nil {
		be = svc.dailyBossSkill.BossSkill[int64(roguelikeId)].Skills
	}
	if svc.dailyBossSkill.MonsterSkill[int64(roguelikeId)] != nil {
		me = svc.dailyBossSkill.MonsterSkill[int64(roguelikeId)].Skills
	}*/
	centerReq := &newcenter.NewCenter_CreateRoguelikeRequest{
		MapId:   mapId,
		RoomId:  int64(r.RoomId()),
		Bots:    robots,
		CardMap: r.CardMap(),
	}
	centerResp := &newcenter.NewCenter_CreateRoguelikeResponse{}
	centerId := env.GetCenterServerId()
	err = svc.svc.GetNatsClient().RequestWithOut(c, centerId, centerReq, centerResp)
	if err != nil {
		svc.log.Error("create rl battle fail", zap.Int64("center_svc_id", centerId), zap.Int64("map", mapId), zap.String("err", err.ErrMsg))
		return nil, err
	}
	copyResp := &cppbattle.NSNB_GetRlBattleRoomsResponse{}
	if err = svc.svc.GetNatsClient().RequestWithOut(c, centerResp.BattleId, &cppbattle.NSNB_GetRlBattleRoomsRequest{}, copyResp); err != nil {
		svc.log.Error("get rl rooms fail", zap.Int64("battleId", r.BattleId()), zap.Int64("map", mapId), zap.String("err", err.ErrMsg))
		return nil, err
	}

	r.StartFight(centerResp.BattleId)
	c.PushMessageToRoles(r.GetParty(), &roguelike_match.RoguelikeMatch_RLRoguelikeStartPush{
		RoomId:   int64(r.RoomId()),
		BattleId: centerResp.BattleId,
		MapId:    mapId,
		RlRooms:  copyResp.RlRooms,
	})
	if isChange {
		dao.SaveCnt(c, cnt)
	}
	return &roguelike_match.RoguelikeMatch_RLStartBattleResponse{}, nil
}

func (svc *Service) ReGetBattle(c *ctx.Context, _ *roguelike_match.RoguelikeMatch_RLReGetBattleRequest) (*roguelike_match.RoguelikeMatch_RLReGetBattleResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, errmsg.NewErrNotIn()
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if !r.IsStarted() {
		return nil, errmsg.NewErrDungeonNotStart()
	}
	var mapId values.Integer = 0
	maps := rule.MustGetReader(c).MapScene.List()
	for _, ma := range maps {
		if ma.RoguelikeId == values.Integer(r.RoguelikeId()) {
			mapId = ma.Id
		}
	}
	copyResp := &cppbattle.NSNB_GetRlBattleRoomsResponse{}
	if err := svc.svc.GetNatsClient().RequestWithOut(c, r.BattleId(), &cppbattle.NSNB_GetRlBattleRoomsRequest{}, copyResp); err != nil {
		svc.log.Error("ReGetBattle rl battle fail", zap.Int64("battleId", r.BattleId()), zap.Int64("map", mapId), zap.String("err", err.ErrMsg))
		return nil, err
	}
	return &roguelike_match.RoguelikeMatch_RLReGetBattleResponse{
		RoomId:   int64(roomId),
		MapId:    mapId,
		BattleId: r.BattleId(),
		RlRooms:  copyResp.RlRooms,
		CurIdx:   copyResp.CurIdx,
	}, nil
}

func (svc *Service) GetRandomRoom(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLGetRandomRoomRequest) (*roguelike_match.RoguelikeMatch_RLGetRandomRoomResponse, *errmsg.ErrMsg) {
	sort.Slice(req.RoguelikeId, func(i, j int) bool {
		return req.RoguelikeId[i] > req.RoguelikeId[j]
	})
	finalRooms := make([]*room_mgr.Room, 0, req.Num)
	roomCnt := 0
	for _, roguelikeId := range req.RoguelikeId {
		rooms, err := svc.mgr.GetRandomRooms(values.RoguelikeId(roguelikeId), int(req.Num)-roomCnt, svc.mgr.CloseAfter)
		if err != nil {
			return &roguelike_match.RoguelikeMatch_RLGetRandomRoomResponse{
				Rooms: nil,
			}, err
		}
		finalRooms = append(finalRooms, rooms...)
		roomCnt += len(rooms)
		if roomCnt >= int(req.Num) {
			break
		}
	}

	res := make([]*models.RoguelikeRoom, len(finalRooms))
	for idx := range finalRooms {
		res[idx] = &models.RoguelikeRoom{
			RoomId:      int64(finalRooms[idx].RoomId()),
			OwnerId:     finalRooms[idx].OwnerId(),
			RoguelikeId: int64(finalRooms[idx].RoguelikeId()),
			RoleInfos:   svc.fillParty(c, finalRooms[idx].GetPartyInfo()),
			IsOpen:      finalRooms[idx].IsOpen(),
			Status:      finalRooms[idx].Status(),
			UpdateAt:    finalRooms[idx].CloseAt(svc.mgr.CloseAfter),
			CombatNeed:  finalRooms[idx].CombatNeed(),
		}
	}
	svc.mgr.PutPubRooms(finalRooms)
	return &roguelike_match.RoguelikeMatch_RLGetRandomRoomResponse{
		Rooms: res,
	}, nil
}

func (svc *Service) JoinRoom(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLJoinRoomRequest) (*roguelike_match.RoguelikeMatch_RLJoinRoomResponse, *errmsg.ErrMsg) {
	group_lock.LockRoom(values.MatchRoomId(req.RoomId))
	defer group_lock.UnlockRoom(values.MatchRoomId(req.RoomId))
	r := svc.mgr.GetRoom(values.MatchRoomId(req.RoomId), values.RoguelikeId(req.RoguelikeId))
	if r == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if !req.IsPrivate && !r.IsOpen() {
		return nil, errmsg.NewErrRLDungeonNotOpen()
	}
	if r.IsStarted() {
		return nil, errmsg.NewErrRoguelikeAlreadyStart()
	}
	if r.CheckClose(svc.mgr.CloseAfter) {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if r.CombatNeed() > 0 && req.UserFight < r.CombatNeed() {
		return nil, errmsg.NewErrRLCombatNotEnough()
	}
	err := svc.mgr.Join(values.MatchRoomId(req.RoomId), values.RoguelikeId(req.RoguelikeId), c.RoleId, c.InServerId)
	if err != nil {
		return nil, err
	}
	svc.roomChangePushWithRoom(c, r)
	svc.svc.AfterFuncCtx(c, time.Microsecond, func(c2 *ctx.Context) {
		im.DefaultClient.JoinRoom(c2, &im.RoomRole{
			RoomID:  strconv.FormatInt(req.RoomId, 10),
			RoleIDs: []string{c2.RoleId},
		})
	})
	return &roguelike_match.RoguelikeMatch_RLJoinRoomResponse{
		Room: &models.RoguelikeRoom{
			RoomId:      int64(r.RoomId()),
			OwnerId:     r.OwnerId(),
			RoguelikeId: int64(r.RoguelikeId()),
			RoleInfos:   svc.fillParty(c, r.GetPartyInfo()),
			IsOpen:      r.IsOpen(),
			Status:      r.Status(),
			UpdateAt:    r.CloseAt(svc.mgr.CloseAfter),
			CombatNeed:  r.CombatNeed(),
		},
	}, nil
}

func (svc *Service) LeaveRoom(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLLeaveRoomRequest) (*roguelike_match.RoguelikeMatch_RLLeaveRoomResponse, *errmsg.ErrMsg) {
	group_lock.LockRoom(values.MatchRoomId(req.RoomId))
	defer group_lock.UnlockRoom(values.MatchRoomId(req.RoomId))
	err := svc.mgr.Leave(values.MatchRoomId(req.RoomId), c.RoleId)
	if err != nil {
		return nil, err
	}
	r := svc.mgr.GetRoom(values.MatchRoomId(req.RoomId), values.RoguelikeId(req.RoguelikeId))
	if r == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if r.IsEmpty() {
		if err = svc.delRoom(r); err != nil {
			return nil, err
		}
	}
	svc.roomChangePushWithRoom(c, r)
	svc.svc.AfterFuncCtx(c, time.Microsecond, func(c1 *ctx.Context) {
		im.DefaultClient.LeaveRoom(c1, &im.RoomRole{
			RoomID:  strconv.FormatInt(req.RoomId, 10),
			RoleIDs: []string{c1.RoleId},
		})
	})
	return &roguelike_match.RoguelikeMatch_RLLeaveRoomResponse{}, nil
}

func (svc *Service) delRoom(r *room_mgr.Room) *errmsg.ErrMsg {
	robotIds := r.GetRobots()
	err := svc.mgr.DelRoom(r.RoomId(), r.RoguelikeId())
	if err != nil {
		return err
	}
	for _, robotId := range robotIds {
		svc.robotInfoMgr.Del(robotId)
	}
	return nil
}

func (svc *Service) Ready(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLReadyRequest) (*roguelike_match.RoguelikeMatch_RLReadyResponse, *errmsg.ErrMsg) {
	isChange, cnt, err := svc.handleCnt(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if req.Ready {
		if cnt.TodayJoin >= svc.mgr.OneDayStartLimit+cnt.ExtraCnt {
			return nil, errmsg.NewErrRLNotEnoughJoinCnt()
		}
	}
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, errmsg.NewErrNotIn()
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if r.IsStarted() {
		return nil, errmsg.NewErrRoguelikeAlreadyStart()
	}
	if err := r.Ready(c.RoleId, req.Ready); err != nil {
		return nil, err
	}
	svc.roomChangePushWithRoom(c, r)
	if isChange {
		dao.SaveCnt(c, cnt)
	}
	return &roguelike_match.RoguelikeMatch_RLReadyResponse{}, nil
}

func (svc *Service) ChooseHero(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLChooseHeroRequest) (*roguelike_match.RoguelikeMatch_RLChooseHeroResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, errmsg.NewErrNotIn()
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if r.IsStarted() {
		return nil, errmsg.NewErrRoguelikeAlreadyStart()
	}
	if err := r.ChooseHero(c.RoleId, req.ConfigId, req.CardId); err != nil {
		return nil, err
	}
	svc.roomChangePushWithRoom(c, r)
	return &roguelike_match.RoguelikeMatch_RLChooseHeroResponse{}, nil
}

func (svc *Service) GetRoleRoom(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLGetRoleRoomRequest) (*roguelike_match.RoguelikeMatch_RLGetRoleRoomResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(req.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return &roguelike_match.RoguelikeMatch_RLGetRoleRoomResponse{
			Room: nil,
		}, nil
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return &roguelike_match.RoguelikeMatch_RLGetRoleRoomResponse{
			Room: nil,
		}, nil
	}
	return &roguelike_match.RoguelikeMatch_RLGetRoleRoomResponse{
		Room: &models.RoguelikeRoom{
			RoomId:      int64(r.RoomId()),
			OwnerId:     r.OwnerId(),
			RoguelikeId: int64(r.RoguelikeId()),
			RoleInfos:   svc.fillParty(c, r.GetPartyInfo()),
			IsOpen:      r.IsOpen(),
			Status:      r.Status(),
			UpdateAt:    r.CloseAt(svc.mgr.CloseAfter),
			CombatNeed:  r.CombatNeed(),
		},
	}, nil
}

func (svc *Service) DismissRoom(c *ctx.Context, _ *roguelike_match.RoguelikeMatch_RLDismissRoomRequest) (*roguelike_match.RoguelikeMatch_RLDismissRoomResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return &roguelike_match.RoguelikeMatch_RLDismissRoomResponse{}, nil
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return &roguelike_match.RoguelikeMatch_RLDismissRoomResponse{}, nil
	}
	if r.OwnerId() != c.RoleId && !r.CheckClose(svc.mgr.CloseAfter) {
		return nil, errmsg.NewErrRLNotOwner()
	}
	roleIds := r.GetParty()
	if err := svc.delRoom(r); err != nil {
		return nil, err
	}
	c.PushMessageToRoles(roleIds, &roguelike_match.RoguelikeMatch_RLRoomClosePush{
		RoomId: int64(roomId),
	})
	svc.svc.AfterFuncCtx(c, time.Microsecond, func(c1 *ctx.Context) {
		im.DefaultClient.LeaveRoom(c1, &im.RoomRole{
			RoomID:  strconv.FormatInt(int64(r.RoomId()), 10),
			RoleIDs: roleIds,
		})
	})
	return &roguelike_match.RoguelikeMatch_RLDismissRoomResponse{}, nil
}

func (svc *Service) DismissRoomEvent(c *ctx.Context, _ *roguelike_match.RoguelikeMatch_RLDismissEvent) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return
	}
	if r.OwnerId() != c.RoleId {
		return
	}
	roleIds := r.GetParty()
	if err := svc.delRoom(r); err != nil {
		return
	}
	c.PushMessageToRoles(roleIds, &roguelike_match.RoguelikeMatch_RLRoomClosePush{
		RoomId: int64(roomId),
	})
	im.DefaultClient.LeaveRoom(c, &im.RoomRole{
		RoomID:  strconv.FormatInt(int64(r.RoomId()), 10),
		RoleIDs: roleIds,
	})
}

func (svc *Service) SetPub(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLPubRoomRequest) (*roguelike_match.RoguelikeMatch_RLPubRoomResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, nil
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil || c.RoleId != r.OwnerId() {
		// 不在房间内, 或者不是房主
		return nil, errmsg.NewErrRLNotOwner()
	}
	svc.mgr.SetPub(r.RoomId(), r.RoguelikeId(), req.IsOpen)
	svc.roomChangePushWithRoom(c, r)
	return &roguelike_match.RoguelikeMatch_RLPubRoomResponse{}, nil
}

func (svc *Service) GetTodayBoss(c *ctx.Context, _ *roguelike_match.RoguelikeMatch_RLGetTodayBossRequest) (*roguelike_match.RoguelikeMatch_RLGetTodayBossResponse, *errmsg.ErrMsg) {
	isChange, cnt, err := svc.handleCnt(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if isChange {
		dao.SaveCnt(c, cnt)
	}
	return &roguelike_match.RoguelikeMatch_RLGetTodayBossResponse{
		DungeonDay: svc.dailyBossSkill.DungeonDay,
		JoinCnt:    cnt.TodayJoin,
		ExtraCnt:   cnt.ExtraCnt,
	}, nil
}

func (svc *Service) BattleFinishAck(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLBattleFinishEvent) {
	roomId := uint64(req.RoomId)
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, values.RoguelikeId(req.RoguelikeId))
	if r == nil {
		return
	}
	for _, v := range r.GetParty() {
		if v != "" {
			serverId := svc.mgr.GetRoleSvcId(v)
			if rewards, exist := req.RoleRewards[v]; exist {
				if len(rewards.Rewards) > 0 && serverId != -1 {
					svc.svc.GetNatsClient().Publish(serverId,
						ctx.NewHeader(v, svc.serverId, svc.serverType, c),
						&service2.BattleEvent_AddItemsEvent{
							Items: rewards.Rewards,
						})
					svc.svc.GetNatsClient().Publish(serverId,
						ctx.NewHeader(v, svc.serverId, svc.serverType, c),
						&service2.Task_FinishRLEvent{
							RoguelikeId: int64(r.RoguelikeId()),
						})
					svc.svc.GetNatsClient().Publish(serverId,
						ctx.NewHeader(v, svc.serverId, svc.serverType, c),
						&service2.Statistic_TrackingPush{
							EventStr: "Multiplayer_copy",
							Data: map[string]string{
								"copy":     strconv.FormatInt(req.RoguelikeId, 10),
								"duration": strconv.FormatInt(req.Duration, 10),
								"is_succ":  strconv.FormatBool(req.IsSuc),
							},
						})
					svc.svc.GetNatsClient().Publish(serverId,
						ctx.NewHeader(v, svc.serverId, svc.serverType, c),
						&service2.Journey_RoguelikeFinishPush{
							Success:     req.IsSuc,
							RoguelikeId: req.RoguelikeId,
						})
					if _, cnt, err := svc.handleCnt(c, v); err == nil {
						cnt.TodayJoin += 1
						cnt.LastJoinAt = timer.StartTime(c.StartTime).Unix()
						dao.SaveCnt(c, cnt)
					}
				}
				c.PushMessageToRole(v, &roguelike_match.RoguelikeMatch_RLRoguelikeFinishPush{
					RoomId:    int64(roomId),
					BattleId:  r.BattleId(),
					HasReward: req.IsSuc,
					Rewards:   rewards.Rewards,
				})
			} else {
				c.PushMessageToRole(v, &roguelike_match.RoguelikeMatch_RLRoguelikeFinishPush{
					RoomId:    int64(roomId),
					BattleId:  r.BattleId(),
					HasReward: req.IsSuc,
				})
			}
		}
	}
	r.StopFight()
	svc.roomChangePushWithRoom(c, r)
	return
}

func (svc *Service) BattleExit(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLBattleExitEvent) {
	if req.RoleId == "" {
		return
	}
	roomId := uint64(req.RoomId)
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	serverId := svc.mgr.GetRoleSvcId(req.RoleId)
	if len(req.Rewards) > 0 {
		if serverId != -1 {
			svc.svc.GetNatsClient().Publish(serverId,
				ctx.NewHeader(req.RoleId, svc.serverId, svc.serverType, c),
				&service2.BattleEvent_AddItemsEvent{
					Items: req.Rewards,
				})
			svc.svc.GetNatsClient().Publish(serverId,
				ctx.NewHeader(req.RoleId, svc.serverId, svc.serverType, c),
				&service2.Task_FinishRLEvent{
					RoguelikeId: req.RoguelikeId,
				})
			svc.svc.GetNatsClient().Publish(serverId,
				ctx.NewHeader(req.RoleId, svc.serverId, svc.serverType, c),
				&service2.Journey_RoguelikeFinishPush{
					Success:     false,
					RoguelikeId: req.RoguelikeId,
				})
		}
		if _, cnt, err := svc.handleCnt(c, req.RoleId); err == nil {
			cnt.TodayJoin += 1
			cnt.LastJoinAt = timer.StartTime(c.StartTime).Unix()
			dao.SaveCnt(c, cnt)
		}
	}
	if err := svc.mgr.Leave(roomId, req.RoleId); err != nil {
		return
	}
	svc.roomChangePush(c, roomId, values.RoguelikeId(req.RoguelikeId))
	im.DefaultClient.LeaveRoom(c, &im.RoomRole{
		RoomID:  strconv.FormatInt(req.RoomId, 10),
		RoleIDs: []string{req.RoleId},
	})
}

func (svc *Service) MatchOffline(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLMatchOfflineRequest) (*roguelike_match.RoguelikeMatch_RLMatchOfflineResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, errmsg.NewErrNotIn()
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if r.Remain() == 0 {
		return nil, errmsg.NewErrRLRoomFull()
	}
	remain := int(r.Remain())
	reader := rule.MustGetReader(c)
	roguelikeCfg, ok := reader.RoguelikeDungeon.GetRoguelikeDungeonById(values.Integer(roguelikeId))
	if !ok {
		return nil, errmsg.NewErrRoguelikeNotExist()
	}
	lv := roguelikeCfg.DungeonLv
	if len(lv) != 2 {
		return nil, errmsg.NewErrRoguelikeNotExist()
	}
	var configMap map[values.Integer]bool
	if len(req.ConfigIds) > 0 {
		configMap = map[values.Integer]bool{}
		for _, configId := range req.ConfigIds {
			for _, deriveId := range reader.DeriveHeroMap(configId) {
				configMap[deriveId] = true
			}
		}
	}
	cfgLen := reader.Robot.Len()
	cfgList := reader.Robot.List()
	if cfgLen == 0 {
		return &roguelike_match.RoguelikeMatch_RLMatchOfflineResponse{}, nil
	}
	if len(req.ConfigIds) <= 1 {
		for i := 0; i < remain; i++ {
			for matchTime := 0; matchTime < maxMatchCnt; matchTime++ {
				robot := cfgList[rand.Intn(cfgLen)]
				if robot.Lv >= lv[0] && robot.Lv <= lv[1] {
					if len(configMap) > 0 && !configMap[robot.ConfigId] {
						continue
					}
					robotRoleId := svc.genRobotRoleId(robot.Id)
					if err := svc.mgr.JoinRobot(roomId, roguelikeId, robotRoleId, robot.ConfigId); err != nil {
						return nil, err
					}
					svc.robotInfoMgr.Set(robotRoleId, robot.Id, rule.MustGetReader(c).GenRandomRobotNickname())
					break
				}
			}
		}
	} else {
		used := map[values.Integer]bool{}
		for i := 0; i < remain; i++ {
			for matchTime := 0; matchTime < maxMatchCnt; matchTime++ {
				robot := cfgList[rand.Intn(cfgLen)]
				originId := reader.OriginHero(robot.ConfigId)
				if originId <= 0 {
					continue
				}
				if robot.Lv >= lv[0] && robot.Lv <= lv[1] {
					if used[originId] || (len(configMap) > 0 && !configMap[robot.ConfigId]) {
						continue
					}
					robotRoleId := svc.genRobotRoleId(robot.Id)
					if err := svc.mgr.JoinRobot(roomId, roguelikeId, robotRoleId, robot.ConfigId); err != nil {
						return nil, err
					}
					svc.robotInfoMgr.Set(robotRoleId, robot.Id, rule.MustGetReader(c).GenRandomRobotNickname())
					used[originId] = true
					break
				}
			}
		}
	}
	svc.roomChangePushWithRoom(c, r)
	return &roguelike_match.RoguelikeMatch_RLMatchOfflineResponse{}, nil
}

func (svc *Service) genRobotRoleId(id values.Integer) values.RoleId {
	var curr = atomic.LoadInt64(&svc.unqRobot)
	now := timer.Unix()
	idStr := fmt.Sprintf("robot%d%d%d", id, now-now/1000*1000, curr)
	atomic.AddInt64(&svc.unqRobot, 1)
	if curr >= 100000 {
		atomic.StoreInt64(&svc.unqRobot, 0)
	}
	return idStr
}

func (svc *Service) Kick(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLKickRequest) (*roguelike_match.RoguelikeMatch_RLKickResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, errmsg.NewErrNotIn()
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	room := svc.mgr.GetRoom(roomId, roguelikeId)
	if room == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if room.OwnerId() != c.RoleId {
		return nil, errmsg.NewErrRLNotOwner()
	}
	if room.IsStarted() {
		return nil, errmsg.NewErrRoguelikeAlreadyStart()
	}
	if room.IsRobot(req.RoleId) {
		room.KickRobot(req.RoleId)
		svc.robotInfoMgr.Del(req.RoleId)
	} else {
		err := svc.mgr.Leave(roomId, req.RoleId)
		if err != nil {
			return nil, err
		}
		c.PushMessageToRole(req.RoleId, &roguelike_match.RoguelikeMatch_RLBeKickPush{
			RoomId: int64(roomId),
		})
		svc.svc.AfterFuncCtx(c, time.Microsecond, func(c1 *ctx.Context) {
			im.DefaultClient.LeaveRoom(c1, &im.RoomRole{
				RoomID:  strconv.FormatInt(int64(roomId), 10),
				RoleIDs: []string{req.RoleId},
			})
		})
	}
	svc.roomChangePushWithRoom(c, room)
	return &roguelike_match.RoguelikeMatch_RLKickResponse{}, nil
}

func (svc *Service) ChangeOwner(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLChangeOwnerRequest) (*roguelike_match.RoguelikeMatch_RLChangeOwnerResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, errmsg.NewErrNotIn()
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	room := svc.mgr.GetRoom(roomId, roguelikeId)
	if room == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if room.OwnerId() != c.RoleId {
		return nil, errmsg.NewErrRLNotOwner()
	}
	if room.IsStarted() {
		return nil, errmsg.NewErrRoguelikeAlreadyStart()
	}
	if room.IsRobot(req.RoleId) {
		return nil, errmsg.NewErrRLRobotCantBeOwner()
	}
	if err := room.ChangeOwner(req.RoleId); err != nil {
		return nil, err
	}
	svc.roomChangePushWithRoom(c, room)
	return &roguelike_match.RoguelikeMatch_RLChangeOwnerResponse{}, nil
}

func GenIMRoleInfoExtra(role *room_mgr.RoleSimpleInfo) string {
	extra := make(map[string]interface{})
	extra[imutil.ExtraRoleLvKey] = role.Level
	extra[imutil.ExtraRoleAvatarIdKey] = role.AvatarId
	extra[imutil.ExtraRoleAvatarFrameKey] = role.AvatarFrame
	_extra, _ := jsoniter.MarshalToString(extra)
	return _extra
}

func (svc *Service) Invite(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLInviteRequest) (*roguelike_match.RoguelikeMatch_RLInviteResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, errmsg.NewErrNotIn()
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	room := svc.mgr.GetRoom(roomId, roguelikeId)
	if room == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if room.OwnerId() != c.RoleId {
		return nil, errmsg.NewErrOnlyOwnerCanInvite()
	}
	info := svc.getRoleInfo(c, c.RoleId)
	if info == nil {
		return nil, errmsg.NewInternalErr("role not found")
	}
	if err := im.DefaultClient.SendMessage(c, &im.Message{
		Type:      im.MsgTypePrivate,
		RoleID:    info.RoleId,
		RoleName:  info.Nickname,
		TargetID:  req.RoleId,
		Content:   fmt.Sprintf(`{"language_id":%d,"room_id":%d,"roguelike_id":%d}`, 1750, room.RoomId(), room.RoguelikeId()),
		ParseType: im.ParseRoguelikeInvite,
		Extra:     GenIMRoleInfoExtra(info),
	}); err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	serverId := svc.mgr.GetRoleSvcId(info.RoleId)
	err := svc.svc.GetNatsClient().RequestWithOut(c, serverId, &lessservicepb.User_AddRecentChatIdsRequest{RoleId: req.RoleId}, &lessservicepb.User_AddRecentChatIdsResponse{})
	if err != nil {
		return nil, err
	}
	return &roguelike_match.RoguelikeMatch_RLInviteResponse{}, nil
}

func (svc *Service) PubRoomToChat(c *ctx.Context, req *roguelike_match.RoguelikeMatch_RLPubRoomToChatRequest) (*roguelike_match.RoguelikeMatch_RLPubRoomToChatResponse, *errmsg.ErrMsg) {
	roomId, roguelikeId := svc.mgr.GetRoleRoom(c.RoleId)
	if roomId == room_mgr.UnExistRoom {
		return nil, errmsg.NewErrNotIn()
	}
	group_lock.LockRoom(roomId)
	defer group_lock.UnlockRoom(roomId)
	room := svc.mgr.GetRoom(roomId, roguelikeId)
	if room == nil {
		return nil, errmsg.NewErrRLRoomNotExist()
	}
	if room.OwnerId() != c.RoleId {
		return nil, errmsg.NewErrOnlyOwnerCanInvite()
	}
	info := svc.getRoleInfo(c, c.RoleId)
	if info == nil {
		return nil, errmsg.NewInternalErr("role not found")
	}
	if err := im.DefaultClient.SendMessage(c, &im.Message{
		Type:      im.MsgTypeBroadcast,
		RoleID:    info.RoleId,
		RoleName:  info.Nickname,
		Content:   fmt.Sprintf(`{"language_id":%d,"room_id":%d,"roguelike_id":%d}`, 1750, room.RoomId(), room.RoguelikeId()),
		ParseType: im.ParseRoguelikeInvite,
		Extra:     GenIMRoleInfoExtra(info),
	}); err != nil {
		return nil, errmsg.NewInternalErr(err.Error())
	}
	return &roguelike_match.RoguelikeMatch_RLPubRoomToChatResponse{}, nil
}

func (svc *Service) fillParty(c *ctx.Context, arr []*models.RoguelikeRoleInfo) []*models.RoguelikeRoleInfo {
	for _, v := range arr {
		if !v.IsRobot {
			info := svc.getRoleInfo(c, v.RoleId)
			if info != nil {
				v.GuildId = info.GuildId
				v.GuildName = info.GuildName
				v.Nickname = info.Nickname
				v.Level = info.Level
				v.AvatarId = info.AvatarId
				v.AvatarFrame = info.AvatarFrame
				if v.ConfigId != 0 {
					for _, hero := range info.Heroes {
						if hero.ConfigId == v.ConfigId {
							v.Power = hero.Power
						}
					}
				}
			}
		} else {
			robot := svc.robotInfoMgr.Get(v.RoleId)
			if robot != nil {
				info, ok := rule.MustGetReader(c).Robot.GetRobotById(robot.RobotId)
				if ok {
					v.Level = info.Lv
					v.Nickname = robot.Nickname
					v.Power = info.Combat
				}
			}
		}
	}
	return arr
}

func (svc *Service) getRoleInfo(c *ctx.Context, roleId values.RoleId) *room_mgr.RoleSimpleInfo {
	info := svc.roleInfoMgr.Get(roleId)
	if info == nil {
		simpleInfoResp := &lessservicepb.User_GetUserSimpleInfoResponse{}
		svcId := svc.mgr.GetRoleSvcId(roleId)
		if svcId < 0 {
			return nil
		}
		if err := svc.svc.GetNatsClient().RequestWithOut(c, svcId, &lessservicepb.User_GetUserSimpleInfoRequest{
			RoleId: roleId,
		}, simpleInfoResp); err != nil {
			return nil
		}
		info = svc.roleInfoMgr.Set(roleId, simpleInfoResp.Info)
	}
	return info
}

func (svc *Service) roomChangePush(c *ctx.Context, roomId values.MatchRoomId, roguelikeId values.RoguelikeId) {
	r := svc.mgr.GetRoom(roomId, roguelikeId)
	if r == nil {
		return
	}
	svc.roomChangePushWithRoom(c, r)
}

func (svc *Service) roomChangePushWithRoom(c *ctx.Context, r *room_mgr.Room) {
	c.PushMessageToRoles(r.GetParty(), &roguelike_match.RoguelikeMatch_RLRoomChangePush{
		Room: &models.RoguelikeRoom{
			RoomId:      int64(r.RoomId()),
			OwnerId:     r.OwnerId(),
			RoguelikeId: int64(r.RoguelikeId()),
			RoleInfos:   svc.fillParty(c, r.GetPartyInfo()),
			IsOpen:      r.IsOpen(),
			UpdateAt:    r.CloseAt(svc.mgr.CloseAfter),
			CombatNeed:  r.CombatNeed(),
		},
	})
}

func (svc *Service) roomChangePushWithModel(c *ctx.Context, room *models.RoguelikeRoom) {
	roleIds := make([]values.RoleId, 0, len(room.RoleInfos))
	for _, p := range room.RoleInfos {
		if p.RoleId != "" {
			roleIds = append(roleIds, p.RoleId)
		}
	}
	c.PushMessageToRoles(roleIds, &roguelike_match.RoguelikeMatch_RLRoomChangePush{
		Room: room,
	})
}

func (svc *Service) handleCnt(c *ctx.Context, roleId values.RoleId) (bool, *pbdao.RoguelikeJoinCnt, *errmsg.ErrMsg) {
	cnt, err := dao.GetCnt(c, roleId)
	if err != nil {
		return false, nil, err
	}
	n := svc.getCurrDayFreshTime(c)
	if cnt.LastJoinAt < n.Unix() {
		cnt.TodayJoin = 0
		cnt.LastJoinAt = n.Unix()
		return true, cnt, nil
	}
	return false, cnt, nil
}

//---------------------------------------------------cheat------------------------------------------------------------//

func (svc *Service) CheatChangeBoss(c *ctx.Context, _ *roguelike_match.RoguelikeMatch_CheatChangeBossRequest) (*roguelike_match.RoguelikeMatch_CheatChangeBossResponse, *errmsg.ErrMsg) {
	svc.genBossSkill(c)
	dao.SaveBossSkill(c, svc.dailyBossSkill)
	return &roguelike_match.RoguelikeMatch_CheatChangeBossResponse{}, nil
}

func (svc *Service) CheatClearTodayCnt(c *ctx.Context, _ *roguelike_match.RoguelikeMatch_CheatClearTodayCntRequest) (*roguelike_match.RoguelikeMatch_CheatClearTodayCntResponse, *errmsg.ErrMsg) {
	_, cnt, err := svc.handleCnt(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	cnt.TodayJoin = 0
	dao.SaveCnt(c, cnt)
	return &roguelike_match.RoguelikeMatch_CheatClearTodayCntResponse{}, nil
}

//---------------------------------------------------loadTest------------------------------------------------------------//

/*func (svc *Service) loadTest() {
	// 压力测试， 1W用户不停创建删除查询
	go func() {
		cnt := int64(0)
		for {
			time.Sleep(time.Second)
			curr := atomic.LoadInt64(&(svc.cnt))
			fmt.Println(curr-cnt, svc.mgr.Fmt())
			cnt = curr
		}
	}()
	ownerRoleIds := make([]values.RoleId, 100000)
	joinerRoleIds := make([]values.RoleId, 200000)
	for i := 0; i < 100000; i++ {
		roleIdInt, _ := idgenerate.GenerateID(context.Background(), idgenerate.RoleIDKey)
		ownerRoleIds[i] = utils.Base34EncodeToString(uint64(roleIdInt))
	}
	for i := 0; i < 200000; i++ {
		roleIdInt, _ := idgenerate.GenerateID(context.Background(), idgenerate.RoleIDKey)
		joinerRoleIds[i] = utils.Base34EncodeToString(uint64(roleIdInt))
	}
	for _, owner := range ownerRoleIds {
		go func(rId values.RoleId) {
			for {
				r, _ := svc.mgr.CreateRoom(&ctx.Context{}, rId, 3, 30, 31)
				if r.OwnerId() != rId {
					panic(fmt.Sprintf("roleId not equal, %s, %s \n", r.OwnerId(), rId))
				}
				group_lock.LockRoom(r.RoomId())
				roomId, roguelikeId := svc.mgr.GetRoleRoom(rId)
				r1 := svc.mgr.GetRoom(roomId, roguelikeId)
				if r1.OwnerId() != rId || r1.OwnerId() != r.OwnerId() || r1.RoomId() != roomId {
					panic(fmt.Sprintf("GetRoom data fail, %s, %s, %s \n", r1.OwnerId(), rId, r1.RoomId()))
				}
				group_lock.UnlockRoom(roomId)
				time.Sleep(time.Duration(rand.Intn(10)+5) * time.Second)
				group_lock.LockRoom(roomId)
				if err := svc.mgr.DelRoom(roomId, roguelikeId); err != nil {
					panic(err)
				}
				r3 := svc.mgr.GetRoom(roomId, roguelikeId)
				if r3 != nil {
					panic(fmt.Sprintf("GetRoom should get nil, %s, %s \n", r3.OwnerId(), r3.RoomId()))
				}
				group_lock.UnlockRoom(roomId)
				atomic.AddInt64(&(svc.cnt), 1)
			}
		}(owner)
	}
	for _, joiner := range joinerRoleIds {
		go func(rId values.RoleId) {
			for {
				time.Sleep(time.Duration(rand.Intn(30)+70) * time.Millisecond)
				roomList, err := svc.mgr.GetRandomRooms(3, 3)
				if err != nil {
					panic(err)
				}
				if len(roomList) == 0 {
					continue
				}
				// atomic.AddInt64(&(svc.cnt), 1)
				r := roomList[0]
				roomId := r.RoomId()
				dungenId := r.roguelikeId()
				group_lock.LockRoom(roomId)
				err = svc.mgr.Join(roomId, dungenId, rId, 30)
				if err != nil {
					// fmt.Println(err)
					group_lock.UnlockRoom(roomId)
					continue
				}
				if err := svc.mgr.Leave(roomId, rId); err != nil {
					fmt.Println(err)
				}
				group_lock.UnlockRoom(roomId)
				atomic.AddInt64(&(svc.cnt), 1)
			}
		}(joiner)
	}
}

*/
