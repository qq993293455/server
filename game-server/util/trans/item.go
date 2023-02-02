package trans

import (
	"fmt"

	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
)

func ItemD2M(item *dao.Item) *models.Item {
	return (*models.Item)(item)
}

func ItemsD2M(items []*dao.Item) []*models.Item {
	ret := make([]*models.Item, 0, len(items))
	for _, v := range items {
		ret = append(ret, ItemD2M(v))
	}
	return ret
}

func ItemMapToProto(items map[values.Integer]values.Integer) []*models.Item {
	res, idx := make([]*models.Item, len(items)), 0
	for itemId, v := range items {
		res[idx] = &models.Item{
			ItemId: itemId,
			Count:  v,
		}
		idx++
	}
	return res
}

func ItemProtoToMap(items []*models.Item) map[values.Integer]values.Integer {
	ret := make(map[values.Integer]values.Integer)
	for _, v := range items {
		ret[v.ItemId] += v.Count
	}
	return ret
}

func ItemSliceToMap(items []values.Integer) map[values.Integer]values.Integer {
	if len(items) < 2 || len(items)%2 != 0 {
		panic(fmt.Sprintf("item slice config error: %v", items))
	}
	ret := make(map[values.Integer]values.Integer, len(items)/2)
	for i := 0; i < len(items); i += 2 {
		ret[items[i]] += items[i+1]
	}
	return ret
}

func ItemSliceToPb(items []values.Integer) []*models.Item {
	if len(items) < 2 || len(items)%2 != 0 {
		panic(fmt.Sprintf("item slice config error: %v", items))
	}
	ret := make([]*models.Item, 0, len(items)/2)
	for i := 0; i < len(items); i += 2 {
		ret = append(ret, &models.Item{ItemId: items[i], Count: items[i+1]})
	}
	return ret
}

func ItemSliceToPbMulti(items []values.Integer, multi int64) []*models.Item {
	if len(items) < 2 || len(items)%2 != 0 {
		panic(fmt.Sprintf("item slice config error: %v", items))
	}
	ret := make([]*models.Item, 0, len(items)/2)
	for i := 0; i < len(items); i += 2 {
		ret = append(ret, &models.Item{ItemId: items[i], Count: items[i+1] * multi})
	}
	return ret
}
