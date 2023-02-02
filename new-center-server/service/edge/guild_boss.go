package edge

import (
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	modelspb "coin-server/common/proto/models"
	centerpb "coin-server/common/proto/newcenter"
	"coin-server/common/values"
	"coin-server/rule"
)

type GuildBossInfo struct {
	battleServerId values.Integer
}

type GuildBossHash struct {
	all map[string]GuildBossInfo //union_boss_id：xxx
	log *logger.Logger
}

var gGuildBossHash *GuildBossHash

func initGuildBossHash(log *logger.Logger) {
	gGuildBossHash = &GuildBossHash{
		all: make(map[string]GuildBossInfo, 256),
		log: log,
	}
}

func (s *GuildBossHash) Check(msg *centerpb.NewCenter_SyncEdgeInfoPush) *errmsg.ErrMsg {
	for _, edge := range msg.Edges {
		r := rule.MustGetReader(nil)
		cnf, ok := r.MapScene.GetMapSceneById(edge.MapId)
		if !ok {
			return errmsg.NewInternalErr("invalid map id")
		}
		if cnf.MapType != values.Integer(modelspb.BattleType_UnionBoss) {
			continue
		}
		if edge.GuildBossId == "" {
			return errmsg.NewInternalErr("GuildBossId empty!")
		}
		if edge.BattleId <= 0 {
			return errmsg.NewInternalErr("GuildBossId invalid BattleId!")
		}
	}
	return nil
}

func (s *GuildBossHash) Sync(msg *centerpb.NewCenter_EdgeInfo) {
	r := rule.MustGetReader(nil)
	cnf, ok := r.MapScene.GetMapSceneById(msg.MapId)
	if !ok || cnf.MapType != values.Integer(modelspb.BattleType_UnionBoss) {
		return
	}
	if msg.GuildBossId == "" || msg.BattleId <= 0 {
		return
	}
	s.all[msg.GuildBossId] = GuildBossInfo{battleServerId: msg.BattleId}
}

func (s *GuildBossHash) Get(guildBossId string) (*GuildBossInfo, bool) {
	val, ok := s.all[guildBossId]
	if ok {
		return &val, true
	}
	return nil, false
}

func (s *GuildBossHash) Remove(guildBossId string) {
	delete(s.all, guildBossId)
}

func (s *GuildBossHash) RemoveByBattleId(battleId values.Integer) {
	for k, v := range s.all {
		if v.battleServerId == battleId {
			delete(s.all, k)
			return
		}
	}
}
