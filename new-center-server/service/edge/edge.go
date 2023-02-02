package edge

import (
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/network/stdtcp"
	modelspb "coin-server/common/proto/models"
	centerpb "coin-server/common/proto/newcenter"
	"coin-server/common/values"
	"go.uber.org/zap"
	"math"
)

type EdgeInfo struct {
	edgeType        modelspb.EdgeType
	servers         map[values.Integer]values.Integer //所有的BattleServerId:MapId
	remainCpuWeight values.Integer
}

type EdgesHash struct {
	all map[*stdtcp.Session]EdgeInfo
	log *logger.Logger
}

var gEdgesHash *EdgesHash

func initEdgeHash(log *logger.Logger) {
	gEdgesHash = &EdgesHash{
		all: make(map[*stdtcp.Session]EdgeInfo, 128),
		log: log,
	}
}

func (e *EdgesHash) Sync(s *stdtcp.Session, msg *centerpb.NewCenter_SyncEdgeInfoPush) *errmsg.ErrMsg {
	curEdge, ok := e.all[s]
	if msg.IsAll {
		info := EdgeInfo{
			edgeType:        msg.Typ,
			remainCpuWeight: msg.RemainCpuWeight,
			servers:         make(map[values.Integer]values.Integer, len(msg.Edges)),
		}
		for _, val := range msg.Edges {
			info.servers[val.BattleId] = val.MapId
		}
		e.all[s] = info
		e.log.Info("add edge", zap.String("addr", s.RemoteAddr()), zap.Any("info", info))
		return nil
	}
	if ok {
		if curEdge.edgeType != msg.Typ {
			return errmsg.NewInternalErr("edgeType not match!")
		}
		curEdge.remainCpuWeight = msg.RemainCpuWeight
		for _, val := range msg.Edges {
			if val.IsEnd {
				delete(curEdge.servers, val.BattleId)
				continue
			}
			if val.IsAdd {
				curEdge.servers[val.BattleId] = val.MapId
			}
		}
		e.all[s] = curEdge
		e.log.Info("update edge", zap.String("addr", s.RemoteAddr()), zap.Any("info", curEdge))
		return nil
	}
	return errmsg.NewInternalErr("msg not IsAll!but edge not exist")
}

func (e *EdgesHash) GetServers(s *stdtcp.Session) (map[values.Integer]values.Integer, bool) {
	if v, ok := e.all[s]; ok {
		return v.servers, true
	}
	return nil, false
}

func (e *EdgesHash) Get(s *stdtcp.Session) (*EdgeInfo, bool) {
	if v, ok := e.all[s]; ok {
		return &v, true
	}
	return nil, false
}

func (e *EdgesHash) SetType(s *stdtcp.Session, edgeType modelspb.EdgeType) bool {
	if v, ok := e.all[s]; ok {
		v.edgeType = edgeType
		e.all[s] = v
		return true
	}
	return false
}

func (e *EdgesHash) Remove(s *stdtcp.Session) {
	delete(e.all, s)
}

func (e *EdgesHash) FindOne(edgeType modelspb.EdgeType, weight values.Integer) *stdtcp.Session {
	var target *stdtcp.Session = nil
	var minDiff values.Integer = math.MinInt64
	for k, v := range e.all {
		if v.edgeType != edgeType {
			continue
		}
		diff := v.remainCpuWeight - weight
		if diff < 0 {
			continue
		}
		if target == nil || minDiff < diff {
			target = k
			minDiff = diff
		}
	}
	return target
}

func (e *EdgesHash) FindOneEmptyEdge() *stdtcp.Session {
	for k, v := range e.all {
		if v.edgeType == modelspb.EdgeType_StatelessServer {
			return k
		}
	}
	return nil
}
