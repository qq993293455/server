package trans

import (
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
)

func DivinationD2M(d *dao.Divination) *models.Divination {
	return &models.Divination{
		TotalCount:     d.TotalCount,
		AvailableCount: d.AvailableCount,
		ResetAt:        d.ResetAt,
	}
}
