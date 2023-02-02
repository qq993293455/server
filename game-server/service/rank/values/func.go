package values

import (
	"coin-server/common/proto/models"
	"coin-server/common/values/enum"
)

type RankCompare struct {
	F RankFunc
}

func (r *RankCompare) Equal(a, b interface{}) bool {
	v1, v2 := a.(*models.RankValue), b.(*models.RankValue)
	if len(v1.Extra) != len(v2.Extra) {
		return false
	}
	if v1.Extra != nil && v2.Extra != nil {
		for k, v := range v1.Extra {
			if v2.Extra[k] != v {
				return false
			}
		}
	}
	if v1.OwnerId == v2.OwnerId && v1.Value1 == v2.Value1 && v1.Value2 == v2.Value2 &&
		v1.Value3 == v2.Value3 && v1.Value4 == v2.Value4 {
		return true
	}
	return false
}

func (r *RankCompare) Less(a, b interface{}) bool {
	return r.F(a.(*models.RankValue), b.(*models.RankValue))
}

// 每一种排行榜对应的排序配置
var AllRankFuncs = map[enum.RankType]RankFunc{
	enum.RankNormal: GameRankSort,
}

type RankFunc func(a, b *models.RankValue) bool

// 游戏内排行榜排序规则
func GameRankSort(a, b *models.RankValue) bool {
	// 比较积分，积分大的在前面
	if a.Value1 != b.Value1 {
		return a.Value1 > b.Value1
	}
	// 积分相同比较时间，时间小的在前面
	if a.Value2 != b.Value2 {
		return a.Value2 < b.Value2
	}
	return a.OwnerId < b.OwnerId
}
