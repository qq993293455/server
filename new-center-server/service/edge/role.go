package edge

import (
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	centerpb "coin-server/common/proto/newcenter"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/rule"
	"go.uber.org/zap"
	"time"
)

type URoleID = uint64

type RoleInfo struct {
	BattleServerId values.Integer
	MapId          values.Integer
	HungUpServerId values.Integer
	HungUpMapId    values.Integer
	AddTime        values.Integer
	Status         values.Integer //  1 正常进入  2 curBattle占位
}

type RoleHash struct {
	all [RoleShardNum]map[URoleID]RoleInfo
	log *logger.Logger
}

const RoleShardNum int = 32

var gRoleHash *RoleHash

func initRoleHash(log *logger.Logger) {
	gRoleHash = &RoleHash{log: log}
	for i := 0; i < RoleShardNum; i++ {
		gRoleHash.all[i] = make(map[URoleID]RoleInfo, 1024)
	}
}

func (r *RoleHash) Sync(msg *centerpb.NewCenter_EdgeInfo, isAll bool) {
	if len(msg.Roles) == 0 {
		return
	}
	rl := rule.MustGetReader(nil)
	cnf, exist := rl.MapScene.GetMapSceneById(msg.MapId)
	if !exist {
		return
	}
	for _, role := range msg.Roles {
		uRoleId := utils.Base34DecodeString(role.RoleId)
		shard := int(uRoleId % URoleID(RoleShardNum))
		curVal, ok := r.all[shard][uRoleId]
		//增量删除
		if !msg.IsAdd && !isAll {
			if curVal.BattleServerId == msg.BattleId {
				delete(r.all[shard], uRoleId)
				r.log.Debug("del role", zap.Uint64("uRoleId", uRoleId), zap.Any("old", curVal))
			} else {
				r.log.Debug("del role ignore", zap.Uint64("uRoleId", uRoleId), zap.Any("old", curVal))
			}
			continue
		}

		//添加
		if !ok {
			info := RoleInfo{
				BattleServerId: msg.BattleId,
				MapId:          msg.MapId,
				AddTime:        role.AddTime,
				Status:         1,
			}
			if cnf.MapType == values.Integer(models.BattleType_HangUp) {
				info.HungUpServerId, info.HungUpMapId = msg.BattleId, msg.MapId
			}
			r.all[shard][uRoleId] = info
		} else {
			curVal.BattleServerId, curVal.MapId, curVal.Status = msg.BattleId, msg.MapId, 1
			if exist && cnf.MapType == values.Integer(models.BattleType_HangUp) {
				curVal.HungUpServerId, curVal.HungUpMapId = msg.BattleId, msg.MapId
			}
			r.all[shard][uRoleId] = curVal
		}
		r.log.Debug("add role", zap.Uint64("uRoleId", uRoleId))
	}
}

func (r *RoleHash) Get(uRoleId URoleID) (*RoleInfo, bool) {
	shard := int(uRoleId % URoleID(RoleShardNum))
	v, ok := r.all[shard][uRoleId]
	if ok {
		return &v, true
	}
	return nil, false
}

func (r *RoleHash) AddDefault(uRoleId URoleID, hungUpMapId values.Integer, hungUpBattleId values.Integer) {
	shard := int(uRoleId % URoleID(RoleShardNum))
	info, ok := r.all[shard][uRoleId]
	old := info
	if ok {
		info.HungUpMapId = hungUpMapId
		info.HungUpServerId = hungUpBattleId
		info.AddTime = time.Now().UnixMilli()
		info.Status = 2
	} else {
		info = RoleInfo{
			BattleServerId: hungUpBattleId,
			MapId:          hungUpMapId,
			HungUpMapId:    hungUpMapId,
			HungUpServerId: hungUpBattleId,
			AddTime:        time.Now().UnixMilli(),
			Status:         2,
		}
	}
	r.all[shard][uRoleId] = info
	r.log.Debug("add role status2", zap.Uint64("uRoleId", uRoleId), zap.Bool("ok", ok),
		zap.Any("old", old), zap.Any("info", info))
}

func (r *RoleHash) Remove(uRoleId URoleID) {
	shard := int(uRoleId % URoleID(RoleShardNum))
	delete(r.all[shard], uRoleId)
}

func (r *RoleHash) ClearRoleBattle(uRoleId URoleID) {
	shard := int(uRoleId % URoleID(RoleShardNum))
	v, ok := r.all[shard][uRoleId]
	if ok {
		rl := rule.MustGetReader(nil)
		cnf, exist := rl.MapScene.GetMapSceneById(v.MapId)
		if exist && cnf.MapType == values.Integer(models.BattleType_HangUp) {
			v.HungUpServerId, v.HungUpMapId = 0, 0
		}
		v.BattleServerId, v.MapId, v.Status = 0, 0, 0
		r.all[shard][uRoleId] = v
	}
}
