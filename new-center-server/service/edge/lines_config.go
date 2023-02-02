package edge

import (
	"coin-server/common/consulkv"
	"coin-server/common/errmsg"
	"coin-server/common/idgenerate"
	"coin-server/common/logger"
	modelspb "coin-server/common/proto/models"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/env"
	"coin-server/rule"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

//挂机地图分线逻辑

var MaxLineBattleId values.Integer
var CreateLineLock sync.Mutex

//挂机地图初始分线配置
type BattleLineConfig struct {
	Name              string                            `json:"name"`
	IcRoomId          values.Integer                    `json:"ic_room_id"`                 // ic报警的群id
	IcToken           string                            `json:"ic_token"`                   // ic报警token
	OpenMonitor       values.Integer                    `json:"open_monitor"`               //是否开启监控，当分线异常或到达预警值时发报警信息
	Threshold         values.Integer                    `json:"monitor_threshold"`          // 地图总人数到达80%开启新分线
	BossHallThreshold values.Integer                    `json:"monitor_bossHall_threshold"` // 每个boss总人数到达80%开启新分线
	Lines             map[values.Integer]values.Integer `json:"hangup_lines"`               // mapConfigId：lineNum
	BossHall          map[values.Integer]values.Integer `json:"boss_hall"`                  // mapConfigId：lineNum
	StaticWeight      values.Integer                    `json:"static_weight"`              //静态战斗权重
	DynamicWeight     values.Integer                    `json:"dynamic_weight"`             //动态战斗权重
}

type ServerLineConfig struct {
	srvLines      BattleLineConfig
	lineBattles   map[values.Integer]map[values.Integer]values.Integer //mapConfigId: lineId:battleId
	deleteBattles map[values.Integer]values.Integer
	adds          map[values.Integer]map[values.Integer]values.Integer //额外动态增加的分线 mapConfigId: lineId:battleId
	log           *logger.Logger
}

type Pos struct {
	X float32
	Y float32
}

var gLineConfig *ServerLineConfig //DeveloperServerId :xxxx  比如  稳定开发服Id:xxxx

func InitLineConfig(cnf *consulkv.Config, log *logger.Logger) {
	// 最大分线BattleId
	val, ok := idgenerate.GetKeyInitValue(idgenerate.CenterBattleId)
	if !ok {
		panic("can not find CenterBattleId")
	}
	MaxLineBattleId = val - 1
	gLineConfig = &ServerLineConfig{
		lineBattles:   map[values.Integer]map[values.Integer]values.Integer{},
		deleteBattles: map[values.Integer]values.Integer{},
		adds:          map[values.Integer]map[values.Integer]values.Integer{},
		log:           log,
	}
	tmpConfigs := map[values.Integer]BattleLineConfig{}
	utils.Must(cnf.Unmarshal("battle-lines", &tmpConfigs))
	if len(tmpConfigs) == 0 {
		panic("init battle-lines config fail")
	}
	serverId := env.GetServerId()
	srvLines, ok1 := tmpConfigs[serverId]
	if !ok1 {
		panic(fmt.Sprintf("serverId:%d config not exist", serverId))
	}
	if srvLines.StaticWeight <= 0 {
		srvLines.StaticWeight = 10
	}
	if srvLines.DynamicWeight <= 0 {
		srvLines.DynamicWeight = 2
	}
	gLineConfig.srvLines = srvLines
	gLineConfig.checkLineConfig()
	log.Info("load consul ok", zap.Int64("serverId", serverId), zap.Any("config", srvLines))
}

func (s *ServerLineConfig) checkLineConfig() {
	//检查battleServerId是否唯一
	serverId := env.GetServerId()
	for mapId, lineNum := range s.srvLines.Lines {
		r := rule.MustGetReader(nil)
		mapCnf, mapOk := r.MapScene.GetMapSceneById(mapId)
		if !mapOk || mapCnf.MapType != values.Integer(modelspb.BattleType_HangUp) {
			panic(fmt.Sprintf("invalid hungUp map id! %d %d", serverId, mapId))
		}
		if mapCnf.LineMaxPersonsNum <= 0 {
			panic(fmt.Sprintf("invalid hangup map person num! %d", mapId))
		}
		if lineNum < 1 {
			panic(fmt.Sprintf("invalid line num! serverId%d mapId:%d %d %d", serverId, mapId, lineNum))
		}
	}

	for bossId, lineNum := range s.srvLines.BossHall {
		r := rule.MustGetReader(nil)
		cnf, ok := r.BossHall.GetBossHallById(bossId)
		if !ok {
			panic(fmt.Sprintf("invalid bossHall id! %d %d", serverId, bossId))
		}
		mapCnf, mapOk := r.MapScene.GetMapSceneById(cnf.BossMap)
		if !mapOk || mapCnf.MapType != values.Integer(modelspb.BattleType_BossHall) {
			panic(fmt.Sprintf("invalid bossHall map id! %d %d %d", bossId, cnf.BossMap, mapCnf.MapType))
		}
		if mapCnf.LineMaxPersonsNum <= 0 {
			panic(fmt.Sprintf("invalid bossHall map person num! %d %d", bossId, cnf.BossMap))
		}
		if lineNum < 1 {
			panic(fmt.Sprintf("invalid bossHall lineNum!%d %d %d", serverId, cnf.BossMap, lineNum))
		}
	}
}

func (s *ServerLineConfig) addNewMapLine(mapId values.Integer, lineId values.Integer, battleId values.Integer) *errmsg.ErrMsg {
	r := rule.MustGetReader(nil)
	mapCnf, mapOk := r.MapScene.GetMapSceneById(mapId)
	if !mapOk || (mapCnf.MapType != values.Integer(modelspb.BattleType_HangUp)) {
		return errmsg.NewInternalErr(fmt.Sprintf("mapId:%d not hangup or bosshall", mapId))
	}

	maxLineNum, ok := s.getLineConfig(mapId)
	if !ok {
		return errmsg.NewInternalErr(fmt.Sprintf("mapId:%d not hangup", mapId))
	}
	if _, ok1 := s.adds[mapId]; !ok1 {
		s.adds[mapId] = map[values.Integer]values.Integer{}
	}
	if _, ok2 := s.lineBattles[mapId]; !ok2 {
		s.lineBattles[mapId] = map[values.Integer]values.Integer{}
	}
	if lineId > maxLineNum {
		s.adds[mapId][lineId] = battleId
	}
	s.lineBattles[mapId][lineId] = battleId
	return nil
}

func (s *ServerLineConfig) getLineConfig(mapId values.Integer) (values.Integer, bool) {
	lineNum, ok1 := s.srvLines.Lines[mapId]
	if ok1 {
		return lineNum, true
	}
	return 0, false
}

func (s *ServerLineConfig) getAllMapLines(mapId values.Integer) map[values.Integer]values.Integer {
	res := map[values.Integer]values.Integer{}
	mapLines, ok1 := gLineConfig.lineBattles[mapId]
	if ok1 {
		for lineId, battleSrvId := range mapLines {
			res[lineId] = battleSrvId
		}
	}
	return res
}

func GetServerName() string {
	return gLineConfig.srvLines.Name
}

func GetEdgeTypeWeight(typ modelspb.EdgeType) (values.Integer, *errmsg.ErrMsg) {
	if typ == modelspb.EdgeType_StaticServer {
		return gLineConfig.srvLines.StaticWeight, nil
	}
	return gLineConfig.srvLines.DynamicWeight, nil
}

func GetIcInfo() (values.Integer, string) {
	return gLineConfig.srvLines.IcRoomId, gLineConfig.srvLines.IcToken
}
