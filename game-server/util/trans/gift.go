package trans

import (
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
)

func DrawRecordD2M(record *dao.DrawRecord) *models.DrawRecord {
	return &models.DrawRecord{
		RoleId:   record.RoleId,
		Item:     ItemD2M(record.Item),
		DrawTime: record.DrawTime,
		Role:     (*models.Role)(record.Role),
	}
}

func DrawRecordsD2M(records []*dao.DrawRecord) []*models.DrawRecord {
	ret := make([]*models.DrawRecord, 0, len(records))
	for _, v := range records {
		ret = append(ret, DrawRecordD2M(v))
	}
	return ret
}

func GiftD2M(gift *dao.Gift) *models.Gift {
	return &models.Gift{
		GiftId:    gift.GiftId,
		GiftNo:    gift.GiftNo,
		RoleId:    gift.RoleId,
		DrawCount: gift.DrawCount,
		Records:   DrawRecordsD2M(gift.Records),
		CreateAt:  gift.CreateAt,
	}
}
