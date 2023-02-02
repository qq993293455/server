package gacha

import (
	"unsafe"

	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
)

func GachaModels2Dao(gachaModel []*models.Gacha) []*dao.Gacha {
	return *(*[]*dao.Gacha)(unsafe.Pointer(&gachaModel))
}

func GachaDao2Models(gachaDao []*dao.Gacha) []*models.Gacha {
	return *(*[]*models.Gacha)(unsafe.Pointer(&gachaDao))
}

func GachaModel2Dao(gachaModel *models.Gacha) *dao.Gacha {
	return (*dao.Gacha)(gachaModel)
}

func GachaDao2Model(gachaDao *dao.Gacha) *models.Gacha {
	return (*models.Gacha)(gachaDao)
}
