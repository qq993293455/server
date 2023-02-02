package edge

import (
	"fmt"
	"math"
	"time"
	"unsafe"

	"coin-server/common/errmsg"
	"coin-server/common/eventloop"
	"coin-server/common/logger"
	"coin-server/common/network/stdtcp"
	edgepb "coin-server/common/proto/edge"
	modelspb "coin-server/common/proto/models"
	centerpb "coin-server/common/proto/newcenter"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/rule"

	"go.uber.org/zap"
)

var gEl *eventloop.ChanEventLoop

func InitEdge(log *logger.Logger) {
	gEl = eventloop.NewChanEventLoop(log)
	gEl.Start(func(e interface{}) {

	})
	initEdgeHash(log)
	initBattleHash(log)
	initRoleHash(log)
	initGuildBossHash(log)
	initBossHall(log)
}

type SyncPush struct {
	session     *stdtcp.Session
	msg         *centerpb.NewCenter_SyncEdgeInfoPush
	killPlayers *edgepb.Edge_KillPlayerRequest
}

type CloseSession struct {
	session *stdtcp.Session
	typ     modelspb.EdgeType
}

type ReqBossInfo struct {
	BossId int64
	MaxCnt int32
}

type GetRoleInfo struct {
	uRoleId URoleID
	info    *RoleInfo
}

type FindSession struct {
	edgeType modelspb.EdgeType
	weight   values.Integer
	session  *stdtcp.Session
}

type SetEdgeTyp struct {
	session  *stdtcp.Session
	edgeType modelspb.EdgeType
}

type GetGuildBoss struct {
	guildBossId string
	battleSrvId values.Integer
	exist       bool
}

type GuildBossCount struct {
	guildBossId string
	count       values.Integer
}

type AddBattleLine struct {
	mapId    values.Integer
	lineId   values.Integer
	battleId values.Integer
}

type MapNum struct {
	CurNum    values.Integer
	MaxNum    values.Integer
	Name      string
	NewLineId values.Integer
}

type CheckCreateBattleLine struct {
	monitorMap map[string]values.Integer
	maps       map[values.Integer]MapNum
}

type CheckLinesFreeArgs struct {
	maps map[values.Integer]bool
}

type FindEmptyLineArgs struct {
	MapId    values.Integer
	LineId   values.Integer
	BattleId values.Integer
	Session  *stdtcp.Session
}

type NeedInitBattleLine struct {
	serverId     values.Integer
	defaultLines map[values.Integer]map[values.Integer]bool // mapId:lineId
	addLines     map[values.Integer]map[values.Integer]bool // mapId:lineId
}

type CheckLineInit struct {
	mapId    values.Integer
	battleId values.Integer

	lineId       values.Integer
	lineBattleId values.Integer
	needInit     bool
}

type TargetBattleServer struct {
	battleId    values.Integer
	resBattleId values.Integer
}

type AllMapLines struct {
	serverId values.Integer
	mapId    values.Integer
	resp     map[values.Integer]*modelspb.LineInfo
}

type TargetMapLines struct {
	roleId       string
	mapId        values.Integer
	battleId     values.Integer
	maxNum       values.ServerId
	resp         *modelspb.LineInfo
	scaleDeleted bool //该分线battle已经被删除 （线上做了缩容操作）
}

type CheckRoleValidArgs struct {
}

type NeedInitBossHallLine struct {
	Lines map[values.Integer]map[values.Integer]values.Integer // mapId:{ lineId:BattleId}
}

type GetBossHallArg struct {
	RoleId      string
	BossId      values.Integer
	MaxNum      values.Integer
	BattleSrvId values.Integer
	LineId      values.Integer
}

type CheckCreateBossHallLine struct {
	monitorInterVal values.Integer
	monitorMap      map[string]values.Integer
	maps            map[values.Integer]values.Integer
}

type FindEmptyBossHallLineArgs struct {
	BossId   values.Integer
	LineId   values.Integer
	BattleId values.Integer
	Session  *stdtcp.Session
}

func SyncEdge(s *stdtcp.Session, msg *centerpb.NewCenter_SyncEdgeInfoPush) (*edgepb.Edge_KillPlayerRequest, *errmsg.ErrMsg) {
	req := &SyncPush{
		session: s, msg: msg,
		killPlayers: &edgepb.Edge_KillPlayerRequest{Players: make(map[values.Integer]*edgepb.Edge_KillPlayers)},
	}
	res, err := eventloop.CallChanEventLoop(gEl, req, doSyncPush)
	if err != nil {
		return nil, err
	}
	return res.killPlayers, nil
}

func OnSessionClose(s *stdtcp.Session) (modelspb.EdgeType, *errmsg.ErrMsg) {
	req := &CloseSession{session: s}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *CloseSession) (*CloseSession, *errmsg.ErrMsg) {
		edgeInfo, ok := gEdgesHash.Get(r.session)
		if ok {
			for battleSrvId, _ := range edgeInfo.servers {
				roles, exist := gBattleHash.GetRoles(battleSrvId)
				if exist {
					for uRoleId, _ := range roles {
						gRoleHash.ClearRoleBattle(uRoleId)
					}
				}
				gBattleHash.Remove(battleSrvId)
				gGuildBossHash.RemoveByBattleId(battleSrvId)
				gBossHall.RemoveBattle(battleSrvId)
			}
			gEdgesHash.Remove(r.session)
		}
		gEdgesHash.log.Info("edge closed", zap.String("addr", r.session.RemoteAddr()), zap.Any("edge", edgeInfo))
		return r, nil
	})
	if err != nil {
		return modelspb.EdgeType_StatelessServer, err
	}
	return res.typ, nil
}

func FindOneEdge(edgeType modelspb.EdgeType, weight values.Integer) (*stdtcp.Session, *errmsg.ErrMsg) {
	req := &FindSession{edgeType: edgeType, weight: weight}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *FindSession) (*FindSession, *errmsg.ErrMsg) {
		if r.edgeType == modelspb.EdgeType_StatelessServer {
			se := gEdgesHash.FindOneEmptyEdge()
			r.session = se
			return r, nil
		}
		se := gEdgesHash.FindOne(r.edgeType, r.weight)
		r.session = se
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	return res.session, nil
}

func SetEdgeType(session *stdtcp.Session, edgeType modelspb.EdgeType) *errmsg.ErrMsg {
	req := &SetEdgeTyp{session: session, edgeType: edgeType}
	_, err := eventloop.CallChanEventLoop(gEl, req, func(r *SetEdgeTyp) (*SetEdgeTyp, *errmsg.ErrMsg) {
		ok := gEdgesHash.SetType(r.session, r.edgeType)
		if !ok {
			return nil, errmsg.NewInternalErr("set edgeType fail!edge session not exist")
		}
		for k, v := range gEdgesHash.all {
			gEdgesHash.log.Info("edge", zap.String("addr", k.RemoteAddr()), zap.Any("edgeType", v.edgeType))
		}
		return r, nil
	})
	return err
}

func GetGuildBossBattle(guildBossId string) (values.Integer, bool, *errmsg.ErrMsg) {
	req := &GetGuildBoss{guildBossId: guildBossId}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *GetGuildBoss) (*GetGuildBoss, *errmsg.ErrMsg) {
		val, ok := gGuildBossHash.Get(r.guildBossId)
		r.exist = false
		if ok {
			r.exist = true
			r.battleSrvId = val.battleServerId
		}
		return r, nil
	})
	if err != nil {
		return 0, false, err
	}
	return res.battleSrvId, res.exist, nil
}

func GetGuildBossCount(guildBossId string) (values.Integer, *errmsg.ErrMsg) {
	req := &GuildBossCount{guildBossId: guildBossId}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *GuildBossCount) (*GuildBossCount, *errmsg.ErrMsg) {
		val, ok := gGuildBossHash.Get(r.guildBossId)
		if !ok {
			return req, nil
		}
		bv, ok1 := gBattleHash.Get(val.battleServerId)
		if !ok1 {
			return nil, errmsg.NewInternalErr("guild boss battleId not exist")
		}
		r.count = values.Integer(len(bv.Roles))
		return r, nil
	})
	if err != nil {
		return 0, err
	}
	return res.count, nil
}

func doSyncPush(req *SyncPush) (*SyncPush, *errmsg.ErrMsg) {
	edgeInfo, ok := gEdgesHash.Get(req.session)
	if !req.msg.IsAll && !ok {
		return req, errmsg.NewInternalErr("not IsAll but stdSession not init")
	}
	if ok && edgeInfo.edgeType != req.msg.Typ {
		return req, errmsg.NewInternalErr("edge type not match")
	}

	// 全量更新,先清除当前数据
	if req.msg.IsAll {
		err := doSyncPushIsAll(req)
		return req, err
	}

	//某个battle结束
	if !req.msg.IsAll && len(req.msg.Edges) == 1 && req.msg.Edges[0].IsEnd {
		err := doSyncBattleEnd(req)
		return req, err
	}

	//玩家进入、离开
	if !req.msg.IsAll && len(req.msg.Edges) == 1 && !req.msg.Edges[0].IsEnd {
		err := doSyncRole(req)
		return req, err
	}

	return req, errmsg.NewInternalErr("invalid msg")
}

//全量更新
func doSyncPushIsAll(req *SyncPush) *errmsg.ErrMsg {
	gGuildBossHash.log.Debug("enter doSyncPushIsAll")

	//角色已经到了其他战斗
	checkFlag := false
	addr := req.session.RemoteAddr()
	for _, edge := range req.msg.Edges {
		for _, role := range edge.Roles {
			uRoleId := utils.Base34DecodeString(role.RoleId)
			roleInfo, ok := gRoleHash.Get(uRoleId)
			if !ok {
				continue
			}
			curAddr := ""
			battleInfo, ok1 := gBattleHash.Get(roleInfo.BattleServerId)
			if ok1 {
				curAddr = battleInfo.Session.RemoteAddr()
			}
			if (edge.MapId > 0 && edge.MapId == roleInfo.MapId &&
				(edge.BattleId != roleInfo.BattleServerId || role.AddTime < roleInfo.AddTime)) || ok1 && curAddr != addr {
				killRoles, exist := req.killPlayers.Players[edge.BattleId]
				if !exist {
					req.killPlayers.Players[edge.BattleId] = &edgepb.Edge_KillPlayers{Roles: []string{role.RoleId}}
				} else {
					killRoles.Roles = append(killRoles.Roles, role.RoleId)
					req.killPlayers.Players[edge.BattleId] = killRoles
				}
				checkFlag = true
				gRoleHash.log.Info("role already in new battle!", zap.String("role", role.GoString()),
					zap.String("addr", addr), zap.String("curAddr", curAddr), zap.Any("oldRoleInfo", roleInfo))
				continue
			}
		}
	}
	if checkFlag {
		return nil
	}

	err := gGuildBossHash.Check(req.msg)
	if err != nil {
		return err
	}

	edgeInfo, ok := gEdgesHash.Get(req.session)
	if ok {
		for battleSrvId, _ := range edgeInfo.servers {
			roles, exist := gBattleHash.GetRoles(battleSrvId)
			if exist {
				for uRoleId, _ := range roles {
					gRoleHash.Remove(uRoleId)
				}
			}
			gBattleHash.log.Debug("doSyncPushIsAll removeRoles", zap.Int64("battleId", battleSrvId), zap.Any("roles", roles))
			gBattleHash.Remove(battleSrvId)

		}
	}
	err = gEdgesHash.Sync(req.session, req.msg)
	if err != nil {
		return err
	}

	gBossHall.Sync(req.msg)

	for _, edge := range req.msg.Edges {
		gRoleHash.Sync(edge, req.msg.IsAll)
		gBattleHash.Sync(req.session, edge, req.msg.IsAll)
		gGuildBossHash.Sync(edge)
	}
	for shard, roles := range gRoleHash.all {
		if len(roles) > 0 {
			gRoleHash.log.Debug("allRoles", zap.Int("shard", shard), zap.Int("allNum", len(roles)), zap.Any("roles", roles))
		}
	}
	return nil
}

//某个battle结束
func doSyncBattleEnd(req *SyncPush) *errmsg.ErrMsg {
	gGuildBossHash.log.Debug("enter doSyncBattleEnd")

	err := gGuildBossHash.Check(req.msg)
	if err != nil {
		return err
	}

	gBossHall.Sync(req.msg)

	for _, edge := range req.msg.Edges {
		if edge.IsEnd {
			roles, exist := gBattleHash.GetRoles(edge.BattleId)
			if exist {
				for uRoleId, _ := range roles {
					gRoleHash.Remove(uRoleId)
				}
			}
			gBattleHash.log.Debug("doSyncBattleEnd removeRoles", zap.Int64("battleId", edge.BattleId),
				zap.Any("roles", roles), zap.Any("edge.Roles", edge.Roles))

			gBattleHash.Remove(edge.BattleId)
			gGuildBossHash.Remove(edge.GuildBossId)

			for _, role := range edge.Roles {
				uRoleId := utils.Base34DecodeString(role.RoleId)
				gRoleHash.Remove(uRoleId)
			}
		}
	}
	err = gEdgesHash.Sync(req.session, req.msg)

	gBossHall.Sync(req.msg)

	return err
}

//更新某个battle的Role
func doSyncRole(req *SyncPush) *errmsg.ErrMsg {
	gGuildBossHash.log.Debug("enter doSyncRole")

	gBossHall.Sync(req.msg)

	err := gGuildBossHash.Check(req.msg)
	if err != nil {
		return err
	}
	err = gEdgesHash.Sync(req.session, req.msg)
	if err != nil {
		return err
	}
	for _, edge := range req.msg.Edges {
		gRoleHash.Sync(edge, req.msg.IsAll)
		gBattleHash.Sync(req.session, edge, req.msg.IsAll)
		gGuildBossHash.Sync(edge)
	}
	//for shard, roles := range gRoleHash.all {
	//	if len(roles) > 0 {
	//		gRoleHash.log.Debug("allRoles", zap.Int("shard", shard), zap.Int("allNum", len(roles)), zap.Any("roles", roles))
	//	}
	//}
	return nil
}

func GetNeedInitBattleLines() (defaultLines map[values.Integer]map[values.Integer]bool, addLines map[values.Integer]map[values.Integer]bool, err *errmsg.ErrMsg) {
	req := &NeedInitBattleLine{
		defaultLines: map[values.Integer]map[values.Integer]bool{},
		addLines:     map[values.Integer]map[values.Integer]bool{},
	}
	res, err1 := eventloop.CallChanEventLoop(gEl, req, func(r *NeedInitBattleLine) (*NeedInitBattleLine, *errmsg.ErrMsg) {
		for mapId, lineNum := range gLineConfig.srvLines.Lines {
			if _, ok0 := r.defaultLines[mapId]; !ok0 {
				r.defaultLines[mapId] = map[values.Integer]bool{}
			}
			lines, ok := gLineConfig.lineBattles[mapId]
			for i := values.Integer(1); i <= lineNum; i++ {
				if !ok {
					r.defaultLines[mapId][i] = true
					continue
				}
				battleId, ok1 := lines[i]
				if !ok1 {
					r.defaultLines[mapId][i] = true
					continue
				}
				_, ok2 := gBattleHash.Get(battleId)
				if !ok2 {
					r.defaultLines[mapId][i] = true
					delete(lines, i)
					gLineConfig.lineBattles[mapId] = lines
					continue
				}
			}
		}

		for mapId, extraLines := range gLineConfig.adds {
			lines, ok := gLineConfig.lineBattles[mapId]
			if !ok {
				continue
			}
			if _, ok0 := r.defaultLines[mapId]; !ok0 {
				r.addLines[mapId] = map[values.Integer]bool{}
			}
			for lId, _ := range extraLines {
				battleId, ok1 := lines[lId]
				if !ok1 {
					r.addLines[mapId][lId] = true
					continue
				}
				_, ok2 := gBattleHash.Get(battleId)
				if ok2 {
					r.addLines[mapId][lId] = true
					delete(lines, lId)
					gLineConfig.lineBattles[mapId] = lines
					continue
				}
			}
		}

		return r, nil
	})
	if err1 != nil {
		return nil, nil, err1
	}
	return res.defaultLines, res.addLines, nil
}

func CheckBossHallNeedInit() (lines map[values.Integer]map[values.Integer]values.Integer, err *errmsg.ErrMsg) {
	req := &NeedInitBossHallLine{Lines: map[values.Integer]map[values.Integer]values.Integer{}}
	res, err1 := eventloop.CallChanEventLoop(gEl, req, func(r *NeedInitBossHallLine) (*NeedInitBossHallLine, *errmsg.ErrMsg) {
		for bossId, lineNum := range gLineConfig.srvLines.BossHall {
			for lineId := values.Integer(1); lineId <= lineNum; lineId++ {
				if _, ok := gBossHall.GetLine(bossId, lineId); !ok {
					if _, ok1 := r.Lines[bossId]; !ok1 {
						r.Lines[bossId] = map[values.Integer]values.Integer{}
					}
					r.Lines[bossId][lineId] = 0
				}
			}
		}
		return r, nil
	})
	if err1 != nil {
		return nil, err1
	}
	return res.Lines, nil
}

func GetBossHallInfo(roleId string, bossId values.Integer, maxNum values.Integer) (battleServerId values.Integer, err *errmsg.ErrMsg) {

	req := &GetBossHallArg{RoleId: roleId, BossId: bossId, MaxNum: maxNum}
	res, err1 := eventloop.CallChanEventLoop(gEl, req, func(r *GetBossHallArg) (*GetBossHallArg, *errmsg.ErrMsg) {
		if _, ok := gLineConfig.srvLines.BossHall[r.BossId]; !ok {
			return nil, errmsg.NewInternalErr("invalid boss id,check consul!")
		}
		info, ok1 := gBossHall.all[r.BossId]
		if !ok1 {
			return nil, errmsg.NewInternalErr("no valid boss line")
		}

		//如果已经在恶魔秘境
		uRoleId := utils.Base34DecodeString(r.RoleId)
		roleInfo, ok2 := gRoleHash.Get(uRoleId)
		gLineConfig.log.Debug("roleInfo", zap.Any("req", r), zap.Any("roleInfo", roleInfo))
		if !ok2 && roleInfo != nil {
			bInfo, ok3 := gBattleHash.Get(roleInfo.BattleServerId)
			//玩家当前所在的恶魔秘境和玩家请求的不是同一个
			if ok3 && bInfo.MapId != info.MapId {
				return nil, errmsg.NewInternalErr(fmt.Sprintf("role already in another hall! %d %d %d", bossId, bInfo.MapId, info.MapId))
			}
			if ok3 && bInfo.MapId == info.MapId {
				r.BattleSrvId = roleInfo.BattleServerId
				r.LineId = bInfo.LineId
				gBossHall.log.Debug("role already in hall!", zap.Any("roleInfo", roleInfo), zap.Any("r", r))
				return r, nil
			}
		}

		targetNum := values.Integer(0)
		targetLine, targetBattleId := values.Integer(-1), values.Integer(-1)

		for lineId, lineInfo := range info.Lines {
			if !lineInfo.CanEnter {
				continue
			}
			bInfo, ok3 := gBattleHash.Get(lineInfo.BattleId)
			if !ok3 {
				gLineConfig.log.Debug("bossHall line invalid", zap.Any("info", info), zap.Any("battleSrvId", lineInfo.BattleId), zap.Any("gBattleHash", gBattleHash.all))
				continue
			}
			curNum := values.Integer(len(bInfo.Roles))
			if curNum >= r.MaxNum {
				gLineConfig.log.Debug("bossHall line full", zap.Any("info", info), zap.Any("battleSrvId", lineInfo.BattleId), zap.Any("curNum", curNum))
				continue
			}
			//尽量分配人多的分线
			if targetNum == 0 || targetNum < curNum {
				targetNum = curNum
				targetBattleId = lineInfo.BattleId
				targetLine = lineId
			}
		}

		gLineConfig.log.Debug("find bossHall res", zap.Any("targetBattleId", targetBattleId), zap.Any("targetLine", targetLine), zap.Any("info", info))
		if targetBattleId > 0 && info.MapId > 0 {
			r.LineId = targetLine
			r.BattleSrvId = targetBattleId

			gBattleHash.all[r.BattleSrvId].Roles[uRoleId] = -1 //先占位
			gRoleHash.AddDefault(uRoleId, info.MapId, r.BattleSrvId)
		}
		return r, nil
	})
	if err1 != nil {
		return 0, err1
	}
	return res.BattleSrvId, nil
}

func SetBossHallCanNotEnter(msg *centerpb.NewCenter_NotCanEnterPush) *errmsg.ErrMsg {
	_, err1 := eventloop.CallChanEventLoop(gEl, msg, func(r *centerpb.NewCenter_NotCanEnterPush) (*centerpb.NewCenter_NotCanEnterPush, *errmsg.ErrMsg) {
		gBossHall.SetCanNotEnter(msg.BossId, msg.BattleId)
		return nil, nil
	})
	return err1
}

func GetTargetBattleServer(battleId values.Integer) (values.Integer, *errmsg.ErrMsg) {
	req := &TargetBattleServer{battleId: battleId}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *TargetBattleServer) (*TargetBattleServer, *errmsg.ErrMsg) {
		bInfo, ok := gBattleHash.Get(r.battleId)
		r.resBattleId = 0
		if ok {
			r.resBattleId = bInfo.BattleServerId
		}
		return r, nil
	})
	if err != nil {
		return -1, err
	}
	return res.resBattleId, nil
}

func GetRoleDetails(roleId string) (*RoleInfo, *errmsg.ErrMsg) {
	uRoleId := utils.Base34DecodeString(roleId)
	req := &GetRoleInfo{uRoleId: uRoleId}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *GetRoleInfo) (*GetRoleInfo, *errmsg.ErrMsg) {
		info, ok := gRoleHash.Get(r.uRoleId)
		if ok {
			r.info = info
		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	return res.info, nil
}

func GetAllMapLines(mapId values.Integer) (map[values.Integer]*modelspb.LineInfo, *errmsg.ErrMsg) {
	req := &AllMapLines{mapId: mapId, resp: map[values.Integer]*modelspb.LineInfo{}}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *AllMapLines) (*AllMapLines, *errmsg.ErrMsg) {
		mapLines, ok := gLineConfig.lineBattles[r.mapId]
		if !ok || len(mapLines) == 0 {
			return nil, errmsg.NewInternalErr(fmt.Sprintf("invalid battle line mapId %d", r.mapId))
		}
		for lineId, battleId := range mapLines {
			bInfo, ok2 := gBattleHash.Get(battleId)
			if ok2 {
				r.resp[lineId] = &modelspb.LineInfo{
					LineId:         lineId,
					BattleServerId: battleId,
					CurNum:         values.Integer(len(bInfo.Roles)),
					Status:         1,
				}
			} else {
				r.resp[lineId] = &modelspb.LineInfo{
					LineId:         lineId,
					BattleServerId: battleId,
					Status:         0, //服务器维护中
				}
			}
		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return res.resp, nil
}

func GetValidLine(roleId string, mapId values.Integer, battleId values.Integer, maxNum values.Integer) (*modelspb.LineInfo, bool, *errmsg.ErrMsg) {
	req := &TargetMapLines{roleId: roleId, mapId: mapId, battleId: battleId, maxNum: maxNum}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *TargetMapLines) (*TargetMapLines, *errmsg.ErrMsg) {
		mapLines := gLineConfig.getAllMapLines(r.mapId)
		if len(mapLines) == 0 {
			return nil, errmsg.NewInternalErr(fmt.Sprintf("invalid battle line mapId %d", r.mapId))
		}
		uRoleId := utils.Base34DecodeString(r.roleId)
		roleInfo, ok := gRoleHash.Get(uRoleId)
		gLineConfig.log.Debug("GetValidLine", zap.Any("mapLines", mapLines), zap.Int64("mapId", r.mapId),
			zap.Int64("battleId", r.battleId), zap.Any("roleInfo", roleInfo))
		if !ok || roleInfo.HungUpServerId != r.battleId {
			r.battleId = 0
		}
		if r.battleId > 0 {
			for lineId, battleSrvId := range mapLines {
				if battleSrvId == r.battleId {
					bInfo, ok2 := gBattleHash.Get(battleSrvId)
					if ok2 {
						r.resp = &modelspb.LineInfo{
							LineId:         lineId,
							BattleServerId: battleId,
							CurNum:         values.Integer(len(bInfo.Roles)),
							Status:         1,
						}
					}
					break
				}
			}
		}
		if r.resp == nil {
			targetNum := values.Integer(math.MaxInt64)
			targetLine, targetBattleId := values.Integer(-1), values.Integer(-1)
			for lineId, battleSrvId := range mapLines {
				bInfo, ok2 := gBattleHash.Get(battleSrvId)
				if !ok2 {
					continue
				}
				curNum := values.Integer(len(bInfo.Roles))
				if curNum < r.maxNum {
					if curNum < targetNum || (curNum == targetNum && lineId < targetLine) {
						targetNum = curNum
						targetLine, targetBattleId = lineId, battleSrvId
					}
				} else {
					gLineConfig.log.Debug("line full", zap.Int64("lineId", lineId), zap.Any("curNum", curNum), zap.Int64("mapId", r.mapId), zap.Int64("battleId", r.battleId))
				}
			}
			if targetLine > 0 && targetBattleId > 0 {
				gBattleHash.all[targetBattleId].Roles[uRoleId] = -1 //先占位
				gRoleHash.AddDefault(uRoleId, r.mapId, targetBattleId)
				r.resp = &modelspb.LineInfo{
					LineId:         targetLine,
					BattleServerId: targetBattleId,
					CurNum:         targetNum,
					Status:         1,
				}
				r.scaleDeleted = true
			}
		}

		if r.resp == nil {
			return nil, errmsg.NewInternalErr("can not find valid battle line")
		}
		return r, nil
	})
	if err != nil {
		return nil, false, err
	}
	if res == nil {
		return nil, false, nil
	}
	return res.resp, res.scaleDeleted, nil
}

func AddNewBattleLines(mapId values.Integer, lineId values.Integer, battleId values.Integer) *errmsg.ErrMsg {
	req := &AddBattleLine{mapId: mapId, lineId: lineId, battleId: battleId}
	_, err := eventloop.CallChanEventLoop(gEl, req, func(r *AddBattleLine) (*AddBattleLine, *errmsg.ErrMsg) {
		err := gLineConfig.addNewMapLine(r.mapId, r.lineId, r.battleId)
		if err != nil {
			return nil, err
		}
		return r, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func CheckCreateBattleLines() (map[values.Integer]MapNum, *errmsg.ErrMsg) {
	req := &CheckCreateBattleLine{maps: map[values.Integer]MapNum{}}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *CheckCreateBattleLine) (*CheckCreateBattleLine, *errmsg.ErrMsg) {
		err := checkCreateLineNums(r)
		if err != nil {
			return nil, err
		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	return res.maps, nil
}

func checkCreateLineNums(req *CheckCreateBattleLine) *errmsg.ErrMsg {
	if gLineConfig.srvLines.OpenMonitor <= 0 {
		return nil
	}

	for mapId, lineNum := range gLineConfig.srvLines.Lines {

		lines, ok := gLineConfig.lineBattles[mapId]
		curLineNum := values.Integer(len(lines))
		if !ok || curLineNum < lineNum {
			continue
		}

		r := rule.MustGetReader(nil)
		mapCnf, mapOk := r.MapScene.GetMapSceneById(mapId)
		if !mapOk || mapCnf.MapType != values.Integer(modelspb.BattleType_HangUp) {
			panic(fmt.Sprintf("invalid hungUp map id %d", mapId))
		}

		curNum := values.Integer(0)
		lineNums := map[values.Integer]values.Integer{}
		maxLineId := values.Integer(0)
		for lineId, battleId := range lines {
			if maxLineId < lineId {
				maxLineId = lineId
			}
			bInfo, ok1 := gBattleHash.Get(battleId)
			if !ok1 {
				curLineNum--
				delete(lines, lineId)
				continue
			}
			curNum += values.Integer(len(bInfo.Roles))
			lineNums[battleId] = values.Integer(len(bInfo.Roles))
		}

		maxNum := mapCnf.LineMaxPersonsNum * curLineNum
		if maxNum <= 0 {
			continue
		}

		per := values.Integer((float64(curNum) / float64(maxNum)) * 10000)
		if per >= gLineConfig.srvLines.Threshold {
			req.maps[mapId] = MapNum{CurNum: curNum, MaxNum: maxNum, Name: mapCnf.Title, NewLineId: maxLineId + 1}
		}
		gLineConfig.log.Info("line nums check", zap.Int64("mapId", mapId), zap.Int("battleNum", len(lineNums)),
			zap.Int64("curNum", curNum), zap.Int64("perMaxNum", mapCnf.LineMaxPersonsNum), zap.Int64("MaxNum", maxNum),
			zap.Any("lineNums", lineNums), zap.Int64("threshold", gLineConfig.srvLines.Threshold))
	}
	return nil
}

func CheckLinesFree() (map[values.Integer]bool, *errmsg.ErrMsg) {
	req := &CheckLinesFreeArgs{maps: map[values.Integer]bool{}}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *CheckLinesFreeArgs) (*CheckLinesFreeArgs, *errmsg.ErrMsg) {

		tmpLines := map[values.Integer]map[values.Integer]values.Integer{}
		for mapId, lineNum := range gLineConfig.srvLines.Lines {

			lines, ok := gLineConfig.lineBattles[mapId]
			curLineNum := values.Integer(len(lines))
			if !ok || curLineNum < lineNum {
				continue
			}

			for lineId, battleId := range lines {
				_, ok1 := gBattleHash.Get(battleId)
				if ok1 {
					if _, ok2 := tmpLines[mapId]; !ok2 {
						tmpLines[mapId] = map[values.Integer]values.Integer{}
					}
					tmpLines[mapId][lineId] = battleId
				}
			}

		}

		for mapId, mapLines := range tmpLines {

			rd := rule.MustGetReader(nil)
			mapCnf, mapOk := rd.MapScene.GetMapSceneById(mapId)
			if !mapOk || mapCnf.MapType != values.Integer(modelspb.BattleType_HangUp) {
				return nil, errmsg.NewInternalErr(fmt.Sprintf("invalid hungUp map id %d", mapId))
			}

			haveBusy, haveEmpty := false, false
			curNum := values.Integer(0)
			flagNum := values.Integer(float64(mapCnf.LineMaxPersonsNum) * 0.5)
			for _, battleId := range mapLines {
				bInfo, ok1 := gBattleHash.Get(battleId)
				if !ok1 {
					continue
				}
				curNum += values.Integer(len(bInfo.Roles))
				if curNum == 0 {
					haveEmpty = true
				}
				if curNum >= flagNum {
					haveBusy = true
				}
			}

			gLineConfig.log.Debug("CheckLinesFree!", zap.Any("mapId", mapId),
				zap.Any("haveBusy", haveBusy), zap.Any("haveEmpty", haveEmpty))

			if haveBusy || !haveEmpty {
				continue
			}

			//没有繁忙的且有空闲的 且关闭一个空闲的后总人数仍小于80%
			maxNum := mapCnf.LineMaxPersonsNum * values.Integer(len(mapLines)-1)
			val := values.Integer((float64(curNum) / float64(maxNum)) * 10000)
			if val < gLineConfig.srvLines.Threshold {
				r.maps[mapId] = true
			}

		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	return res.maps, nil
}

func FindEmptyLine(mapId values.Integer) (*FindEmptyLineArgs, *errmsg.ErrMsg) {
	req := &FindEmptyLineArgs{MapId: mapId}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *FindEmptyLineArgs) (*FindEmptyLineArgs, *errmsg.ErrMsg) {

		lines, ok := gLineConfig.adds[mapId]
		if !ok {
			return r, nil
		}
		for lineId, battleId := range lines {
			bInfo, ok2 := gBattleHash.Get(battleId)
			if !ok2 || len(bInfo.Roles) > 0 {
				continue
			}
			if r.LineId == 0 || r.LineId < lineId {
				r.LineId = lineId
				r.BattleId = battleId
				r.Session = bInfo.Session
			}
		}
		if r.LineId > 0 {
			gLineConfig.deleteBattles[r.BattleId] = r.LineId
		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CheckCreateBossHallLines() (map[values.Integer]values.Integer, *errmsg.ErrMsg) {
	req := &CheckCreateBossHallLine{maps: map[values.Integer]values.Integer{}}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *CheckCreateBossHallLine) (*CheckCreateBossHallLine, *errmsg.ErrMsg) {
		err := checkCreateBossHallLineNums(r)
		if err != nil {
			return nil, err
		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	return res.maps, nil
}

func checkCreateBossHallLineNums(req *CheckCreateBossHallLine) *errmsg.ErrMsg {
	if gLineConfig.srvLines.OpenMonitor <= 0 {
		return nil
	}

	rd := rule.MustGetReader(nil)
	for bossId, linNum := range gLineConfig.srvLines.BossHall {
		bossCnf, bossOk := rd.BossHall.GetBossHallById(bossId)
		if !bossOk {
			return errmsg.NewInternalErr(fmt.Sprintf("invalid bossHall id %d", bossId))
		}
		mapCnf, mapOk := rd.MapScene.GetMapSceneById(bossCnf.BossMap)
		if !mapOk || mapCnf.MapType != values.Integer(modelspb.BattleType_BossHall) {
			return errmsg.NewInternalErr(fmt.Sprintf("invalid bossHall map id %d %d", bossId, bossCnf.BossMap))
		}

		info, ok := gBossHall.all[bossId]
		if !ok {
			continue
		}
		curLineNum := values.Integer(len(info.Lines))
		if linNum == 0 || curLineNum < linNum {
			continue
		}

		curNum := values.Integer(0)
		//lineNums := map[values.Integer]values.Integer{}
		for _, lineInfo := range info.Lines {
			bInfo, ok1 := gBattleHash.Get(lineInfo.BattleId)
			if !ok1 {
				gLineConfig.log.Warn("bossHallLineInvalid", zap.Any("info", info), zap.Any("lineInfo", lineInfo), zap.Any("gBattleHash", gBattleHash.all))
				gBossHall.RemoveBattle(lineInfo.BattleId)
				curLineNum--
				continue
			}
			curNum += values.Integer(len(bInfo.Roles))
			//lineNums[lineInfo.BattleId] = values.Integer(len(bInfo.Roles))
		}

		if addLines, ok2 := gBossHall.adds[bossId]; ok2 {
			for lid, addBattleId := range addLines {
				bInfo, ok1 := gBattleHash.Get(addBattleId)
				if !ok1 {
					gLineConfig.log.Warn("bossHallAddLineInvalid", zap.Any("info", info), zap.Any("lid", lid), zap.Any("addBattleId", addBattleId))
					gBossHall.RemoveBattle(addBattleId)
					continue
				}
				curLineNum++
				curNum += values.Integer(len(bInfo.Roles))
			}
		}

		maxNum := mapCnf.LineMaxPersonsNum * curLineNum
		if maxNum <= 0 {
			continue
		}

		per := values.Integer((float64(curNum) / float64(maxNum)) * 10000)
		if gLineConfig.srvLines.BossHallThreshold > 0 && per >= gLineConfig.srvLines.BossHallThreshold {
			newLineId := gBossHall.GetNewLineId(bossId)
			req.maps[bossId] = newLineId
		}

		gLineConfig.log.Info("bossHallLineNumsCheck", zap.Int64("bossId", bossId), zap.Int64("curNum", curNum),
			zap.Int64("maxNum", maxNum), zap.Any("per", per),
			zap.Int64("threshold", gLineConfig.srvLines.BossHallThreshold), zap.Any("req.maps", req.maps))
	}
	return nil
}

func CheckBossHallLinesFree() (map[values.Integer]bool, *errmsg.ErrMsg) {
	req := &CheckLinesFreeArgs{maps: map[values.Integer]bool{}}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *CheckLinesFreeArgs) (*CheckLinesFreeArgs, *errmsg.ErrMsg) {

		tmpLines := map[values.Integer]map[values.Integer]values.Integer{}
		for bossId, bossInfo := range gBossHall.all {
			bossHallLineNum, ok := gLineConfig.srvLines.BossHall[bossId]
			if !ok || values.Integer(len(bossInfo.Lines)) < bossHallLineNum {
				continue
			}
			for lineId, info := range bossInfo.Lines {
				_, ok1 := gBattleHash.Get(info.BattleId)
				if ok1 {
					if _, ok2 := tmpLines[bossId]; !ok2 {
						tmpLines[bossId] = map[values.Integer]values.Integer{}
					}
					tmpLines[bossId][lineId] = info.BattleId
				}
			}

			if addLines, ok3 := gBossHall.adds[bossId]; ok3 {
				for lineId, addBattleId := range addLines {
					_, ok1 := gBattleHash.Get(addBattleId)
					if ok1 {
						if _, ok2 := tmpLines[bossId]; !ok2 {
							tmpLines[bossId] = map[values.Integer]values.Integer{}
						}
						tmpLines[bossId][lineId] = addBattleId
					}
				}
			}

		}

		rd := rule.MustGetReader(nil)
		for bossId, mapLines := range tmpLines {
			bossCnf, bossOk := rd.BossHall.GetBossHallById(bossId)
			if !bossOk {
				continue
			}
			mapCnf, mapOk := rd.MapScene.GetMapSceneById(bossCnf.BossMap)
			if !mapOk || mapCnf.MapType != values.Integer(modelspb.BattleType_BossHall) {
				continue
			}

			haveBusy, haveEmpty := false, false
			curNum := values.Integer(0)
			flagNum := values.Integer(float64(mapCnf.LineMaxPersonsNum) * 0.5)
			for _, battleId := range mapLines {
				bInfo, ok1 := gBattleHash.Get(battleId)
				if !ok1 {
					continue
				}
				curNum += values.Integer(len(bInfo.Roles))
				if curNum == 0 {
					haveEmpty = true
				}
				if curNum >= flagNum {
					haveBusy = true
				}
			}

			gLineConfig.log.Info("CheckBossHallLinesFree!", zap.Any("bossId", bossId),
				zap.Any("haveBusy", haveBusy), zap.Any("haveEmpty", haveEmpty))

			if haveBusy || !haveEmpty {
				continue
			}

			//没有繁忙的且有空闲的 且关闭一个空闲的后总人数仍小于80%
			maxNum := mapCnf.LineMaxPersonsNum * values.Integer(len(mapLines)-1)
			val := values.Integer((float64(curNum) / float64(maxNum)) * 10000)
			if val < gLineConfig.srvLines.BossHallThreshold {
				r.maps[bossId] = true
			}
			gLineConfig.log.Info("CheckBossHallLinesFree!", zap.Any("curNum", curNum),
				zap.Any("maxNum", maxNum), zap.Any("val", val),
				zap.Any("threshold", gLineConfig.srvLines.BossHallThreshold), zap.Any("maps", r.maps))
		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	return res.maps, nil
}

func FindEmptyBossHallLine(bossId values.Integer) (*FindEmptyBossHallLineArgs, *errmsg.ErrMsg) {
	req := &FindEmptyBossHallLineArgs{BossId: bossId}
	res, err := eventloop.CallChanEventLoop(gEl, req, func(r *FindEmptyBossHallLineArgs) (*FindEmptyBossHallLineArgs, *errmsg.ErrMsg) {

		lines, ok := gBossHall.adds[r.BossId]
		if !ok {
			return r, nil
		}
		for lineId, battleId := range lines {
			bInfo, ok2 := gBattleHash.Get(battleId)
			if !ok2 || len(bInfo.Roles) > 0 {
				continue
			}
			if r.LineId == 0 || r.LineId < lineId {
				r.LineId = lineId
				r.BattleId = battleId
				r.Session = bInfo.Session
			}
		}
		if r.LineId > 0 {
			gLineConfig.deleteBattles[r.BattleId] = r.LineId
		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CheckRoleValid() {
	req := &CheckRoleValidArgs{}
	_, _ = eventloop.CallChanEventLoop(gEl, req, func(r *CheckRoleValidArgs) (*CheckRoleValidArgs, *errmsg.ErrMsg) {
		now := time.Now().UnixMilli()
		for i := 0; i < RoleShardNum; i++ {
			for uRoleId, v := range gRoleHash.all[i] {
				if v.Status == 2 && now-v.AddTime >= 300000 {
					gRoleHash.log.Info("delete user status2", zap.Uint64("uRoleId", uRoleId), zap.Int64("now", now), zap.Any("info", v))
					if _, ok := gBattleHash.all[v.BattleServerId]; ok {
						delete(gBattleHash.all[v.BattleServerId].Roles, uRoleId)
					}
					gRoleHash.ClearRoleBattle(uRoleId)
				}
			}
		}
		return nil, nil
	})
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
