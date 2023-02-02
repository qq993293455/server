// Code generated by exporter. DO NOT EDIT.
package rule_model

import (
	"errors"

	"coin-server/common/values"
)

// struct
type GoodslistLv struct {
	ShopgoodslistId values.Integer                    `mapstructure:"shopgoodslist_id" json:"shopgoodslist_id"`
	Id              values.Integer                    `mapstructure:"id" json:"id"`
	ItemIdNum       []values.Integer                  `mapstructure:"item_id_num" json:"item_id_num"`
	Weight          values.Integer                    `mapstructure:"weight" json:"weight"`
	Cost            map[values.Integer]values.Integer `mapstructure:"cost" json:"cost"`
	Discount        map[values.Integer]values.Integer `mapstructure:"discount" json:"discount"`
	GoodsLevel      []values.Integer                  `mapstructure:"goods_level" json:"goods_level"`
}

// parse func
func ParseGoodslistLv(data *Data) {
	if err := data.UnmarshalKey("goodslist_lv", &h.goodslistLv); err != nil {
		panic(errors.New("parse table GoodslistLv err:\n" + err.Error()))
	}
	for i, el := range h.goodslistLv {
		if _, ok := h.goodslistLvMap[el.ShopgoodslistId]; !ok {
			h.goodslistLvMap[el.ShopgoodslistId] = map[values.Integer]int{el.Id: i}
		} else {
			h.goodslistLvMap[el.ShopgoodslistId][el.Id] = i
		}
	}
}

func (i *GoodslistLv) Len() int {
	return len(h.goodslistLv)
}

func (i *GoodslistLv) List() []GoodslistLv {
	return h.goodslistLv
}

func (i *GoodslistLv) GetGoodslistLvById(parentId, id values.Integer) (*GoodslistLv, bool) {
	item, ok := h.goodslistLvMap[parentId]
	if !ok {
		return nil, false
	}
	index, ok := item[id]
	if !ok {
		return nil, false
	}
	return &h.goodslistLv[index], true
}