package trans

import (
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
)

func GuideD2M(d *dao.Guide) *models.Guide {
	return &models.Guide{
		GuideId:          d.GuideId,
		StepId:           d.StepId,
		FinishedGuideIds: d.FinishedGuideIds,
	}
}
