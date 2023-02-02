package im

import (
	"coin-server/common/proto/dao"
	"coin-server/common/timer"
	wr "coin-server/common/utils/weightedrand"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

func GenGift(giftID values.GiftId, roleID values.RoleId, giftNo values.Integer) *dao.Gift {
	reader := rule.MustGetReader(nil)
	//_, giftMap := r.Gift()
	gcfg, ok := reader.Gift.GetGiftById(giftNo)
	if !ok {
		return nil
	}
	ret := &dao.Gift{
		GiftId:   giftID,
		GiftNo:   giftNo,
		RoleId:   roleID,
		Items:    make([]*dao.Item, 0, gcfg.Num),
		Records:  make([]*dao.DrawRecord, 0),
		CreateAt: timer.UnixMilli(),
	}

	choices := make([]*wr.Choice[rulemodel.GiftItem, int64], 0)
	//giftItems, ok := r.GitItemMap()[giftNo]
	giftItems, ok := reader.GitItemMap()[giftNo]
	if !ok {
		return nil
	}
	for _, v := range giftItems {
		choices = append(choices, wr.NewChoice(v, v.ItemWeight))
	}

	chooser, _ := wr.NewChooser[rulemodel.GiftItem, int64](choices...)
	for i := 0; i < int(gcfg.Num); i++ {
		gi := chooser.Pick()
		ret.Items = append(ret.Items, &dao.Item{
			ItemId: gi.Id,
			Count:  gi.ItemNum,
		})
	}

	return ret
}

const (
	GiftDayNum = "GiftDayNum"
)

func GetGiftDayLimit() int64 {
	reader := rule.MustGetReader(nil)
	number, _ := reader.KeyValue.GetInt64(GiftDayNum)
	return number
}
