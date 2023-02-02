package event

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type MapEventTriggered struct {
	Event *models.MapEvent
}

type MapEventFinished struct {
	EventId   values.EventId
	IsSuccess bool
	Rewards   map[values.ItemId]values.Integer
	StoryId   values.StoryId
	Piece     int64
	MapId     values.MapId
	Ratio     values.Integer
}
