package event

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type RelicsSuitUpdate struct {
	RelicsSuit *models.RelicsSuit
}

type RelicsFuncAttrUpdate struct {
	Attr map[values.AttrId]values.Integer
}
