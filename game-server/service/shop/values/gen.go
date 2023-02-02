package values

import (
	"math/rand"

	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	rulemodel "coin-server/rule/rule-model"
)

func genDetail(lvl values.Level, typ ShopTyp) (map[values.Integer]*models.TypeShopList, *errmsg.ErrMsg) {
	reader := rulemodel.GetReader()
	listRule := reader.GetShopTypMap(values.Integer(typ))
	if listRule == nil {
		return nil, errmsg.NewErrShopExcelWrong()
	}
	shopGoodsPermanentId, ok := reader.KeyValue.GetInt64("ShopGoodsPermanentId")
	if !ok {
		return nil, errmsg.NewErrShopExcelWrong()
	}
	for _, id := range listRule {
		v, ok := reader.Shopgoodslist.GetShopgoodslistById(id)
		if !ok {
			continue
		}
		playLv := v.PlayLv
		if len(playLv) != 2 {
			return nil, errmsg.NewErrShopExcelWrong()
		}
		if lvl >= playLv[0] && lvl <= playLv[1] {
			// 获得了商品组Id
			goodList := reader.GetShopMap()[v.Id]
			weightList := make([]values.Integer, len(goodList))
			mustList := make([]values.Integer, 0, 8)
			for idx, good := range goodList {
				if good.Weight > 0 {
					weightList[idx] = good.Weight
				} else {
					mustList = append(mustList, values.Integer(idx))
				}
			}
			chooseList := randWeight(weightList, int(v.GoodsNum))
			finialList := make([]values.Integer, 0, len(chooseList)+len(mustList))
			for _, must := range mustList {
				finialList = append(finialList, must)
			}
			for _, choose := range chooseList {
				finialList = append(finialList, values.Integer(choose))
			}
			details := make(map[values.Integer]*models.TypeShopList)
			for _, final := range finialList {
				goodRule := goodList[final]
				if len(goodRule.ItemIdNum) != 3 {
					return nil, errmsg.NewErrShopExcelWrong()
				}
				good := &models.Item{
					ItemId: goodRule.ItemIdNum[0],
					Count:  randCnt([2]values.Integer{goodRule.ItemIdNum[1], goodRule.ItemIdNum[2]}),
				}
				itemRule, ok := reader.Item.GetItemById(good.ItemId)
				if !ok {
					continue
				}
				detail := &models.ShopDetail{
					Good:   good,
					IsSale: false,
					Cost:   map[int64]int64{},
				}
				if len(goodRule.Discount) == 0 {
					detail.Discount = 10
					for k, val := range goodRule.Cost {
						detail.Cost[k] = val * good.Count
					}
				} else {
					ids := make([]values.Integer, 0)
					weights := make([]values.Integer, 0)
					weight := int64(0)
					discount := int64(0)
					for k, w := range goodRule.Discount {
						weight += w
						ids = append(ids, k)
						weights = append(weights, weight)
					}
					r := rand.Int63n(weight)
					for i := range weights {
						if r < weights[i] {
							discount = ids[i]
							break
						}
					}
					detail.Discount = 10 - discount
					detail.Cost = map[values.ItemId]values.Integer{}
					for k := range goodRule.Cost {
						detail.Cost[k] = goodRule.Cost[k] * good.Count
						detail.Cost[k] = (detail.Cost[k] * detail.Discount) / 10
					}
				}
				if goodRule.Weight == -1 {
					if _, exist := details[shopGoodsPermanentId]; !exist {
						details[shopGoodsPermanentId] = &models.TypeShopList{
							List: make([]*models.ShopDetail, 0, 8),
						}
					}
					details[shopGoodsPermanentId].List = append(details[shopGoodsPermanentId].List, detail)
				} else {
					if _, exist := details[itemRule.ItemShopShow]; !exist {
						details[itemRule.ItemShopShow] = &models.TypeShopList{
							List: make([]*models.ShopDetail, 0, 8),
						}
					}
					details[itemRule.ItemShopShow].List = append(details[itemRule.ItemShopShow].List, detail)
				}
			}
			return details, nil
		}
	}
	return nil, errmsg.NewErrShopExcelWrong()
}
