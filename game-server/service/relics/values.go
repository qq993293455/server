package relics

import (
	"unsafe"

	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
)

const (
	DefaultRelicsLevel = 0
	DefaultRelicsStar  = 0
)

func RelicsModels2Dao(relicsModel []*models.RelicsSuit) []*dao.RelicsSuit {
	return *(*[]*dao.RelicsSuit)(unsafe.Pointer(&relicsModel))
}

func RelicsDao2Models(relicsDao []*dao.RelicsSuit) []*models.RelicsSuit {
	return *(*[]*models.RelicsSuit)(unsafe.Pointer(&relicsDao))
}

func RelicsModel2Dao(relicsModel *models.RelicsSuit) *dao.RelicsSuit {
	return (*dao.RelicsSuit)(relicsModel)
}

func RelicsDao2Model(relicsDao *dao.RelicsSuit) *models.RelicsSuit {
	return (*models.RelicsSuit)(relicsDao)
}
