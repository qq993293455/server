package values

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

// 内存排行榜数据变化
type MemRankValueChangeData struct {
	RankId     values.RankId
	AddList    []*models.RankValue
	UpdateList []*models.RankValue
	DeleteList []*models.RankValue
}
