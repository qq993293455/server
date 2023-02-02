package values

import (
	"sync"

	"coin-server/common/values"
)

type MemDb interface {
	// 返回排行榜
	GetRank(rankId values.RankId) RankAgg
	// 添加或更新排行榜
	SaveRank(rank RankAgg)
	// 删除排行榜
	DeleteRank(rankId values.RankId)
}

type MemRankDb struct {
	ranksDb sync.Map
}

func NewMemRankDb() MemDb {
	res := &MemRankDb{}
	return res
}

func (m *MemRankDb) GetRank(rankId values.RankId) RankAgg {
	tmpRank, ok := m.ranksDb.Load(rankId)
	if !ok {
		return nil
	}
	return tmpRank.(RankAgg)
}

func (m *MemRankDb) SaveRank(rank RankAgg) {
	rankId := rank.GetRankId()
	m.ranksDb.Store(rankId, rank)
}

func (m *MemRankDb) DeleteRank(rankId values.RankId) {
	m.ranksDb.Delete(rankId)
}
