package edge

import (
	"coin-server/common/logger"
	centerpb "coin-server/common/proto/newcenter"
	"coin-server/common/values"
	"go.uber.org/zap"
)

type BossHallLine struct {
	BattleId values.Integer
	CanEnter bool
}
type BossHallInfo struct {
	MapId values.Integer
	Lines map[values.Integer]*BossHallLine //lineId:battleId
}
type BossHallHash struct {
	all  map[values.Integer]*BossHallInfo                     // bossId:BossHallHash
	adds map[values.Integer]map[values.Integer]values.Integer //额外动态增加的分线 bossId: lineId:battleId
	log  *logger.Logger
}

var gBossHall *BossHallHash

func initBossHall(log *logger.Logger) {
	gBossHall = &BossHallHash{
		all:  make(map[values.Integer]*BossHallInfo, 16),
		adds: map[values.Integer]map[values.Integer]values.Integer{},
		log:  log,
	}
}

func (b *BossHallHash) Sync(msg *centerpb.NewCenter_SyncEdgeInfoPush) {
	b.log.Debug("enter bossHall Sync ", zap.Any("msg", msg))
	for _, edge := range msg.Edges {
		if edge.BossHallId <= 0 {
			continue
		}
		if msg.IsAll || edge.IsAdd {
			info, ok := b.all[edge.BossHallId]
			if !ok {
				info = &BossHallInfo{MapId: edge.MapId, Lines: make(map[values.Integer]*BossHallLine, 32)}
				b.all[edge.BossHallId] = info
			}
			bossHallLineNum, ok := gLineConfig.srvLines.BossHall[edge.BossHallId]
			if ok && edge.LineId > bossHallLineNum {
				if _, ok1 := b.adds[edge.BossHallId]; !ok1 {
					b.adds[edge.BossHallId] = map[values.Integer]values.Integer{}
				}
				b.adds[edge.BossHallId][edge.LineId] = edge.BattleId
			}
			info.Lines[edge.LineId] = &BossHallLine{BattleId: edge.BattleId, CanEnter: !edge.CanNotEnter}
			b.log.Info("add bossHall line", zap.Any("BossHallId", edge.BossHallId), zap.Any("battleId", edge.BattleId), zap.Any("info", info))
			continue
		}
		if edge.IsEnd {
			info, ok := b.all[edge.BossHallId]
			if ok && edge.LineId > 0 {
				delete(info.Lines, edge.LineId)
				b.log.Info("delete bossHall line", zap.Any("BossHallId", edge.BossHallId), zap.Any("battleId", edge.BattleId), zap.Any("info", info))
			}
			addInfo, ok1 := b.adds[edge.BossHallId]
			if ok1 {
				delete(addInfo, edge.LineId)
			}
		}
	}
}

func (b *BossHallHash) GetLine(bossId values.Integer, lineId values.Integer) (lineInfo *BossHallLine, exist bool) {
	if info, ok := b.all[bossId]; ok {
		if lInfo, ok2 := info.Lines[lineId]; ok2 {
			return lInfo, true
		}
	}
	return nil, false
}

func (b *BossHallHash) RemoveBattle(battleId values.Integer) {
	for bossId, info := range b.all {
		_ = bossId
		for lineId, lineInfo := range info.Lines {
			if lineInfo.BattleId == battleId {
				delete(info.Lines, lineId)
			}
		}
	}

	for _, lines := range b.adds {
		for lineId, addBattleId := range lines {
			if addBattleId == battleId {
				delete(lines, lineId)
			}
		}
	}
}

func (b *BossHallHash) SetCanNotEnter(bossId values.Integer, battleId values.Integer) {
	info, ok := b.all[bossId]
	if !ok {
		return
	}
	for _, lineInfo := range info.Lines {
		if lineInfo.BattleId == battleId {
			lineInfo.CanEnter = false
		}
	}
}

func (b *BossHallHash) GetNewLineId(bossId values.Integer) (newLineId values.Integer) {
	newLineId = 0
	bossHallLineNum, ok := gLineConfig.srvLines.BossHall[bossId]
	if info, ok1 := b.all[bossId]; ok1 {
		for lId, _ := range info.Lines {
			if lId > newLineId {
				newLineId = lId
			}
		}
	}
	if lines, ok2 := b.adds[bossId]; ok2 {
		for lId, _ := range lines {
			if lId > newLineId {
				newLineId = lId
			}
		}
	}
	newLineId += 1
	if ok && newLineId <= bossHallLineNum {
		newLineId = bossHallLineNum + 1
	}
	return newLineId
}
