// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type ExchangeItem struct {
	ExchangeId values.Integer `mapstructure:"exchange_id" json:"exchange_id"`
	Id         values.Integer `mapstructure:"id" json:"id"`
	ItemId     values.Integer `mapstructure:"item_id" json:"item_id"`
	ItemCount  values.Integer `mapstructure:"item_count" json:"item_count"`
	ItemWeight values.Integer `mapstructure:"item_weight" json:"item_weight"`
}

// parse func
func ParseExchangeItem(data *Data) {
	if err := data.UnmarshalKey("exchange_item", &h.exchangeItem); err != nil {
		panic(errors.New("parse table ExchangeItem err:\n" + err.Error()))
	}
	for i, el := range h.exchangeItem {
		if _, ok := h.exchangeItemMap[el.ExchangeId]; !ok {
			h.exchangeItemMap[el.ExchangeId] = map[values.Integer]int{el.Id: i}
		} else {
			h.exchangeItemMap[el.ExchangeId][el.Id] = i
		}
	}
}

func (i *ExchangeItem) Len() int {
	return len(h.exchangeItem)
}

func (i *ExchangeItem) List() []ExchangeItem {
	return h.exchangeItem
}

func (i *ExchangeItem) GetExchangeItemById(parentId, id values.Integer) (*ExchangeItem, bool) {
	item, ok := h.exchangeItemMap[parentId]
	if !ok {
		return nil, false
	}
	index, ok := item[id]
	if !ok {
		return nil, false
	}
	return &h.exchangeItem[index], true
}
