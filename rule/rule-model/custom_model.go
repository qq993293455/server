package rule_model

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

type CustomDropList struct {
	DropListId   values.Integer
	MiniId       values.Integer
	DropListMini []DropListsMini
}

type CustomDrop struct {
	DropId   values.Integer `mapstructure:"drop_id" json:"drop_id"`
	MiniId   values.Integer `mapstructure:"id" json:"id"`
	DropMini []DropMini
}

type CustomDivination struct {
	TotalWeight values.Integer
	List        []Divination
}

type CustomGachaWeight struct {
	TotalWeight  values.Integer
	GachaIdx     []values.Integer
	GachaWeights []values.Integer
}

type CustomExchangeWeight struct {
	Typ             models.ExchangeType
	TotalWeight     values.Integer
	ExchangeWeights []ExchangeItem
}

type CustomDialogRelation struct {
	HeadDialogId values.Integer
	IsEnd        bool
}

type CustomNpcDialogRelationMap map[values.Integer]CustomDialogRelation

type MaxSoulContract struct {
	Rank  values.Integer
	Level values.Integer
}
