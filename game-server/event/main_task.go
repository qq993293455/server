package event

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type MainTaskUpdate struct {
	Task *models.MainTask
}

type MainTaskFinished struct {
	TaskNo       int64
	TaskIdx      int64
	ExpProfit    values.Integer
	Illustration values.Integer
}
