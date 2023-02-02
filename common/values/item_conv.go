package values

import "coin-server/common/proto/models"

func Items2Map(items []*models.Item) map[ItemId]Integer {
	m := make(map[ItemId]Integer, len(items))
	for idx := range items {
		m[items[idx].ItemId] = items[idx].Count
	}
	return m
}

//func Map2Items(items map[ItemId]Integer) []*models.Item {
//	m := make([]*models.Item, len(items))
//	i:=0
//	for k := range items {
//		m[i] = &models.Item{
//			ItemId: k,
//			Count:  items[k],
//			Expire: 0,
//			Lock:   false,
//		}
//	}
//	return m
//}
