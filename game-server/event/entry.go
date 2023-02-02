package event

import (
	"coin-server/common/values"
	"coin-server/common/values/enum"
)

// EntrySpecialAddition 词条加成
type EntrySpecialAddition struct {
	EntryId enum.EntryId
	Value   values.Integer
}
