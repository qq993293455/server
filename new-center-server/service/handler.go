package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"coin-server/common/idgenerate"
	"coin-server/common/network/stdtcp"
	modelspb "coin-server/common/proto/models"
	"coin-server/common/service"
	"coin-server/common/values"
	env2 "coin-server/common/values/env"
	"coin-server/new-center-server/service/edge"
	icUtils "coin-server/pikaviewer/utils"
	"coin-server/rule"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	edgepb "coin-server/common/proto/edge"
	centerpb "coin-server/common/proto/newcenter"
	"coin-server/common/timer"
	"coin-server/common/utils"

	"go.uber.org/zap"
)

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("获取本机监听地址", this_.HandleSelfAddr)
	this_.svc.RegisterFunc("验证edgeserver", this_.HandleAuthEdge)
	this_.svc.RegisterEvent("同步Edge信息", this_.HandleSyncEdgeInfoPush)
	this_.svc.RegisterFunc("CurBattle", this_.HandleCurBattleInfo)
	this_.svc.RegisterFunc("工会boss信息", this_.HandleCurGuildBossInfo)
	this_.svc.RegisterFunc("工会boss在线人数", this_.HandleCurGuildBossCount)
	this_.svc.RegisterFunc("创建Roguelike副本", this_.HandleCreateRoguelike)
	this_.svc.RegisterFunc("获得所有挂机分线信息", this_.HandleGetAllMapLines)
	this_.svc.RegisterFunc("获得某个具体的分线信息", this_.HandleGetTargetLine)
	this_.svc.RegisterFunc("获得恶魔秘境Boss信息", this_.HandleGetBossHallInfo)
	this_.svc.RegisterEvent("恶魔秘境不可进入push", this_.HandleBossHallCanNotPush)

}

func (this_ *Service) HandleSelfAddr(c *ctx.Context, _ *centerpb.NewCenter_SelfAddrRequest) (*centerpb.NewCenter_SelfAddrResponse, *errmsg.ErrMsg) {
	return &centerpb.NewCenter_SelfAddrResponse{Addr: this_.selfAddr}, nil
}

func (this_ *Service) HandleAuthEdge(c *ctx.Context, req *centerpb.NewCenter_AuthEdgeRequest) (*centerpb.NewCenter_AuthEdgeResponse, *errmsg.ErrMsg) {
	now := timer.Now().Unix()
	if req.Token == "" || req.Now+5 < now {
		return nil, errmsg.NewErrInvalidRequestParam()
	}
	token := utils.MD5String("edge-center-" + strconv.Itoa(int(req.Now)))
	if token != req.Token {
		return nil, errmsg.NewErrInvalidRequestParam()
	}

	timer.NowString()
	token = utils.MD5String("center-edge-" + strconv.Itoa(int(now)))
	return &centerpb.NewCenter_AuthEdgeResponse{
		Token: token,
		Now:   now,
	}, nil
}

func (this_ *Service) HandleSyncEdgeInfoPush(c *ctx.Context, msg *centerpb.NewCenter_SyncEdgeInfoPush) {
	v := c.GetValue(service.TCPSessionKey)
	session, ok := v.(*stdtcp.Session)
	if !ok {
		return
	}
	killPlayers, err := edge.SyncEdge(session, msg)
	if killPlayers != nil && len(killPlayers.Players) > 0 {
		res := &edgepb.Edge_KillPlayerResponse{}
		this_.log.Info("start kill player", zap.String("addr", session.RemoteAddr()), zap.String("req", killPlayers.GoString()))
		err1 := session.RPCRequestOut(nil, killPlayers, res)
		if err1 != nil {
			this_.log.Warn("kill player fail", zap.String("msg", res.GoString()), zap.Error(err))
			return
		}
	}
	if err != nil {
		this_.log.Warn("", zap.Error(err))
		return
	}

	// 尝试启动默认分线服务
	//全量更新或者某个battle结束
	var flag = !msg.IsAll && len(msg.Edges) == 1 && msg.Edges[0].IsEnd
	if msg.IsAll || flag {
		if msg.Typ == modelspb.EdgeType_StaticServer || msg.Typ == modelspb.EdgeType_StatelessServer {
			this_.checkBattleLines(false)
			this_.checkBossHallLines(false)
		}
		if msg.Typ == modelspb.EdgeType_DynamicServer {
			for _, e := range msg.Edges {
				if e.BossHallId > 0 {
					this_.checkBossHallLines(false)
				}
			}
		}
	}

}

func (this_ *Service) HandleCurBattleInfo(c *ctx.Context, req *centerpb.NewCenter_CurBattleInfoRequest) (*centerpb.NewCenter_CurBattleInfoResponse, *errmsg.ErrMsg) {
	if req.HungUpMapId == 0 || req.MapId == 0 {
		return nil, errmsg.NewInternalErr("invalid args, mapId empty")
	}

	line, err := this_.getValidLine(c.RoleId, req.HungUpMapId, req.HungUpServerId)
	if err != nil {
		return nil, err
	}

	res := &centerpb.NewCenter_CurBattleInfoResponse{
		HungUpServerId: line.BattleServerId,
		HungUpMapId:    req.HungUpMapId,
		LineInfo:       line,
	}
	if req.BattleServerId > 0 {
		r := rule.MustGetReader(nil)
		cnf, ok := r.MapScene.GetMapSceneById(req.MapId)
		if ok && cnf.MapType == values.Integer(modelspb.BattleType_HangUp) {
			res.BattleServerId = res.HungUpServerId
			res.MapId = res.HungUpMapId
		} else {
			battleId, err1 := edge.GetTargetBattleServer(req.BattleServerId)
			if err1 != nil {
				return nil, err1
			}
			if battleId == req.BattleServerId {
				res.MapId = req.MapId
				res.BattleServerId = battleId
			} else {
				res.BattleServerId = res.HungUpServerId
				res.MapId = res.HungUpMapId
			}
		}
	}
	if req.BattleServerId == 0 {
		res.BattleServerId = res.HungUpServerId
		res.MapId = res.HungUpMapId
	}
	return res, nil
}

func (this_ *Service) HandleCurGuildBossInfo(c *ctx.Context, req *centerpb.NewCenter_CurGuildBossInfoRequest) (*centerpb.NewCenter_CurGuildBossInfoResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(nil)
	cnf, ok := r.MapScene.GetMapSceneById(req.MapId)
	if !ok || cnf.MapType != values.Integer(modelspb.BattleType_UnionBoss) {
		return nil, errmsg.NewInternalErr("invalid guild boss map id")
	}
	battleId, exist, err := edge.GetGuildBossBattle(req.UnionBossId)
	if err != nil {
		return nil, err
	}
	this_.log.Info("cur guild boss battle", zap.Int64("battleId", battleId), zap.Bool("exist", exist))
	isNew := false
	if !exist {
		edgeReq := &edgepb.Edge_CreateServerRequest{
			MapSceneId:  req.MapId,
			GuildId:     req.UnionId,
			GuildDayId:  req.UnionBossId,
			TotalDamage: req.TotalDamages,
			Damages:     req.Damages,
			Typ:         modelspb.EdgeType_DynamicServer,
		}
		res, err1 := this_.createBattle(edgeReq)
		if err1 != nil {
			return nil, err1
		}
		battleId = res.BattleServerId
		isNew = true
	}
	res := &centerpb.NewCenter_CurGuildBossInfoResponse{
		BattleServerId: battleId,
		IsNew:          isNew,
	}
	return res, nil
}

func (this_ *Service) HandleCurGuildBossCount(c *ctx.Context, req *centerpb.NewCenter_UnionBossOnlineCountRequest) (*centerpb.NewCenter_UnionBossOnlineCountResponse, *errmsg.ErrMsg) {
	count, err := edge.GetGuildBossCount(req.UnionBossId)
	if err != nil {
		return nil, err
	}
	return &centerpb.NewCenter_UnionBossOnlineCountResponse{Count: count}, nil
}

func (this_ *Service) HandleCreateRoguelike(c *ctx.Context, req *centerpb.NewCenter_CreateRoguelikeRequest) (*centerpb.NewCenter_CreateRoguelikeResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(nil)
	cnf, ok := r.MapScene.GetMapSceneById(req.MapId)
	if !ok || cnf.MapType != values.Integer(modelspb.BattleType_Roguelike) {
		return nil, errmsg.NewInternalErr("invalid rogue like map id")
	}
	edgeReq := &edgepb.Edge_CreateServerRequest{
		MapSceneId:     req.MapId,
		RoomId:         req.RoomId,
		Bots:           req.Bots,
		MonsterEffects: req.MonsterEffects,
		BossEffects:    req.BossEffects,
		MatchServerId:  c.ServerId,
		Typ:            modelspb.EdgeType_DynamicServer,
		CardMap:        req.CardMap,
	}
	res, err := this_.createBattle(edgeReq)
	if err != nil {
		return nil, err
	}

	return &centerpb.NewCenter_CreateRoguelikeResponse{
		MapId:    res.MapId,
		RoomId:   req.RoomId,
		BattleId: res.BattleServerId,
	}, nil
}

func (this_ *Service) HandleGetAllMapLines(c *ctx.Context, req *centerpb.NewCenter_GetMapAllLinesRequest) (*centerpb.NewCenter_GetMapAllLinesResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(nil)
	cnf, ok := r.MapScene.GetMapSceneById(req.MapId)
	if !ok || cnf.MapType != values.Integer(modelspb.BattleType_HangUp) {
		return nil, errmsg.NewInternalErr("invalid HangUp map id")
	}
	allLines, err := edge.GetAllMapLines(req.MapId)
	if err != nil {
		return nil, err
	}
	res := &centerpb.NewCenter_GetMapAllLinesResponse{
		AllLineInfo: &modelspb.AllLineInfo{
			MapId:    req.MapId,
			MaxNum:   cnf.LineMaxPersonsNum,
			AllLines: allLines,
		},
	}
	return res, err
}

func (this_ *Service) HandleGetTargetLine(c *ctx.Context, req *centerpb.NewCenter_GetTargetLineRequest) (*centerpb.NewCenter_GetTargetLineResponse, *errmsg.ErrMsg) {
	if req.MapId == 0 {
		return nil, errmsg.NewInternalErr("invalid args, mapId empty")
	}
	r := rule.MustGetReader(nil)
	cnf, ok := r.MapScene.GetMapSceneById(req.MapId)
	if !ok || cnf.MapType != values.Integer(modelspb.BattleType_HangUp) {
		return nil, errmsg.NewInternalErr("invalid HangUp map id")
	}

	line1, err1 := this_.getValidLine(c.RoleId, req.MapId, req.BattleServerId)
	if err1 != nil {
		return nil, err1
	}
	res := &centerpb.NewCenter_GetTargetLineResponse{LineInfo: line1}
	//角色是否已经在分线内
	if req.BattleServerId > 0 {
		roleInfo, err2 := edge.GetRoleDetails(c.RoleId)
		if err2 != nil {
			return nil, err2
		}
		this_.log.Debug("roleInfo", zap.Any("info", roleInfo), zap.Any("line1", line1))
		if roleInfo != nil && roleInfo.HungUpMapId == req.MapId &&
			roleInfo.HungUpServerId == req.BattleServerId {
			res.RoleInLine = true
		}
	}
	return res, nil
}

func (this_ *Service) HandleGetBossHallInfo(c *ctx.Context, req *centerpb.NewCenter_CurBossHallRequest) (*centerpb.NewCenter_CurBossHallResponse, *errmsg.ErrMsg) {
	r := rule.MustGetReader(nil)
	cnf, ok := r.BossHall.GetBossHallById(req.BossId)
	if !ok {
		return nil, errmsg.NewInternalErr("invalid boss id")
	}
	cnf2, ok2 := r.MapScene.GetMapSceneById(cnf.BossMap)
	if !ok2 {
		return nil, errmsg.NewInternalErr("invalid map scene id")
	}
	battleSrvId, err := edge.GetBossHallInfo(c.RoleId, req.BossId, cnf2.LineMaxPersonsNum)
	if err != nil {
		return nil, err
	}
	res := &centerpb.NewCenter_CurBossHallResponse{MapScenceId: cnf.BossMap, BattleServerId: battleSrvId}
	return res, nil
}

func (this_ *Service) HandleBossHallCanNotPush(c *ctx.Context, msg *centerpb.NewCenter_NotCanEnterPush) {
	this_.log.Debug("NotCanEnterPush", zap.Any("msg", msg))
	if msg.BossId <= 0 || msg.BattleId <= 0 {
		return
	}
	err := edge.SetBossHallCanNotEnter(msg)
	if err != nil {
		this_.log.Error("error", zap.Error(err))
		return
	}
	//this_.checkBossHallLines(false)
}

func (this_ *Service) getValidLine(roleId string, mapId values.Integer, battlesServerId values.Integer) (*modelspb.LineInfo, *errmsg.ErrMsg) {
	r := rule.MustGetReader(nil)
	cnf, ok := r.MapScene.GetMapSceneById(mapId)
	if !ok || cnf.MapType != values.Integer(modelspb.BattleType_HangUp) {
		return nil, errmsg.NewInternalErr("invalid HangUp map id")
	}
	line, scaleDeleted, err := edge.GetValidLine(roleId, mapId, battlesServerId, cnf.LineMaxPersonsNum)
	if err != nil {
		return nil, err
	}
	this_.log.Debug("getValidLine", zap.Int64("serverId", env2.GetServerId()), zap.Int64("mapId", mapId),
		zap.Uint64("uRoleId", utils.Base34DecodeString(roleId)), zap.Int64("battleId", battlesServerId),
		zap.Bool("scaleDeleted", scaleDeleted), zap.Any("line", line))
	_ = scaleDeleted
	return line, nil
}

//尝试创建战斗分线
func (this_ *Service) checkBattleLines(flag bool) {
	if !flag && this_.initLinesFlag != 2 {
		this_.log.Debug("checkBattleLines ignore!wait start time", zap.Bool("flag", flag), zap.Int("initLinesFlag", this_.initLinesFlag))
		return
	}
	this_.log.Debug("checkBattleLines enter", zap.Bool("flag", flag), zap.Int("initLinesFlag", this_.initLinesFlag))

	edge.CreateLineLock.Lock()
	defer edge.CreateLineLock.Unlock()

	defaultLines, addLines, err := edge.GetNeedInitBattleLines()
	if err != nil {
		this_.log.Warn("GetNeedInitBattleLines error!", zap.Error(err))
		return
	}
	if len(defaultLines) == 0 && len(addLines) == 0 {
		this_.initLinesFlag = 2
		this_.log.Debug("no need init battle lines")
		return
	}
	this_.log.Info("NeedInitBattleLines", zap.Any("defaultLines", defaultLines), zap.Any("addLines", addLines))
	for mapId, dLines := range defaultLines {
		for lineId, _ := range dLines {
			newBattleSrvId, errGen := idgenerate.GenerateID(context.Background(), idgenerate.CenterBattleId)
			if errGen != nil {
				this_.log.Warn("gen battle server id err", zap.Error(errGen))
				return
			}
			edgeReq := &edgepb.Edge_CreateServerRequest{
				BattleId:   newBattleSrvId,
				MapSceneId: mapId,
				Typ:        modelspb.EdgeType_StaticServer,
				LineId:     lineId,
			}
			edgeRes, err2 := this_.createBattle(edgeReq)
			if err2 != nil {
				this_.log.Warn("create defaultLine fail", zap.Int64("lineId", lineId),
					zap.String("edgeReq", edgeReq.GoString()), zap.Error(err2))
				return
			}
			err3 := edge.AddNewBattleLines(mapId, lineId, newBattleSrvId)
			if err3 != nil {
				this_.log.Warn("create new line ok,but set fail", zap.Int64("mapId", mapId),
					zap.Any("lineId", lineId), zap.Any("battleId", newBattleSrvId), zap.Error(err3))
				return
			}
			this_.log.Debug("create defaultLine ok", zap.Int64("lineId", lineId),
				zap.String("edgeReq", edgeReq.GoString()), zap.String("edgeRes", edgeRes.GoString()))
		}

	}

	for mapId, aLines := range addLines {
		for lineId, _ := range aLines {
			newBattleSrvId, errGen := idgenerate.GenerateID(context.Background(), idgenerate.CenterBattleId)
			if errGen != nil {
				this_.log.Warn("gen battle server id err", zap.Error(err))
				return
			}

			edgeReq := &edgepb.Edge_CreateServerRequest{
				BattleId:   newBattleSrvId,
				MapSceneId: mapId,
				Typ:        modelspb.EdgeType_StaticServer,
				LineId:     lineId,
			}
			edgeRes, err2 := this_.createBattle(edgeReq)
			if err2 != nil {
				this_.log.Warn("create addLine fail", zap.Int64("lineId", lineId),
					zap.String("edgeReq", edgeReq.GoString()), zap.Error(err2))
				return
			}
			err3 := edge.AddNewBattleLines(mapId, lineId, newBattleSrvId)
			if err3 != nil {
				this_.log.Warn("create new line ok,but set fail", zap.Int64("mapId", mapId),
					zap.Any("lineId", lineId), zap.Any("battleId", newBattleSrvId), zap.Error(err3))
				return
			}
			this_.log.Debug("create addLine ok", zap.Int64("lineId", lineId),
				zap.String("edgeReq", edgeReq.GoString()), zap.String("edgeRes", edgeRes.GoString()))
		}
	}

	if flag {
		this_.initLinesFlag = 2
	}

}

func (this_ *Service) checkLineValid() {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if this_.initLinesFlag != 2 {
			continue
		}
		this_.checkCreateNewLine()
		this_.checkDeleteEmptyLine()
	}

}

func (this_ *Service) checkBossHallLineValid() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if this_.initLinesFlag != 2 {
			continue
		}
		this_.checkCreateNewBossHallLine()
		this_.checkDeleteEmptyBossHallLine()
	}

}

//尝试开启新分线
func (this_ *Service) checkCreateNewLine() {
	edge.CreateLineLock.Lock()
	defer edge.CreateLineLock.Unlock()

	maps, err := edge.CheckCreateBattleLines()
	if err != nil {
		this_.log.Warn("battle line num check err", zap.Error(err))
		return
	}
	this_.log.Debug("checkCreateNewLine", zap.Any("maps", maps))

	if len(maps) == 0 {
		return
	}
	//this_.log.Debug("checkCreateNewLine", zap.Any("maps", maps))

	for mapId, val := range maps {

		str := fmt.Sprintf("警告！！！ 地图Id:%d %s 当前分线人数: %d/%d 已达阈值，请尽快开启新分线!", mapId, val.Name, val.CurNum, val.MaxNum)
		content := map[string]string{"content": str, "at_user": "all"}
		if err1 := send2IC(edge.GetServerName(), content); err1 != nil {
			this_.log.Warn("send2IC fail", zap.Error(err))
		}

		newBattleSrvId, errGen := idgenerate.GenerateID(context.Background(), idgenerate.CenterBattleId)
		if errGen != nil {
			this_.log.Warn("gen battle server id err", zap.Error(err))
			return
		}

		//尝试自动开启新的分线
		edgeReq := &edgepb.Edge_CreateServerRequest{
			BattleId:   newBattleSrvId,
			MapSceneId: mapId,
			LineId:     val.NewLineId,
			Typ:        modelspb.EdgeType_StaticServer,
		}
		edgeRes, err2 := this_.createBattle(edgeReq)
		if err2 != nil {
			this_.log.Warn("create new line fail", zap.Int64("mapId", mapId), zap.Any("val", val), zap.Error(err2))
			return
		}
		err3 := edge.AddNewBattleLines(mapId, val.NewLineId, newBattleSrvId)
		if err3 != nil {
			this_.log.Warn("create new line ok,but set fail", zap.Int64("mapId", mapId), zap.Any("val", val), zap.Error(err3))
			return
		}

		str1 := fmt.Sprintf("警告！！！ 地图Id:%d %s 已自动开启新分线:%d", mapId, val.Name, val.NewLineId)
		content1 := map[string]string{"content": str1, "at_user": "all"}
		if err1 := send2IC("", content1); err1 != nil {
			this_.log.Warn("send2IC fail", zap.Error(err))
		}

		this_.log.Info("create new line ok", zap.Int64("mapId", mapId), zap.Any("val", val),
			zap.String("edgeRes", edgeRes.GoString()))
	}

}

//尝试删除空闲分线
func (this_ *Service) checkDeleteEmptyLine() {
	edge.CreateLineLock.Lock()
	defer edge.CreateLineLock.Unlock()
	maps, err := edge.CheckLinesFree()
	if err != nil || maps == nil {
		return
	}
	this_.log.Debug("checkDeleteEmptyLine", zap.Any("maps", maps), zap.Any("freeLinesMap", this_.freeLinesMap))

	//原有的空闲地图现在已经不再空闲
	for mapId, _ := range this_.freeLinesMap {
		if _, ok := maps[mapId]; !ok {
			delete(this_.freeLinesMap, mapId)
		}
	}

	//加入新的空闲地图
	for mapId, _ := range maps {
		if _, ok := this_.freeLinesMap[mapId]; !ok {
			this_.freeLinesMap[mapId] = time.Now().Unix()
		}
	}

	//空闲十分钟以上的地图准备删除
	now := time.Now().Unix()
	for mapId, timeVal := range this_.freeLinesMap {
		if now-timeVal >= 600 {
			res, err1 := edge.FindEmptyLine(mapId)
			if err1 != nil {
				this_.log.Warn("FindEmptyLine err!", zap.Any("mapId", mapId), zap.Error(err1))
				continue
			}
			if res.Session == nil {
				this_.log.Warn("FindEmptyLine session nil!", zap.Any("mapId", mapId), zap.Any("res", res))
				continue
			}
			killRes := &edgepb.Edge_KillBattleResponse{}
			err2 := res.Session.RPCRequestOut(nil, &edgepb.Edge_KillBattleRequest{BattleId: res.BattleId}, killRes)
			if err2 != nil {
				this_.log.Warn("kill battle fail!", zap.Any("mapId", mapId), zap.Any("res", res), zap.Error(err2))
				delete(this_.freeLinesMap, mapId)
				continue
			}
			delete(this_.freeLinesMap, mapId)

			r := rule.MustGetReader(nil)
			cnf, cnfOk := r.MapScene.GetMapSceneById(mapId)
			if cnfOk {
				str1 := fmt.Sprintf("警告！！！ 地图Id:%d %s 分线Id:%d 空闲超过10分钟自动关闭", mapId, cnf.Title, res.LineId)
				content1 := map[string]string{"content": str1, "at_user": "all"}
				if err3 := send2IC("", content1); err3 != nil {
					this_.log.Warn("send2IC fail", zap.Error(err3))
				}
			}

			this_.log.Info("kill battle ok!", zap.Any("mapId", mapId), zap.Any("res", res), zap.String("addr", res.Session.RemoteAddr()))
		}

	}

}

//尝试开启新恶魔秘境分线
func (this_ *Service) checkCreateNewBossHallLine() {
	edge.CreateLineLock.Lock()
	defer edge.CreateLineLock.Unlock()

	maps, err := edge.CheckCreateBossHallLines()
	if err != nil {
		this_.log.Warn("battle line num check err", zap.Error(err))
		return
	}
	if len(maps) == 0 {
		return
	}
	this_.log.Info("checkCreateNewBossHallLine", zap.Any("maps", maps))

	for bossId, lineId := range maps {

		r := rule.MustGetReader(nil)
		cnf, ok := r.BossHall.GetBossHallById(bossId)
		if !ok {
			panic(fmt.Sprintf("invalid bossHall id %d", bossId))
		}

		newBattleSrvId, errGen := idgenerate.GenerateID(context.Background(), idgenerate.CenterBattleId)
		if errGen != nil {
			this_.log.Warn("gen battle server id err", zap.Error(err))
			return
		}

		//尝试自动开启新的分线
		edgeReq := &edgepb.Edge_CreateServerRequest{
			BattleId:   newBattleSrvId,
			MapSceneId: cnf.BossMap,
			BossId:     bossId,
			LineId:     lineId,
			Typ:        modelspb.EdgeType_DynamicServer,
		}
		edgeRes, err2 := this_.createBattle(edgeReq)
		if err2 != nil {
			this_.log.Warn("create new line fail", zap.Any("edgeReq", edgeReq), zap.Error(err2))
			return
		}

		str1 := fmt.Sprintf("警告！！！ %s 恶魔秘境BossId:%d  已自动开启新分线:%d", edge.GetServerName(), edgeReq.BossId, edgeReq.LineId)
		content1 := map[string]string{"content": str1, "at_user": "all"}
		if err3 := send2IC("", content1); err3 != nil {
			this_.log.Warn("send2IC fail", zap.Error(err3))
		}

		this_.log.Info("create new bossHall line ok", zap.Any("edgeReq", edgeReq), zap.String("edgeRes", edgeRes.GoString()))
	}

}

//尝试删除空闲恶魔秘境分线
func (this_ *Service) checkDeleteEmptyBossHallLine() {
	edge.CreateLineLock.Lock()
	defer edge.CreateLineLock.Unlock()
	bossIds, err := edge.CheckBossHallLinesFree()
	if err != nil || bossIds == nil {
		return
	}
	this_.log.Info("checkDeleteEmptyBossLine", zap.Any("bossIds", bossIds), zap.Any("freeLinesMap", this_.freeLinesMap))

	//原有的空闲地图现在已经不再空闲
	for bossId, _ := range this_.freeBossHallLines {
		if _, ok := bossIds[bossId]; !ok {
			delete(this_.freeBossHallLines, bossId)
		}
	}

	//加入新的空闲地图
	for bossId, _ := range bossIds {
		if _, ok := this_.freeBossHallLines[bossId]; !ok {
			this_.freeBossHallLines[bossId] = time.Now().Unix()
		}
	}

	//空闲十分钟以上的地图准备删除
	now := time.Now().Unix()
	for bossId, timeVal := range this_.freeBossHallLines {
		if now-timeVal >= 600 {
			res, err1 := edge.FindEmptyBossHallLine(bossId)
			if err1 != nil {
				this_.log.Warn("FindEmptyLine err!", zap.Any("bossId", bossId), zap.Error(err1))
				continue
			}
			if res.Session == nil {
				this_.log.Warn("FindEmptyLine session nil!", zap.Any("bossId", bossId), zap.Any("res", res))
				continue
			}
			killRes := &edgepb.Edge_KillBattleResponse{}
			err2 := res.Session.RPCRequestOut(nil, &edgepb.Edge_KillBattleRequest{BattleId: res.BattleId}, killRes)
			if err2 != nil {
				this_.log.Warn("kill battle fail!", zap.Any("bossId", bossId), zap.Any("res", res), zap.Error(err2))
				delete(this_.freeBossHallLines, bossId)
				continue
			}
			delete(this_.freeBossHallLines, bossId)

			str1 := fmt.Sprintf("警告！！！ %s 恶魔秘境BossId:%d 分线Id:%d 空闲超过10分钟自动关闭", edge.GetServerName(), bossId, res.LineId)
			content1 := map[string]string{"content": str1, "at_user": "all"}
			if err1 := send2IC("", content1); err1 != nil {
				this_.log.Warn("send2IC fail", zap.Error(err))
			}

			this_.log.Info("kill bossHall battle ok!", zap.Any("bossId", bossId), zap.Any("res", res), zap.String("addr", res.Session.RemoteAddr()))
		}

	}

}

func (this_ *Service) checkRoleValid() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if this_.initLinesFlag != 2 {
			continue
		}
		edge.CheckRoleValid()
	}
}

func (this_ *Service) checkBossHallLines(flag bool) {
	if !flag && this_.initBossHallFlag != 2 {
		this_.log.Debug("checkBossHallLines ignore!wait start time", zap.Bool("flag", flag), zap.Int("initBossHallFlag", this_.initBossHallFlag))
		return
	}

	edge.CreateLineLock.Lock()
	defer edge.CreateLineLock.Unlock()

	addLines, err := edge.CheckBossHallNeedInit()
	if err != nil {
		this_.log.Warn("get fail", zap.Error(err))
		return
	}
	if len(addLines) == 0 {
		this_.log.Debug("no need init boss hall")
		return
	}
	this_.log.Info("needInitBossHallLines", zap.Any("lines", addLines))
	for bossId, lines := range addLines {
		for lineId, _ := range lines {
			battleServerId, errGen := idgenerate.GenerateID(context.Background(), idgenerate.CenterBattleId)
			if errGen != nil {
				this_.log.Warn("GenerateID fail", zap.Error(errGen))
				continue
			}
			r := rule.MustGetReader(nil)
			cnf, ok := r.BossHall.GetBossHallById(bossId)
			if !ok {
				panic(fmt.Sprintf("invalid bossHall id %d", bossId))
			}
			edgeReq := &edgepb.Edge_CreateServerRequest{
				BattleId:   battleServerId,
				MapSceneId: cnf.BossMap,
				BossId:     bossId,
				LineId:     lineId,
				Typ:        modelspb.EdgeType_DynamicServer,
			}
			res, err1 := this_.createBattle(edgeReq)
			if err1 != nil {
				this_.log.Warn("create bossHall fail", zap.Any("edgeReq", edgeReq), zap.Error(err1))
				continue
			}
			this_.log.Debug("create bossHall ok", zap.Any("edgeReq", edgeReq), zap.Any("res", res))
		}
	}

	if flag {
		this_.initBossHallFlag = 2
	}

}

func (this_ *Service) createBattle(edgeReq *edgepb.Edge_CreateServerRequest) (*centerpb.NewCenter_CurBattleInfoResponse, *errmsg.ErrMsg) {
	if edgeReq.Typ == modelspb.EdgeType_StatelessServer {
		return nil, errmsg.NewInternalErr("not set EdgeType!")
	}
	if edgeReq.BattleId == 0 {
		id, err := idgenerate.GenerateID(context.Background(), idgenerate.CenterBattleId)
		if err != nil {
			return nil, err
		}
		edgeReq.BattleId = id
	}
	edgeType := edgeReq.Typ

	if edgeReq.CpuWeight == 0 {
		//根据mapId找对应的权重
		weight, err := edge.GetEdgeTypeWeight(edgeType)
		if err != nil {
			return nil, err
		}
		edgeReq.CpuWeight = weight
	}
	weight := edgeReq.CpuWeight
	if weight == 0 {
		return nil, errmsg.NewInternalErr("create battle fail!battle weight == 0")
	}

	//找一个可用的edge
	session, err := edge.FindOneEdge(edgeType, weight)
	if err != nil {
		return nil, err
	}
	//没找到，尝试找一个空的edge
	if session == nil {
		session, err = edge.FindOneEdge(modelspb.EdgeType_StatelessServer, 0)
		if err != nil {
			return nil, err
		}
		if session == nil {
			return nil, errmsg.NewInternalErr("can not find valid edge session")
		}
		// 通知edge设置类型
		setRes := &edgepb.Edge_SetEdgeTypeResponse{}
		err1 := session.RPCRequestOut(nil, &edgepb.Edge_SetEdgeTypeRequest{Typ: edgeType}, setRes)
		if err1 != nil {
			return nil, err1
		}
		err2 := edge.SetEdgeType(session, edgeType)
		if err2 != nil {
			return nil, err2
		}
	}

	edgeRes := &edgepb.Edge_CreateServerResponse{}
	this_.log.Info("start edge create rpc", zap.String("addr", session.RemoteAddr()), zap.String("req", edgeReq.GoString()))
	err1 := session.RPCRequestOut(nil, edgeReq, edgeRes)
	if err1 != nil {
		return nil, err1
	}
	res := &centerpb.NewCenter_CurBattleInfoResponse{
		BattleServerId: edgeRes.BattleId,
		MapId:          edgeRes.MapSceneId,
	}
	return res, nil
}

func send2IC(tag string, params map[string]string) error {
	return nil
	roomId, token := edge.GetIcInfo()
	if roomId == 0 || token == "" {
		return nil
	}
	data := map[string]string{
		"token":        token,
		"target":       "group",
		"room":         strconv.FormatInt(roomId, 10),
		"title":        "分线服务报警 " + tag,
		"content_type": "1",
	}
	for k, v := range params {
		data[k] = v
	}
	_, err := icUtils.NewRequest("http://im-api.skyunion.net/msg").Post(data)
	return err
}
