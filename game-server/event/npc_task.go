package event

import (
	"coin-server/common/proto/models"
)

type NpcTaskUpdate struct {
	Task *models.NpcTask
}
