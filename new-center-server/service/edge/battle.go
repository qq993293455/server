package edge

import (
	"coin-server/common/logger"
	"coin-server/common/network/stdtcp"
	centerpb "coin-server/common/proto/newcenter"
	utils2 "coin-server/common/utils"
	"coin-server/common/values"
	"go.uber.org/zap"
)

type BattleInfo struct {
	BattleServerId values.Integer
	MapId          values.Integer
	LineId         values.Integer
	Roles          map[URoleID]values.Integer
	Session        *stdtcp.Session
}

type BattleHash struct {
	all map[values.Integer]BattleInfo
	log *logger.Logger
}

var gBattleHash *BattleHash

func initBattleHash(log *logger.Logger) {
	gBattleHash = &BattleHash{log: log}
	gBattleHash.all = make(map[values.Integer]BattleInfo, 1024)
}

func (s *BattleHash) Sync(session *stdtcp.Session, msg *centerpb.NewCenter_EdgeInfo, isAll bool) {
	battleSrvId := msg.BattleId
	curVal, ok := s.all[battleSrvId]
	s.log.Debug("enter battleSync", zap.Any("ok", ok), zap.Any("isAll", isAll), zap.Any("msg.IsAdd", msg.IsAdd))
	// 增量删除
	if !isAll && !msg.IsAdd && ok {
		for _, role := range msg.Roles {
			uRoleId := utils2.Base34DecodeString(role.RoleId)
			delete(curVal.Roles, uRoleId)
		}
		s.log.Debug("battleSync del", zap.Int64("battleSrvId", battleSrvId), zap.Int("roleNum", len(curVal.Roles)))
		return
	}

	// 增量或者全量添加
	if !ok && (isAll || msg.IsAdd) {
		delete(gLineConfig.deleteBattles, battleSrvId)
		info := BattleInfo{
			BattleServerId: battleSrvId,
			MapId:          msg.MapId,
			Roles:          make(map[URoleID]values.Integer, len(msg.Roles)),
			LineId:         msg.LineId,
			Session:        session,
		}
		if len(msg.Roles) > 0 {
			for _, role := range msg.Roles {
				roleId := utils2.Base34DecodeString(role.RoleId)
				info.Roles[roleId] = role.AddTime
			}
		}
		s.all[battleSrvId] = info
		_ = gLineConfig.addNewMapLine(info.MapId, info.LineId, info.BattleServerId)
		s.log.Debug("battleSync add1", zap.Int64("battleSrvId", battleSrvId), zap.Int("roleNum", len(info.Roles)))
		return
	}

	//增量更新
	if ok && msg.IsAdd {
		delete(gLineConfig.deleteBattles, battleSrvId)

		curVal.Session = session
		curVal.BattleServerId, curVal.MapId = msg.BattleId, msg.MapId
		for _, role := range msg.Roles {
			roleId := utils2.Base34DecodeString(role.RoleId)
			curVal.Roles[roleId] = role.AddTime
		}
		s.all[battleSrvId] = curVal
		_ = gLineConfig.addNewMapLine(curVal.MapId, curVal.LineId, curVal.BattleServerId)
		s.log.Debug("battleSync add2", zap.Int64("battleSrvId", battleSrvId), zap.Int("roleNum", len(curVal.Roles)))
	}

}

func (s *BattleHash) GetRoles(battleSrvId values.Integer) (map[URoleID]values.Integer, bool) {
	v, ok := s.all[battleSrvId]
	if ok {
		return v.Roles, true
	}
	return nil, false
}

func (s *BattleHash) GetRolesNum(battleSrvId values.Integer) (values.Integer, bool) {
	v, ok := s.all[battleSrvId]
	if ok {
		return values.Integer(len(v.Roles)), true
	}
	return 0, false
}

func (s *BattleHash) Get(battleSrvId values.Integer) (*BattleInfo, bool) {
	v, ok := s.all[battleSrvId]
	if ok {
		return &v, true
	}
	return nil, false
}

func (s *BattleHash) Remove(battleSrvId values.Integer) {
	delete(s.all, battleSrvId)
}

func (s *BattleHash) CheckRoleInBattle(roleId URoleID, battleSrvId values.Integer, mapId values.Integer) bool {
	v, ok := s.all[battleSrvId]
	if ok {
		if v.MapId != mapId {
			return false
		}
		_, ok2 := v.Roles[roleId]
		if ok2 {
			return true
		}
	}
	return false
}
